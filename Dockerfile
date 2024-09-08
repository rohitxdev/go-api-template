# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

ARG GO_VERSION=1.23
ARG IMAGE_OS=alpine
ARG IMAGE_OS_VERSION=3.19
ARG TARGETARCH
ARG TARGETOS


# Development image
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION AS development

WORKDIR /app

COPY go.mod go.sum tasks ./

RUN  ./tasks init

CMD ["./tasks","watch"]


# Production builder image
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION-$IMAGE_OS$IMAGE_OS_VERSION AS builder

WORKDIR /app

RUN apk add git && apk add bash

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOARCH=$TARGETARCH GOOS=$TARGETOS ./tasks build


# Production image
FROM --platform=$BUILDPLATFORM $IMAGE_OS:$IMAGE_OS_VERSION AS production

COPY --from=builder /app/bin/release_build /app/release_build

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

EXPOSE 8000

CMD ["/app/build" ]