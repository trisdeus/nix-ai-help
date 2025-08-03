#!/bin/bash

# Script to build the example plugin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building example plugin...${NC}"

# Create plugins directory if it doesn't exist
PLUGINS_DIR="./plugins"
if [ ! -d "$PLUGINS_DIR" ]; then
    echo -e "${YELLOW}Creating plugins directory...${NC}"
    mkdir -p "$PLUGINS_DIR"
fi

# Build the plugin
PLUGIN_NAME="example-plugin"
echo -e "${YELLOW}Building plugin: $PLUGIN_NAME${NC}"

# Build as a Go plugin
go build -buildmode=plugin -o "$PLUGINS_DIR/${PLUGIN_NAME}.so" ./examples/plugins/example_plugin.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Plugin built successfully!${NC}"
    echo -e "${GREEN}Plugin location: $PLUGINS_DIR/${PLUGIN_NAME}.so${NC}"
    
    # Show plugin info
    echo -e "${YELLOW}Plugin information:${NC}"
    echo -e "  Name: $PLUGIN_NAME"
    echo -e "  Location: $PLUGINS_DIR/${PLUGIN_NAME}.so"
    echo -e "  Size: $(du -h "$PLUGINS_DIR/${PLUGIN_NAME}.so" | cut -f1)"
else
    echo -e "${RED}Plugin build failed!${NC}"
    exit 1
fi