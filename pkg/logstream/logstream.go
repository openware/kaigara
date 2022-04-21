package logstream

import "io"

type LogStream interface {
	Publish(channel string, stream io.ReadCloser) error
	HeartBeat(name string, quit chan int)
}
