#!/bin/bash
set -x

# install build dependencies
sudo apt-get install -y libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev

# get rocksdb sources
git clone https://github.com/facebook/rocksdb.git /home/runner/rocksdb

cd /home/runner/rocksdb

git checkout "$ROCKSDB_VERSION"

# install rocksdb locally
sudo make -j $(nproc --all) install-shared
