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

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/ika"
)

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	SecretStore  string `yaml:"secret-store" env:"KAIGARA_SECRET_STORE" env-default:"vault"`
	VaultToken   string `yaml:"vault-token" env:"KAIGARA_VAULT_TOKEN"`
	VaultAddr    string `yaml:"vault-addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`
	AppName      string `yaml:"vault-app-name" env:"KAIGARA_APP_NAME"`
	DeploymentID string `yaml:"deployment-id" env:"KAIGARA_DEPLOYMENT_ID"`
	Scopes       string `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public"`
}

var cnf = &KaigaraConfig{}
var secretStore types.SecretStore

func initConfig() {
	err := ika.ReadConfig("", cnf)
	if err != nil {
		panic(err)
	}
}

func getVaultService() *vault.Service {
	return vault.NewService(cnf.VaultAddr, cnf.VaultToken, cnf.AppName, cnf.DeploymentID)
}

func getVaultSecretStore() types.SecretStore {
	var ss types.SecretStore
	vs := getVaultService()
	ss = vs

	return ss
}

func parseScopes() []string {
	return strings.Split(cnf.Scopes, ",")
}

func kaigaraRun(ls logstream.LogStream, secretStore types.SecretStore, cmd string, cmdArgs []string) {
	log.Printf("Starting command: %s %v", cmd, cmdArgs)
	c := exec.Command(cmd, cmdArgs...)
	env := config.BuildCmdEnv(secretStore, os.Environ(), parseScopes())
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

func main() {
	if len(os.Args) < 1 {
		panic("Usage: kaigara CMD [ARGS...]")
	}
	ls := initLogStream()
	initConfig()
	secretStore = getVaultSecretStore()

	kaigaraRun(ls, secretStore, os.Args[1], os.Args[2:])
}
