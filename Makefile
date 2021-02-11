build:
	go build -o bin/kaigara cmd/kaigara/kaigara.go
	go build -o bin/kaitail cmd/kaitail/kaitail.go
	go build -o bin/kaisave cmd/kaisave/kaisave.go

clean:
	rm -rf bin/*
