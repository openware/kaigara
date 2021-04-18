package logstream

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
	buf := make([]byte, 64)
	for {
		n, err := stream.Read(buf)
		os.Stdout.Write(buf[:n])

		if r.client != nil {
			e := r.client.Publish(channel, buf).Err()
			if e != nil {
				panic(e)
			}
		}
		if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed){
			break
		}

		// Debug block in case of unexpected error is returned
		if n == 0 {
			log.Println(fmt.Sprintf("ERR: %v", err))
			log.Println(fmt.Sprintf("Additional infromation:\n Bytes read: %d\n,Buf: %s", n, buf))
		}
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
