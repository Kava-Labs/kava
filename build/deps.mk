## Derived from work and ideas in https://github.com/ovrclk/akash/blob/8a29be1c3404843c80f0c861b2e577067607474b/make/setup-cache.mk

################################################################################
###                             Required Variables                           ###
################################################################################
ifndef BUILD_CACHE_DIR
$(error BUILD_CACHE_DIR not set)
endif

ifndef BIN_DIR
$(error BIN_DIR not set)
endif

ifndef OS_FAMILY
$(error OS_FAMILY not set)
endif

ifndef MACHINE
$(error MACHINE not set)
endif

################################################################################
###                             Dir Setup                                    ###
################################################################################
DIRS := $(BUILD_CACHE_DIR) $(BIN_DIR)

################################################################################
###                             Tool Versions                                ###
################################################################################
PROTOC_VERSION ?= v21.9
BUF_VERSION ?= v1.9.0
PROTOC_GEN_GOCOSMOS_VERSION ?= v0.3.1
PROTOC_GEN_GRPC_GATEWAY_VERSION ?= $(shell go list -m github.com/grpc-ecosystem/grpc-gateway| sed 's:.* ::')
PROTOC_GEN_DOC_VERSION ?= v1.5.1
SWAGGER_COMBINE_VERSION ?= v1.4.0

################################################################################
###                             Protoc Install                               ###
################################################################################
PROTOC_VERSION_FILE := $(BUILD_CACHE_DIR)/protoc-$(PROTOC_VERSION).version

ifeq ($(OS_FAMILY),Linux)
PROTOC_PLATFORM := linux
endif
ifeq ($(OS_FAMILY),Darwin)
PROTOC_PLATFORM := osx
endif
PROTOC_MACHINE := $(MACHINE)
ifeq ($(MACHINE),amd64)
PROTOC_MACHINE := x86_64
endif
ifeq ($(MACHINE),aarch64)
PROTOC_MACHINE := aarch_64
endif
ifeq ($(MACHINE),arm64)
PROTOC_MACHINE := aarch_64
endif

