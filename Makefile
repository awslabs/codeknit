.PHONY: lint fmt fmt-check build test test-unit test-e2e clean setup fieldalign \
       release release-dry-run deps deps-check license license-check \
       third-party-licenses third-party-check third-party-list changelog \
       docs docs-generate docs-translate docs-build docs-preview

# Binary name and entry point
BINARY    := codeknit
CMD_PATH  := ./cmd/codeknit
BUILD_DIR := bin

# Version info (override via env or CLI: make build VERSION=1.2.3)
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS   := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

# Run golangci-lint (auto-formats first to avoid gofumpt drift)
lint: fmt
	golangci-lint run ./...

# Format code and ensure license headers
fmt: license
	golangci-lint fmt ./...

# Check formatting without writing changes
fmt-check:
	golangci-lint fmt --diff ./...

# Build the binary for the host platform (development)
build:
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

# --- Release (GoReleaser) -----------------------------------------------------

# Test the release process locally without publishing (no tag required)
release-dry-run:
	goreleaser release --snapshot --clean

# Run a full release (requires a git tag: git tag v1.0.0 && git push --tags)
release:
	goreleaser release --clean

# Run unit tests
test-unit:
	go test $(shell go list ./... | grep -v /e2e)

# Run e2e tests (builds binary automatically via TestMain)
test-e2e:
	go test -v ./e2e/...

# Run all tests (unit + e2e)
test: test-unit test-e2e

# Auto-fix struct field alignment (review changes before committing)
fieldalign:
	fieldalignment -fix ./...
	golangci-lint fmt ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)/ $(BINARY) dist/

# Add license headers to all Go files
license:
	addlicense -f .license-header.txt .

# Check that all Go files have license headers
license-check:
	addlicense -f .license-header.txt -check .

# Generate THIRD-PARTY-LICENSES file from Go module dependencies
third-party-licenses:
	go-licenses report ./... --include_tests --template=NOTICE.tpl > THIRD-PARTY-LICENSES
	@echo "Generated THIRD-PARTY-LICENSES"

# List third-party dependencies with their licenses
third-party-list:
	@go-licenses report ./... --include_tests 2>/dev/null | awk -F, '{printf "%s - %s\n", $$1, $$3}'

# Check for disallowed dependency licenses (reciprocal/restricted)
third-party-check:
	go-licenses check ./... --include_tests

# Scaffold a new version entry in CHANGELOG.md (usage: make changelog V=0.2.0)
changelog:
	@if [ -z "$(V)" ]; then echo "Usage: make changelog V=0.2.0"; exit 1; fi
	@DATE=$$(date +%Y-%m-%d); \
	sed -i '' "s/## \[Unreleased\]/## [Unreleased]\n\n## [$(V)] - $$DATE\n\n### Added\n\n### Changed\n\n### Fixed\n/" CHANGELOG.md
	@echo "Scaffolded [$(V)] in CHANGELOG.md — fill in the details before committing"

# --- Dependency management -----------------------------------------------------

# Install build dependencies (Go tooling + cross-compilers for local releases)
deps:
	@echo "==> Installing Go tooling..."
	@command -v fieldalignment >/dev/null 2>&1 || { \
		echo "    Installing fieldalignment..."; \
		go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; \
	}
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "    Installing golangci-lint..."; \
		brew install golangci-lint 2>/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	@command -v addlicense >/dev/null 2>&1 || { \
		echo "    Installing addlicense..."; \
		go install github.com/google/addlicense@latest; \
	}
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "    Installing goreleaser..."; \
		brew install goreleaser 2>/dev/null || go install github.com/goreleaser/goreleaser/v2@latest; \
	}
	@command -v go-licenses >/dev/null 2>&1 || { \
		echo "    Installing go-licenses..."; \
		go install github.com/google/go-licenses/v2@latest; \
	}
	@echo "==> Installing cross-compilers (needed for local multi-platform releases)..."
ifeq ($(shell uname -s),Darwin)
	@echo "    [macOS] Using Homebrew"
	brew install FiloSottile/musl-cross/musl-cross 2>/dev/null || true
	brew install messense/macos-cross-toolchains/aarch64-unknown-linux-musl 2>/dev/null || true
	brew install mingw-w64 2>/dev/null || true
else ifeq ($(shell uname -s),Linux)
	@echo "    [Linux] Using apt (requires sudo)"
	sudo apt-get update -qq
	sudo apt-get install -y -qq gcc-aarch64-linux-gnu gcc-mingw-w64-x86-64
endif
	@echo "==> Done. Run 'make deps-check' to verify."

# Verify dependency availability
deps-check:
	@echo "Checking build dependencies..."
	@printf "  %-35s" "go" && (command -v go >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "golangci-lint" && (command -v golangci-lint >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "goreleaser" && (command -v goreleaser >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "go-licenses" && (command -v go-licenses >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "x86_64-linux-musl-gcc (linux-amd64)" && (command -v x86_64-linux-musl-gcc >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "aarch64-linux-musl-gcc (linux-arm64)" && (command -v aarch64-linux-musl-gcc >/dev/null 2>&1 && echo "OK" || echo "MISSING")
	@printf "  %-35s" "x86_64-w64-mingw32-gcc (windows)" && (command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1 && echo "OK" || echo "MISSING")

# Set up git hooks and tooling
setup: deps
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks/"

# --- Documentation (Astro Starlight + Amazon Bedrock) ---------------------------

# Bedrock model (override via env: make docs MODEL_ID=deepseek.v3.2)
MODEL_ID       ?= qwen.qwen3-235b-a22b-2507-v1:0
DOCS_FORCE     ?=
DOCS_BT        ?=

# Generate English docs from prompts using Bedrock
docs-generate: build
	@echo "==> Generating English documentation with $(MODEL_ID)..."
	go run ./docs/gen/... \
		--model $(MODEL_ID) \
		--skip-translate \
		$(if $(DOCS_FORCE),--force)

# Translate existing English docs into target languages
docs-translate:
	@echo "==> Translating documentation with $(MODEL_ID)..."
	go run ./docs/gen/... \
		--model $(MODEL_ID) \
		--skip-generate \
		$(if $(DOCS_FORCE),--force) \
		$(if $(DOCS_BT),--backtranslate)

# Build the Astro static site
docs-build:
	cd docs && npm run build

# Full pipeline: generate → translate → build (sequential — models must not overlap)
docs:
	$(MAKE) docs-generate
	$(MAKE) docs-translate
	$(MAKE) docs-build
	@echo "==> Documentation site ready at docs/dist/"

# Start Astro dev server for previewing docs (run manually)
docs-preview:
	cd docs && npm run dev
