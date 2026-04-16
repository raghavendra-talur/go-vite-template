#!/usr/bin/env bash
# =============================================================================
# Template init script — interactively customizes a fresh clone of this template.
#
# Usage:
#   bash scripts/init.sh
#
# EXTENSIBILITY
# -------------
# Each customization is a "step" defined by two functions:
#
#   collect_<step>()  — Prompts the user for input and sets variables.
#   apply_<step>()    — Applies the collected values to files in the repo.
#
# To add a new step:
#   1. Define collect_<step> and apply_<step> functions below.
#   2. Append "<step>" to the STEPS array.
#
# The main loop calls collect_* for all steps first (so the user answers
# everything upfront), shows a summary, asks for confirmation, then runs
# apply_* for each step.
# =============================================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# ---- Helpers ----------------------------------------------------------------

prompt() {
  local var_name="$1" prompt_text="$2" default="$3"
  local value
  if [[ -n "$default" ]]; then
    printf "%s [%s]: " "$prompt_text" "$default" >&2
  else
    printf "%s: " "$prompt_text" >&2
  fi
  read -r value
  value="${value:-$default}"
  printf -v "$var_name" '%s' "$value"
}

# Title-case a slug: "my-cool-app" → "My Cool App"
title_case() {
  echo "$1" | tr '-' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2)} 1'
}

# Detect Go module base path from git remote (e.g., github.com/user/repo)
detect_module_path() {
  local remote
  remote="$(git remote get-url origin 2>/dev/null || true)"
  if [[ -z "$remote" ]]; then
    return
  fi
  # SSH: git@github.com:user/repo.git → github.com/user/repo
  # HTTPS: https://github.com/user/repo.git → github.com/user/repo
  remote="${remote%.git}"
  if [[ "$remote" == git@* ]]; then
    remote="${remote#git@}"
    remote="${remote/://}"
  elif [[ "$remote" == https://* ]]; then
    remote="${remote#https://}"
  fi
  echo "$remote"
}

# Portable in-place sed (macOS vs Linux)
sedi() {
  if sed --version >/dev/null 2>&1; then
    sed -i "$@"
  else
    sed -i '' "$@"
  fi
}

# Replace a placeholder in specific files.
# Usage: replace_in <placeholder> <value> <file> [<file> ...]
replace_in() {
  local placeholder="$1" value="$2"
  shift 2
  local sep
  # Pick a sed delimiter that doesn't appear in the value
  for sep in '|' '#' '%' '@'; do
    if [[ "$value" != *"$sep"* ]]; then
      break
    fi
  done
  for f in "$@"; do
    if [[ -f "$f" ]]; then
      sedi "s${sep}${placeholder}${sep}${value}${sep}g" "$f"
    fi
  done
}

# ---- Steps ------------------------------------------------------------------
# Add new steps by defining collect_<name> + apply_<name> and appending to STEPS.

STEPS=(naming agent)

# -- Step: naming -------------------------------------------------------------

collect_naming() {
  local dir_name
  dir_name="$(basename "$REPO_ROOT")"

  local default_module
  default_module="$(detect_module_path)"

  echo ""
  echo "=== Project Setup ==="
  echo ""

  prompt APP_NAME   "App name (slug, used for binary/package/data dirs)" "$dir_name"
  prompt DISPLAY_NAME "Display name (human-readable, used in UI and docs)" "$(title_case "$APP_NAME")"
  prompt MODULE_PATH  "Go module base path (without /server-go)" "${default_module:-github.com/youruser/$APP_NAME}"
  prompt DESCRIPTION  "Short description (HTML meta, one line)" "${DISPLAY_NAME} — a Go + Vite full-stack app"
}

apply_naming() {
  echo "Applying naming..."

  # __APP_NAME__ — machine slug
  replace_in "__APP_NAME__" "$APP_NAME" \
    Makefile \
    package.json \
    .env.example \
    CLAUDE.md \
    README.md \
    server-go/config/config.go \
    client/src/pages/TokenEntry.tsx \
    scripts/smoke-server.sh \
    scripts/release/package-server.sh \
    .github/workflows/ci.yml \
    .github/workflows/release.yml

  # __DISPLAY_NAME__ — human-friendly
  replace_in "__DISPLAY_NAME__" "$DISPLAY_NAME" \
    README.md \
    client/index.html \
    client/src/App.tsx \
    client/src/pages/Landing.tsx

  # __MODULE_PATH__ — Go module base (imports use __MODULE_PATH__/server-go/...)
  replace_in "__MODULE_PATH__" "$MODULE_PATH" \
    server-go/go.mod \
    server-go/main.go \
    server-go/modules/tokens/storage.go \
    server-go/modules/terminal/handler.go \
    server-go/middleware/auth.go

  # __DESCRIPTION__ — HTML meta description
  replace_in "__DESCRIPTION__" "$DESCRIPTION" \
    client/index.html

  # Regenerate package-lock.json with the new package name
  if command -v npm >/dev/null 2>&1; then
    echo "Regenerating package-lock.json..."
    npm install --package-lock-only --ignore-scripts --silent 2>/dev/null || true
  fi
}

# -- Step: agent --------------------------------------------------------------

collect_agent() {
  echo ""
  echo "=== AI Agent ==="
  echo ""

  prompt AGENT_CMD "AI agent command for the web terminal (e.g., claude, aider, goose)" "claude"
}

apply_agent() {
  echo "Configuring agent..."

  # Append AGENT_CMD to .env (create if it doesn't exist)
  if [[ -f .env ]]; then
    # Remove any existing AGENT_CMD line
    sedi '/^AGENT_CMD=/d' .env
  fi
  echo "AGENT_CMD=$AGENT_CMD" >> .env
}

# ---- Main -------------------------------------------------------------------

main() {
  # Check that placeholders still exist (i.e., init hasn't already run)
  if ! grep -q "__APP_NAME__" Makefile 2>/dev/null; then
    echo "It looks like this template has already been initialized." >&2
    echo "The __APP_NAME__ placeholder was not found in Makefile." >&2
    exit 1
  fi

  # Collect all inputs
  for step in "${STEPS[@]}"; do
    "collect_${step}"
  done

  # Summary
  echo ""
  echo "=== Summary ==="
  echo ""
  echo "  App name:      $APP_NAME"
  echo "  Display name:  $DISPLAY_NAME"
  echo "  Module path:   $MODULE_PATH"
  echo "  Description:   $DESCRIPTION"
  echo "  Agent command: $AGENT_CMD"
  echo ""

  prompt CONFIRM "Proceed? (y/n)" "y"
  if [[ "$CONFIRM" != [yY]* ]]; then
    echo "Aborted."
    exit 0
  fi

  echo ""

  # Apply all steps
  for step in "${STEPS[@]}"; do
    "apply_${step}"
  done

  # Clean up
  echo ""
  echo "Done! You can now remove this script with: rm scripts/init.sh"
  echo ""
  echo "Next steps:"
  echo "  npm install"
  echo "  make dev"
}

main "$@"
