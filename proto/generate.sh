#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

PROTO_DIR="."
GO_OUT=".."

mapfile -t PROTO_FILES < <(
  find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f
)

if [ ${#PROTO_FILES[@]} -eq 0 ]; then
  echo "Error: No .proto files found in '$SCRIPT_DIR'"
  exit 1
fi

for proto in "${PROTO_FILES[@]}"; do
  echo "Generating: $proto"

  protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$GO_OUT" \
    --go-grpc_out="$GO_OUT" \
    "$proto"

done
