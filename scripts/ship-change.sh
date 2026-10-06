#!/usr/bin/env bash
# Ship one OpenSpec change: gate, archive, commit, push.
#
#   scripts/ship-change.sh <change-name> "<commit subject>" [path...]
#
# Trailing paths scope the commit to those files, so several changes can be
# shipped one at a time out of a single working tree. With no paths the whole
# working copy is committed.
#
# The gate is non-negotiable and runs BEFORE anything is archived or committed:
# every task checked off, `nix flake check` green, and Go test coverage at or
# above the thresholds below. A failing gate leaves the tree untouched.
set -euo pipefail

COVERAGE_MIN_TOTAL=70
COVERAGE_MIN_CORE=80
CORE_PACKAGES="seed extract compat generate module version"

die() { printf '\n\033[31mship: %s\033[0m\n' "$*" >&2; exit 1; }
step() { printf '\n\033[1m==> %s\033[0m\n' "$*"; }

change="${1:-}"
subject="${2:-}"
[ -n "$change" ]  || die "usage: scripts/ship-change.sh <change-name> \"<commit subject>\""
[ -n "$subject" ] || die "usage: scripts/ship-change.sh <change-name> \"<commit subject>\""

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo"

step "Change: $change"
openspec validate "$change" >/dev/null || die "openspec validate failed for $change"

step "Gate 1/3: every task checked off"
status="$(openspec status --change "$change" 2>&1)" || die "openspec status failed"
tasks_file="$(printf '%s\n' "$status" | sed -n 's/^Change root: //p')/tasks.md"
[ -f "$tasks_file" ] || die "tasks.md not found at $tasks_file"
if grep -q '^\s*- \[ \]' "$tasks_file"; then
  printf 'unchecked tasks:\n' >&2
  grep -n '^\s*- \[ \]' "$tasks_file" >&2
  die "$change has unchecked tasks"
fi
printf 'all tasks checked\n'

step "Gate 2/3: nix flake check"
nix flake check || die "nix flake check failed"

step "Gate 3/3: Go test coverage"
profile="$(mktemp)"
trap 'rm -f "$profile"' EXIT
( cd tools && go test -coverprofile="$profile" ./... >/dev/null ) || die "go test failed"

total="$(cd tools && go tool cover -func="$profile" | awk '/^total:/ { gsub("%","",$3); print $3 }')"
awk -v got="$total" -v min="$COVERAGE_MIN_TOTAL" \
  'BEGIN { exit !(got + 0 >= min) }' \
  || die "overall coverage ${total}% is below ${COVERAGE_MIN_TOTAL}%"
printf 'overall %s%% (min %s%%)\n' "$total" "$COVERAGE_MIN_TOTAL"

for pkg in $CORE_PACKAGES; do
  pct="$(cd tools && go test -cover "./$pkg/" 2>/dev/null \
        | sed -n 's/.*coverage: \([0-9.]*\)% of statements.*/\1/p')"
  [ -n "$pct" ] || die "could not read coverage for core package $pkg"
  awk -v got="$pct" -v min="$COVERAGE_MIN_CORE" -v p="$pkg" \
    'BEGIN { exit !(got + 0 >= min) }' \
    || die "core package $pkg at ${pct}% is below ${COVERAGE_MIN_CORE}%"
  printf '  %-10s %s%%\n' "$pkg" "$pct"
done

step "Archiving $change"
openspec archive "$change" --yes

step "Committing"
shift 2 || true
if [ "$#" -gt 0 ]; then
  jj commit -m "$subject" "$@"
else
  jj commit -m "$subject"
fi
jj bookmark set main -r @-
jj log -r @- --no-graph -T 'commit_id.short() ++ " " ++ author.name() ++ " <" ++ author.email() ++ ">\n"'

step "Pushing main"
jj git push --bookmark main

printf '\n\033[32mship: %s is live\033[0m\n' "$change"
