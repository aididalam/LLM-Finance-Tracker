dev:
	air -c .air.toml

run:
	go run ./cmd/main

build:
	go build -o bin/main ./cmd/main