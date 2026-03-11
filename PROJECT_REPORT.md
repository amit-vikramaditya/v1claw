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
- Added migration support for the newer provider/channel surfaces (`deepseek`, `nvidia`, `moonshot`, `ollama`, `github_copilot`, `slack`, `line`, `onebot`) so imports no longer drop real config.
- Aligned provider setup across `configure`, `onboard`, `onboard --auto`, runtime provider creation, `doctor`, `gateway`, and `status` for `zhipu`, `moonshot`, `ollama`, `vllm`, and `github_copilot`.
- Fixed provider onboarding order so providers without built-in default models are configured first, then modeled, then validated instead of failing before a model is chosen.
- Added explicit runtime support and regression coverage for keyless/self-hosted providers: `ollama`, `vllm`, and `github_copilot`, plus explicit `moonshot` provider creation.
- Synced README setup docs with the real non-interactive/local-provider flows, including `--api-base` support and keyless/local provider examples.
- Hardened release automation so the manual release workflow now reruns formatting, vet, test, and installer smoke jobs before tagging, refuses duplicate tags, and no longer hard-fails when Docker Hub credentials are intentionally absent.
- Made Docker publishing configuration optional-safe in both GoReleaser and the Docker workflow, so GitHub releases and GHCR publishing are not blocked by missing Docker Hub secrets or repository variables.
- Improved local-provider diagnostics in `doctor` and onboarding so down `ollama`, `vllm`, or `github_copilot` endpoints report actionable local fixes instead of generic API-key/internet advice.
- Updated installer end-of-run guidance to show both cloud-key and local-provider setup examples.
- Added `make release-check` and `scripts/release_check.sh` so publish readiness can be revalidated locally with one command.
- Added `RELEASING.md` to document the exact publish path, workflow expectations, and post-release verification.

## Validation completed

- `go test ./...`
- `go vet ./...`
- `make build`
- `bash -n install.sh`
- Source installer smoke test
- Clean-home onboarding smoke test with `V1CLAW_HOME`
- Custom-workspace onboarding smoke test
- Targeted provider/setup regression tests covering CLI setup, migration, and provider creation
- Clean local-provider smoke test with a stub OpenAI-compatible endpoint: installer, `onboard --auto`, `doctor`, and `agent -m` all passed end-to-end
- Workflow/YAML parse validation for build, PR, release, Docker, and GoReleaser config

## Remaining work when resuming

- Run live Windows/PowerShell install verification on a Windows machine.
- Publish a tagged GitHub release and verify the installer against the real release artifacts once they exist.
- Fix GitHub auth on the publishing machine so `main` and the release tag can actually be pushed.
