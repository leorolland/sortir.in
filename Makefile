BINARY_NAME := sortir

GO := $(shell which go 2>/dev/null || echo "go")
MODD := $(shell which modd 2>/dev/null || echo "modd")
NPX := $(shell which npx 2>/dev/null || echo "npx")
NPM := $(shell which npm 2>/dev/null || echo "npm")

.PHONY: install dev dev-ui build clean

install:
	$(GO) mod download
	$(GO) install github.com/cortesi/modd/cmd/modd@latest
	cd ui && $(NPX) pnpm install

dev: ui/build
	$(MODD)

dev-ui:
	cd ui && $(NPM) run dev

# Rebuild the UI only when any source file changes
UI_SRC_FILES := $(shell find ui -type f -not -path "ui/node_modules/*" -not -path "ui/build/*")
ui/build: $(UI_SRC_FILES)
	cd ui && $(NPM) run build

# Build the binary, UI must be built first as it is embedded in the binary
build: ui/build
	$(GO) build -o $(BINARY_NAME) cmd/main.go

clean:
	rm -f $(BINARY_NAME)
	rm -rf ui/build
