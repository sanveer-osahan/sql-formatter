.PHONY: build run

build:
	go build -o bin/sqlfmt ./cmd/sqlfmt

run:
	go run ./cmd/sqlfmt $(ARGS)
