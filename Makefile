build:
	CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara
	CGO_ENABLED=0 go build -a -ldflags '-w' -o bin/kaitail ./cmd/kaitail
	sh ./build-kaisave.sh 

clean:
	rm -rf bin/*
