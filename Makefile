# ---- Docker Hub / multi-arch build ----
# Default to GitHub Container Registry for this repo:
#   https://github.com/Sharkboy-j/shop
REGISTRY ?= ghcr.io
DOCKERHUB_USER ?= sharkboy-j
TAG ?= latest
VERSION ?=1.0
PLATFORMS ?= linux/amd64,linux/arm64

API_IMAGE ?= $(REGISTRY)/$(DOCKERHUB_USER)/cws
BUILDER ?= cws-builder

# golangci-lint version used by CI and local installs
GOLANGCI_LINT_VERSION ?= v2.7.2

API_TAGS := -t $(API_IMAGE):latest
ifneq ($(VERSION),)
  API_TAGS := $(API_TAGS) -t $(API_IMAGE):$(VERSION)
endif

.PHONY: lint lint-install

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with: make lint-install" && exit 1)
	golangci-lint fmt
	golangci-lint run

lint-install-mac:
	@echo "Installing golangci-lint..."
	brew install golangci-lint
	brew upgrade golangci-lint	
	@echo "golangci-lint installed successfully"

# Cross-platform installer that installs the fixed version defined in GOLANGCI_LINT_VERSION.
# On macOS prefers brew, otherwise uses the official install script and places the binary to /usr/local/bin.
lint-install:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@if command -v brew >/dev/null 2>&1; then \
	  brew install golangci-lint || brew upgrade golangci-lint; \
	else \
	  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin $(GOLANGCI_LINT_VERSION); \
	fi
	@echo "golangci-lint installation finished"
# Build and push multi-arch image
.PHONY: docker-build docker-push docker-build-push

docker-build:
	@echo "Building multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) for $(PLATFORMS)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push=false --load=false .

docker-push: lint
	@echo "Pushing multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) to $(REGISTRY)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push .

docker-build-push: lint
	@echo "Building and pushing multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) for $(PLATFORMS)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push .


.PHONY: pull
pull:
	docker compose pull
	docker compose up -d