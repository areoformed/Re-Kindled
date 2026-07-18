#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
SOURCE_FONT="$ROOT_DIR/build/source-fonts/PTM55FT.ttf"
OUTPUT_DIR="$ROOT_DIR/build/type"

if [ ! -f "$SOURCE_FONT" ]; then
    "$ROOT_DIR/scripts/fetch-reference-font.sh"
fi

mkdir -p "$OUTPUT_DIR"
python3 "$ROOT_DIR/scripts/materialize-type-recipe.py" \
    --source-font "$SOURCE_FONT" \
    --source-preset "$ROOT_DIR/kindle/reader/presets/mac/rekindled-mono-air.json" \
    --recipe "$ROOT_DIR/typography-lab/reference-recipe.json" \
    --output-font "$OUTPUT_DIR/ReKindledMonoAir-Regular.ttf" \
    --output-preset "$OUTPUT_DIR/rekindled-mono-air.json" \
    --preset-id rekindled-mono-air \
    --label "ReKindled Mono Air / physical reference" \
    --header "ReKindled / Mono Air / airy" \
    --derived-family "ReKindled Mono Air"

echo "Prepared the local reference face and preset in $OUTPUT_DIR"
