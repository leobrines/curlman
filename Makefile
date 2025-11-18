.PHONY: all build install clean test run help deps

# Binary name
BINARY_NAME=curlman

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package path
MAIN_PATH=.

# Install location (user's local bin directory)
INSTALL_PATH=$(HOME)/bin

## deps: Download and install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## build: Build the application
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the application to user's local bin directory
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete! Run '$(BINARY_NAME)' to start."
	@echo "Note: Make sure $(INSTALL_PATH) is in your PATH environment variable."

## uninstall: Remove the application from user's local bin directory
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled successfully."
