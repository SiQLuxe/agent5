# Build Script Design

## Goal
Create a consistent build script for the agent-tui Go project, producing a reliably named binary with embedded version information.

## Changes

### 1. `cmd/agent/main.go` — Add version variables
Add package-level `version`, `commit`, and `date` variables, populated at build time via `-ldflags`.

### 2. `scripts/build.sh` — New build script
- Fixed output binary name: `agent`
- Injects git tag/commit/date via `-ldflags`
- Falls back gracefully when git is unavailable
- Follows existing `scripts/` conventions (`set -euo pipefail`, ROOT detection)

### 3. `.gitignore` — Update artifact pattern
Change `agent-tui` to `/agent` to match the new binary name and scope it to root only.

## Build Command
```bash
VERSION=v1.0.0 ./scripts/build.sh    # explicit version
./scripts/build.sh                    # auto-detect from git
```

## Output
```
agent   (binary at project root, e.g. agent-tui/agent)
```
