#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <command> [args...]"
  exit 1
fi

CMD="$1"
shift

docker compose \
  -f "$SCRIPT_DIR/docker-compose.yml" \
  run --rm \
  api \
  go run ./cmd/"$CMD"/main.go "$@"
