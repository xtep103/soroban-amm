#!/usr/bin/env bash
set -euo pipefail

# This script generates the ABI for all deployable contracts and outputs it as JSON.
# It expects the contracts to be built first.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WASM_DIR="$ROOT_DIR/target/wasm32v1-none/release"
ABI_FILE="$ROOT_DIR/docs/abi.json"
TEMP_FILE="$ROOT_DIR/docs/abi.tmp.json"

if ! command -v stellar >/dev/null 2>&1; then
    echo "ERROR: 'stellar' CLI not found. Please install it (e.g., cargo install --locked stellar-cli)." >&2
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: 'jq' is required but not installed." >&2
    exit 1
fi

shopt -s nullglob
wasms=("$WASM_DIR"/*.wasm)
if [ ${#wasms[@]} -eq 0 ]; then
    echo "ERROR: No WASM artifacts found in $WASM_DIR." >&2
    echo "Did you run 'make build' first?" >&2
    exit 1
fi

echo '{"contracts": {}}' > "$TEMP_FILE"

for wasm_path in "${wasms[@]}"; do
    contract_name=$(basename "$wasm_path" .wasm)
    
    echo "Generating ABI for $contract_name..."
    abi_json=$(stellar contract info interface --wasm "$wasm_path" --output json)
    
    # Check if the output is valid JSON
    if ! echo "$abi_json" | jq . >/dev/null 2>&1; then
        echo "ERROR: 'stellar contract info interface' failed for $contract_name" >&2
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    # Merge into temp file
    jq --arg name "$contract_name" --argjson abi "$abi_json" \
       '.contracts[$name] = $abi' "$TEMP_FILE" > "${TEMP_FILE}.new"
    mv "${TEMP_FILE}.new" "$TEMP_FILE"
done

# Format using jq for consistent spacing
jq . "$TEMP_FILE" > "$ABI_FILE"
rm -f "$TEMP_FILE"

echo "Successfully generated $ABI_FILE"
