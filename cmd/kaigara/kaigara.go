package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
)

func kaigaraRun(ls logstream.LogStream, cnf config.Config, channelName, cmd string, cmdArgs []string) {
	log.Printf("Starting command: %s %v", cmd, cmdArgs)
	c := exec.Command(cmd, cmdArgs...)
	c.Env = cmdEnv(cnf)
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	channelOut := fmt.Sprintf("log.%s.%s", channelName, "stdout")
	channelErr := fmt.Sprintf("log.%s.%s", channelName, "stderr")
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
		ls.HeartBeat(channelName, quit)
		wg.Done()
	}()

	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Printf("exit status 0\n")
	quit <- 0

	wg.Wait()
}

func cmdEnv(cnf config.Config) []string {
	env := []string{}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "KAIGARA_") {
			env = append(env, v)
		}
	}
	if cnf == nil {
		return env
	}
	for k, v := range cnf.ListEntries() {
		env = append(env, strings.ToUpper(k)+"="+v.(string))
	}
	return env
}

func initConfig() config.Config {
	addr := os.Getenv("KAIGARA_VAULT_ADDR")
	token := os.Getenv("KAIGARA_VAULT_TOKEN")
	path := os.Getenv("KAIGARA_VAULT_CONFIG_PATH")
	missingParam := false
	if addr == "" {
		log.Println("KAIGARA_VAULT_ADDR unset")
		missingParam = true
	}
	if token == "" {
		log.Println("KAIGARA_VAULT_TOKEN unset")
		missingParam = true
	}
	if path == "" {
		log.Println("KAIGARA_VAULT_CONFIG_PATH unset")
		missingParam = true
	}
	if missingParam {
		log.Println("Do not start use remote config")
		return nil
	}
	return config.NewVaultConfig(addr, token, path)
}

func initLogStream() logstream.LogStream {
	url := os.Getenv("KAIGARA_REDIS_URL")
	return logstream.NewRedisClient(url)
}

func main() {
	svc := os.Getenv("KAIGARA_SERVICE_NAME")
	if len(os.Args) < 1 {
		panic("Usage: kaigara CMD [ARGS...]")
	}
	ls := initLogStream()
	cnf := initConfig()
	kaigaraRun(ls, cnf, svc, os.Args[1], os.Args[2:])
}
