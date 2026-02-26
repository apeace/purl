#!/usr/bin/env bash
set -euo pipefail

if ! grep -qi ubuntu /etc/os-release 2>/dev/null; then
  echo "error: deploy/cmd.sh is for prod only. Locally, use ./cmd.sh instead." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <command> [args...]"
  exit 1
fi

CMD="$1"
shift

EXTRA_ARGS=()
CLIENTS_FILE="$SCRIPT_DIR/../api/clients.json"
if [ -f "$CLIENTS_FILE" ]; then
  EXTRA_ARGS+=("-e" "PURL_CLIENTS_JSON=$(tr -d '\n' < "$CLIENTS_FILE")")
fi

docker compose \
  -f "$SCRIPT_DIR/docker-compose.prod.yml" \
  run --rm \
  "${EXTRA_ARGS[@]}" \
  api \
  ./bin/"$CMD" "$@"
