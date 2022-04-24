package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/openware/kaigara/pkg/logstream"
)

var redisChannel = "log.*"
var showName bool

func kaitailCmd() error {
	ls, err := logstream.NewRedisClient(conf.RedisURL)
	if err != nil {
		return err
	}

	ch := ls.Subscribe(redisChannel)

	re, err := regexp.Compile(`^Message<(log\.[A-z.]+?): (.+?)>$`)
	if err != nil {
		return err
	}

	for msg := range ch {
		rs := re.FindStringSubmatch(msg.String())

		if len(rs) < 2 {
			log.Printf("WRN: could not parse message: %s\n", msg)
			continue
		}

		if showName {
			fmt.Printf("%s: %s\n", rs[1], rs[2])
		} else {
			fmt.Printf("%s\n", rs[2])
		}
	}

	return nil
}
