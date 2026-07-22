#!/usr/bin/env bash
set -euo pipefail

# Install the gcsl VS Code extension from the local checkout.
# Run this from the repo root:  bash editors/vscode/install.sh

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# Detect VS Code variant
VSCODE_HOME="${VSCODE_HOME:-}"
if [ -z "$VSCODE_HOME" ]; then
  if [ -d "$HOME/.vscode-oss/extensions" ]; then
    VSCODE_HOME="$HOME/.vscode-oss"
  elif [ -d "$HOME/.vscode-server/extensions" ]; then
    VSCODE_HOME="$HOME/.vscode-server"
  elif [ -d "$HOME/.vscode/extensions" ]; then
    VSCODE_HOME="$HOME/.vscode"
  else
    echo "could not find VS Code extensions directory"
    echo "try: VSCODE_HOME=\$HOME/.vscode $0"
    exit 1
  fi
fi

EXT_DIR="$VSCODE_HOME/extensions/gcsim.gcsl-0.1.0"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"

# Step 1: regenerate extension JSON files from Go sources
echo "--- regenerating extension files from Go sources ---"
(cd "$REPO_ROOT" && go run ./cmd/gcslgen/)

# Step 2: npm install the LSP client dependency
echo "--- installing npm dependencies ---"
npm install --prefix "$SRC_DIR" --omit=dev

# Step 3: build and bundle gcsls for the current platform
echo "--- building bundled gcsls ---"
ext=""
[ "$(uname -s)" = "MINGW"* ] || [ "$(uname -s)" = "MSYS"* ] && ext=".exe"
go build -ldflags="-s -w" -o "$SRC_DIR/server/gcsls${ext}" "$REPO_ROOT/cmd/gcsls/"
echo "bundled: server/gcsls${ext}"

# Step 4: symlink extension into VS Code
echo "--- installing extension ---"
if [ -L "$EXT_DIR" ]; then
  rm "$EXT_DIR"
elif [ -d "$EXT_DIR" ]; then
  rm -rf "$EXT_DIR"
fi

ln -s "$SRC_DIR" "$EXT_DIR"
echo "installed: $EXT_DIR -> $SRC_DIR"
echo ""
echo "Done. Reload VS Code (Ctrl+Shift+P → Developer: Reload Window)."
echo "Open a .gcsl file — you should see diagnostics, completions, and hover docs from gcsls."
