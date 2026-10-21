#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Ensure we have the current built contracts' ABI to check against
bash "$SCRIPT_DIR/generate_abi.sh"

# Use git diff to see if docs/abi.json changed
if git diff --exit-code "$ROOT_DIR/docs/abi.json" > /dev/null; then
    echo "Success: docs/abi.json is up-to-date."
    exit 0
else
    echo "ERROR: docs/abi.json is out of date!" >&2
    echo "Please run 'make check-abi' or 'bash scripts/generate_abi.sh' to update it, and commit the changes." >&2
    git diff "$ROOT_DIR/docs/abi.json" >&2
    # Restore the original file so we don't leave the working tree dirty in CI, though CI will exit anyway
    git checkout -- "$ROOT_DIR/docs/abi.json"
    exit 1
fi
