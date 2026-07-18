---
name: rekindled-display
description: Build, connect, operate, tune, or recover a ReKindled cable-local Kindle text display and its lightweight MCP server. Use when an agent needs to deploy the Go reader, configure USB SSH, send paginated text, navigate with MCP or touch, select a typography preset, calculate physical font size and baseline pitch, materialize a local font recipe, diagnose FBInk layout rejection, or safely restore the stock Kindle UI.
---

# ReKindled Display

Operate ReKindled as a narrow physical display, not a general remote shell.

## Establish scope

Locate the release `code/` directory. Confirm the Kindle is already jailbroken, has USB networking and FBInk, and is owned by the user. Do not obtain or run a jailbreak from this skill. Do not copy private keys, device backups, commercial fonts, personal documents, or logs into a public tree.

Choose the smallest relevant workflow:

- For a new host/device connection or deployment, read [setup-and-recovery.md](references/setup-and-recovery.md).
- For text display or MCP integration, read [mcp-protocol.md](references/mcp-protocol.md).
- For body size, visual angle, line height, font preparation, or preset experiments, read [typography.md](references/typography.md).
- For code changes or boundary questions, read [architecture.md](references/architecture.md).

## Operate through MCP

Call `rekindled_status` before recovery work. Call `rekindled_presets` before selecting type. Use `rekindled_show` with deliberately paginated complete thoughts; never assume scrolling will rescue an overflowing page. Treat a layout error as a useful safety result: split the named page at a semantic boundary and retry.

Use `rekindled_navigate` for cable-driven page turns. Kindle touch remains available: triple-tap right or swipe up for next, triple-tap left or swipe down for previous, and swipe right from the far-left edge to exit.

Read `rekindled://help` for the complete embedded manual and `rekindled://typography` for the physical experiment loop. Prefer those live resources over assumptions when the MCP is running.

## Tune physical typography

Keep body pixels and absolute baseline pitch as independent dials. Use `rekindled_type_lab`, `scripts/type-lab.py`, or `typography-lab/index.html`; all three are non-destructive calculators. Express line-pitch changes relative to the source preset's physical baseline, not as a multiplier on the resized body.

Export a recipe, then use `scripts/materialize-type-recipe.py` only with a legally obtained local font. Give every derivative a new internal family name and a new preset ID. Run FBInk preflight on a representative multi-page specimen before activation. Judge the physical panel at a recorded distance; never call a browser preview proof of Kindle rendering.

## Preserve invariants

- Keep SSH key-only and cable-local.
- Keep strict host-key checking enabled after intentional enrollment.
- Never add arbitrary shell, arbitrary remote-path, network-fetch, or general file-write MCP tools.
- Upload content to a temporary path, preflight every page, and atomically promote only on success.
- Resume the stock window manager whenever the reader stops or fails.
- Preserve the previous visible document and preset when layout validation fails.
- Keep fonts and generated binaries outside the public source release.

## Verify completion

Run `code/scripts/test.sh`. For a device deployment, additionally verify welcome-page rendering, both triple-tap directions, both vertical swipes, the edge-exit gesture, stock-UI touch recovery, MCP status, and a rejected overflow page that leaves the old display intact.
