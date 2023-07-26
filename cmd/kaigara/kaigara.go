package main

import (
	"encoding/base64"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/env"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
)

var (
	conf    *config.KaigaraConfig
	restart chan int
)

func parseScopes() []string {
	return strings.Split((*conf).Scopes, ",")
}

func parseAppNames() []string {
	return strings.Split((*conf).AppNames, ",")
}

func kaigaraRun(ss types.Storage, cmd string, cmdArgs []string) {
	scopes := parseScopes()
	c := exec.Command(cmd, cmdArgs...)
	envs, err := env.BuildCmdEnv(parseAppNames(), ss, os.Environ(), scopes)
	if err != nil {
		panic(err)
	}

	c.Env = envs.Vars
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr

	for _, file := range envs.Files {
		if err := os.MkdirAll(path.Dir(file.Path), 0750); err != nil {
			log.Printf("ERR: failed to make dir %s: %s", file.Path, err.Error())
		}

		contents, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			log.Printf("ERR: failed to decode string %s: %s", file.Content, err.Error())
		}

		if err := os.WriteFile(file.Path, contents, 0640); err != nil {
			log.Printf("ERR: failed to write file %s: %s", file.Path, err.Error())
		}
	}

	log.Printf("INF: starting command: %s %v\n", cmd, cmdArgs)
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	go exitWhenSecretsOutdated(c, ss, scopes)

	if err := c.Wait(); err != nil {
		log.Printf("Process completed: %s\n", err)
	}
}

func exitWhenSecretsOutdated(c *exec.Cmd, ss types.Storage, scopes []string) {
	appNames := parseAppNames()

	if ignore, ok := os.LookupEnv("KAIGARA_IGNORE_GLOBAL"); !ok || ignore != "true" {
		appNames = append(appNames, "global")
	}

	for range time.Tick(time.Second * 5) {
		for _, appName := range appNames {
			for _, scope := range scopes {
				current, err := ss.GetCurrentVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				latest, err := ss.GetLatestVersion(appName, scope)
				if err != nil {
					log.Println(err.Error())
					break
				}
				if current != latest {
					log.Printf("INF: found secrets updated on '%v' scope. from: v%v, to: v%v, restarting process...\n", scope, current, latest)
					if err := c.Process.Signal(os.Kill); err != nil {
						log.Printf("FTL: failed to kill process: %s", err.Error())
					}
					restart <- 1
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

	var err error
	conf, err = config.NewKaigaraConfig()
	if err != nil {
		panic(err)
	}

	ss, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	restart = make(chan int)

	for {
		kaigaraRun(ss, os.Args[1], os.Args[2:])

		select {
		case <-restart:
			continue
		default:
			return
		}
	}
}
