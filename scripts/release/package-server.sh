#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 5 ]]; then
  echo "usage: scripts/release/package-server.sh <binary> <os> <arch> <version> <out-dir>" >&2
  exit 1
fi

BINARY_PATH="$1"
TARGET_OS="$2"
TARGET_ARCH="$3"
VERSION="$4"
OUT_DIR="$5"

if [[ ! -x "$BINARY_PATH" ]]; then
  echo "binary not found or not executable: $BINARY_PATH" >&2
  exit 1
fi

# Change this to your app name
APP_NAME="go-vite-template"
ARCHIVE_BASENAME="${APP_NAME}-server-${TARGET_OS}-${TARGET_ARCH}"
STAGE_DIR="$OUT_DIR/$ARCHIVE_BASENAME"

rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR/bin"

cp "$BINARY_PATH" "$STAGE_DIR/bin/$APP_NAME"
chmod +x "$STAGE_DIR/bin/$APP_NAME"

cat > "$STAGE_DIR/VERSION" <<EOF
${VERSION}
EOF

(
  cd "$OUT_DIR"
  tar -czf "${ARCHIVE_BASENAME}.tar.gz" "$ARCHIVE_BASENAME"
)

echo "Packaged: $OUT_DIR/${ARCHIVE_BASENAME}.tar.gz"
