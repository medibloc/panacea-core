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

proto_file=${*: -1}
case "$proto_file" in
  *panacea/pnft/*)
    output_dir=../github.com/medibloc/panacea-core/x/pnft/types
    output_file=pnft.pb.go
    ;;
  *panacea/nft/*)
    output_dir=../github.com/medibloc/panacea-core/v2/x/nft/types
    output_file=nft.pb.go
    ;;
  *)
    echo "unexpected proto file: $proto_file" >&2
    exit 1
    ;;
esac

mkdir -p "$output_dir"
touch "$output_dir/$output_file"
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

echo "protocgen path test passed"
