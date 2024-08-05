################################################################################
###                             Required Variables                           ###
################################################################################
ifndef DOCKER
$(error DOCKER not set)
endif

ifndef BUILD_DIR
$(error BUILD_DIR not set)
endif

################################################################################
###                             Lint Settings                                ###
################################################################################

LINT_FROM_REV ?= $(shell git merge-base origin/master HEAD)

GOLANGCI_VERSION ?= $(shell cat .golangci-version)
GOLANGCI_IMAGE_TAG ?= golangci/golangci-lint:$(GOLANGCI_VERSION)

GOLANGCI_DIR ?= $(CURDIR)/$(BUILD_DIR)/.golangci-lint

GOLANGCI_CACHE_DIR ?= $(GOLANGCI_DIR)/$(GOLANGCI_VERSION)-cache
GOLANGCI_MOD_CACHE_DIR ?= $(GOLANGCI_DIR)/go-mod

################################################################################
###                             Lint Target                                  ###
################################################################################

.PHONY: lint
lint: $(GOLANGCI_CACHE_DIR) $(GOLANGCI_MOD_CACHE_DIR)
	@echo "Running lint from rev $(LINT_FROM_REV), use LINT_FROM_REV var to override."
	$(DOCKER) run -t --rm \
		-v $(GOLANGCI_CACHE_DIR):/root/.cache \
		-v $(GOLANGCI_MOD_CACHE_DIR):/go/pkg/mod \
		-v $(CURDIR):/app \
		-w /app \
		$(GOLANGCI_IMAGE_TAG) \
		golangci-lint run -v --new-from-rev $(LINT_FROM_REV)

$(GOLANGCI_CACHE_DIR):
	@mkdir -p $@

$(GOLANGCI_MOD_CACHE_DIR):
	@mkdir -p $@
