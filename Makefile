# Constants.
VERSION_FILE=version.txt
BINARY_NAME=thunder

BACKEND_BASE_DIR := backend
REPOSITORY_DIR=$(BACKEND_BASE_DIR)/cmd/server/repository

OUTPUT_DIR=target
BUILD_DIR=$(OUTPUT_DIR)/.build

FRONTEND_DIR := frontend/loginportal
BACKEND_DIR := $(BACKEND_BASE_DIR)/cmd/server

SERVER_SCRIPTS_DIR=$(BACKEND_BASE_DIR)/scripts
SERVER_DB_SCRIPTS_DIR=$(BACKEND_BASE_DIR)/dbscripts



# Variable constants.
VERSION=$(shell cat $(VERSION_FILE))
# ZIP_FILE_NAME=${BINARY_NAME_PREFIX}-$(VERSION)
PRODUCT_FOLDER=$(BINARY_NAME)-$(VERSION)

# Default target.
all: clean build test

# Clean up build artifacts.
clean:
	rm -rf $(OUTPUT_DIR)

# Build project and package it.
build: _build _build-frontend _package

# Build the Go project.
_build:
	mkdir -p $(BUILD_DIR) && \
	go build -C $(BACKEND_BASE_DIR) -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

_build-frontend:
	@echo "Building frontend..."
	npm install --prefix $(FRONTEND_DIR)
	npm run build --prefix $(FRONTEND_DIR)

# Package the binary and repository directory into a zip file.
_package:
	mkdir -p $(OUTPUT_DIR)/$(PRODUCT_FOLDER) && \
	cp $(BUILD_DIR)/$(BINARY_NAME) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r $(REPOSITORY_DIR) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp $(VERSION_FILE) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r $(SERVER_SCRIPTS_DIR) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r $(SERVER_DB_SCRIPTS_DIR) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r $(FRONTEND_DIR)/build $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/dist/ && \
	cd $(OUTPUT_DIR) && zip -r $(PRODUCT_FOLDER).zip $(PRODUCT_FOLDER) && \
	rm -rf $(PRODUCT_FOLDER) && \
	rm -rf $(BUILD_DIR)

# Run all tests.
test: _integration-test

# Run integration tests.
_integration-test:
	@echo "Running integration tests..."
	@go run -C ./tests/integration ./main.go

run: _build-frontend
	@echo "Removing old build artifacts..."
	@rm -rf $(BACKEND_DIR)/dist
	@echo "Copying frontend build to backend static directory..."
	@mkdir -p $(BACKEND_DIR)/dist
	@cp -r $(FRONTEND_DIR)/build/* $(BACKEND_DIR)/dist/
	@echo "Running the application..."
	@go run -C $(BACKEND_DIR) .

help:
	@echo "Makefile targets:"
	@echo "  all          - Clean, build, and test the project."
	@echo "  clean        - Remove build artifacts."
	@echo "  build        - Build the Go project."
	@echo "  test         - Run all tests."
	@echo "  help         - Show the help message."

.PHONY: all clean build _build _package test _integration-test help
