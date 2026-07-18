#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
mkdir -p "$ROOT_DIR/build/go-cache" "$ROOT_DIR/build/pycache"

FORMAT_DIFF=$(find "$ROOT_DIR/mcp/server" "$ROOT_DIR/kindle/reader" -name '*.go' -type f -exec gofmt -d {} +)
if [ -n "$FORMAT_DIFF" ]; then
    echo "$FORMAT_DIFF"
    echo "Go sources need gofmt" >&2
    exit 1
fi

(cd "$ROOT_DIR/mcp/server" && GOCACHE="$ROOT_DIR/build/go-cache" go test ./...)
(cd "$ROOT_DIR/kindle/reader" && GOCACHE="$ROOT_DIR/build/go-cache" go test ./...)
PYTHONPYCACHEPREFIX="$ROOT_DIR/build/pycache" python3 -m unittest discover -s "$ROOT_DIR/tests" -p 'test_*.py'
PYTHONPYCACHEPREFIX="$ROOT_DIR/build/pycache" python3 -m py_compile \
    "$ROOT_DIR/scripts/type-lab.py" \
    "$ROOT_DIR/scripts/materialize-type-recipe.py"

"$ROOT_DIR/scripts/build-mcp.sh"
"$ROOT_DIR/scripts/build-reader.sh"
python3 "$ROOT_DIR/../agent-kit/tools/audit-public-release.py" "$ROOT_DIR/.."

echo "All ReKindled source, geometry, build, and release-boundary checks passed."
