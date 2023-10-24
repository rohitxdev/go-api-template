.PHONY: build build-prod start run dev test bench

build : 
	rm -rf ./bin/* && go build -o ./bin/main ./cmd/main

build-prod:
	rm -rf ./bin/* && go build -trimpath -ldflags="-extldflags=-static -s -w" -tags osusergo,netgo -buildmode=pie -o ./bin ./cmd/main

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