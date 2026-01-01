# Makefile for managing the shop application with docker-compose

# ---- Docker Hub / multi-arch build (Raspberry Pi) ----
# Default to GitHub Container Registry for this repo:
#   https://github.com/Sharkboy-j/shop
REGISTRY ?= ghcr.io
# IMPORTANT: docker image repository names must be lowercase (GHCR enforces this).
DOCKERHUB_USER ?= sharkboy-j
TAG ?= latest
VERSION ?=1.0
PLATFORMS ?= linux/amd64,linux/arm64

API_IMAGE ?= $(REGISTRY)/$(DOCKERHUB_USER)/cws
BUILDER ?= cws-builder

# Build tags: always include latest, add version tag if VERSION is set
API_TAGS := -t $(API_IMAGE):latest
ifneq ($(VERSION),)
  API_TAGS := $(API_TAGS) -t $(API_IMAGE):$(VERSION)
endif

# Build and push multi-arch image
.PHONY: docker-build docker-push docker-build-push

docker-build:
	@echo "Building multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) for $(PLATFORMS)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push=false --load=false .

docker-push:
	@echo "Pushing multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) to $(REGISTRY)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push .

docker-build-push:
	@echo "Building and pushing multi-arch image $(API_IMAGE):latest$(if $(VERSION), and $(API_IMAGE):$(VERSION),) for $(PLATFORMS)"
	docker buildx create --use --name $(BUILDER) 2>/dev/null || true
	docker buildx build --platform $(PLATFORMS) $(API_TAGS) --push .
