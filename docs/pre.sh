#!/usr/bin/env bash

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "./$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "./$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/* "$_" && mv "./$_/README.md" "./$_/00_README.md" 
  fi
done