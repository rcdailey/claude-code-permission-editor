#!/bin/bash

# Quick test script for claude-permissions
cd "$(dirname "$0")/.."

echo "Downloading Go modules..."
go mod tidy

echo "Building claude-permissions..."
go build -o claude-permissions .

echo "Running with test data in debug mode..."
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
