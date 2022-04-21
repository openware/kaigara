package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/pkg/ika"
)

var conf = &config.KaigaraConfig{}

func main() {
	channel := flag.String("c", "log.*", "Redis channel pattern to subscribe")
	showName := flag.Bool("n", false, "Show channel name")
	flag.Parse()

	if err := ika.ReadConfig("", conf); err != nil {
		panic(err)
	}

	ls, err := logstream.NewRedisClient(conf.RedisURL)
	if err != nil {
		panic(err)
	}

	ch := ls.Subscribe(*channel)

	re := regexp.MustCompile(`^Message<(log\.[A-z.]+?): (.+?)>$`)

	for msg := range ch {
		rs := re.FindStringSubmatch(msg.String())

		if len(rs) < 2 {
			log.Printf("WRN: could not parse message: %s\n", msg)
			continue
		}

		if *showName {
			fmt.Printf("%s: %s\n", rs[1], rs[2])
		} else {
			fmt.Printf("%s\n", rs[2])
		}
	}
}
