#!/usr/bin/env bash

GO_VERSION=$(awk 'NR==3 {print $2}' go.mod)
PACKAGE_NAME=$(awk 'NR==1 {print $2}' go.mod)
APP_NAME=$(echo "$PACKAGE_NAME" | awk '{split($0, arr, "/"); print arr[3]}')
APP_VERSION=$(git describe --tags --always --long --dirty)
DOCKERHUB_REPO_NAME=$(echo "$PACKAGE_NAME" | awk '{split($0, arr, "/"); print arr[2] "/" arr[3]}')

set -e

if [[ -z $ENV_FILE ]]; then
    ENV_FILE=".env"
    fallback=true
fi

if [[ -r $ENV_FILE ]]; then
    for line in $(<"$ENV_FILE"); do
        export "${line[*]}"
    done
else
    if [[ ! $fallback ]]; then
        printf "\033[93m%s\033[0m\n" "Warning: Cannot read from \"$ENV_FILE\". Make sure the file path is correct & has required permissions." >&2
    fi
fi

if [[ $APP_ENV == "production" ]]; then
    APP_VERSION=$(git describe --tags --always --long)
fi

export GO_VERSION PACKAGE_NAME APP_NAME APP_VERSION DOCKERHUB_REPO_NAME ENV_FILE

menu="
 1) docker-dev    -  Run development server in docker
 2) docker-prod   -  Run production server in docker
 3) docker-build  -  Build docker image
 4) docker-push   -  Push docker image to dockerhub
 5) watch         -  Run go app and watch for changes
 6) run           -  Run go app without building
 7) start         -  Run go app build
 8) build         -  Build go app (pass \"--release\" to optimize for release)
 9) test          -  Run tests (pass \"--cover\" to show coverage)
10) bench         -  Run benchmarks
11) checkpoint    -  Create a git checkpoint and push to remote
"

script_name=${0#$"./"}
script_path=$(readlink -f "$script_name")
task=$1
flag=$2

if [[ -z $task ]]; then
    printf "\033[1;95m%s\033[94m%s\033[0m%s\n" "Pick a task to run " "(Enter Q to quit)" "$menu"
    printf "\033[1m%s\033[0m" "Enter option: "
    read -r task flag
fi

build_release_args=(--ldflags="-s -w -X $PACKAGE_NAME/config.name=$APP_NAME -X $PACKAGE_NAME/config.version=$APP_VERSION -X $PACKAGE_NAME/config.build=release -extldflags=-static" --trimpath --buildmode=pie)
build_debug_args=(--ldflags="-X $PACKAGE_NAME/config.name=$APP_NAME -X $PACKAGE_NAME/config.version=$APP_VERSION -X $PACKAGE_NAME/config.build=debug" --race)

while true; do
    case $task in
    "docker-dev" | 1)
        compose_file="compose.dev.yml"
        docker compose -f "$compose_file" down --remove-orphans
        docker compose -f "$compose_file" up --build
        break
        ;;
    "docker-prod" | 2)
        compose_file="compose.prod.yml"
        docker compose -f "$compose_file" down --remove-orphans
        docker compose -f "$compose_file" up -d --build
        break
        ;;
    "docker-build" | 3)
        docker build --tag="$APP_NAME":"$APP_VERSION" --target="prod" --build-arg="GO_VERSION=$GO_VERSION" .
        break
        ;;
    "docker-push" | 4)
        docker login
        docker tag "$APP_NAME":"$APP_VERSION" "$DOCKERHUB_REPO_NAME":"$APP_VERSION"
        docker push "$DOCKERHUB_REPO_NAME":"$APP_VERSION"
        break
        ;;
    "watch" | 5)
        air
        break
        ;;
    "run" | 6)
        go run "${build_debug_args[@]}" ./main.go
        break
        ;;
    "start" | 7)
        build="debug"
        if [[ $flag == "--release" ]]; then
            build="release"
        fi
        if [[ -x "./bin/build_$build" ]]; then
            ./bin/build_"$build"
        else
            echo "Build not found"
            sleep 1
            $script_path build --$build && ./bin/build_"$build"
        fi
        break
        ;;
    "build" | 8)
        build="debug"
        if [[ $flag == "--release" ]]; then
            build="release"
        fi
        echo "Building app..."
        if [[ $build == "release" ]]; then
            CGO_ENABLED=0 go build "${build_release_args[@]}" -o ./bin/build_$build ./main.go
        else
            go build "${build_debug_args[@]}" -o ./bin/build_$build ./main.go
        fi
        echo "Built app successfully ✔"
        echo "name: $APP_NAME | version: $APP_VERSION | build: $build"
        break
        ;;
    "test" | 9)
        if [[ $flag != "--cover" ]]; then
            go test --count=2 -v ./...
        else
            go test --coverprofile=./tmp/coverage.out ./... && go tool cover --html=./tmp/coverage.out
        fi
        break
        ;;
    "bench" | 10)
        go test --count=2 -v --bench=. ./...
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
