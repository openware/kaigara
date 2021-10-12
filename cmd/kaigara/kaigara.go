package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/ika"
)

var cnf = &config.KaigaraConfig{}

func initConfig() {
	err := ika.ReadConfig("", cnf)
	if err != nil {
		panic(err)
	}
}

func getVaultService() *vault.Service {
	return vault.NewService(cnf.VaultAddr, cnf.VaultToken, cnf.DeploymentID)
}

func parseScopes() []string {
	return strings.Split(cnf.Scopes, ",")
}

func parseAppNames() []string {
	return strings.Split(cnf.AppNames, ",")
}

func appNamesToLoggingName() string {
	return strings.Join(parseAppNames(), "&")
}

func buildProcess(secretStore types.SecretStore, cmd string, cmdArgs []string) *exec.Cmd {
	log.Printf("Starting command: %s %v", cmd, cmdArgs)
	scopes := parseScopes()
	c := exec.Command(cmd, cmdArgs...)
	env := config.BuildCmdEnv(parseAppNames(), secretStore, os.Environ(), scopes)

	c.Env = env.Vars

	for _, file := range env.Files {
		os.MkdirAll(path.Dir(file.Path), 0750)
		err := ioutil.WriteFile(file.Path, []byte(file.Content), 0640)
		if err != nil {
			panic(fmt.Sprintf("Failed to write file %s: %s", file.Path, err.Error()))
		}
	}

	stdin, err := c.StdinPipe()
	if err != nil {
		panic(err)
	}

	// Read STDIN and write it to the command
	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			line, isPrefix, err := r.ReadLine()
			if err == io.EOF {
				log.Printf("Reached EOF on STDIN")
				stdin.Close()
				break
			} else if err != nil {
				panic(err)
			}
			stdin.Write(line)
			if !isPrefix {
				stdin.Write([]byte("\n"))
			}
		}
	}()

	return c
}

func publishProcess(c *exec.Cmd, ls logstream.LogStream, wg *sync.WaitGroup) {
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	channelOut := fmt.Sprintf("log.%s.%s", appNamesToLoggingName(), "stdout")
	channelErr := fmt.Sprintf("log.%s.%s", appNamesToLoggingName(), "stderr")
	log.Printf("Publishing on %s and %s\n", channelOut, channelErr)

	go func() {
		wg.Add(1)
		defer wg.Done()
		ls.Publish(channelOut, stdout)
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		ls.Publish(channelErr, stderr)
	}()
}

func kaigaraRun(ls logstream.LogStream, secretStore types.SecretStore, cmd string, cmdArgs []string) {
	for {
		c := buildProcess(secretStore, cmd, cmdArgs)

		wg := &sync.WaitGroup{}
		publishProcess(c, ls, wg)

		if err := c.Start(); err != nil {
			log.Fatal(err)
		}

		scopes := parseScopes()
		restart := make(chan int, 1)
		go exitWhenSecretsOutdated(c, secretStore, scopes, restart)

		quit := make(chan int)
		go func() {
			wg.Add(1)
			defer wg.Done()
			ls.HeartBeat(appNamesToLoggingName(), quit)
		}()

		if err := c.Wait(); err != nil {
			log.Fatal(err)
		}
		quit <- 0
		wg.Wait()

		select {
		case <-restart:
			continue
		default:
			log.Println("Process exited code 0")
			return
		}
	}
}

func initLogStream() logstream.LogStream {
	url := os.Getenv("KAIGARA_REDIS_URL")
	return logstream.NewRedisClient(url)
}

func exitWhenSecretsOutdated(c *exec.Cmd, secretStore types.SecretStore, scopes []string, restart chan int) {
	appNames := append(parseAppNames(), "global")

	if ignore, ok := os.LookupEnv("KAIGARA_IGNORE_GLOBAL"); ok && ignore == "true" {
		appNames = appNames[:len(appNames)-1]
	}

	for range time.Tick(time.Second * 20) {
		for _, appName := range appNames {
			for _, scope := range scopes {
				current, err := secretStore.GetCurrentVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				latest, err := secretStore.GetLatestVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				if current != latest {
					log.Printf("Found secrets updated on '%v' scope. from: v%v, to: v%v. restarting process...\n", scope, current, latest)
					restart <- 0
					if err := c.Process.Signal(syscall.SIGINT); err != nil {
						log.Fatal("Failed to send interrupt signal to process", err)
					}
					return
				}
			}
		}
	}
}

func main() {
	log.SetPrefix("[Kaigara] ")
	if len(os.Args) < 2 {
		panic("Usage: kaigara CMD [ARGS...]")
	}
	ls := initLogStream()
	initConfig()
	secretStore := getVaultService()

	kaigaraRun(ls, secretStore, os.Args[1], os.Args[2:])
}
