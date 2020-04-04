package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/openware/kaigara/pkg/broker"
)

var (
	cmd  = flag.String("exec", "date", "Your command")
	name = flag.String("name", "", "stream name")
)

func runCommand(cmdName, channelName string, cmdArgs []string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	channel := fmt.Sprintf("logs.%s.%s", channelName, "stdout")
	log.Printf("Publishing on %s\n", channel)
	broker.RedisPublish(channel, stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()
	var channelName string
	if *name == "" {
		channelName = *cmd
	} else {
		channelName = *name
	}
	runCommand(*cmd, channelName, flag.Args())
}
