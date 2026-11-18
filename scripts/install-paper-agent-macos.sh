#!/bin/bash
# Opt-in 24/7 public market paper monitoring. Runs only after the owner executes
# this script locally on Mac. Does not change the SkewLab CI runner or OmniGaming.
set -euo pipefail

if [ "$(uname -s)" != "Darwin" ]; then echo "Must run on macOS" >&2; exit 1; fi
SOURCE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
command -v go >/dev/null || { echo "Go 1.23+ is required: brew install go" >&2; exit 1; }
command -v python3 >/dev/null || { echo "python3 is required (Xcode Command Line Tools)" >&2; exit 1; }

APP_HOME="$HOME/Library/Application Support/SkewLab-paper"
DATA="$APP_HOME/data"
BIN="$APP_HOME/bin/skewlab-bot"
LOGS="$HOME/Library/Logs/SkewLab-paper"
PLIST="$HOME/Library/LaunchAgents/com.twappleon.skewlab.paper.plist"
mkdir -p "$DATA" "$APP_HOME/bin" "$LOGS" "$HOME/Library/LaunchAgents"
chmod 700 "$DATA" "$APP_HOME" "$LOGS"

( cd "$SOURCE_DIR"; go build -o "$BIN.new" ./cmd/bot )
mv -f "$BIN.new" "$BIN"
chmod 700 "$BIN"

# Use plistlib to escape nonstandard home-directory characters correctly.
python3 - "$PLIST" "$BIN" "$DATA" "$LOGS" <<'PY'
import plistlib, pathlib, sys
plist, binary, data, logs = sys.argv[1:]
record = {
    'Label': 'com.twappleon.skewlab.paper',
    'ProgramArguments': [binary, '--data-dir', data],
    'RunAtLoad': True,
    'KeepAlive': True,
    'ThrottleInterval': 15,
    'EnvironmentVariables': {'BIND_ADDR':'127.0.0.1','PORT':'8181'},
    'StandardOutPath': str(pathlib.Path(logs)/'stdout.log'),
    'StandardErrorPath': str(pathlib.Path(logs)/'stderr.log'),
}
with open(plist, 'wb') as f: plistlib.dump(record, f)
PY
chmod 600 "$PLIST"
plutil -lint "$PLIST"
SERVICE="gui/$(id -u)/com.twappleon.skewlab.paper"
launchctl bootout "$SERVICE" >/dev/null 2>&1 || true
launchctl bootstrap "gui/$(id -u)" "$PLIST"
launchctl kickstart -k "$SERVICE"
echo "Installed the LOCAL-ONLY read-only paper app (no orders/wallets)."
echo "Dashboard: http://127.0.0.1:8181"
echo "Data: $DATA"
echo "Status: launchctl print $SERVICE"
echo "To stop: launchctl bootout $SERVICE"
echo "Mac must remain powered on, logged in, awake, and connected for 7-14 days."
