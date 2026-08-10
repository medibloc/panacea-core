#!/usr/bin/env bash

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
fixture_root=$(mktemp -d "${TMPDIR:-/tmp}/panacea-protocgen-test.XXXXXX")
trap 'rm -rf "$fixture_root"' EXIT

mkdir -p \
  "$fixture_root/repo/proto/scripts" \
  "$fixture_root/repo/proto/panacea/nft/v1" \
  "$fixture_root/repo/proto/panacea/pnft/v2" \
  "$fixture_root/repo/x/nft/types" \
  "$fixture_root/repo/github.com/private-data" \
  "$fixture_root/bin"

cp "$repo_root/proto/scripts/protocgen.sh" "$fixture_root/repo/proto/scripts/protocgen.sh"
cat >"$fixture_root/repo/go.mod" <<'EOF'
module github.com/medibloc/panacea-core/v2
EOF
touch \
  "$fixture_root/repo/proto/buf.gen.gogo.yaml" \
  "$fixture_root/repo/proto/panacea/nft/v1/nft.proto" \
  "$fixture_root/repo/proto/panacea/pnft/v2/pnft.proto" \
  "$fixture_root/repo/x/nft/types/stale.pb.go" \
  "$fixture_root/repo/x/nft/types/handwritten.go" \
  "$fixture_root/repo/github.com/private-data/sentinel.txt"

cat >"$fixture_root/bin/buf" <<'EOF'
#!/usr/bin/env bash

set -euo pipefail

if [[ "$*" != "generate --template buf.gen.gogo.yaml --path ./panacea" ]]; then
  echo "unexpected buf arguments: $*" >&2
  exit 1
fi

if [[ "${BUF_FAIL_PNFT:-0}" == 1 ]]; then
  mkdir -p ../github.com/medibloc/panacea-core/v2/x/nft/types
  printf 'partial output\n' \
    >../github.com/medibloc/panacea-core/v2/x/nft/types/nft.pb.go
  echo "injected pnft generation failure" >&2
  exit 42
fi

mkdir -p \
  ../github.com/medibloc/panacea-core/x/pnft/types \
  ../github.com/medibloc/panacea-core/v2/x/nft/types
touch \
  ../github.com/medibloc/panacea-core/x/pnft/types/pnft.pb.go \
  ../github.com/medibloc/panacea-core/v2/x/nft/types/nft.pb.go
EOF
chmod +x "$fixture_root/bin/buf"

(
  cd "$fixture_root/repo"
  PATH="$fixture_root/bin:$PATH" sh ./proto/scripts/protocgen.sh
)

failures=0

assert_file() {
  if [[ ! -f "$1" ]]; then
    echo "expected file to exist: $1" >&2
    failures=1
  fi
}

assert_absent() {
  if [[ -e "$1" ]]; then
    echo "expected path to be absent: $1" >&2
    failures=1
  fi
}

assert_file "$fixture_root/repo/x/nft/types/nft.pb.go"
assert_file "$fixture_root/repo/x/pnft/types/pnft.pb.go"
assert_absent "$fixture_root/repo/x/nft/types/stale.pb.go"
assert_file "$fixture_root/repo/x/nft/types/handwritten.go"
assert_file "$fixture_root/repo/github.com/private-data/sentinel.txt"
assert_absent "$fixture_root/repo/v2"

if ((failures != 0)); then
  exit 1
fi

printf 'original nft generated file\n' >"$fixture_root/repo/x/nft/types/nft.pb.go"
printf 'original pnft generated file\n' >"$fixture_root/repo/x/pnft/types/pnft.pb.go"
printf 'original stale generated file\n' >"$fixture_root/repo/x/nft/types/stale.pb.go"

if (
  cd "$fixture_root/repo"
  BUF_FAIL_PNFT=1 PATH="$fixture_root/bin:$PATH" sh ./proto/scripts/protocgen.sh
); then
  echo "partial protobuf generation unexpectedly succeeded" >&2
  exit 1
fi

if [[ $(<"$fixture_root/repo/x/nft/types/nft.pb.go") != "original nft generated file" ]]; then
  echo "failed protobuf generation changed the existing nft generated file" >&2
  exit 1
fi
if [[ $(<"$fixture_root/repo/x/pnft/types/pnft.pb.go") != "original pnft generated file" ]]; then
  echo "failed protobuf generation changed the existing pnft generated file" >&2
  exit 1
fi
if [[ $(<"$fixture_root/repo/x/nft/types/stale.pb.go") != "original stale generated file" ]]; then
  echo "failed protobuf generation deleted the existing stale generated file" >&2
  exit 1
fi

echo "protocgen path test passed"
