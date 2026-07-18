# Security model

ReKindled assumes a device owner knowingly operates an already-jailbroken Kindle. Jailbreaking removes parts of the stock security boundary; understand the device-specific consequences before continuing.

## Intended boundary

- Bind SSH to the private USB network, not Wi-Fi or a public interface.
- Use one dedicated key with `IdentitiesOnly yes`; never copy the private key onto the Kindle.
- Keep password authentication disabled.
- Keep host-key checking enabled after the first intentional enrollment.
- Expose only the seven documented MCP tools. Do not add a general shell or arbitrary file-write tool.
- Treat content sent to the display as potentially sensitive. It remains on the host and Kindle until replaced or deleted.

The included `ssh_config.example` intentionally points to a cable-local address and a dedicated key. Review it for the network layout of your device.

## Safe content update

The host uploads to `content.next`. The device reader asks FBInk to compute every page with truncation detection. Only after all pages fit does the host atomically replace `content.txt` and signal the live reader. A failed preflight leaves the visible document unchanged.

## Recovery invariant

The launch script pauses the stock window manager so the reader can claim touch input. The stop path and reader cleanup resume it. If a reader crashes, run `kindle/stop-reader.sh` over SSH or execute `killall -CONT awesome` on the Kindle.

## What to report

Do not attach device backups, private keys, personal documents, or full logs to a public issue. Reduce a report to the device generation, firmware family, command, and sanitized error.
