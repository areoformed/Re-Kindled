#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUT_DIR="$ROOT_DIR/build/macos"
GO_CACHE="$ROOT_DIR/build/go-cache"

mkdir -p "$OUT_DIR" "$GO_CACHE"
cd "$ROOT_DIR/mcp/server"

GOCACHE="$GO_CACHE" CGO_ENABLED=0 go build \
    -buildvcs=false -trimpath -ldflags='-s -w' \
    -o "$OUT_DIR/rekindled-mcp" .

file "$OUT_DIR/rekindled-mcp"
du -h "$OUT_DIR/rekindled-mcp"
