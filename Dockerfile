# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

ARG GO_VERSION=1.23
ARG IMAGE_OS=alpine
ARG IMAGE_OS_VERSION=3.19
ARG TARGETARCH
ARG TARGETOS


# Development Image
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION AS dev

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./

RUN go mod download

CMD ["./tasks.sh", "watch"]


# Multi-stage Build Image
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION-$IMAGE_OS$IMAGE_OS_VERSION AS build

WORKDIR /app

RUN apk add git && apk add bash

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOARCH=$TARGETARCH GOOS=$TARGETOS ./tasks.sh build --release


# Production Image
FROM --platform=$BUILDPLATFORM $IMAGE_OS:$IMAGE_OS_VERSION AS prod

COPY --from=build /app/bin/build_release /app/build_release

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

EXPOSE 8443

CMD ["/app/build_release" ]