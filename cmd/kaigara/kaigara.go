package main

import (
	"fmt"
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

func getVaultService(appName string) *vault.Service {
	return vault.NewService(cnf.VaultAddr, cnf.VaultToken, appName, cnf.DeploymentID)
}

func parseScopes() []string {
	return strings.Split(cnf.Scopes, ",")
}

func kaigaraRun(ls logstream.LogStream, secretStores []types.SecretStore, cmd string, cmdArgs []string) {
	log.Printf("Starting command: %s %v", cmd, cmdArgs)
	scopes := parseScopes()
	c := exec.Command(cmd, cmdArgs...)
	env := config.BuildCmdEnv(secretStores, os.Environ(), scopes)

	c.Env = env.Vars

	for _, file := range env.Files {
		os.MkdirAll(path.Dir(file.Path), 0750)
		err := ioutil.WriteFile(file.Path, []byte(file.Content), 0640)
		if err != nil {
			panic(fmt.Sprintf("Failed to write file %s: %s", file.Path, err.Error()))
		}
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	channelOut := fmt.Sprintf("log.%s.%s", cnf.AppName, "stdout")
	channelErr := fmt.Sprintf("log.%s.%s", cnf.AppName, "stderr")
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

	go exitWhenSecretsOutdated(c, secretStores, scopes)

	quit := make(chan int)
	go func() {
		ls.HeartBeat(cnf.AppName, quit)
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

func exitWhenSecretsOutdated(c *exec.Cmd, secretStores []types.SecretStore, scopes []string) {
	for range time.Tick(time.Minute * 1) {
		for _, secretStore := range secretStores {
			for _, scope := range scopes {
				current, err := secretStore.GetCurrentVersion(scope)
				if err != nil {
					log.Fatal(err)
					break
				}
				latest, err := secretStore.GetLatestVersion(scope)
				if err != nil {
					log.Fatal(err)
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
	if len(os.Args) < 1 {
		panic("Usage: kaigara CMD [ARGS...]")
	}
	ls := initLogStream()
	initConfig()
	secretStores := []types.SecretStore{
		getVaultService(cnf.AppName),
		getVaultService("global"),
	}

	kaigaraRun(ls, secretStores, os.Args[1], os.Args[2:])
}
