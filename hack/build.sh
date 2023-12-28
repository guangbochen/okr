#!/bin/bash
set -e

cd $(dirname $0)/..

if [ -z "$DEST_IP" ]; then
    echo "DEST_IP is not set"
    exit 1
fi

mkdir -p bin
if [ "$(uname)" = "Linux" ]; then
    OTHER_LINKFLAGS="-extldflags -static -s"
fi
LINKFLAGS="-X github.com/oneblock-ai/okr/pkg/version.Version=$VERSION"
LINKFLAGS="-X github.com/oneblock-ai/okr/pkg/version.GitCommit=$COMMIT $LINKFLAGS"
GOOS=linux CGO_ENABLED=0 go build -ldflags "$LINKFLAGS $OTHER_LINKFLAGS" -o bin/okr ./cmd/main.go

echo "Uploading okr to $DEST_IP"
scp ./bin/okr ubuntu@$DEST_IP:/home/ubuntu
