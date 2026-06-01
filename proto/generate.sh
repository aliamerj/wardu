#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

PROTO_DIR="." # Look for .proto files in the same directory as the script
GO_OUT=".."   # Output generated Go files in the parent directory (project root)

PROTO_SRC=$(find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f)

if [ -z "$PROTO_SRC" ]; then
  echo "Error: No .proto files found in '$SCRIPT_DIR'"
  exit 1
fi

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$GO_OUT" \
  --go-grpc_out="$GO_OUT" \
  "$PROTO_SRC"
