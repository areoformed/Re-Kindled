#!/usr/bin/env python3
"""Fail if a public ReKindled tree contains common private-release artifacts."""

from __future__ import annotations

import argparse
import re
from pathlib import Path


FORBIDDEN_SUFFIXES = {
    ".ttf", ".otf", ".ttc", ".pem", ".key", ".p12",
    ".jpg", ".jpeg", ".png", ".heic", ".bin", ".zip", ".tar", ".gz",
}
FORBIDDEN_PARTS = {"build", ".git", "backups", "device-backup", "vendor"}
TEXT_PATTERNS = {
    "macOS personal path": re.compile(r"/Users/[A-Za-z0-9._-]+/"),
    "Linux personal path": re.compile(r"/home/[A-Za-z0-9._-]+/"),
    "private key": re.compile(r"BEGIN (?:OPENSSH |RSA |EC )?PRIVATE KEY"),
    "clipboard temporary path": re.compile("codex-" + "clipboard-"),
}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", type=Path)
    args = parser.parse_args()
    root = args.root.resolve()
    failures: list[str] = []
    files = 0
    total_bytes = 0

    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root)
        if any(part in FORBIDDEN_PARTS for part in relative.parts):
            continue
        if not path.is_file():
            continue
        files += 1
        total_bytes += path.stat().st_size
        if path.suffix.lower() in FORBIDDEN_SUFFIXES:
            failures.append(f"forbidden public artifact: {relative}")
            continue
        raw = path.read_bytes()
        if b"\x00" in raw:
            failures.append(f"unexpected binary file: {relative}")
            continue
        text = raw.decode("utf-8", errors="replace")
        for label, pattern in TEXT_PATTERNS.items():
            if pattern.search(text):
                failures.append(f"{label} in {relative}")

    if failures:
        raise SystemExit("Public release audit failed:\n- " + "\n- ".join(failures))
    print(f"Public release audit passed: {files} source files, {total_bytes} bytes; no private artifacts found.")


if __name__ == "__main__":
    main()
