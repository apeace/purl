#!/usr/bin/env bash
set -euo pipefail

if ! grep -qi ubuntu /etc/os-release 2>/dev/null; then
  echo "error: d.sh is for prod only. You appear to be running locally." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

docker compose \
  -f "$SCRIPT_DIR/docker-compose.prod.yml" \
  "$@"
