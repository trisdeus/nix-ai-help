#!/usr/bin/env bash
# Full local test suite for nixai
# Run this locally for comprehensive testing before committing

set -euo pipefail

echo "🧪 Running Full nixai Test Suite"
echo "================================="
echo ""

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "internal" ]]; then
    echo "❌ Error: Please run this script from the nixai project root"
    exit 1
fi

# Function to run tests with better output
run_test_package() {
    local package=$1
    local description=$2
    
    echo "🔍 Testing: $description"
    echo "   Package: $package"
    
    if go test -v "$package" 2>&1; then
        echo "✅ $description: PASSED"
    else
        echo "❌ $description: FAILED"
        return 1
    fi
    echo ""
}

# Core package tests (these run in CI too)
echo "📦 Core Package Tests (also run in CI)"
echo "======================================"
run_test_package "./internal/ai/function/..." "AI Function Integration"
run_test_package "./internal/ai/context/..." "AI Context Management"
run_test_package "./internal/ai/" "AI Provider Manager"
run_test_package "./internal/config/..." "Configuration Management"
run_test_package "./internal/mcp/..." "MCP Documentation Server"
run_test_package "./internal/nixos/..." "NixOS System Integration"
run_test_package "./pkg/..." "Utility Packages"

# Extended tests (local only)
echo "🏠 Extended Local Tests"
echo "======================"
run_test_package "./internal/ai/agent/..." "AI Agent System (local only)"
run_test_package "./internal/ai/validation/..." "AI Validation System (local only)"
run_test_package "./internal/cli/..." "CLI Commands"
run_test_package "./internal/community/..." "Community Features"
run_test_package "./internal/devenv/..." "Development Environment"
run_test_package "./internal/hardware/..." "Hardware Detection"
run_test_package "./internal/learning/..." "Learning Modules"
run_test_package "./internal/machines/..." "Machine Management"
run_test_package "./internal/neovim/..." "Neovim Integration"
run_test_package "./internal/packaging/..." "Package Analysis"

# Build tests
echo "🔨 Build Tests"
echo "=============="
echo "🔍 Testing: Go Build"
if go build ./...; then
    echo "✅ Go Build: PASSED"
else
    echo "❌ Go Build: FAILED"
    exit 1
fi

# Check if nix is available for Nix build test
if command -v nix >/dev/null 2>&1; then
    echo "🔍 Testing: Nix Flake Build"
    if nix build 2>/dev/null; then
        echo "✅ Nix Flake Build: PASSED"
    else
        echo "❌ Nix Flake Build: FAILED"
        exit 1
    fi
else
    echo "⚠️  Nix not available, skipping Nix build test"
fi

echo ""
echo "🎉 Full test suite completed successfully!"
echo ""
echo "💡 Tips:"
echo "   • Run individual test packages with: go test -v ./internal/cli/..."
echo "   • Run tests with coverage: go test -cover ./..."
echo "   • Use 'go test -short' to skip long-running tests"
echo "   • Check CI-only tests in .github/workflows/ci.yaml"
