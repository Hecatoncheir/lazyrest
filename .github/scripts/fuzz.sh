#!/usr/bin/env bash
# Runs every fuzz target in the module. Both the push job and the weekly job in
# go-test.yml call this, so the two differ only in how long they fuzz.
#
# Usage: .github/scripts/fuzz.sh <fuzztime> <parallel>
#
# The targets are discovered rather than listed, so a target added to the tree
# cannot silently go unfuzzed. A discovery that comes back empty is an error
# rather than a run that quietly fuzzes nothing.

set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <fuzztime> <parallel>" >&2
  exit 2
fi

fuzztime="$1"
parallel="$2"

fuzz_targets=$(go test ./... -list='^Fuzz' | awk '
  /^ok[ \t]/ { for (i = 1; i <= count; i++) print $2, targets[i]; count = 0; next }
  /^Fuzz/    { targets[++count] = $1 }
')

if [ -z "$fuzz_targets" ]; then
  echo "no fuzz targets discovered" >&2
  exit 1
fi

printf '%s\n' "$fuzz_targets"

while read -r package_name fuzz_target; do
  go test "$package_name" -run='^$' -fuzz="^${fuzz_target}\$" -fuzztime="$fuzztime" -parallel="$parallel"
done <<<"$fuzz_targets"
