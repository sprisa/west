#!/usr/bin/env bash
set -euo pipefail

# Function to clean up background processes
cleanup() {
  echo "Stopping background upload processes..."
  # Kill all background jobs
  kill $(jobs -p) 2>/dev/null || true
  exit 1
}

# Trap Ctrl+C and other termination signals
trap cleanup SIGINT SIGTERM

echo "Building Snap Release..."
goreleaser release --snapshot --clean

pids=()
for file in ./dist/west_*.snap; do
  echo "Uploading: $file"
  snapcraft upload $@ $file &
  pids+=($!)
done


wait
echo "✨ Done"
