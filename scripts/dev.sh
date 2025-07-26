#!/bin/bash

# Hot reload development script for claude-permissions using tui-hotreload
cd "$(dirname "$0")/.."

echo "Downloading Go modules..."
go mod tidy

echo "Creating tmp directory for builds..."
mkdir -p tmp

echo "Starting hot reload development with tui-hotreload..."
echo "Press Ctrl+C to stop"
echo ""

# Ensure Go bin directory is in PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Check if tui-hotreload is installed
if ! command -v tui-hotreload &> /dev/null; then
    echo "tui-hotreload not found. Installing..."
    go install github.com/jhowrez/tui-hotreload@latest
fi

# Start tui-hotreload which will watch for changes and rebuild/restart
tui-hotreload
