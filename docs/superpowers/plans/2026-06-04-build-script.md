# Build Script Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a consistent build script (`scripts/build.sh`) that produces a reliably named `agent` binary with embedded version information.

**Architecture:** Three targeted changes: add version variables to `cmd/agent/main.go`, create `scripts/build.sh` with ldflags injection, and update `.gitignore` to match the new binary name.

**Tech Stack:** Go, shell script

---

### Task 1: Add version variables to main.go

**Files:**
- Modify: `cmd/agent/main.go:1` (after `package main`)

- [ ] **Step 1: Add version variables**

Add these package-level variables right after `package main` (before the import block):

```go
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./cmd/agent`
Expected: no errors, binary `agent` appears

- [ ] **Step 3: Commit**

```bash
git add cmd/agent/main.go
git commit -m "feat: add version ldflags variables to main package"
```

---

### Task 2: Create build script

**Files:**
- Create: `scripts/build.sh`

- [ ] **Step 1: Create scripts/build.sh**

```bash
#!/bin/bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "none")"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo "Building agent..."
echo "  Version: $VERSION"
echo "  Commit:  $COMMIT"
echo "  Date:    $DATE"

go build -ldflags="\
  -X main.version=$VERSION \
  -X main.commit=$COMMIT \
  -X main.date=$DATE" \
  -o agent ./cmd/agent

echo "Done: ./agent"
```

- [ ] **Step 2: Make executable and test**

```bash
chmod +x scripts/build.sh
./scripts/build.sh
```
Expected: builds `agent` binary at project root with version info

- [ ] **Step 3: Verify ldflags work (optional quick check)**

```bash
go tool nm ./agent | grep -E 'main\.(version|commit|date)'
```
Expected: shows the three symbols with injected values

- [ ] **Step 4: Commit**

```bash
git add scripts/build.sh
git commit -m "feat: add build script with version injection"
```

---

### Task 3: Update .gitignore

**Files:**
- Modify: `.gitignore:9`

- [ ] **Step 1: Replace agent-tui with /agent**

Change line 9 from `agent-tui` to `/agent`:

```
# 编译产物
/agent
```

- [ ] **Step 2: Verify git status**

Run: `git status`
Expected: `agent` binary not shown as untracked

- [ ] **Step 3: Commit**

```bash
git add .gitignore
git commit -m "chore: update gitignore to match agent binary name"
```
