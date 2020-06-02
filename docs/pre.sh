#!/usr/bin/env bash

mkdir -p Modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "Modules/$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "Modules/$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/* "$_" && mv "./$_/README.md" "./$_/00_README.md" 
  fi
done

baseGitUrl="https://raw.githubusercontent.com/Kava-Labs"

# Client docs (JavaScript SDK)
clientGitRepo="javascript-sdk"
clientDir="Building"

mkdir -p "./${clientDir}"
curl "${baseGitUrl}/${clientGitRepo}/master/README.md" -o "./${clientDir}/${clientGitRepo}.md"

# Kava Tools docs
toolsGitRepo="kava-tools"
toolsDir="Tools"
toolDocs=("auction" "oracle")

mkdir -p "./${toolsDir}"
for T in ${toolDocs[@]}; do
  curl "${baseGitUrl}/${toolsGitRepo}/master/${T}/README.md" -o "./${toolsDir}/${T}.md"
done

