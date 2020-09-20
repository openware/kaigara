package logstream

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

type RedisLogStream struct {
	client *redis.Client
}

func NewRedisClient(url string) *RedisLogStream {
	if url == "" {
		log.Println("KAIGARA_REDIS_URL unset, do not connect to redis")
		return &RedisLogStream{client: nil}
	}
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
	return &RedisLogStream{client: client}
}

func (r *RedisLogStream) Publish(channel string, stream io.ReadCloser) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if r.client != nil {
			err := r.client.Publish(channel, scanner.Text()).Err()
			if err != nil {
				panic(err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		stream.Close()
	}
}

// TODO: return a generic type to include Subscribe to the interface
func (r *RedisLogStream) Subscribe(channel string) <-chan *redis.Message {
	return r.client.PSubscribe(channel).Channel()
}

func (r *RedisLogStream) HeartBeat(name string, quit chan int) {
	key := fmt.Sprintf("service.%s", name)

	if r.client != nil {
		r.client.Set(key, time.Now(), 20*time.Second)
	}

	for {
		select {
		case <-quit:
			if r.client != nil {
				r.client.Del(key)
			}
			return

		case <-time.After(10 * time.Second):
			if r.client != nil {
				r.client.Expire(key, 20*time.Second)
			}
		}
	}
}
