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
