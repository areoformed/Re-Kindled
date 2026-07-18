# ReKindled agent kit

This directory lets a fresh coding agent understand the display without loading the entire implementation into context.

## Install the skill

Copy `skills/rekindled-display` into the skills directory supported by the agent client. For Codex, a typical personal installation is:

```sh
cp -R skills/rekindled-display "$CODEX_HOME/skills/"
```

Then ask the agent to use `$rekindled-display` to set up, operate, tune, or diagnose the device. The skill routes the agent to only the reference needed for the current task.

`AI-HANDOFF.md` is the standalone orientation document for agents that do not support installable skills. `tools/audit-public-release.py` enforces the privacy boundary; `tools/make-release-manifest.py` produces deterministic checksums.
