# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

ARG GO_VERSION=1.23
ARG IMAGE_OS_NAME=alpine
ARG IMAGE_OS_VERSION=3.20


# Development image
FROM golang:${GO_VERSION} AS development

WORKDIR /app

RUN go install github.com/air-verse/air@latest && go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN go mod download

ENTRYPOINT ["./run","watch"]


# Production builder image
FROM golang:${GO_VERSION}-${IMAGE_OS_NAME}${IMAGE_OS_VERSION} AS builder

WORKDIR /app

RUN apk add git && apk add bash

RUN go install github.com/air-verse/air@latest && go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN ./run build


# Production image
FROM ${IMAGE_OS_NAME}:${IMAGE_OS_VERSION} AS production

WORKDIR /app

COPY --from=builder /app/bin/main ./bin/main

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    non_root_user

USER non_root_user

ENTRYPOINT ["./bin/main"]