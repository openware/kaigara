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
	"time"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/ika"
)

var cnf = &config.KaigaraConfig{}

func parseScopes() []string {
	return strings.Split(cnf.Scopes, ",")
}

func parseAppNames() []string {
	return strings.Split(cnf.AppNames, ",")
}

func appNamesToLoggingName() string {
	return strings.Join(parseAppNames(), "&")
}

func kaigaraRun(ls logstream.LogStream, store types.Storage, cmd string, cmdArgs []string) {
	log.Printf("Starting command: %s %v", cmd, cmdArgs)
	scopes := parseScopes()
	c := exec.Command(cmd, cmdArgs...)
	env := config.BuildCmdEnv(parseAppNames(), store, os.Environ(), scopes)

	c.Env = env.Vars

	for _, file := range env.Files {
		err := os.MkdirAll(path.Dir(file.Path), 0750)
		if err != nil {
			panic(fmt.Sprintf("Failed to make dir %s: %s", file.Path, err.Error()))
		}

		err = ioutil.WriteFile(file.Path, []byte(file.Content), 0640)
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
			_, err = stdin.Write(line)
			if err != nil {
				panic(err)
			}

			if !isPrefix {
				_, err = stdin.Write([]byte("\n"))
				if err != nil {
					panic(err)
				}
			}
		}
	}()

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

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		ls.Publish(channelOut, stdout)
		wg.Done()
	}()

	go func() {
		ls.Publish(channelErr, stderr)
		wg.Done()
	}()

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	go exitWhenSecretsOutdated(c, store, scopes)

	quit := make(chan int)
	go func() {
		ls.HeartBeat(appNamesToLoggingName(), quit)
		wg.Done()
	}()

	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Printf("exit status 0\n")
	quit <- 0

	wg.Wait()
}

func initLogStream() logstream.LogStream {
	url := os.Getenv("KAIGARA_REDIS_URL")
	return logstream.NewRedisClient(url)
}

func exitWhenSecretsOutdated(c *exec.Cmd, store types.Storage, scopes []string) {
	appNames := append(parseAppNames(), "global")

	if ignore, ok := os.LookupEnv("KAIGARA_IGNORE_GLOBAL"); ok && ignore == "true" {
		appNames = appNames[:len(appNames)-1]
	}

	for range time.Tick(time.Second * 20) {
		for _, appName := range appNames {
			for _, scope := range scopes {
				current, err := store.GetCurrentVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				latest, err := store.GetLatestVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				if current != latest {
					log.Printf("Found secrets updated on '%v' scope. from: v%v, to: v%v. killing process...\n", scope, current, latest)
					if err := c.Process.Kill(); err != nil {
						log.Fatal("Failed to kill process", err)
					}
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

	if err := ika.ReadConfig("", cnf); err != nil {
		panic(err)
	}

	store, err := config.GetStorageService(cnf)
	if err != nil {
		panic(err)
	}

	kaigaraRun(ls, store, os.Args[1], os.Args[2:])
}
