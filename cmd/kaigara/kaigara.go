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
	pid = flag.String("pid", "", "process unique name")
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

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	channelOut := fmt.Sprintf("logs.%s.%s", channelName, "stdout")
	channelErr := fmt.Sprintf("logs.%s.%s", channelName, "stderr")
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

	c := make(chan string, 1)
	go broker.RedisHeartBeat(*pid, c)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	c <- "stop"

	wg.Wait()
}

func main() {
	flag.Parse()
	if *pid == "" {
		pid = cmd
	}
	runCommand(*cmd, *pid, flag.Args())
}
