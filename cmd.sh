#!/usr/bin/env bash
set -euo pipefail

if grep -qi ubuntu /etc/os-release 2>/dev/null; then
  echo "error: cmd.sh is for local dev only. On prod, use deploy/cmd.sh instead." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <command> [args...]"
  exit 1
fi

CMD="$1"
shift

# If clients.json exists, pass its contents to the container via -e.
# Newlines are stripped so the JSON fits in a single env var value.
EXTRA_ARGS=()
CLIENTS_FILE="$SCRIPT_DIR/api/clients.json"
if [ -f "$CLIENTS_FILE" ]; then
  EXTRA_ARGS+=("-e" "PURL_CLIENTS_JSON=$(tr -d '\n' < "$CLIENTS_FILE")")
fi

docker compose \
  -f "$SCRIPT_DIR/docker-compose.yml" \
  run --rm \
  "${EXTRA_ARGS[@]}" \
  api \
  go run ./cmd/"$CMD"/main.go "$@"
