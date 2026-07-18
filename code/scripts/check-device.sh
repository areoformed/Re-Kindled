#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
SSH_CONFIG=${REKINDLED_SSH_CONFIG:-"$ROOT_DIR/device-config/ssh_config"}
HOST=${REKINDLED_HOST:-rekindled-pw5}

if [ ! -f "$SSH_CONFIG" ]; then
    echo "Missing $SSH_CONFIG; copy device-config/ssh_config.example and add your key." >&2
    exit 1
fi

ssh -F "$SSH_CONFIG" "$HOST" 'BASE=/mnt/us/rekindled
printf "connected\tyes\n"
if [ -x /mnt/us/libkh/bin/fbink ]; then printf "fbink\tready\n"; else printf "fbink\tmissing\n"; fi
if [ -f "$BASE/reader.pid" ] && kill -0 "$(cat "$BASE/reader.pid")" 2>/dev/null; then
    printf "reader\trunning\n"
else
    printf "reader\tstopped\n"
fi
printf "preset\t"
if [ -f "$BASE/active-preset" ]; then cat "$BASE/active-preset"; else printf "not selected\n"; fi
if [ -f "$BASE/bin/rekindled-reader" ]; then wc -c "$BASE/bin/rekindled-reader"; fi
if [ -f "$BASE/reader.log" ]; then tail -n 5 "$BASE/reader.log"; fi'
