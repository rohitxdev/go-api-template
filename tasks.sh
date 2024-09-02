#!/usr/bin/env bash

set -o allexport

ENV_FILE="${ENV_FILE:-.env}"

if [[ -r $ENV_FILE ]]; then
    # shellcheck disable=SC1090
    source "$ENV_FILE"
else
    printf "\033[93m%s\033[0m\n" "Warning: Cannot read from \"$ENV_FILE\". Make sure the file path is correct & has required permissions." >&2
fi

GO_VERSION=$(go list -m -f "{{.GoVersion}}")
PACKAGE_NAME=$(go list -m)
APP_NAME=${PACKAGE_NAME##*/}
APP_VERSION=$(git describe --tags --always --dirty)
DOCKERHUB_REPO_NAME=$(echo "$PACKAGE_NAME" | awk -F'/' '{print $2 "/" $3}')

set +o allexport

set -e

menu=" 1) docker-watch    -  Run development server in docker
 2) docker-build    -  Build docker image
 3) docker-push     -  Push docker image to dockerhub
 4) watch           -  Run go app and watch for changes
 5) run             -  Run go app without building (for development)
 6) start           -  Run go app build
 7) build           -  Build go app (for release)
 8) test            -  Run tests
 9) test-cover      -  Run tests and show coverage
10) bench           -  Run benchmarks
11) checkpoint      -  Create a git checkpoint and push to remote
"

script_name=${0#$"./"}
script_path=$(readlink -f "$script_name")
task=$1

if [[ -z $task ]]; then
    printf "%s\n\033[1;95m%s\033[94m%s\n\033[0m\n" """$menu""" "Pick a task to run " "(Enter Q|q to quit)"
    printf "\033[1m%s\033[0m" "Enter option: "
    read -r task
fi

build_debug_args=(--ldflags="-X $PACKAGE_NAME/pkg/config.BuildInfo=$APP_NAME::$APP_VERSION::debug" --race)
build_release_args=(--ldflags="-s -w -X $PACKAGE_NAME/pkg/config.BuildInfo=$APP_NAME::$APP_VERSION::release -extldflags=-static" --trimpath --buildmode=pie)

while true; do
    case $task in
    "docker-watch" | 1)
        compose_file="docker-compose.yml"
        docker compose -f "$compose_file" down --remove-orphans
        docker compose -f "$compose_file" up --build
        break
        ;;
    "docker-build" | 2)
        docker build --tag="$APP_NAME":"$APP_VERSION" --target="prod" --build-arg="GO_VERSION=$GO_VERSION" .
        break
        ;;
    "docker-push" | 3)
        docker login
        docker tag "$APP_NAME":"$APP_VERSION" "$DOCKERHUB_REPO_NAME":"$APP_VERSION"
        docker push "$DOCKERHUB_REPO_NAME":"$APP_VERSION"
        break
        ;;
    "watch" | 4)
        watchexec --stop-timeout="${SHUTDOWN_TIMEOUT:-10s}" --stop-signal=SIGINT --shell=none --restart ./tasks.sh run
        break
        ;;
    "run" | 5)
        CGO_ENABLED=0 go run "${build_debug_args[@]}" .
        break
        ;;
    "start" | 6)
        if [[ -x "./bin/build" ]]; then
            ./bin/"build"
        else
            echo "Build not found. Building..."
            sleep 1
            $script_path build && ./bin/build
        fi
        break
        ;;
    "build" | 7)
        echo "Building app..." && CGO_ENABLED=0 go build "${build_release_args[@]}" -o ./bin/build ./main.go && echo "Built app successfully ✔"
        break
        ;;
    "test" | 8)
        go test --race --count=2 -v ./...
        break
        ;;
    "test-cover" | 9)
        go test --race --coverprofile=./tmp/coverage.out ./... && go tool cover --html=./tmp/coverage.out
        break
        ;;
    "bench" | 10)
        go test --race --count=2 -v -benchmem --bench=. ./...
        break
        ;;
    "checkpoint" | 11)
        git add . && git commit -m "Checkpoint at $(date "+%Y-%m-%dT%H:%M:%S%z")" && git push && echo "Checkpoint created and pushed to remote ✔"
        break
        ;;
    *)
        if [[ $task =~ ^[qQ]$ ]]; then
            printf "\033[1;34m%s\033[0m\n" "Quitting..."
            break
        fi
        printf "\033[1;31m%s\033[0m\n" "Invalid option"
        printf "\033[1m%s \033[0m" "Enter option:"
        read -r task
        ;;
    esac
done
