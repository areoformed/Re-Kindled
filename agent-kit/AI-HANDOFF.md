# ReKindled AI handoff

## Objective

Maintain a lightweight system that lets any MCP-capable agent render beautiful, deliberately paginated text on an already-jailbroken Kindle connected to a host by USB. Preserve touch navigation on the Kindle and offer physical typography calibration that another agent can reason about numerically.

## What exists

- A native Linux/ARMv7 Go reader that parses explicit pages, asks FBInk to preflight and draw them, owns touch only while active, responds to signals, and restores the stock UI on exit.
- A dependency-free Go stdio MCP server with seven narrow tools and four self-documenting resources.
- A reference preset representing the physically accepted 49 px body, 57 px baseline pitch, and approximately 0.325° cap height at 18 inches.
- Three matching type-calibration surfaces: MCP tool, Python CLI, and single-file HTML lab with dials.
- A font materializer that changes `hhea.lineGap`, removes old primary naming records, installs a new internal family name, and updates a preset.
- Transactional deployment, public-boundary audit, checksum manifest, tests, and an installable skill.

## Non-negotiable decisions

1. MCP is not a remote shell. Keep paths and operations fixed-purpose.
2. USB key-only SSH is the network boundary.
3. Every page is semantically pre-paginated by the sender.
4. FBInk computes the layout before visible content or type is replaced.
5. Body size and baseline pitch are independent physical dials.
6. Browser preview is not evidence of e-ink rendering quality.
7. Fonts are local inputs under their own licenses; public source ships no font binary.
8. The reader's touch grab must be reversible; stock UI recovery outranks display continuity.

## File map

- `../code/mcp/server/`: MCP transport, schemas, resources, device adapter, Type Lab.
- `../code/kindle/reader/`: document, preset, render, touch, power, signal logic.
- `../code/kindle/reader/presets/mac/`: public physical reference preset.
- `../code/typography-lab/`: visual instrument and exact first-trial recipe.
- `../code/scripts/`: builds, CLI lab, font materializer, deploy, device check, tests.
- `skills/rekindled-display/`: compact agent workflow.
- `tools/`: release audit and manifest generator.

## How to reproduce

Read `../code/REPLICATION.md`. The short version is: establish supported USB networking and key-only SSH, verify FBInk/input/window-manager paths, run tests, prepare the reference type locally, deploy transactionally, verify touch and exit behavior, build the host MCP, and connect it through absolute local paths.

## How to tune type

Read `../code/TYPOGRAPHY.md`. Start from an existing preset. Set body and absolute baseline pitch independently, export a JSON recipe, materialize to a newly named derivative, preflight multiple real pages, and judge a controlled physical comparison. Preserve unsuccessful trials as data, but do not promote them to the default.

## Definition of done for changes

- `code/scripts/test.sh` passes.
- Native MCP and ARMv7 reader build without third-party Go modules.
- Public audit finds no personal paths, keys, backups, binaries, photographs, logs, or font files.
- MCP help and skill references match the actual tool count, names, defaults, gestures, and recovery path.
- A device-impacting change has a physical test plan and a reversible stock-UI recovery path.
- `RELEASE-MANIFEST.json` and `SHA256SUMS` are regenerated last.
