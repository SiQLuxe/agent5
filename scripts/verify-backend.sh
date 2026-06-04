#!/bin/bash
# verify-backend.sh — 一键验证 agent-backend 模块
#
# Usage:
#   ./scripts/verify-backend.sh          # unit tests only
#   ./scripts/verify-backend.sh --full   # unit + integration
#   ./scripts/verify-backend.sh --coverage  # with coverage

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  ❌ $1"; }

echo "=========================================="
echo "  Agent Backend Verification"
echo "=========================================="
echo ""

# ── 1. Build ──────────────────────────────────
echo "--- Step 1: Build ---"
if go build ./... 2>&1; then
    pass "go build ./..."
else
    fail "go build ./..."
fi

# ── 2. Lint (if golangci-lint available) ──────
echo ""
echo "--- Step 2: Lint ---"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./... 2>&1; then
        pass "golangci-lint"
    else
        fail "golangci-lint"
    fi
else
    echo "  ⚠️  golangci-lint not found, skipping"
fi

# Verify vet
if go vet ./... 2>&1; then
    pass "go vet ./..."
else
    fail "go vet ./..."
fi

# ── 3. Unit Tests (short) ─────────────────────
echo ""
echo "--- Step 3: Unit Tests ---"
if go test -short -count=1 -timeout=60s ./... 2>&1; then
    pass "go test -short (all packages)"
else
    fail "go test -short (all packages)"
fi

# ── 4. Race Detector ──────────────────────────
echo ""
echo "--- Step 4: Race Detector ---"
if go test -race -short -count=1 -timeout=120s ./... 2>&1; then
    pass "go test -race -short"
else
    fail "go test -race -short"
fi

# ── 5. Integration Tests (--full only) ────────
if [[ "${1:-}" == "--full" || "${1:-}" == "--coverage" ]]; then
    echo ""
    echo "--- Step 5: Integration Tests ---"
    if command -v opencode &> /dev/null; then
        if go test -run 'Test(Start|Stop|Session|Integration|SendAndGet|Search|Events|ExternalAgent|ExecuteCommand|ReadFile)' -count=1 -timeout=300s -v ./internal/backend/opencode/ 2>&1; then
            pass "integration tests (opencode backend)"
        else
            fail "integration tests (opencode backend)"
        fi
    else
        echo "  ⚠️  opencode not found in PATH, skipping integration tests"
    fi

    # Just run the server tests without filter
    echo ""
    echo "--- Step 5b: Server package ---"
    if go test -count=1 -timeout=60s ./internal/server/... 2>&1; then
        pass "server package tests"
    else
        fail "server package tests"
    fi
fi

# ── 6. Coverage (--coverage only) ─────────────
if [[ "${1:-}" == "--coverage" ]]; then
    echo ""
    echo "--- Step 6: Coverage ---"
    mkdir -p /tmp/agent-tui-coverage
    if go test -short -count=1 -coverprofile=/tmp/agent-tui-coverage/cover.out -covermode=atomic ./... 2>&1; then
        go tool cover -func=/tmp/agent-tui-coverage/cover.out
        pass "coverage report generated"
    else
        fail "coverage"
    fi
fi

# ── Summary ────────────────────────────────────
echo ""
echo "=========================================="
echo "  Result: $PASS passed, $FAIL failed"
echo "=========================================="

exit $FAIL
