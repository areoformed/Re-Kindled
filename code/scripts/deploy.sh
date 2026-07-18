#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
SSH_CONFIG=${REKINDLED_SSH_CONFIG:-"$ROOT_DIR/device-config/ssh_config"}
HOST=${REKINDLED_HOST:-rekindled-pw5}
BASE=/mnt/us/rekindled
FONT="$ROOT_DIR/build/type/ReKindledMonoAir-Regular.ttf"
PRESET="$ROOT_DIR/build/type/rekindled-mono-air.json"
READER="$ROOT_DIR/build/kindle/rekindled-reader"

if [ ! -f "$SSH_CONFIG" ]; then
    echo "Missing $SSH_CONFIG; copy device-config/ssh_config.example and add your key." >&2
    exit 1
fi
if [ ! -f "$FONT" ] || [ ! -f "$PRESET" ]; then
    "$ROOT_DIR/scripts/prepare-reference-type.sh"
fi
if [ ! -x "$READER" ]; then
    "$ROOT_DIR/scripts/build-reader.sh"
fi

ssh -F "$SSH_CONFIG" "$HOST" "set -eu
BASE=$BASE
if [ -x \"\$BASE/stop-reader.sh\" ]; then \"\$BASE/stop-reader.sh\"; fi
mkdir -p \"\$BASE/bin\" \"\$BASE/fonts\" \"\$BASE/presets/mac\"
test -x /mnt/us/libkh/bin/fbink"

scp -F "$SSH_CONFIG" "$READER" "$HOST:$BASE/bin/rekindled-reader.next"
scp -F "$SSH_CONFIG" "$FONT" "$HOST:$BASE/fonts/ReKindledMonoAir-Regular.ttf.next"
scp -F "$SSH_CONFIG" "$PRESET" "$HOST:$BASE/presets/mac/rekindled-mono-air.json.next"
scp -F "$SSH_CONFIG" "$ROOT_DIR/kindle/launch-reader.sh" "$ROOT_DIR/kindle/stop-reader.sh" "$HOST:$BASE/"
scp -F "$SSH_CONFIG" "$ROOT_DIR/welcome.txt" "$HOST:$BASE/content.next"

ssh -F "$SSH_CONFIG" "$HOST" "set -eu
BASE=$BASE
chmod 0755 \"\$BASE/bin/rekindled-reader.next\" \"\$BASE/launch-reader.sh\" \"\$BASE/stop-reader.sh\"
mv \"\$BASE/bin/rekindled-reader.next\" \"\$BASE/bin/rekindled-reader\"
mv \"\$BASE/fonts/ReKindledMonoAir-Regular.ttf.next\" \"\$BASE/fonts/ReKindledMonoAir-Regular.ttf\"
mv \"\$BASE/presets/mac/rekindled-mono-air.json.next\" \"\$BASE/presets/mac/rekindled-mono-air.json\"
\"\$BASE/bin/rekindled-reader\" -check-layout -content \"\$BASE/content.next\" -preset \"\$BASE/presets/mac/rekindled-mono-air.json\"
mv \"\$BASE/content.next\" \"\$BASE/content.txt\"
printf '%s\n' \"\$BASE/presets/mac/rekindled-mono-air.json\" > \"\$BASE/active-preset\"
\"\$BASE/launch-reader.sh\""

echo "ReKindled deployed and preflighted."
"$ROOT_DIR/scripts/check-device.sh"
