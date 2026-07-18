# ReKindled code

This directory contains a small host-to-Kindle display stack:

```text
agent -> stdio MCP -> ssh/scp over USB -> native reader -> FBInk -> e-ink
                                                -> evdev -> page gestures
```

The MCP and reader are separate, dependency-free Go programs. The HTML Type Lab has no build step or external asset. `fontTools` is needed only when writing line-gap data into a local OpenType derivative.

See [`FOOTPRINT.md`](FOOTPRINT.md) for measured artifact sizes: the host MCP is 2.82 MB and the incremental Kindle payload with one local face is about 2.54 MB.

## Prerequisites

- An already-jailbroken Kindle with a supported USB-networking package.
- Key-based root SSH reachable over the private USB link.
- FBInk installed as `/mnt/us/libkh/bin/fbink`.
- Go 1.22 or newer on the host.
- Python 3 plus optional `fontTools` for the reference font workflow.
- `ssh`, `scp`, and `curl` on the host.

This repository does not choose or perform a jailbreak. Firmware/device support changes and a failed exploit can damage or reset a device; use a current device-specific guide and make a verified backup before reaching this stage.

## Build

```sh
./scripts/build-mcp.sh
./scripts/build-reader.sh
```

The first output is a native macOS stdio MCP binary. The second is a stripped Linux/ARMv7 reader for the Kindle.

## Typography dials

Open `typography-lab/index.html` in a browser. Its four primary dials are body pixels, absolute baseline pitch, viewing distance, and side margins. It calculates angular cap height, rounded line-gap pixels, the minimal `hhea.lineGap` units that produce them, and estimated lines per page. Load any compatible preset JSON or local font for preview; export the result as an agent-readable recipe.

For reproducible or automated visual trials, the lab also accepts query parameters such as `?body=47&pitch=63&distance=18&margin=120`.

The same calculation is available without a browser:

```sh
python3 scripts/type-lab.py \
  kindle/reader/presets/mac/rekindled-mono-air.json \
  --body-change-percent -5 \
  --line-pitch-change-percent 10 \
  --write-recipe build/type/my-recipe.json
```

The browser and CLI do not modify a font or Kindle. Read [`TYPOGRAPHY.md`](TYPOGRAPHY.md) before treating a preview as evidence.

## Prepare the open reference face

```sh
python3 -m pip install -r requirements-fonts.txt
./scripts/prepare-reference-type.sh
```

The script downloads and checksums the upstream PT Mono source plus its OFL, renames the derivative internally to comply with the font's Reserved Font Name clause, applies the physical reference line gap, and writes ignored results beneath `build/type/`. No font binary is checked in.

## Connect and deploy

```sh
cp device-config/ssh_config.example device-config/ssh_config
# Edit HostName and IdentityFile, then enroll the Kindle host key intentionally.
./scripts/check-device.sh
./scripts/deploy.sh
```

Deployment stops the existing reader safely, uploads to temporary names, swaps files, preflights the welcome document with FBInk, then starts the reader. See [`REPLICATION.md`](REPLICATION.md) for the complete sequence.

## Run the MCP server

```sh
./build/macos/rekindled-mcp -manual
./build/macos/rekindled-mcp \
  -root "$PWD" \
  -ssh-config "$PWD/device-config/ssh_config"
```

The second command speaks MCP over stdin/stdout. Copy `mcp-config.example.json` into your agent client's MCP configuration and replace its absolute paths.

`examples/an-old-page.txt` is a tiny saved specimen in the reader's native document format. It can be used for a first render or a quiet moment after setup.

## Tests

```sh
./scripts/test.sh
```

The test suite covers protocol discovery, content safety, gestures, preset validation, the physical type calculation, the Python CLI, cross-compilation, public-boundary scanning, and a built-in HTML smoke test. A physical panel trial remains a manual acceptance test.
