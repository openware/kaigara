APPS := kaienv kaitail kaidump kaidel

build: all

all:	$(APPS) kaigara kaisave

$(APPS):
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/$@ ./cmd/$@/*.go

kaigara:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara

kaisave:
	chmod +x ./build-kaisave.sh
	./build-kaisave.sh

clean:
	rm -rf bin/*
