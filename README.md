# ReKindled

ReKindled turns an already-jailbroken Kindle into a quiet, cable-local text display that agents can operate through a small MCP server.

It has two deliberately separate parts:

- [`code/`](code/) contains the Go reader, stdio MCP bridge, deployment scripts, typography presets, and a dependency-free Type Lab.
- [`agent-kit/`](agent-kit/) contains an installable agent skill and the concise references needed to build, operate, tune, or recover the display.

The design favors a few strong guarantees: no network service beyond the private USB link, no general shell tool exposed through MCP, no visible content swap until every page passes FBInk layout preflight, and no font binary in the public release.

The measured stripped MCP is 2.82 MB on the host; the Kindle-side reader plus one locally prepared reference face is about 2.54 MB. See [`code/FOOTPRINT.md`](code/FOOTPRINT.md).

## Start here

1. Read [`code/README.md`](code/README.md).
2. Open [`code/typography-lab/index.html`](code/typography-lab/index.html) to explore physical type geometry with dials.
3. Follow [`code/REPLICATION.md`](code/REPLICATION.md) when the Kindle side is ready.
4. Give another agent [`agent-kit/AI-HANDOFF.md`](agent-kit/AI-HANDOFF.md), or install the skill from `agent-kit/skills/rekindled-display`.

## Release boundary

This package does **not** contain a jailbreak, firmware image, device backup, SSH private key, generated binary, device log, photograph, personal path, or commercial font. It begins at the point where a Kindle already has a supported jailbreak, USB networking, FBInk, and key-based SSH.

See [`THIRD_PARTY.md`](THIRD_PARTY.md) for font and dependency responsibilities and [`SECURITY.md`](SECURITY.md) before connecting a device.

## License

ReKindled's original source and documentation are MIT licensed. Third-party programs and fonts retain their own licenses and are not covered by this license.
