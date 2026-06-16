BINARY := gentr
BUILD_DIR := build

PREFIX ?= $(HOME)/.local
BIN_DIR ?= $(PREFIX)/bin

GO ?= go

.PHONY: all build test test-race vet fmt fmt-check check install uninstall clean run

all: check build

build:
	@mkdir -p "$(BUILD_DIR)"
	$(GO) build -trimpath -o "$(BUILD_DIR)/$(BINARY)" .

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { \
		echo "[x] These files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	}

check: fmt-check vet test

run:
	$(GO) run .

install: build
	@mkdir -p "$(BIN_DIR)"
	@tmp="$(BIN_DIR)/.$(BINARY).tmp"; \
	cp "$(BUILD_DIR)/$(BINARY)" "$$tmp"; \
	chmod 0755 "$$tmp"; \
	mv -f "$$tmp" "$(BIN_DIR)/$(BINARY)"
	@echo "[v] Installed $(BINARY) to $(BIN_DIR)/$(BINARY)"
	@if ! echo ":$$PATH:" | grep -q ":$(BIN_DIR):"; then \
		echo "[!] $(BIN_DIR) is not in PATH"; \
	fi

uninstall:
	@rm -f "$(BIN_DIR)/$(BINARY)"
	@echo "[v] Removed $(BIN_DIR)/$(BINARY)"

clean:
	rm -rf "$(BUILD_DIR)"
	rm -f coverage.out coverage.html

