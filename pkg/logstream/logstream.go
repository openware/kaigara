package logstream

import "io"

type LogStream interface {
	Publish(channel string, stream io.ReadCloser)
	HeartBeat(name string, quit chan int)
}
