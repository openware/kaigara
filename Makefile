APPS := kaienv kaitail kaidump kaidel
KAIGARA_VERSION ?= "1.0.0"

build: all

all:	$(APPS) kaigara kaicli kaisave

$(APPS):
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/$@ ./cmd/$@/*.go

kaicli:
	CGO_ENABLED=0 go build -a -ldflags "-w -X main.Version=${KAIGARA_VERSION}" -o bin/kai ./cmd/kaicli

kaigara:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara

kaisave:
	chmod +x ./build-kaisave.sh
	./build-kaisave.sh

clean:
	rm -rf bin/*
