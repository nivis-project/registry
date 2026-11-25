#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
fail=0

contract() {
  local dir="$1" resources="$2"
  mkdir -p "$dir/docs/providers/hashicorp/http/3.6.3"
  printf '[{"namespace":"hashicorp","name":"http","version":"3.6.3"}]\n' > "$dir/catalog.json"
  if [ "$resources" != "absent" ]; then
    printf '{"id":"3.6.3","docs":{"resources":%s,"datasources":[],"functions":[],"guides":[]}}\n' "$resources" \
      > "$dir/docs/providers/hashicorp/http/3.6.3/index.json"
  fi
}

expect() {
  local name="$1" want="$2" pattern="$3" dir="$work/$1"
  local got=pass
  bash "$HERE/verify-contract.sh" "$dir" > "$work/$name.log" 2>&1 || got=fail
  if [ "$got" != "$want" ]; then
    echo "test-verify-contract: FAIL $name: got $got, expected $want"; sed 's/^/    /' "$work/$name.log"; fail=1
  elif [ -n "$pattern" ] && ! grep -qE "$pattern" "$work/$name.log"; then
    echo "test-verify-contract: FAIL $name: the message does not match /$pattern/"; sed 's/^/    /' "$work/$name.log"; fail=1
  else
    echo "test-verify-contract: ok $name ($want)"
  fi
}

contract "$work/complete" '[{"name":"http_x","title":"http_x"}]'
expect complete pass ""

contract "$work/empty-lists" '[]'
expect empty-lists pass ""

contract "$work/null-list" 'null'
expect null-list fail 'hashicorp/http.*resources'

contract "$work/missing-index" absent
expect missing-index fail 'no index document'

exit "$fail"
