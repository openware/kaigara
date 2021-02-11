build:
	go build -o bin/kaigara ./cmd/kaigara
	go build -o bin/kaitail ./cmd/kaitail
	go build -o bin/kaisave ./cmd/kaisave

clean:
	rm -rf bin/*
