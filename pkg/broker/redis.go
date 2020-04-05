package broker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/openware/kaigara/pkg/utils"
)

var RedisClient *redis.Client = redisClient()

func redisClient() *redis.Client {
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

func RedisPublish(channel string, stream io.ReadCloser) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		err := RedisClient.Publish(channel, scanner.Text()).Err()
		if err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		stream.Close()
	}
}

func RedisHeartBeat(name string, quit chan int) {
	key := fmt.Sprintf("service.%s", name)
	RedisClient.Set(key, time.Now(), 10*time.Second)

	for {
		select {
		case <-quit:
			RedisClient.Del(key)
			return

		case <-time.After(10 * time.Second):
			RedisClient.Expire(key, 10*time.Second)
		}
	}
}
