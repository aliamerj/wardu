#!/usr/bin/env bash
set -euo pipefail

old_pwd="$(pwd)"
script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$script_path"

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
oapi-codegen --config cfg.yaml openapi.yml

cd "$old_pwd"
