-include .env

protocol=http
ifeq ($(HTTPS),true)
	protocol=https
endif

.PHONY: build start run dev doc-gen bench test test-cover pprof

build:
	make doc-gen && rm -rf ./bin && CGO_ENABLED=0 go build -trimpath -buildmode=pie -o ./bin/main ./cmd/main

start:
	./bin/main

run:
	go run ./cmd/main

dev:
	air

doc-gen:
	swag init -q -g ./handler/router.go -o ./docs

bench:
	go test -count=4 -v -bench=. ./...

test:
	go test -v ./...

test-cover:
	go test -coverprofile=/tmp/coverage.out ./... && go tool cover -html=/tmp/coverage.out

pprof:
	curl -o ./cmd/main/default.pgo $(protocol)://$(HOST):$(PORT)/debug/pprof/profile?seconds=30
