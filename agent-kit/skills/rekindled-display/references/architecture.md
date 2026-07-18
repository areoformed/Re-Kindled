# Architecture reference

ReKindled has two small Go processes with a private USB boundary:

```text
agent -> stdio MCP on host -> ssh/scp -> Kindle reader -> FBInk -> e-ink
                                                    -> evdev -> gestures
```

The host MCP owns JSON-RPC, schemas, embedded help, document validation, and fixed-purpose SSH commands. The device reader owns document parsing, FBInk preflight/draw, signals, touch input, and reversible power/UI cleanup. Both use only the Go standard library.

## Seven-tool surface

- `rekindled_show`: validate, upload, preflight, and display explicit pages.
- `rekindled_navigate`: signal one next/previous page turn.
- `rekindled_set_type`: preflight and activate a preset.
- `rekindled_status`: inspect cable, reader, preset, footprint, and recent diagnostics.
- `rekindled_presets`: list type geometry.
- `rekindled_type_lab`: calculate a non-destructive recipe.
- `rekindled_help`: return operating guidance by topic.

Resources are `rekindled://help`, `rekindled://presets`, `rekindled://status`, and `rekindled://typography`.

## Transaction invariant

Validate host input, upload `content.next`, run the reader's `-check-layout`, require FBInk's computed output to report no truncation, then rename to `content.txt` and signal/restart the reader. Never swap visible state before validation.

## Device adapter

Current defaults assume Linux ARMv7, `/mnt/us/rekindled`, `/mnt/us/libkh/bin/fbink`, `/dev/input/event1`, and an `awesome` stock window manager. Verify each on a new Kindle generation. Keep generation-specific differences in this adapter boundary rather than leaking them into MCP tool semantics.

## Touch and signals

The reader grabs evdev only while active. `SIGUSR1` is next, `SIGUSR2` previous, and `SIGHUP` reloads a preflighted document. Cleanup releases touch and resumes the stock UI. Single and double taps have no display action; triple taps and swipes are intentionally distinct.
