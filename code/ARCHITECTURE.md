# Architecture and invariants

## Components

```text
Host                                         Kindle
-----------------------------------          --------------------------------
agent client                                 /mnt/us/rekindled/
  | stdio JSON-RPC                              content.txt
rekindled-mcp                                   active-preset
  | ssh + scp over USB                          fonts/
  |                                             presets/mac/
  +------------------------------------------>  bin/rekindled-reader
                                                   | FBInk compute + draw
                                                   | evdev touch grab
                                                   v
                                                e-ink panel
```

`mcp/server` implements the MCP transport, schemas, embedded help, status inspection, and a narrow SSH device adapter. `kindle/reader` owns document parsing, FBInk preflight/rendering, signals, power leases, and touch gestures. Neither program has a third-party Go dependency.

## MCP surface

The server exposes seven tools:

| Tool | Mutates | Purpose |
| --- | --- | --- |
| `rekindled_show` | yes | Preflight and atomically display explicit pages |
| `rekindled_navigate` | yes | Turn exactly one page |
| `rekindled_set_type` | yes | Preflight and activate a prepared preset |
| `rekindled_status` | no | Inspect cable, reader, preset, logs, and footprint |
| `rekindled_presets` | no | List local typography geometry |
| `rekindled_type_lab` | no | Calculate a candidate type recipe |
| `rekindled_help` | no | Read concise operating help |

There is no arbitrary command, arbitrary remote path, web request, or general filesystem tool.

## Content transaction

1. Validate title, page count, page length, total length, and the reserved separator on the host.
2. Upload the complete document as `content.next`.
3. Run the reader in `-check-layout` mode with the active or requested preset.
4. Ask FBInk to compute each page and reject `truncated=1`.
5. Only on success, rename `content.next` to `content.txt` and reload or restart the reader.

The old visible page is retained whenever steps 1–4 fail.

## Reader and input

The Kindle window manager normally owns touch input. `launch-reader.sh` pauses it reversibly, then the reader claims the configured evdev device. Cleanup releases the device, restores power state, and resumes the stock UI.

The gesture detector deliberately ignores single and double taps. A triple tap on either half changes pages; vertical swipes also navigate; a rightward swipe beginning at the far-left edge exits. MCP navigation sends `SIGUSR1` or `SIGUSR2`, so cable and touch inputs are peers.

## Typography path

FBInk is the device truth. The preset supplies physical calibration, font metrics, FBInk body pixels, margins, and device font paths. The font's own `hhea.lineGap` supplies interline space; the Type Lab converts an absolute baseline target into units that survive integer rounding. See [`TYPOGRAPHY.md`](TYPOGRAPHY.md).

## Portability boundary

The reader currently targets Linux ARMv7 and assumes the Kindle's framebuffer stack, `/dev/input/event1`, the `awesome` window manager, and the standard `/mnt/us` userstore. Device generations can differ. Treat those values as a device adapter to verify, not universal Kindle constants.
