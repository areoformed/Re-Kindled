#!/bin/sh
set -eu

BASE=/mnt/us/rekindled
PIDFILE="$BASE/reader.pid"
LOGFILE="$BASE/reader.log"
ACTIVE_PRESET_FILE="$BASE/active-preset"
if [ -n "${REKINDLED_PRESET:-}" ]; then
    PRESET=$REKINDLED_PRESET
elif [ -f "$ACTIVE_PRESET_FILE" ]; then
    PRESET=$(cat "$ACTIVE_PRESET_FILE")
else
    PRESET="$BASE/presets/mac/rekindled-mono-air.json"
fi
CONTENT=${REKINDLED_CONTENT:-$BASE/content.txt}

if [ -f "$PIDFILE" ]; then
    OLD_PID=$(cat "$PIDFILE")
    if kill -0 "$OLD_PID" 2>/dev/null; then
        echo "ReKindled reader is already running as PID $OLD_PID"
        exit 0
    fi
fi

# Kindle's X server exclusively owns the touch device while the stock window
# manager is live. KOReader uses the same reversible handoff: pause Awesome,
# leave X and the framework running, and let the native reader claim evdev.
killall -STOP awesome 2>/dev/null || true

nohup "$BASE/bin/rekindled-reader" \
    -content "$CONTENT" \
    -preset "$PRESET" \
    >"$LOGFILE" 2>&1 </dev/null &
echo $! >"$PIDFILE"
PID=$!
usleep 300000
if ! kill -0 "$PID" 2>/dev/null; then
    killall -CONT awesome 2>/dev/null || true
    rm -f "$PIDFILE"
    echo "ReKindled reader failed to start; stock input restored" >&2
    if [ -f "$LOGFILE" ]; then
        tail -n 8 "$LOGFILE" >&2
    fi
    exit 1
fi
echo "Started ReKindled reader as PID $PID"
