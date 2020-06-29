FROM ubuntu:18.04

ARG GOVERSION=1.14.4

WORKDIR /binaries

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file build-essential wget git

RUN wget https://dl.google.com/go/go${GOVERSION}.linux-amd64.tar.gz  --quiet --continue && \
    tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz && \
    mkdir -p go/{bin,pkg,src} && \
    git clone https://github.com/kava-labs/kava



RUN cd kava && git checkout v0.8.1 &&\
 /usr/local/go/bin/go build -mod=readonly -tags "netgo ledger" -ldflags '-X github.com/cosmos/cosmos-sdk/version.Name=kava -X github.com/cosmos/cosmos-sdk/version.ServerName=kvd -X github.com/cosmos/cosmos-sdk/version.ClientName=kvcli -X github.com/cosmos/cosmos-sdk/version.Version=0.8.1 -X github.com/cosmos/cosmos-sdk/version.Commit=869189054d68d6ec3e6446156ea0a91eb45af09c -X "github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger"' -o build/kvd ./cmd/kvd && \
/usr/local/go/bin/go build -mod=readonly -tags "netgo ledger" -ldflags '-X github.com/cosmos/cosmos-sdk/version.Name=kava -X github.com/cosmos/cosmos-sdk/version.ServerName=kvd -X github.com/cosmos/cosmos-sdk/version.ClientName=kvcli -X github.com/cosmos/cosmos-sdk/version.Version=0.8.1 -X github.com/cosmos/cosmos-sdk/version.Commit=869189054d68d6ec3e6446156ea0a91eb45af09c -X "github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger"' -o build/kvcli ./cmd/kvcli

