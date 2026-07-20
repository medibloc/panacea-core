#!/bin/sh

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


set -eu

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd)
repo_root=$(CDPATH='' cd "$script_dir/../.." && pwd)
generation_root=$(mktemp -d "${TMPDIR:-/tmp}/panacea-protocgen.XXXXXX")

cleanup() {
  rm -rf -- "$generation_root"
}

trap cleanup EXIT
trap 'exit 1' HUP INT TERM

echo "Generating gogo proto code"
cp -R "$repo_root/proto" "$generation_root/proto"
cd "$generation_root/proto"
find ./panacea -name '*.proto' \
  -exec buf generate --template buf.gen.gogo.yaml {} \;

# move proto files to the right places
module_path=$(awk '$1 == "module" { print $2; exit }' "$repo_root/go.mod")
case "$module_path" in
  */*) ;;
  *)
    echo "failed to determine a valid module path from go.mod" >&2
    exit 1
    ;;
esac

case "$module_path" in
  /*|../*|*/../*|*/..|./*|*/./*|*/.)
    echo "refusing to use unsafe module path: $module_path" >&2
    exit 1
    ;;
esac

generated_module_dir=$generation_root/$module_path
module_version=${module_path##*/}
staged_repo_root=$generation_root/staged-repo
mkdir -p "$staged_repo_root"

# Older protos still generate below the pre-v2 import path. Copy those entries
# individually so the module-version directory itself is not copied to ./v2.
case "$module_version" in
  v[2-9]|v[1-9][0-9]*)
    generated_legacy_dir=$generation_root/${module_path%/*}
    if [ -d "$generated_legacy_dir" ]; then
      for generated_entry in "$generated_legacy_dir"/*; do
        [ -e "$generated_entry" ] || continue
        [ "$generated_entry" = "$generated_module_dir" ] && continue
        cp -R "$generated_entry" "$staged_repo_root"/
      done
    fi
    ;;
esac

if [ -d "$generated_module_dir" ]; then
  cp -R "$generated_module_dir"/. "$staged_repo_root"/
fi

generated_manifest=$generation_root/generated-files.txt
(
  cd "$staged_repo_root"
  find ./x -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' \) -print \
    | LC_ALL=C sort >"$generated_manifest"
)

if [ ! -s "$generated_manifest" ]; then
  echo "protobuf generation produced no Go files" >&2
  exit 1
fi

cp -R "$staged_repo_root"/. "$repo_root"/

# All protobuf Go files below x/ are generated from proto/panacea. Remove files
# that were not produced by this run so deleted or renamed protos cannot linger.
(
  cd "$repo_root"
  find ./x -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' \) -print |
    while IFS= read -r generated_file; do
      if ! grep -Fqx "$generated_file" "$generated_manifest"; then
        rm -f -- "$generated_file"
      fi
    done
)
