#!/usr/bin/env bash

## Create modules pages
mkdir -p docs/modules
cp modules_category.json docs/modules/_category_.json

for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=docs/modules/$(echo $D | awk -F/ '{print $NF}')
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
  fi
done