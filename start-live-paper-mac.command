#!/bin/bash
set -e
cd "$(dirname "$0")"
if ! command -v go >/dev/null 2>&1; then
  echo 'Please install Go 1.23 or newer: https://go.dev/dl/'
  read -r -p 'Press Enter to exit...'
  exit 1
fi
printf '\nSkewLab - REAL PUBLIC MARKET DATA, PAPER TRADING ONLY.\n'
(go run ./cmd/bot) &
PID=$!
trap 'kill "$PID" 2>/dev/null || true' EXIT
sleep 3
open http://localhost:8080
wait "$PID"
