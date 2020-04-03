build:
	go build -o bin/kaigara cmd/kaigara/kaigara.go
	go build -o bin/kaitail cmd/kaitail/kaitail.go

clean:
	rm -rf bin/*
