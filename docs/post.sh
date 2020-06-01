#!/usr/bin/env bash

# Module specs
for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "./$(echo $D | awk -F/ '{print $NF}')"
  fi
done

# JavaScript SDK docs
 rm -rf "./building-on-kava"

 # Kava Tools docs
 rm -rf "./kava-tools"
