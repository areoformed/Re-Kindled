#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUT_DIR="$ROOT_DIR/build/kindle"
GO_CACHE="$ROOT_DIR/build/go-cache"

mkdir -p "$OUT_DIR" "$GO_CACHE"
cd "$ROOT_DIR/kindle/reader"

GOCACHE="$GO_CACHE" CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
    go build -buildvcs=false -trimpath -ldflags='-s -w' \
    -o "$OUT_DIR/rekindled-reader" .

file "$OUT_DIR/rekindled-reader"
du -h "$OUT_DIR/rekindled-reader"
