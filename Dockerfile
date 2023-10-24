FROM golang:1.21.3-alpine3.17

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/app ./cmd/main

CMD ["app"]
