package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/openware/kaigara/pkg/broker"
)

var (
	cmd = flag.String("exec", "date", "Your command")
	svc = flag.String("name", "", "process unique name")
)

func runCommand(cmdName, channelName string, cmdArgs []string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	channelOut := fmt.Sprintf("log.%s.%s", channelName, "stdout")
	channelErr := fmt.Sprintf("log.%s.%s", channelName, "stderr")
	log.Printf("Publishing on %s and %s\n", channelOut, channelErr)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		broker.RedisPublish(channelOut, stdout)
		wg.Done()
	}()

	go func() {
		broker.RedisPublish(channelErr, stderr)
		wg.Done()
	}()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	quit := make(chan int)
	go func() {
		broker.RedisHeartBeat(channelName, quit)
		wg.Done()
	}()

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Printf("exit status 0\n")
	quit <- 0

	wg.Wait()
}

func main() {
	flag.Parse()
	if *svc == "" {
		svc = cmd
	}
	runCommand(*cmd, *svc, flag.Args())
}
