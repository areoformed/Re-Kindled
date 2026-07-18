# Replication runbook

This runbook begins with a working, already-jailbroken Kindle. It intentionally does not reproduce a firmware-specific exploit.

## 1. Establish the cable boundary

1. Enable the device's supported USB-network mode.
2. Create a dedicated SSH key on the host.
3. Put only its public key in the Kindle root account.
4. Copy `device-config/ssh_config.example` to `device-config/ssh_config`.
5. Verify the USB address and edit `HostName` if necessary.
6. Enroll the expected host key deliberately, then keep strict checking enabled.
7. Run `./scripts/check-device.sh`.

Do not expose root SSH on Wi-Fi for this project.

## 2. Confirm the device adapter

Before deploying, verify:

- `uname -m` matches the ARMv7 build target or adjust `build-reader.sh`.
- `/mnt/us/libkh/bin/fbink` exists and is executable.
- the touchscreen evdev node is `/dev/input/event1`; use device diagnostics if it differs.
- the stock window manager process is named `awesome`.

Changing any of these is a code/configuration adaptation, not a typography adjustment.

## 3. Build and test on the host

```sh
python3 -m pip install -r requirements-fonts.txt
./scripts/test.sh
./scripts/prepare-reference-type.sh
```

Review the fetched font license under `build/source-fonts/` and the derived outputs under `build/type/`.

## 4. Deploy transactionally

```sh
./scripts/deploy.sh
```

The deployer uploads binaries, font, preset, scripts, and welcome content to `.next` paths, promotes them only after transfer, runs FBInk layout preflight, then launches. If preflight fails, correct the named page or preset; do not bypass it.

## 5. Verify physical behavior

1. Confirm the welcome page appears without truncation.
2. Triple-tap slowly on the right half; the second page should appear only after the third distinct tap.
3. Triple-tap left to return.
4. Swipe up and down to navigate.
5. Swipe right from the far-left edge to exit and confirm the stock UI regains touch.
6. Start the reader again with `/mnt/us/rekindled/launch-reader.sh` over SSH.

## 6. Connect an MCP client

Build the host server, run `-manual`, then adapt `mcp-config.example.json` with absolute local paths. On initialization, the server self-documents its seven tools and four resources. A prudent first agent call is `rekindled_status`; a normal content flow is `rekindled_presets`, then `rekindled_show` with explicit semantic pages.

## 7. Calibrate typography

Use the visual lab or `rekindled_type_lab`, export a recipe, materialize it to a newly named local font, make a new preset ID, preflight a representative specimen, and compare photographs taken at the same distance and lighting. Change no more than two dials between trials.

## Recovery

- Connection failure: leave USB-network mode active and run `ssh -F device-config/ssh_config -vv rekindled-pw5 true`.
- Reader stopped: run `./scripts/deploy.sh` or invoke the remote launch script.
- Stock UI has no touch: invoke the remote stop script; as a last reversible step, run `killall -CONT awesome` on the Kindle.
- Content rejected: split the named page at a paragraph boundary.
- Type rejected: return to the previous preset; the display transaction preserves it.