PROTOC_ARCHIVE_NAME := protoc-$(shell echo $(PROTOC_VERSION) | sed s/^v//)-$(PROTOC_PLATFORM)-$(PROTOC_MACHINE).zip
PROTOC_DOWNLOAD_URL := https://github.com/protocolbuffers/protobuf/releases/download/$(PROTOC_VERSION)/$(PROTOC_ARCHIVE_NAME)

$(PROTOC_VERSION_FILE):
	@echo "installing protoc..."
	@mkdir -p $(DIRS)
	@touch $(PROTOC_VERSION_FILE)
	@cd $(BUILD_CACHE_DIR); \
	mkdir -p protoc && cd protoc; \
	curl -sOL $(PROTOC_DOWNLOAD_URL); \
	unzip -q $(PROTOC_ARCHIVE_NAME) bin/protoc
	@cp $(BUILD_CACHE_DIR)/protoc/bin/protoc $(BIN_DIR)/protoc
	@rm -rf $(BUILD_CACHE_DIR)/protoc

PROTOC := $(BIN_DIR)/protoc
BUILD_DEPS = $(PROTOC)

$(PROTOC): $(PROTOC_VERSION_FILE)
	@echo "installed $(shell $(PROTOC) --version)"

################################################################################
###                             Buf Install                                  ###
################################################################################
BUF_VERSION_FILE := $(BUILD_CACHE_DIR)/buf-$(BUF_VERSION).version

BUF_ARCHIVE_NAME := buf-$(OS_FAMILY)-$(MACHINE).tar.gz
BUF_DOWNLOAD_URL := https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/$(BUF_ARCHIVE_NAME)

$(BUF_VERSION_FILE):
	@echo "installing buf..."
	@mkdir -p $(DIRS)
	@touch $(BUF_VERSION_FILE)
	@cd $(BUILD_CACHE_DIR); \
	mkdir -p buf && cd buf; \
	curl -sOL $(BUF_DOWNLOAD_URL); \
	tar -xzf $(BUF_ARCHIVE_NAME) buf/bin/buf
	@cp $(BUILD_CACHE_DIR)/buf/buf/bin/buf $(BIN_DIR)/buf
	@rm -rf $(BUILD_CACHE_DIR)/buf

BUF := $(BIN_DIR)/buf
BUILD_DEPS += $(BUF)

$(BUF): $(BUF_VERSION_FILE)
	@echo "installed buf $(shell $(BUF) --version)"

################################################################################
###                             gocomos proto plugin                         ###
################################################################################
PROTOC_GEN_GOCOSMOS_VERSION_FILE := $(BUILD_CACHE_DIR)/protoc-gen-gocosmos-$(PROTOC_GEN_GOCOSMOS_VERSION).version

$(PROTOC_GEN_GOCOSMOS_VERSION_FILE):
	@echo "installing protoc-gen-gocosmos..."
	@mkdir -p $(DIRS)
	@touch $(PROTOC_GEN_GOCOSMOS_VERSION_FILE)
	@cd $(BUILD_CACHE_DIR); \
	mkdir -p protoc-gen-gocosmos && cd protoc-gen-gocosmos; \
	git clone -q https://github.com/regen-network/cosmos-proto.git; \
	cd cosmos-proto; \
	git checkout -q $(PROTOC_GEN_GOCOSMOS_VERSION); \
	GOBIN=$(ROOT_DIR)/$(BIN_DIR) go install ./protoc-gen-gocosmos
	@rm -rf $(BUILD_CACHE_DIR)/protoc-gen-gocosmos

PROTOC_GEN_GOCOSMOS := $(BIN_DIR)/protoc-gen-gocosmos
BUILD_DEPS += $(PROTOC_GEN_GOCOSMOS)

$(PROTOC_GEN_GOCOSMOS): $(PROTOC_GEN_GOCOSMOS_VERSION_FILE)
	@echo "installed protoc-gen-gocosmos $(PROTOC_GEN_GOCOSMOS_VERSION)"

################################################################################
###                        grpc gateway proto plugin                         ###
################################################################################
PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE := $(BUILD_CACHE_DIR)/protoc-gen-grpc-gateway-$(PROTOC_GEN_GRPC_GATEWAY_VERSION).version

$(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE):
	@echo "installing protoc-gen-grpc-gateway..."
	@mkdir -p $(DIRS)
	@touch $(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE)
	@cd $(BUILD_CACHE_DIR); \
	mkdir -p protoc-gen-grpc-gateway && cd protoc-gen-grpc-gateway; \
	git clone -q https://github.com/grpc-ecosystem/grpc-gateway.git; \
	cd grpc-gateway; \
	git checkout -q $(PROTOC_GEN_GRPC_GATEWAY_VERSION); \
	GOBIN=$(ROOT_DIR)/$(BIN_DIR) go install ./protoc-gen-grpc-gateway; \
	GOBIN=$(ROOT_DIR)/$(BIN_DIR) go install ./protoc-gen-swagger
	@rm -rf $(BUILD_CACHE_DIR)/protoc-gen-grpc-gateway

PROTOC_GEN_GRPC_GATEWAY := $(BIN_DIR)/protoc-gen-grpc-gateway
BUILD_DEPS += $(PROTOC_GEN_GRPC_GATEWAY)

$(PROTOC_GEN_GRPC_GATEWAY): $(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE)
	@echo "installed protoc-gen-grpc-gateway $(PROTOC_GEN_GRPC_GATEWAY_VERSION)"

PROTOC_GEN_SWAGGER := $(BIN_DIR)/protoc-gen-swagger
BUILD_DEPS += $(PROTOC_GEN_SWAGGER)

$(PROTOC_GEN_SWAGGER): $(PROTOC_GEN_GRPC_GATEWAY_VERSION_FILE)
	@echo "installed protoc-gen-swagger $(PROTOC_GEN_GRPC_GATEWAY_VERSION)"

################################################################################
###                        Proto Gen Doc Install                             ###
################################################################################
PROTOC_GEN_DOC_VERSION_FILE := $(BUILD_CACHE_DIR)/protoc-gen-doc-$(PROTOC_GEN_DOC_VERSION).version

ifeq ($(OS_FAMILY),Linux)
PROTOC_GEN_DOC_PLATFORM := linux
endif
ifeq ($(OS_FAMILY),Darwin)
PROTOC_GEN_DOC_PLATFORM := darwin
endif
PROTOC_GEN_DOC_MACHINE := $(MACHINE)
ifeq ($(MACHINE),x86_64)
PROTOC_GEN_DOC_MACHINE := amd64
endif
ifeq ($(MACHINE),aarch64)
PROTOC_GEN_DOC_MACHINE := arm64
endif

PROTOC_GEN_DOC_ARCHIVE_NAME := protoc-gen-doc_$(shell echo $(PROTOC_GEN_DOC_VERSION) | sed s/^v//)_$(PROTOC_GEN_DOC_PLATFORM)_$(PROTOC_GEN_DOC_MACHINE).tar.gz
PROTOC_GEN_DOC_DOWNLOAD_URL := https://github.com/pseudomuto/protoc-gen-doc/releases/download/$(PROTOC_GEN_DOC_VERSION)/$(PROTOC_GEN_DOC_ARCHIVE_NAME)

$(PROTOC_GEN_DOC_VERSION_FILE):
	@echo "installing protoc-gen-doc..."
	@mkdir -p $(DIRS)
	@touch $(PROTOC_GEN_DOC_VERSION_FILE)
	@cd $(BUILD_CACHE_DIR); \
	mkdir -p protoc-gen-doc && cd protoc-gen-doc; \
	curl -sOL $(PROTOC_GEN_DOC_DOWNLOAD_URL); \
	tar -xzf $(PROTOC_GEN_DOC_ARCHIVE_NAME) protoc-gen-doc
	@cp $(BUILD_CACHE_DIR)/protoc-gen-doc/protoc-gen-doc $(BIN_DIR)/protoc-gen-doc
	@rm -rf $(BUILD_CACHE_DIR)/protoc-gen-doc

PROTOC_GEN_DOC := $(BIN_DIR)/protoc-gen-doc
BUILD_DEPS += $(PROTOC_GEN_DOC)

$(PROTOC_GEN_DOC): $(PROTOC_GEN_DOC_VERSION_FILE)
	@echo "installed protoc-gen-doc $(shell $(PROTOC_GEN_DOC) --version)"

################################################################################
###                        Swagger Combine                                   ###
################################################################################
SWAGGER_COMBINE_VERSION_FILE := $(BUILD_CACHE_DIR)/swagger-combine-$(SWAGGER_COMBINE_VERSION).version

$(SWAGGER_COMBINE_VERSION_FILE):
	@echo "installing swagger-combine..."
	@mkdir -p $(DIRS) $(BUILD_CACHE_DIR)/node_modules
	@touch $(SWAGGER_COMBINE_VERSION_FILE)
	@npm install --silent --no-progress --prefix $(BUILD_CACHE_DIR) swagger-combine@$(shell echo $(SWAGGER_COMBINE_VERSION) | sed s/^v//)
	@ln -sf ../.cache/node_modules/.bin/swagger-combine $(BIN_DIR)/swagger-combine

SWAGGER_COMBINE := $(BIN_DIR)/swagger-combine
BUILD_DEPS += $(SWAGGER_COMBINE)

$(SWAGGER_COMBINE): $(SWAGGER_COMBINE_VERSION_FILE)
	@echo "installed swagger-combine $(shell $(SWAGGER_COMBINE) -v)"

################################################################################
###                        Build Deps                                        ###
################################################################################
.PHONY: install-build-deps
install-build-deps: $(BUILD_DEPS)
