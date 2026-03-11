# V1Claw Project Report

Date: 2026-03-11
Status: hardening pass paused; work committed for resume later

## What was completed

- Fixed release/install blockers: source fallback in `install.sh`, added `install.ps1`, updated README install paths, and added installer smoke coverage in CI.
- Reworked workspace embedding and onboarding so builds do not depend on generated workspace copies and refresh no longer overwrites user files.
- Hardened `web_fetch` and redirect handling, added dial-time private IP blocking, and ensured fetched content is returned to the model.
- Replaced partial GitHub skill installs with archive-based directory installs, including scripts and references.
- Tightened execution boundaries: stronger `exec` tool filtering, sandbox-aware CLI provider flags, and workspace-aware shell policy for agent and cron execution.
- Hardened API/device transport: WebSocket registration token binding, bounded read/write behavior, disconnect cleanup, and dashboard auth no longer accepts API keys in query strings.
- Improved provider/auth behavior: timeout-aware OAuth requests, more reliable CLI worker argument ordering, and working GitHub Copilot `stdio` mode.
- Centralized cross-platform home/config resolution with `V1CLAW_HOME`, fixed history/config/auth paths, and cleaned up builtin skill fallback paths.
- Fixed migration portability: OpenClaw workspace migration now respects the target home instead of doing a blind string replacement.
- Added `onboard --auto --skip-test` for CI/offline setup, synced installer hints, and verified clean-home onboarding.
- Normalized `workspace.path` and `agents.defaults.workspace` so setup/config changes cannot drift the runtime workspace.
- Expanded `configure` to cover all advertised channels and added a dedicated permissions section for non-technical users.
- Fixed `doctor` so credential checks are provider-aware and do not falsely fail valid external-auth providers like Vertex and Bedrock.
- Added `notifications` and `screen` to the persisted permissions schema and runtime permission loading.

## Validation completed

- `go test ./...`
- `go vet ./...`
- `make build`
- `bash -n install.sh`
- Source installer smoke test
- Clean-home onboarding smoke test with `V1CLAW_HOME`
- Custom-workspace onboarding smoke test

## Remaining work when resuming

- Run live Windows/PowerShell install verification on a Windows machine.
- Continue publish-readiness audit for remaining feature-claim gaps and first-run UX edges.
- Final release pass: tag/release flow verification, install verification against published artifacts, and final README tightening.
