# Setup and recovery

Start only after the owner has a supported jailbreak, USB networking, FBInk, and a verified backup. Firmware-specific jailbreaking is outside this release.

## Host setup

From the release `code/` directory:

```sh
cp device-config/ssh_config.example device-config/ssh_config
# Set the cable-local HostName and dedicated IdentityFile.
./scripts/check-device.sh
./scripts/test.sh
python3 -m pip install -r requirements-fonts.txt
./scripts/prepare-reference-type.sh
./scripts/deploy.sh
```

The reference preparation fetches PT Mono plus its OFL from Google Fonts, verifies a pinned SHA-256 digest, renames the modified face to ReKindled Mono Air, and keeps outputs in ignored `build/` paths.

Build the MCP separately with `./scripts/build-mcp.sh`. Replace placeholders in `mcp-config.example.json` with absolute paths and install that entry in the agent client.

## Device checks

Verify architecture, FBInk executable path, touch evdev node, and stock window-manager process on every new device generation. Do not guess an input device and grab it blindly.

## Recovery order

1. Run `rekindled_status` or `./scripts/check-device.sh`.
2. Keep USB-network mode active and check the SSH alias directly.
3. If the reader is stopped, invoke `/mnt/us/rekindled/launch-reader.sh`.
4. If content or type is rejected, keep the old display and correct the named page/preset.
5. If the stock UI lacks touch, invoke `/mnt/us/rekindled/stop-reader.sh`.
6. If that script is unavailable, run `killall -CONT awesome` as the reversible UI-resume step.

Never factory-reset, replace firmware, erase user storage, disable host-key checking, or expose root SSH on Wi-Fi as routine recovery.

## Physical acceptance

Confirm page one renders, triple-tap navigation works in both halves, up/down swipes work, the far-left edge exit restores the stock UI, and an intentionally overlong page is rejected without replacing the old display.
