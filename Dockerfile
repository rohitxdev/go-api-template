# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

ARG GO_VERSION
ARG IMAGE_OS_NAME
ARG IMAGE_OS_VERSION


# Development image
FROM golang:${GO_VERSION} AS development

WORKDIR /app

COPY go.mod go.sum run .git ./

RUN ./run init

ENTRYPOINT ["./run","watch"]


# Production builder image
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-${IMAGE_OS_NAME}${IMAGE_OS_VERSION} AS builder

WORKDIR /app

RUN apk add git && apk add bash

COPY go.mod go.sum run .git ./

RUN ./run init

COPY . .

RUN ./run build


# Production image
FROM --platform=${BUILDPLATFORM} ${IMAGE_OS_NAME}:${IMAGE_OS_VERSION} AS production

WORKDIR /app

COPY --from=builder /app/bin/main /app/bin/main

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

ENTRYPOINT ["/app/bin/main"]