.PHONY: build build-prod start run dev test bench

include .env
export 

PROJECT_ROOT=${PWD}

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