SHELL := /bin/bash

# =============================================================================
# 🎯 Project Configuration
# =============================================================================
PROJECT_NAME := flexphish
BIN_DIR := $(shell pwd)/bin
BUILD_DIR := $(shell pwd)/build
CMD_MAIN_PATH := ./cmd/flexphish/main.go

UI_DIR := ui
CONFIG_DIR := configs
TEMPLATE_DIR := templates
FRONTEND_DIR := app

GO ?= go
CGO_ENABLED ?= 1

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_BY := $(shell whoami)

# Detect host OS/ARCH automaticamente
HOST_OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH := $(shell uname -m)

# Normaliza nomes
ifeq ($(HOST_ARCH),x86_64)
    HOST_ARCH := amd64
endif
ifeq ($(HOST_ARCH),aarch64)
    HOST_ARCH := arm64
endif
ifeq ($(HOST_ARCH),armv7l)
    HOST_ARCH := arm
    HOST_SUBARCH := 7
endif

# Normaliza variantes de Windows no uname
ifneq (,$(findstring mingw,$(HOST_OS)))
    HOST_OS := windows
endif
ifneq (,$(findstring msys,$(HOST_OS)))
    HOST_OS := windows
endif
ifneq (,$(findstring cygwin,$(HOST_OS)))
    HOST_OS := windows
endif

# Default target platform is host unless overridden
GOOS ?= $(HOST_OS)
GOARCH ?= $(HOST_ARCH)
PLATFORMS ?= $(GOOS)/$(GOARCH)$(if $(GOARM),/$(GOARM))

