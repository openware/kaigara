package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/utils"
)

var (
	channel  = flag.String("c", "log.*", "Redis channel pattern to subscribe")
	showName = flag.Bool("n", false, "Show channel name")
)

func main() {
	flag.Parse()

	ls := logstream.NewRedisClient(utils.GetEnv("KAIGARA_REDIS_URL", "redis://localhost:6379/0"))
	ch := ls.Subscribe(*channel)

	re := regexp.MustCompile(`^Message<(log\.[A-z.]+?): (.+?)>$`)

	for msg := range ch {
		rs := re.FindStringSubmatch(msg.String())

		if len(rs) < 2 {
			log.Printf("Could not parse message: %s\n", msg)
			continue
		}

		if *showName {
			fmt.Printf("%s: %s\n", rs[1], rs[2])
		} else {
			fmt.Printf("%s\n", rs[2])
		}
	}
}
