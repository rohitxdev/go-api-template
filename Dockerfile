FROM golang:1.21.4-alpine3.17

WORKDIR /usr/src/app

RUN apk add --no-cache make

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN make build

CMD make start

EXPOSE 8443