# Colors
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
BOLD := \033[1m
RESET := \033[0m

INFO := bash -c 'echo -e "$(BLUE)ℹ️  $(RESET)$$0"'
SUCCESS := bash -c 'echo -e "$(GREEN)✅ $(RESET)$$0"'
WARN := bash -c 'echo -e "$(YELLOW)⚠️  $(RESET)$$0"'
ERROR := bash -c 'echo -e "$(RED)❌ $(RESET)$$0"'
WORKING := bash -c 'echo -e "$(CYAN)🔨 $(RESET)$$0"'
DEBUG := bash -c 'echo -e "$(MAGENTA)🔍 $(RESET)$$0"'
ROCKET := bash -c 'echo -e "$(GREEN)🚀 $(RESET)$$0"'
PACKAGE := bash -c 'echo -e "$(CYAN)📦 $(RESET)$$0"'
TRASH := bash -c 'echo -e "$(YELLOW)🗑️ $(RESET)$$0"'

LD_FLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.Commit=$(GIT_COMMIT)' \
	-X 'main.Branch=$(GIT_BRANCH)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.BuildBy=$(BUILD_BY)'

# =============================================================================
# 🔨 Build Targets
# =============================================================================
.PHONY: all build clean zip frontend help version

all: clean frontend build zip


init:
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(BUILD_DIR)

frontend: ## Build Angular frontend and copy to ui/
	@echo "[+] Building Angular..."
	cd app && pnpm install && pnpm build --configuration production && cd ..
	@BUILD_SUBDIR=$$(find app/dist -maxdepth 1 -mindepth 1 -type d | head -n 1) && \
	if [ ! -d "$$BUILD_SUBDIR" ]; then \
		$(ERROR) "Build output not found: $$BUILD_SUBDIR" && exit 1; \
	fi && \
	cp -r "$$BUILD_SUBDIR"/browser/* ui/ && \
	$(SUCCESS) "Frontend built and copied to ui/"

build: init ## Build Go binary
	@echo "[+] Building Go binary..."
	@rm -rf "$(BIN_DIR)/$(PROJECT_NAME)_*" 
	@for platform in $(PLATFORMS); do \
		OS=$${platform%%/*}; \
		ARCHVAR=$${platform#*/}; \
		ARCH=$${ARCHVAR%%/*}; \
		SUBARCH=$${ARCHVAR#*/}; \
		OUT_NAME="$(PROJECT_NAME)_$${OS}_$${ARCH}"; \
		if [ "$$SUBARCH" != "$$ARCH" ] && [ "$$SUBARCH" != "-" ]; then \
			OUT_NAME="$${OUT_NAME}_$${SUBARCH}"; \
		fi; \
		if [ "$$OS" = "windows" ]; then \
			BIN_FILE="$(PROJECT_NAME).exe"; \
		else \
			BIN_FILE="$(PROJECT_NAME)"; \
		fi; \
		OUT_PATH="$(BIN_DIR)/$$OUT_NAME"; \
		echo "[*] Building for $${OS}/$${ARCH}/$${SUBARCH}..."; \
		mkdir -p "$$OUT_PATH"; \
		if [ "$$ARCH" = "arm" ] && [ "$$SUBARCH" != "" ]; then \
			GOOS=$$OS GOARCH=$$ARCH GOARM=$$SUBARCH CGO_ENABLED=$(CGO_ENABLED) \
			$(GO) build -ldflags="$(LD_FLAGS)" -o "$$OUT_PATH/$$BIN_FILE" $(CMD_MAIN_PATH); \
		else \
			GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=$(CGO_ENABLED) \
			$(GO) build -ldflags="$(LD_FLAGS)" -o "$$OUT_PATH/$$BIN_FILE" $(CMD_MAIN_PATH); \
		fi; \
	done
	@echo "[✅] Binary builds completed."

zip: build frontend ## Package everything into zips for existing platforms
	@echo "[📦] Creating zip archives for available platforms..."
	@for platform in $(PLATFORMS); do \
		OS=$${platform%%/*}; \
		ARCHVAR=$${platform#*/}; \
		ARCH=$${ARCHVAR%%/*}; \
		SUBARCH=$${ARCHVAR#*/}; \
		OUT_NAME="$(PROJECT_NAME)_$${OS}_$${ARCH}"; \
		if [ "$$SUBARCH" != "" ] && [ "$$SUBARCH" != "$$ARCH" ] && [ "$$SUBARCH" != "-" ]; then \
			OUT_NAME="$${OUT_NAME}_$${SUBARCH}"; \
		fi; \
		if [ "$$OS" = "windows" ]; then \
			BIN_FILE="$(PROJECT_NAME).exe"; \
		else \
			BIN_FILE="$(PROJECT_NAME)"; \
		fi; \
		BIN_PATH="$(BIN_DIR)/$$OUT_NAME/$$BIN_FILE"; \
		if [ ! -f "$$BIN_PATH" ]; then \
			$(WARN) "Skipping zip for $$OUT_NAME: binary not found at $$BIN_PATH"; \
			continue; \
		fi; \
		ZIP_NAME="$(PROJECT_NAME)_$(VERSION)_$${OS}_$${ARCH}"; \
		if [ "$$SUBARCH" != "" ] && [ "$$SUBARCH" != "$$ARCH" ] && [ "$$SUBARCH" != "-" ]; then \
			ZIP_NAME="$${ZIP_NAME}_$${SUBARCH}"; \
		fi; \
		ZIP_PATH="$(BUILD_DIR)/$$ZIP_NAME.zip"; \
		mkdir -p $(BUILD_DIR)/temp_zip/; \
		cp "$$BIN_PATH" $(BUILD_DIR)/temp_zip/; \
		cp -r $(UI_DIR) $(BUILD_DIR)/temp_zip/ui; \
		cp -r $(CONFIG_DIR) $(BUILD_DIR)/temp_zip/configs; \
		cp -r $(TEMPLATE_DIR) $(BUILD_DIR)/temp_zip/templates; \
		(cd $(BUILD_DIR)/temp_zip && zip -r "$$ZIP_PATH" . > /dev/null); \
		sha256sum "$$ZIP_PATH" | awk '{print $$1}' > "$(BUILD_DIR)/$$ZIP_NAME.sha256"; \
		rm -rf $(BUILD_DIR)/temp_zip; \
		$(SUCCESS) "[✅] Created zip: $$ZIP_PATH"; \
		$(SUCCESS) "[🔐] SHA256 saved to: $(BUILD_DIR)/$$ZIP_NAME.sha256"; \
	done

clean: ## Clean build artifacts
	@echo "[🗑️] Cleaning build and bin folders..."
	@rm -rf $(BIN_DIR)/* $(BUILD_DIR)/*
	@rm -rf $(UI_DIR)/*
	@echo "[✅] Clean complete"

help: ## Show help
	@echo "\nAvailable targets:\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

version: ## Show version info
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Branch:     $(GIT_BRANCH)"
	@echo "Built:      $(BUILD_TIME)"
	@echo "Built by:   $(BUILD_BY)"
	@echo "Go version: $(shell $(GO) version)"

.DEFAULT_GOAL := help
