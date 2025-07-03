# System Monitor Plugin Makefile

PLUGIN_NAME = system-monitor
PLUGIN_VERSION = 1.0.0
BUILD_DIR = build
PLUGIN_DIR = ~/.nixai/plugins

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
	@cp README.md $(PLUGIN_DIR)/$(PLUGIN_NAME)/
	@cp plugin.yaml $(PLUGIN_DIR)/$(PLUGIN_NAME)/
	@echo "Plugin installed to $(PLUGIN_DIR)/$(PLUGIN_NAME)"

uninstall:
	@echo "Uninstalling $(PLUGIN_NAME) plugin..."
	@rm -rf $(PLUGIN_DIR)/$(PLUGIN_NAME)
	@echo "Plugin uninstalled"

test:
	@echo "Running tests for $(PLUGIN_NAME) plugin..."
	go test -v ./...
	@echo "Tests complete"

dev-install: build
	@echo "Installing plugin for development..."
	@mkdir -p $(PLUGIN_DIR)/$(PLUGIN_NAME)
	@cp $(BUILD_DIR)/$(PLUGIN_NAME).so $(PLUGIN_DIR)/$(PLUGIN_NAME)/
	@echo "Development installation complete"

package: build
	@echo "Creating plugin package..."
	@mkdir -p $(BUILD_DIR)/package/$(PLUGIN_NAME)
	@cp $(BUILD_DIR)/$(PLUGIN_NAME).so $(BUILD_DIR)/package/$(PLUGIN_NAME)/
	@cp README.md $(BUILD_DIR)/package/$(PLUGIN_NAME)/
	@cp plugin.yaml $(BUILD_DIR)/package/$(PLUGIN_NAME)/
	@cd $(BUILD_DIR)/package && tar -czf $(PLUGIN_NAME)-$(PLUGIN_VERSION).tar.gz $(PLUGIN_NAME)/
	@echo "Package created: $(BUILD_DIR)/package/$(PLUGIN_NAME)-$(PLUGIN_VERSION).tar.gz"

help:
	@echo "Available targets:"
	@echo "  build      - Build the plugin"
	@echo "  clean      - Clean build directory"
	@echo "  install    - Install plugin to user directory"
	@echo "  uninstall  - Remove plugin from user directory"
	@echo "  test       - Run plugin tests"
	@echo "  package    - Create distributable package"
	@echo "  help       - Show this help message"