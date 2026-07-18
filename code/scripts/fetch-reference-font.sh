#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DEST_DIR=${1:-"$ROOT_DIR/build/source-fonts"}
FONT_URL=https://raw.githubusercontent.com/google/fonts/main/ofl/ptmono/PTM55FT.ttf
LICENSE_URL=https://raw.githubusercontent.com/google/fonts/main/ofl/ptmono/OFL.txt
FONT_SHA256=cbe732b3b8fd211fd986ebdfc9b870ddeca4faab0bb5425fc509b37f9b4ac804

mkdir -p "$DEST_DIR"
TMP_FONT=$(mktemp "${TMPDIR:-/tmp}/rekindled-ptmono.XXXXXX")
TMP_LICENSE=$(mktemp "${TMPDIR:-/tmp}/rekindled-ofl.XXXXXX")
trap 'rm -f "$TMP_FONT" "$TMP_LICENSE"' EXIT HUP INT TERM

curl --fail --location --silent --show-error "$FONT_URL" -o "$TMP_FONT"
if command -v shasum >/dev/null 2>&1; then
    ACTUAL=$(shasum -a 256 "$TMP_FONT" | awk '{print $1}')
elif command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "$TMP_FONT" | awk '{print $1}')
else
    echo "Need shasum or sha256sum to verify the download" >&2
    exit 1
fi
if [ "$ACTUAL" != "$FONT_SHA256" ]; then
    echo "PT Mono checksum changed; inspect upstream before using the new file" >&2
    echo "expected $FONT_SHA256" >&2
    echo "actual   $ACTUAL" >&2
    exit 1
fi

curl --fail --location --silent --show-error "$LICENSE_URL" -o "$TMP_LICENSE"
mv "$TMP_FONT" "$DEST_DIR/PTM55FT.ttf"
mv "$TMP_LICENSE" "$DEST_DIR/PT-Mono-OFL.txt"
trap - EXIT HUP INT TERM

echo "Fetched and verified PT Mono from the Google Fonts repository."
echo "Font:    $DEST_DIR/PTM55FT.ttf"
echo "License: $DEST_DIR/PT-Mono-OFL.txt"
echo "The Reserved Font Names are PT Sans, PT Serif, PT Mono, and ParaType; rename derivatives."
