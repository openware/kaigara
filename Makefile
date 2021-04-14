build:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaitail ./cmd/kaitail
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaisave ./cmd/kaisave
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaidump ./cmd/kaidump
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaidel ./cmd/kaidel

clean:
	rm -rf bin/*
