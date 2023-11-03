.PHONY: build build-prod start run dev test bench

include .env
export

DOCKER=docker run --rm -it -p 127.0.0.1:${PORT}:${PORT} -v $(shell pwd):/usr/src/app --env-file=.env go-api

build-image:
	docker build -t go-api .

build : 
	go build -o ./bin ./cmd/main

build-prod:
	go build -trimpath -ldflags="-extldflags=-static -s -w" -tags osusergo,netgo -buildmode=pie -o ./bin ./cmd/main

start:
	./bin/main

run:
	go run ./cmd/main

dev:
	air

test:
	go test -v ./...

bench:
	go test -v -bench=. ./...