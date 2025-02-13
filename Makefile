################################################################################
###                             Project Info                                 ###
################################################################################
PROJECT_NAME := kava# unique namespace for project

GO_BIN ?= go

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT := $(shell git rev-parse --short HEAD)

BRANCH_PREFIX := $(shell echo $(GIT_BRANCH) | sed 's/\/.*//g')# eg release, master, feat

EXACT_TAG := $(shell git describe --tags --exact-match 2> /dev/null)
RECENT_TAG := $(shell git describe --tags)

ifeq ($(BRANCH_PREFIX), release)
# we are on a release branch, set version to the last or current tag
VERSION := $(RECENT_TAG)# use current tag or most recent tag + number of commits + g + abbrivated commit
VERSION_NUMBER := $(shell echo $(VERSION) | sed 's/^v//')# drop the "v" prefix for versions
else ifeq ($(EXACT_TAG), $(RECENT_TAG))
# we have a tag checked out directly
VERSION := $(RECENT_TAG)# use exact tag
VERSION_NUMBER := $(shell echo $(VERSION) | sed 's/^v//')# drop the "v" prefix for versions
else
# we are not on a release branch, and do not have clean tag history (etc v0.19.0-xx-gxx will not make sense to use)
VERSION := $(GIT_COMMIT_SHORT)
VERSION_NUMBER := $(VERSION)
endif

TENDERMINT_VERSION := $(shell $(GO_BIN) list -m github.com/cometbft/cometbft | sed 's:.* ::')
COSMOS_SDK_VERSION := $(shell $(GO_BIN) list -m github.com/cosmos/cosmos-sdk | sed 's:.* ::')

.PHONY: print-git-info
print-git-info:
	@echo "branch $(GIT_BRANCH)\nbranch_prefix $(BRANCH_PREFIX)\ncommit $(GIT_COMMIT)\ncommit_short $(GIT_COMMIT_SHORT)"

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

include $(BUILD_DIR)/lint.mk

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
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TENDERMINT_VERSION)

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

# build on rocksdb-backed kava on macOS with shared libs from brew
# this assumes you are on macOS & these deps have been installed with brew:
# rocksdb, snappy, lz4, and zstd
# use like `make build-rocksdb-brew COSMOS_BUILD_OPTIONS=rocksdb`
build-rocksdb-brew:
	export CGO_CFLAGS := -I$(shell brew --prefix rocksdb)/include
	export CGO_LDFLAGS := -L$(shell brew --prefix rocksdb)/lib -lrocksdb -lstdc++ -lm -lz -L$(shell brew --prefix snappy)/lib -L$(shell brew --prefix lz4)/lib -L$(shell brew --prefix zstd)/lib

install: go.sum
	$(GO_BIN) install -mod=readonly $(BUILD_FLAGS) ./cmd/kava

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download $(GO_BIN) modules to local cache"
	@$(GO_BIN) mod download
PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensuring dependencies have not been modified"
	@$(GO_BIN) mod verify

########################################
### Linting

# Check url links in the repo are not broken.
# This tool checks local markdown links as well.
# Set to exclude riot links as they trigger false positives
link-check:
	@$(GO_BIN) get -u github.com/raviqqe/liche@f57a5d1c5be4856454cb26de155a65a4fd856ee3
	liche -r . --exclude "^http://127.*|^https://riot.im/app*|^http://kava-testnet*|^https://testnet-dex*|^https://kava3.data.kava.io*|^https://ipfs.io*|^https://apps.apple.com*|^https://kava.quicksync.io*"

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

# Build docker image and tag as kava/kava:local
docker-build:
	DOCKER_BUILDKIT=1 $(DOCKER) build -t kava/kava:local --load .

docker-build-rocksdb:
	DOCKER_BUILDKIT=1 $(DOCKER) build -f Dockerfile-rocksdb -t kava/kava:local .

build-docker-local-kava:
	@$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/kvd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/kvd:Z kava/kavanode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --keyring-backend=test ; fi
	$(DOCKER) compose up -d

localnet-stop:
	$(DOCKER) compose down

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
	@$(GO_BIN) test ./app -v
	# basic simulation (seed "4" happens to not unbond all validators before reaching 100 blocks)
	#@$(GO_BIN) test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	# other sim tests
	#@$(GO_BIN) test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	#@$(GO_BIN) test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h
	# AppStateDeterminism does not use Seed flag
	#@$(GO_BIN) test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 4 -v -timeout 24h

# run module tests and short simulations
test-basic: test
	@$(GO_BIN) test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	# other sim tests
	@$(GO_BIN) test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	@$(GO_BIN) test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m
	@# AppStateDeterminism does not use Seed flag
	@$(GO_BIN) test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 4 -v -timeout 2m

# run end-to-end tests (local docker container must be built, see docker-build)
test-e2e: docker-build
	$(GO_BIN) test -failfast -count=1 -v ./tests/e2e/...

# run interchaintest tests (./tests/e2e-ibc)
# Use -count=1 to prevent caching, in case docker-build changes
test-ibc: docker-build
	cd tests/e2e-ibc && KAVA_TAG=local $(GO_BIN) test -failfast -timeout 10m -count=1 .
.PHONY: test-ibc

test:
	@$(GO_BIN) test $$($(GO_BIN) list ./... | grep -v 'contrib' | grep -v 'tests/e2e')

# Run cli integration tests
# `-p 4` to use 4 cores, `-tags cli_test` to tell $(GO_BIN) not to ignore the cli package
# These tests use the `kvd` or `kvcli` binaries in the build dir, or in `$BUILDDIR` if that env var is set.
test-cli: build
	@$(GO_BIN) test ./cli_test -tags cli_test -v -p 4

# Run tests for migration cli command
test-migrate:
	@$(GO_BIN) test -v -count=1 ./migrate/...

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
		-—job-queue "simulation-1-queue-spot" \
		-—array-properties size=1000 \
		-—job-definition kava-sim-master \
		-—container-override environment=[{SIM_NAME=master-$(VERSION)}]

update-kvtool:
	git submodule init || true
	git submodule update
	cd tests/e2e/kvtool && make install

.PHONY: all build-linux install clean build test test-cli test-all test-rest test-basic start-remote-sims
