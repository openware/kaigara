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

func NewRedisClient(url string) (*RedisLogStream, error) {
	if url == "" {
		return nil, fmt.Errorf("redis url is empty")
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	if _, err = client.Ping().Result(); err != nil {
		return nil, err
	}

	return &RedisLogStream{client: client}, nil
}

func (r *RedisLogStream) Publish(channel string, stream io.ReadCloser) error {
	buf := make([]byte, 64)
	for {
		n, err := stream.Read(buf)
		os.Stdout.Write(buf[:n])

		if r.client != nil {
			err := r.client.Publish(channel, buf).Err()
			if err != nil {
				return err
			}
		} else if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) {
			return err
		}

		// Debug block in case of unexpected error is returned
		if n == 0 {
			log.Printf("ERR: %s\n", err.Error())
			log.Printf("INF: bytes read - %d, buf: %s\n", n, buf)
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
