#!/bin/bash

if hash kvtool 2>/dev/null; then
  echo "[install-kvtool.sh] kvtool is already installed. skipping installation."
  exit 0
fi

echo "[install-kvtool.sh] installing kvtool."
cd kvtool || exit 1
make install
