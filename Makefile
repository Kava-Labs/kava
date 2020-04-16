#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true

export GO111MODULE = on

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

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=kava \
		  -X github.com/cosmos/cosmos-sdk/version.ServerName=kvd \
		  -X github.com/cosmos/cosmos-sdk/version.ClientName=kvcli \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'


all: install

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/kvd.exe ./cmd/kvd
	go build -mod=readonly $(BUILD_FLAGS) -o build/kvcli.exe ./cmd/kvcli
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/kvd ./cmd/kvd
	go build -mod=readonly $(BUILD_FLAGS) -o build/kvcli ./cmd/kvcli
endif

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/kvd
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/kvcli

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensuring dependencies have not been modified"
	@go mod verify

clean:
	rm -rf build/

########################################
### Linting

# Check url links in the repo are not broken.
# This tool checks local markdown links as well.
# Set to exclude riot links as they trigger false positives
link-check:
	@go run github.com/raviqqe/liche -r . --exclude "^http://127.*|^https://riot.im/app*"

########################################
### Testing

# TODO tidy up cli tests to use same -Enable flag as simulations, or the other way round
# TODO -mod=readonly ?
# build dependency needed for cli tests
test-all: build test test-cli
	# basic app tests
	@go test ./app -v
	# basic simulation (seed "2" happens to not unbond all validators before reaching 100 blocks)
	@go test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 2 -v -timeout 24h
	# other sim tests
	@go test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 2 -v -timeout 24h
	@go test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 2 -v -timeout 24h
	@# AppStateDeterminism does not use Seed flag
	@go test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed 2 -v -timeout 24h

# run module tests and short simulations
test-basic: test
	# basic simulation (seed "2" happens to not unbond all validators before reaching 100 blocks)
	@go test ./app -run TestFullAppSimulation        -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 2 -v -timeout 2m
	# other sim tests
	@go test ./app -run TestAppImportExport          -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 2 -v -timeout 2m
	@go test ./app -run TestAppSimulationAfterImport -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 2 -v -timeout 2m
	@# AppStateDeterminism does not use Seed flag
	@go test ./app -run TestAppStateDeterminism      -Enabled -Commit -NumBlocks=5 -BlockSize=200 -Seed 2 -v -timeout 2m

test:
	@go test ./...

test_dredd:
	rest_test/./run_all_tests_from_make.sh

test-cli:
	@go test ./cli_test -tags cli_test -v -p 4

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

.PHONY: all build-linux install clean build test test-cli test-all test_dredd test-basic start-remote-sims
