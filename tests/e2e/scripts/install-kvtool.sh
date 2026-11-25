#!/usr/bin/env bash
set -euo pipefail

e2e_dir=$(cd "$(dirname "$0")/.." && pwd)
output="$e2e_dir/../../build/.cache/kvtool"
templates="$output/config/templates"
mkdir -p "$templates"
cp -R "$e2e_dir/kvtool/config/templates/." "$templates/"

cd "$e2e_dir/kvtool"
go build -ldflags "-X 'github.com/kava-labs/kvtool/config/generate.ConfigTemplatesDir=$templates'" -o "$output/kvtool" .
