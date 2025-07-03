# Simple System Info Plugin Makefile

PLUGIN_NAME = simple-system-info
PLUGIN_VERSION = 1.0.0
BUILD_DIR = build
PLUGIN_DIR = ~/.config/nixai/plugins

.PHONY: all build clean install uninstall test

all: build

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@mkdir -p $(BUILD_DIR)
	go build -buildmode=plugin -o $(BUILD_DIR)/$(PLUGIN_NAME).so main.go
	@echo "Plugin built successfully: $(BUILD_DIR)/$(PLUGIN_NAME).so"

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

install: build
	@echo "Installing $(PLUGIN_NAME) plugin..."
	@mkdir -p $(PLUGIN_DIR)/$(PLUGIN_NAME)
	@cp $(BUILD_DIR)/$(PLUGIN_NAME).so $(PLUGIN_DIR)/$(PLUGIN_NAME)/
	@echo "Plugin installed to $(PLUGIN_DIR)/$(PLUGIN_NAME)"

uninstall:
	@echo "Uninstalling $(PLUGIN_NAME) plugin..."
	@rm -rf $(PLUGIN_DIR)/$(PLUGIN_NAME)
	@echo "Plugin uninstalled"

test:
	@echo "Testing $(PLUGIN_NAME) plugin..."
	go run main.go get-info
	@echo "Test complete"

help:
	@echo "Available targets:"
	@echo "  build      - Build the plugin"
	@echo "  clean      - Clean build directory"
	@echo "  install    - Install plugin to user directory"
	@echo "  uninstall  - Remove plugin from user directory"
	@echo "  test       - Test plugin standalone"
	@echo "  help       - Show this help message"