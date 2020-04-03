package broker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-redis/redis/v7"
	"github.com/openware/kaigara/pkg/utils"
)

func RedisClient() *redis.Client {
	url := utils.GetEnv("REDIS_URL", "redis://localhost:6379/0")
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	_, err = client.Ping().Result()
	if err != nil {
		panic(err)
	}
	log.Println("Connected to redis")
	return client
}

func RedisPublish(channel string, stdout io.Reader) {
	client := RedisClient()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		err := client.Publish(channel, scanner.Text()).Err()
		if err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
