BINARY_NAME := sortir

VERSION ?= $(shell git describe --tags --match 'v*' --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/leorolland/sortir.in/pkg/version.Version=$(VERSION)

GO := $(shell which go 2>/dev/null || echo "go")
MODD := $(shell which modd 2>/dev/null || echo "modd")
NPX := $(shell which npx 2>/dev/null || echo "npx")
NPM := $(shell which npm 2>/dev/null || echo "npm")

.PHONY: install dev dev-ui build release clean generate

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
	cd ui && VITE_APP_VERSION=$(VERSION) $(NPM) run build

# Build the binary, UI must be built first as it is embedded in the binary
build: ui/build
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) cmd/main.go

# Build the UI and the release binaries for the supported architectures.
# CI: semantic-release calls this with the computed version (make release VERSION=x.y.z)
release:
	rm -rf ui/build release
	cd ui && VITE_APP_VERSION=$(VERSION) $(NPM) run build
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o release/sortir.linux-amd64 cmd/main.go
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o release/populate.linux-amd64 cmd/populate/populate.go

clean:
	rm -f $(BINARY_NAME)
	rm -rf ui/build pb_data release

generate:
	$(GO) generate ./...
