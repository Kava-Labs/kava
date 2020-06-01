#!/usr/bin/env bash

# Module specs
for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "./$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "./$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/* "$_" && mv "./$_/README.md" "./$_/00_README.md" 
  fi
done

baseGitUrl="https://raw.githubusercontent.com/Kava-Labs"

# Client docs (JavaScript SDK)
clientGitRepo="javascript-sdk"
clientDir="building-on-kava"

mkdir -p "./${clientDir}"
curl "${baseGitUrl}/${clientGitRepo}/master/README.md" -o "./${clientDir}/${clientGitRepo}.md"

# Kava Tools docs
toolsGitRepo="kava-tools"
toolDocs=("auction" "oracle")

mkdir -p "./${toolsGitRepo}"
for T in ${toolDocs[@]}; do
  curl "${baseGitUrl}/${toolsGitRepo}/master/${T}/README.md" -o "./${toolsGitRepo}/${T}.md"
done