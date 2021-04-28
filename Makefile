build:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaitail ./cmd/kaitail
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaidump ./cmd/kaidump
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaidel ./cmd/kaidel
	chmod +x ./build-kaisave.sh
	./build-kaisave.sh

clean:
	rm -rf bin/*
