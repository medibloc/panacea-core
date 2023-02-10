#!/usr/bin/env bash

# Reference:
# - https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/scripts/protocgen.sh
# - https://github.com/cosmos/cosmos-sdk/blob/v0.46.8/scripts/protocgen.sh

set -eo pipefail

# regen-network/cosmos-proto must be used to handle Go interface types as google.protobuf.Any.
# https://github.com/regen-network/cosmos-proto
protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
}

protoc_gen_gocosmos

cd proto

proto_dirs=$(find ./panacea -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    buf generate --template buf.gen.gogo.yaml $file
  done
done

cd -

## move proto files to the right places
find ./x -type f -name "*.pb*.go" -exec rm {} \;
cp -rv github.com/medibloc/panacea-core/v2/x/* ./x/
rm -rfv github.com