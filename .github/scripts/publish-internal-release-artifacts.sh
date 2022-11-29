#!/bin/bash
set -x

archs=(amd64)
dbs=(goleveldb rocksdb)
OS=linux

for db in "${dbs[@]}"
do
    for arch in "${archs[@]}"
    do
        env GOOS=$OS GOARCH="$arch" make build COSMOS_BUILD_OPTIONS="$db" && aws s3 cp ./out/$OS/kava s3://releases.kava.io/kava/$OS/"$db"/"$BUILD_TAG"
    done
done
