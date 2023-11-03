FROM golang:1.21.3-alpine3.17

WORKDIR /usr/src/app

RUN apk add --no-cache make

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o ./bin/main ./cmd/main

CMD ./bin/main
