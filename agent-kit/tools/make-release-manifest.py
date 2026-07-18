#!/usr/bin/env python3
"""Write deterministic checksums and a machine-readable release manifest."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


GENERATED = {"SHA256SUMS", "RELEASE-MANIFEST.json"}
IGNORED_PARTS = {"build", ".git", "__pycache__"}


def digest(path: Path) -> str:
    value = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", type=Path)
    args = parser.parse_args()
    root = args.root.resolve()
    entries = []
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root)
        if not path.is_file() or relative.name in GENERATED:
            continue
        if any(part in IGNORED_PARTS for part in relative.parts):
            continue
        entries.append({
            "path": relative.as_posix(),
            "bytes": path.stat().st_size,
            "sha256": digest(path),
        })

    manifest = {
        "schema_version": 1,
        "name": "ReKindled",
        "version": (root / "VERSION").read_text(encoding="utf-8").strip(),
        "source_files": len(entries),
        "source_bytes": sum(item["bytes"] for item in entries),
        "files": entries,
    }
    (root / "RELEASE-MANIFEST.json").write_text(
        json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    (root / "SHA256SUMS").write_text(
        "".join(f'{item["sha256"]}  {item["path"]}\n' for item in entries), encoding="utf-8"
    )
    print(f"Wrote manifest for {len(entries)} files ({manifest['source_bytes']} bytes).")


if __name__ == "__main__":
    main()
