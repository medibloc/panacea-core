#!/usr/bin/env bash

#== Requirements ==
#
## make sure your `go env GOPATH` is in the `$PATH`
## Install:
## + latest buf (v1.0.0-rc11 or later)
## + protobuf v3
#
## All protoc dependencies must be installed not in the module scope
## currently we must use grpc-gateway v1
# cd ~
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
# go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
# go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
# go get github.com/regen-network/cosmos-proto@latest # doesn't work in install mode


set -eo pipefail

echo "Generating gogo proto code"
cd proto
while IFS= read -r -d '' file; do
  buf generate --template buf.gen.gogo.yaml "$file"
done < <(find ./panacea -name '*.proto' -print0)

cd ..

# move proto files to the right places
module_path=$(awk '$1 == "module" { print $2; exit }' go.mod)
if [[ -z "$module_path" || "$module_path" != */* ]]; then
  echo "failed to determine a valid module path from go.mod" >&2
  exit 1
fi

generated_module_dir=$module_path
module_version=${module_path##*/}

# Older protos still generate below the pre-v2 import path. Copy those entries
# individually so the module-version directory itself is not copied to ./v2.
case "$module_version" in
  v[2-9]|v[1-9][0-9]*)
    generated_legacy_dir=${module_path%/*}
    if [[ -d "$generated_legacy_dir" ]]; then
      for generated_entry in "$generated_legacy_dir"/*; do
        [[ -e "$generated_entry" ]] || continue
        [[ "$generated_entry" == "$generated_module_dir" ]] && continue
        cp -R "$generated_entry" ./
      done
    fi
    ;;
esac

if [[ -d "$generated_module_dir" ]]; then
  cp -R "$generated_module_dir"/. ./
fi

generated_namespace=${module_path%%/*}
if [[ ! "$generated_namespace" =~ ^[[:alnum:]][[:alnum:].-]*$ ]]; then
  echo "refusing to remove invalid generated namespace: $generated_namespace" >&2
  exit 1
fi
rm -rf -- "$generated_namespace"
