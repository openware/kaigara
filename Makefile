build:
	./scripts/build.sh

clean:
	rm -rf bin/*

start-dev:
	docker-compose up -Vd

stop-dev:
	docker-compose down
