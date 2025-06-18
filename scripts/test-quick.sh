#!/usr/bin/env bash
# Quick test suite - equivalent to what runs in CI
# Run this for fast feedback during development

set -euo pipefail

echo "⚡ Quick Test Suite (CI equivalent)"
echo "=================================="
echo ""

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "internal" ]]; then
    echo "❌ Error: Please run this script from the nixai project root"
    exit 1
fi

echo "🔨 Building..."
go build ./...
echo "✅ Build successful"
echo ""

echo "🧪 Running core tests..."
go test -v ./internal/ai/function/...
go test -v ./internal/ai/context/...
go test -v ./internal/ai/
go test -v ./internal/config/...
go test -v ./internal/mcp/...
go test -v ./internal/nixos/...
go test -v ./pkg/...

echo ""
echo "✅ Quick test suite completed!"
echo "💡 Run './scripts/test-local-full.sh' for comprehensive testing"
