#!/bin/bash
# Test script to verify MCP port configuration changes

set -e

echo "🧪 Testing MCP Port Configuration..."

# Test 1: Check that we can read the mcp_port configuration
echo "Test 1: Reading mcp_port configuration"
cd /home/olafkfreund/Source/NIX/nix-ai-help
result=$(go run cmd/nixai/main.go config get mcp_port)
echo "Result: $result"

# Test 2: Verify configuration loading from embedded defaults
echo "Test 2: Testing configuration loading"
go run cmd/nixai/main.go config show | grep "MCP Server"

# Test 3: Test setting mcp_port configuration
echo "Test 3: Setting mcp_port to test value"
go run cmd/nixai/main.go config set mcp_port 12345
new_result=$(go run cmd/nixai/main.go config get mcp_port)
echo "New result: $new_result"

# Test 4: Reset to default
echo "Test 4: Resetting to default port"
go run cmd/nixai/main.go config set mcp_port 39847
final_result=$(go run cmd/nixai/main.go config get mcp_port)
echo "Final result: $final_result"

echo "✅ All tests completed!"
