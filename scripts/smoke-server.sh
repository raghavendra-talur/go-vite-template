#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${1:-$REPO_ROOT/dist/go-vite-template}"

if [[ ! -x "$BINARY" ]]; then
  echo "expected executable binary at $BINARY" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
DATA_DIR="$TMP_DIR/data"
PORT="${PORT:-$((19000 + RANDOM % 1000))}"
SERVER_LOG="$TMP_DIR/server.log"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

DATA_DIR="$DATA_DIR" "$BINARY" --port "$PORT" >"$SERVER_LOG" 2>&1 &
SERVER_PID="$!"

for _ in {1..60}; do
  if curl -fsS "http://127.0.0.1:${PORT}/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.25
done

curl -fsS "http://127.0.0.1:${PORT}/health" >/dev/null
[[ -f "$DATA_DIR/adminAuth.json" ]]

TOKEN="$(DATA_DIR="$DATA_DIR" "$BINARY" admin-token --raw)"
[[ "$TOKEN" == rb_admin_* ]]

curl -fsS \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://127.0.0.1:${PORT}/api/v1/health" >/dev/null

echo "Smoke test passed."
