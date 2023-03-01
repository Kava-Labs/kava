RSYNC_BIN ?= rsync

#
# Versioning for google protobuf dependencies (any, http, etc) and
#   outside (non go.mod) dependencies that we download and vendor
#
GOOGLE_APIS_PROTO_VERSION ?= f10c285cfa79997e018ea62e6f165286b1f04376
GOOGLE_APIS_DOWNLOAD_URL = https://raw.githubusercontent.com/googleapis/googleapis/$(GOOGLE_APIS_PROTO_VERSION)/google/api

PROTOBUF_ANY_PROTO_VERSION ?= v21.9
PROTOBUF_ANY_DOWNLOAD_URL = https://raw.githubusercontent.com/protocolbuffers/protobuf/$(PROTOBUF_ANY_PROTO_VERSION)/src/google/protobuf

#
# Proto dependencies under go.mod
#
GOGO_PATH := $(shell go list -m -f '{{.Dir}}' github.com/gogo/protobuf)
TENDERMINT_PATH := $(shell go list -m -f '{{.Dir}}' github.com/tendermint/tendermint)
COSMOS_PROTO_PATH := $(shell go list -m -f '{{.Dir}}' github.com/cosmos/cosmos-proto)
COSMOS_SDK_PATH := $(shell go list -m -f '{{.Dir}}' github.com/cosmos/cosmos-sdk)
IBC_GO_PATH := $(shell go list -m -f '{{.Dir}}' github.com/cosmos/ibc-go/v6)
ETHERMINT_PATH := $(shell go list -m -f '{{.Dir}}' github.com/evmos/ethermint)

#
# Common target directories
#
GOOGLE_PROTO_TYPES = third_party/proto/google/api
PROTOBUF_GOOGLE_TYPES = third_party/proto/google/protobuf
COSMOS_PROTO_TYPES = third_party/proto/cosmos_proto

.PHONY: check-rsync
check-rsync: ## Fails if rsync does not exist
	@which $(RSYNC_BIN) > /dev/null || (echo "\`$(RSYNC_BIN)\` not found. Please install $(RSYNC_BIN) or ensure it is in PATH."; exit 1)

.PHONY: proto-update-deps
proto-update-deps: check-rsync ## Update all third party proto files
	@echo "Syncing proto file dependencies"
	@mkdir -p $(GOOGLE_PROTO_TYPES)
	@curl -sSL $(GOOGLE_APIS_DOWNLOAD_URL)/annotations.proto > $(GOOGLE_PROTO_TYPES)/annotations.proto
	@curl -sSL $(GOOGLE_APIS_DOWNLOAD_URL)/http.proto > $(GOOGLE_PROTO_TYPES)/http.proto
	@curl -sSL $(GOOGLE_APIS_DOWNLOAD_URL)/httpbody.proto > $(GOOGLE_PROTO_TYPES)/httpbody.proto

	@mkdir -p $(PROTOBUF_GOOGLE_TYPES)
	@curl -sSL $(PROTOBUF_ANY_DOWNLOAD_URL)/any.proto > $(PROTOBUF_GOOGLE_TYPES)/any.proto

	@mkdir -p client/docs
	@cp $(COSMOS_SDK_PATH)/client/docs/swagger-ui/swagger.yaml client/docs/cosmos-swagger.yml
	@cp $(IBC_GO_PATH)/docs/client/swagger-ui/swagger.yaml client/docs/ibc-go-swagger.yml

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@cp $(COSMOS_PROTO_PATH)/proto/cosmos_proto/cosmos.proto $(COSMOS_PROTO_TYPES)/cosmos.proto

	@$(RSYNC_BIN) -r --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --include "*.proto" --include='*/' --exclude='*' $(GOGO_PATH)/gogoproto third_party/proto
	@$(RSYNC_BIN) -r --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --include "*.proto" --include='*/' --exclude='*' $(TENDERMINT_PATH)/proto third_party
	@$(RSYNC_BIN) -r --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --include "*.proto" --include='*/' --exclude='*' $(COSMOS_SDK_PATH)/proto third_party
	@$(RSYNC_BIN) -r --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --include "*.proto" --include='*/' --exclude='*' $(IBC_GO_PATH)/proto third_party
	@$(RSYNC_BIN) -r --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --include "*.proto" --include='*/' --exclude='*' $(ETHERMINT_PATH)/proto third_party
	@cp -f $(IBC_GO_PATH)/third_party/proto/proofs.proto third_party/proto/proofs.proto

.PHONY: check-proto-deps
check-proto-deps: proto-update-deps ## Return error code 1 if proto dependencies are not changed
	@git diff --exit-code third_party > /dev/null || (echo "Protobuf dependencies are not up to date! Please run \`make proto-update-deps\`."; exit 1)
