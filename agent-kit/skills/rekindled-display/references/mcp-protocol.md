# MCP operation

Run `rekindled-mcp -manual` for human-readable help or start it without `-manual` as a newline-delimited stdio JSON-RPC server. Protocol messages use stdout; diagnostics use stderr; closing stdin exits cleanly.

## Normal sequence

1. Optionally call `rekindled_status`.
2. Call `rekindled_presets` when selecting typography.
3. Call `rekindled_show` with semantic pages.
4. Use `rekindled_navigate` for host-driven page turns.

Example tool arguments:

```json
{
  "title": "Morning brief",
  "pages": [
    "The important thing first. End this page at a complete thought.",
    "The second page begins cleanly and can stand on its own."
  ],
  "preset": "rekindled-mono-air"
}
```

Do not place `---PAGE---` in page content; it is the reserved device document separator. Titles must be one line. A tool error on overflow is expected safety behavior, not permission to bypass preflight.

## Type lab arguments

`rekindled_type_lab` requires a preset ID and accepts:

- `body_size_change_percent`, from -40 to 100; or
- `target_cap_height_degrees`, from 0.15 to 1.5; not both;
- `line_pitch_change_percent`, from -20 to 100, relative to old absolute pitch;
- optional `viewing_distance_inches`, from 6 to 60.

It changes no file and does not contact the Kindle. Its `recipe` object can be passed to the local materializer.

## Failures

- Connection error: inspect the returned SSH detail and cable state.
- Layout rejection: split the named page at a paragraph boundary.
- Preset rejection: list exact preset IDs and retry.
- Reader stopped: a successful show call launches it.
- End-of-document navigation: no-op by design.
