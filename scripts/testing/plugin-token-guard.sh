#!/usr/bin/env bash
set -euo pipefail

mode="${PLUGIN_TOKEN_GUARD_MODE:-fail}"
root="${1:-backend}"

if [ ! -d "$root" ]; then
  echo "plugin token guard: path not found: $root" >&2
  exit 2
fi

pattern='PX_TOOL_TOKEN|PX_PLUGIN_TOOL_TOKEN|AllowSignedContext:\s*true|X-PowerX-CTX'
matches="$(rg -n --glob '*.go' "$pattern" "$root" || true)"

if [ -z "$matches" ]; then
  echo "plugin token guard: ok"
  exit 0
fi

echo "plugin token guard: deprecated plugin tool token usage found"
echo "$matches"

if [ "$mode" = "warn" ]; then
  exit 0
fi
exit 1
