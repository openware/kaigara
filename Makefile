APPS := kaienv kaitail kaidump kaidel
KAIGARA_VERSION ?= "1.0.0"

build: all

all: kaigara kaicli

kaigara:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara

kaicli:
	./kaibuild.sh

clean:
	rm -rf bin/*

start-dev:
	docker-compose up -Vd

stop-dev:
	docker-compose down
