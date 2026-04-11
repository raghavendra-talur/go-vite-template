.PHONY: dev dev-frontend dev-backend build run clean test check install uninstall \
       service-start service-stop service-restart service-status service-logs

# Load .env if it exists (values can still be overridden on the command line)
-include .env
export

# Configurable port (override with: make dev PORT=8080)
PORT ?= 9002

# Internal: Vite dev server port (not user-facing, backend proxies to it)
VITE_DEV_PORT = 9005

# Version: set from env or git tag, falls back to commit hash
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go ldflags for embedding version
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# App name — change this when cloning the template
APP_NAME := go-vite-template

# Development: run Go backend + Vite frontend in parallel
dev:
	@echo "Starting dev server on :$(PORT)..."
	@$(MAKE) -j2 dev-backend dev-frontend PORT=$(PORT)

dev-backend:
	cd server-go && DEV=1 PORT=$(PORT) VITE_URL=http://127.0.0.1:$(VITE_DEV_PORT) air

dev-frontend:
	PORT=$(PORT) VITE_PORT=$(VITE_DEV_PORT) npx vite --port $(VITE_DEV_PORT) --strictPort

# Run all tests
test:
	cd server-go && go test ./...
	npm run check

# Run Go backend tests only
test-backend:
	cd server-go && go test ./...

# Run frontend type checks only
check:
	npm run check

# Build for production (frontend must be built first — it's embedded into the Go binary)
build: build-frontend build-backend

build-frontend:
	npm install --prefer-offline
	npx vite build
	@echo "This file exists so go:embed dist/public works on a fresh clone." > server-go/dist/public/placeholder

build-backend:
	cd server-go && go build $(LDFLAGS) -o ../dist/$(APP_NAME) .

# Run production build
run:
	./dist/$(APP_NAME) --port $(PORT)

# --- Service installation (macOS launchd / Linux systemd) ---

INSTALL_DIR ?= $(HOME)/.local/bin
LOG_DIR     ?= $(HOME)/Library/Logs
PLIST_NAME  := com.$(APP_NAME).server
PLIST_SRC   := service/$(PLIST_NAME).plist
PLIST_DST   := $(HOME)/Library/LaunchAgents/$(PLIST_NAME).plist
SYSTEMD_SRC := service/$(APP_NAME).service
SYSTEMD_DST := $(HOME)/.config/systemd/user/$(APP_NAME).service

install: build
	@mkdir -p $(INSTALL_DIR)
	cp dist/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Binary installed to $(INSTALL_DIR)/$(APP_NAME)"
ifeq ($(shell uname),Darwin)
	@mkdir -p $(HOME)/Library/LaunchAgents $(LOG_DIR)
	@sed -e 's|__INSTALL_DIR__|$(INSTALL_DIR)|g' \
	     -e 's|__PORT__|$(PORT)|g' \
	     -e 's|__LOG_DIR__|$(LOG_DIR)|g' \
	     -e 's|__APP_NAME__|$(APP_NAME)|g' \
	     $(PLIST_SRC) > $(PLIST_DST)
	@echo "launchd plist installed to $(PLIST_DST)"
	@echo ""
	@echo "Start:   make service-start"
	@echo "Stop:    make service-stop"
	@echo "Logs:    make service-logs"
	@echo "Remove:  make uninstall"
else
	@mkdir -p $(HOME)/.config/systemd/user
	@if [ -f $(SYSTEMD_SRC) ]; then \
		sed -e 's|__INSTALL_DIR__|$(INSTALL_DIR)|g' \
		    -e 's|__PORT__|$(PORT)|g' \
		    $(SYSTEMD_SRC) > $(SYSTEMD_DST); \
		systemctl --user daemon-reload; \
		echo "systemd unit installed to $(SYSTEMD_DST)"; \
	fi
	@echo ""
	@echo "Start:   make service-start"
	@echo "Stop:    make service-stop"
	@echo "Logs:    make service-logs"
	@echo "Remove:  make uninstall"
endif

uninstall: service-stop
ifeq ($(shell uname),Darwin)
	@rm -f $(PLIST_DST)
	@echo "launchd plist removed"
else
	@rm -f $(SYSTEMD_DST)
	@systemctl --user daemon-reload 2>/dev/null || true
	@echo "systemd unit removed"
endif
	@rm -f $(INSTALL_DIR)/$(APP_NAME)
	@echo "Binary removed from $(INSTALL_DIR)/$(APP_NAME)"

service-start:
ifeq ($(shell uname),Darwin)
	launchctl load $(PLIST_DST)
	@echo "$(APP_NAME) started (port $(PORT))"
else
	systemctl --user start $(APP_NAME)
endif

service-stop:
ifeq ($(shell uname),Darwin)
	-launchctl unload $(PLIST_DST) 2>/dev/null
	@echo "$(APP_NAME) stopped"
else
	-systemctl --user stop $(APP_NAME) 2>/dev/null
endif

service-restart:
ifeq ($(shell uname),Darwin)
	-launchctl unload $(PLIST_DST) 2>/dev/null
	launchctl load $(PLIST_DST)
	@echo "$(APP_NAME) restarted"
else
	systemctl --user restart $(APP_NAME)
endif

service-status:
ifeq ($(shell uname),Darwin)
	@launchctl list | grep $(APP_NAME) || echo "not running"
	@echo "Health: $$(curl -s http://127.0.0.1:$(PORT)/health)"
else
	systemctl --user status $(APP_NAME)
endif

service-logs:
ifeq ($(shell uname),Darwin)
	tail -f $(LOG_DIR)/$(APP_NAME).log
else
	journalctl --user -u $(APP_NAME) -f
endif

# Clean build artifacts
clean:
	rm -rf dist/
	find server-go/dist/public/ -not -name 'placeholder' -not -path server-go/dist/public/ -delete 2>/dev/null || true
