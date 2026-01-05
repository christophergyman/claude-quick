#!/bin/bash

# Build script for Claude Quick
# This script builds the Go project and creates a symlink in ~/.local/bin

set -e  # Exit on any error

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="claude-quick"
BINARY_PATH="${PROJECT_DIR}/${BINARY_NAME}"
INSTALL_DIR="${HOME}/.local/bin"
SYMLINK_PATH="${INSTALL_DIR}/${BINARY_NAME}"

echo "==> Running tests..."
cd "${PROJECT_DIR}"
go test ./...
echo "    All tests passed!"

echo "==> Building Claude Quick..."
go build -o "${BINARY_NAME}" .
echo "    Build successful: ${BINARY_PATH}"

echo "==> Setting up symlink..."

# Create ~/.local/bin if it doesn't exist
if [ ! -d "${INSTALL_DIR}" ]; then
    echo "    Creating directory: ${INSTALL_DIR}"
    mkdir -p "${INSTALL_DIR}"
fi

# Remove existing symlink or file if it exists
if [ -L "${SYMLINK_PATH}" ]; then
    echo "    Removing existing symlink: ${SYMLINK_PATH}"
    rm "${SYMLINK_PATH}"
elif [ -e "${SYMLINK_PATH}" ]; then
    echo "    Removing existing file: ${SYMLINK_PATH}"
    rm "${SYMLINK_PATH}"
fi

# Create the symlink
ln -s "${BINARY_PATH}" "${SYMLINK_PATH}"
echo "    Symlink created: ${SYMLINK_PATH} -> ${BINARY_PATH}"

echo ""
echo "==> Done!"
echo ""
echo "Make sure ${INSTALL_DIR} is in your PATH."
echo "You can add it by adding this line to your ~/.bashrc or ~/.zshrc:"
echo ""
echo "    export PATH=\"\${HOME}/.local/bin:\${PATH}\""
echo ""
echo "Then run 'claude-quick' from anywhere!"
