#!/usr/bin/env bash
set -euo pipefail

mode="${FRAMEWORK_LOGGER_GUARD_MODE:-fail}"
root="${1:-backend}"

if [ ! -d "$root" ]; then
  echo "logger guard: path not found: $root" >&2
  exit 2
fi

pattern='logrus\.New\(|\blog\.Print|\blog\.Fatal|\blog\.Panic'
matches="$(rg -n --glob '*.go' --glob '!**/*_test.go' --glob '!**/internal/logger/**' "$pattern" "$root" || true)"

if [ -z "$matches" ]; then
  echo "logger guard: ok"
  exit 0
fi

echo "logger guard: direct logger usages found"
echo "$matches"

if [ "$mode" = "warn" ]; then
  exit 0
fi
exit 1
