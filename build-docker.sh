#!/bin/bash

set -euo pipefail

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build \
    -mod=vendor \
    -a -installsuffix cgo -ldflags '-extldflags "-static"' \
    -o docker/transferwise-exchange-rate \
    .

cd docker
docker build -t porty/transferwise-slack-bot .
