#!/bin/sh
set -eu

BASE=/mnt/us/rekindled
PIDFILE="$BASE/reader.pid"

if [ ! -f "$PIDFILE" ]; then
    killall -CONT awesome 2>/dev/null || true
    echo "ReKindled reader is not running"
    exit 0
fi

PID=$(cat "$PIDFILE")
if kill -0 "$PID" 2>/dev/null; then
    kill "$PID"
    COUNT=0
    while kill -0 "$PID" 2>/dev/null && [ "$COUNT" -lt 30 ]; do
        usleep 100000
        COUNT=$((COUNT + 1))
    done
	if kill -0 "$PID" 2>/dev/null; then
		killall -CONT awesome 2>/dev/null || true
		echo "ReKindled reader did not stop cleanly as PID $PID" >&2
		exit 1
	fi
fi
rm -f "$PIDFILE"
killall -CONT awesome 2>/dev/null || true
echo "Stopped ReKindled reader"
