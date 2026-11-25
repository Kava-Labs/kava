#!/bin/bash
set -euxo pipefail

sudo apt-get update
sudo apt-get install -y libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev

source_dir=$(mktemp -d)
git clone --depth 1 --branch "$ROCKSDB_VERSION" https://github.com/facebook/rocksdb.git "$source_dir"
sudo make -C "$source_dir" -j "$(nproc)" install-shared
sudo ldconfig
