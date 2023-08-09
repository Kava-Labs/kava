################################################################################
###                             Project Info                                 ###
################################################################################
PROJECT_NAME := kava# unique namespace for project

GO_BIN ?= go

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT := $(shell git rev-parse --short HEAD)

BRANCH_PREFIX := $(shell echo $(GIT_BRANCH) | sed 's/\/.*//g')# eg release, master, feat

ifeq ($(BRANCH_PREFIX),release)
# we are on a release branch, set version to the last or current tag
VERSION := $(shell git describe --tags)# use current tag or most recent tag + number of commits + g + abbrivated commit
VERSION_NUMBER := $(shell echo $(VERSION) | sed 's/^v//')# drop the "v" prefix for versions
else
# we are not on a release branch, and do not have clean tag history (etc v0.19.0-xx-gxx will not make sense to use)
VERSION := $(GIT_COMMIT_SHORT)
VERSION_NUMBER := $(VERSION)
endif

TENDERMINT_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
COSMOS_SDK_VERSION := $(shell go list -m github.com/cosmos/cosmos-sdk | sed 's:.* ::')

.PHONY: print-version
print-version:
	@echo "kava $(VERSION)\ntendermint $(TENDERMINT_VERSION)\ncosmos $(COSMOS_SDK_VERSION)"

################################################################################
###                             Project Settings                             ###
################################################################################
LEDGER_ENABLED ?= true
DOCKER:=docker
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
HTTPS_GIT := https://github.com/Kava-Labs/kava.git

################################################################################
###                             Machine Info                                 ###
################################################################################
OS_FAMILY := $(shell uname -s)
MACHINE := $(shell uname -m)

NATIVE_GO_OS := $(shell echo $(OS_FAMILY) | tr '[:upper:]' '[:lower:]')# Linux -> linux, Darwin -> darwin

NATIVE_GO_ARCH := $(MACHINE)
ifeq ($(MACHINE),x86_64)
NATIVE_GO_ARCH := amd64# x86_64 -> amd64
endif
ifeq ($(MACHINE),aarch64)
NATIVE_GO_ARCH := arm64# aarch64 -> arm64
endif

TARGET_GO_OS ?= $(NATIVE_GO_OS)
TARGET_GO_ARCH ?= $(NATIVE_GO_ARCH)
.PHONY: print-machine-info
print-machine-info:
	@echo "platform $(NATIVE_GO_OS)/$(NATIVE_GO_ARCH)"
	@echo "target $(TARGET_GO_OS)/$(TARGET_GO_ARCH)"

################################################################################
###                             PATHS                                        ###
################################################################################
BUILD_DIR := build# build files
BIN_DIR := $(BUILD_DIR)/bin# for binary dev dependencies
BUILD_CACHE_DIR := $(BUILD_DIR)/.cache# caching for non-artifact outputs
OUT_DIR := out# for artifact intermediates and outputs

ROOT_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))# absolute path to root
export PATH := $(ROOT_DIR)/$(BIN_DIR):$(PATH)# add local bin first in path

.PHONY: print-path
print-path:
	@echo $(PATH)

.PHONY: print-paths
print-paths:
	@echo "build $(BUILD_DIR)\nbin $(BIN_DIR)\ncache $(BUILD_CACHE_DIR)\nout $(OUT_DIR)"

.PHONY: clean
clean:
	@rm -rf $(BIN_DIR) $(BUILD_CACHE_DIR) $(OUT_DIR)

################################################################################
###                             Dev Setup                                    ###
################################################################################
include $(BUILD_DIR)/deps.mk

include $(BUILD_DIR)/proto.mk
include $(BUILD_DIR)/proto-deps.mk

#export GO111MODULE = on
# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=kava \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=kava \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION_NUMBER) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(GIT_COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TENDERMINT_VERSION)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb
  BUILD_TAGS += badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

all: install

build: go.sum
ifeq ($(OS), Windows_NT)
	$(GO_BIN) build -mod=readonly $(BUILD_FLAGS) -o out/$(shell $(GO_BIN) env GOOS)/kava.exe ./cmd/kava
else
	$(GO_BIN) build -mod=readonly $(BUILD_FLAGS) -o out/$(shell $(GO_BIN) env GOOS)/kava ./cmd/kava
endif

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum
	$(GO_BIN) install -mod=readonly $(BUILD_FLAGS) ./cmd/kava

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensuring dependencies have not been modified"
	@go mod verify

########################################
### Linting

# Check url links in the repo are not broken.
# This tool checks local markdown links as well.
# Set to exclude riot links as they trigger false positives
link-check:
	@go get -u github.com/raviqqe/liche@f57a5d1c5be4856454cb26de155a65a4fd856ee3
	liche -r . --exclude "^http://127.*|^https://riot.im/app*|^http://kava-testnet*|^https://testnet-dex*|^https://kava3.data.kava.io*|^https://ipfs.io*|^https://apps.apple.com*|^https://kava.quicksync.io*"


lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify
.PHONY: lint

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/tendermint
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/cosmos/cosmos-sdk
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/kava-labs/kava
.PHONY: format

###############################################################################
###                                Localnet                                 ###
###############################################################################

build-docker-local-kava:
	@$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/kvd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/kvd:Z kava/kavanode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --keyring-backend=test ; fi
	docker-compose up -d

localnet-stop:
	docker-compose down

# Launch a new single validator chain
start:
	./contrib/devnet/init-new-chain.sh
	kava start

#proto-format:
#@echo "Formatting Protobuf files"
#@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
#find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -style=file -i {} \; ; fi

########################################
### Testing

# TODO tidy up cli tests to use same -Enable flag as simulations, or the other way round
# TODO -mod=readonly ?
# build dependency needed for cli tests
test-all: build
	# basic app tests
	@go test ./app -v
	# basic simulation (seed "4" happens to not unbond all validators before reaching 100 blocks)
	#@go test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	# other sim tests
	#@go test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	#@go test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	# AppStateDeterminism does not use Seed flag
	#@go test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h

# run module tests and short simulations
test-basic: test
	@go test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	# other sim tests
	@go test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	@go test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	@# AppStateDeterminism does not use Seed flag
	@go test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m

test:
	@go test $$(go list ./... | grep -v 'contrib')

# Run cli integration tests
# `-p 4` to use 4 cores, `-tags cli_test` to tell go not to ignore the cli package
# These tests use the `kvd` or `kvcli` binaries in the build dir, or in `$BUILDDIR` if that env var is set.
test-cli: build
	@go test ./cli_test -tags cli_test -v -p 4

# Run tests for migration cli command
test-migrate:
	@go test -v -count=1 ./migrate/...

# Kick start lots of sims on an AWS cluster.
# This submits an AWS Batch job to run a lot of sims, each within a docker image. Results are uploaded to S3
start-remote-sims:
	# build the image used for running sims in, and tag it
	docker build -f simulations/Dockerfile -t kava/kava-sim:master .
	# push that image to the hub
	docker push kava/kava-sim:master
	# submit an array job on AWS Batch, using 1000 seeds, spot instances
	aws batch submit-job \
		-—job-name "master-$(VERSION)" \
		-—job-queue “simulation-1-queue-spot" \
		-—array-properties size=1000 \
		-—job-definition kava-sim-master \
		-—container-override environment=[{SIM_NAME=master-$(VERSION)}]

.PHONY: all build-linux install clean build test test-cli test-all test-rest test-basic start-remote-sims
