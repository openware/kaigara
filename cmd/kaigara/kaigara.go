package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/openware/kaigara/pkg/broker"
)

func runCommand(cmdName string, cmdArgs []string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	channel := fmt.Sprintf("logs.%s.%s", cmdName, "stdout")
	log.Printf("Publishing on %s\n", channel)
	broker.RedisPublish(channel, stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var cmd string

	flag.StringVar(&cmd, "exec", "date", "Your command")
	flag.Parse()
	runCommand(cmd, flag.Args())
}
