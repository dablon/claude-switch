#!/bin/bash
set -e

echo "Installing claude-switch..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go build -o claude-switch ./cmd/claude-switch

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
mv claude-switch "$INSTALL_DIR/"

echo "Installed to $INSTALL_DIR/claude-switch"
echo "Add $INSTALL_DIR to your PATH if not already added"
