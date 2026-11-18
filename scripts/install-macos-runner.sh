#!/usr/bin/env bash
# Dedicated Apple Silicon runner for twappleon/SkewLab. Never touches OmniGaming.
set -Eeuo pipefail

REPO="twappleon/SkewLab"
RUNNER_LABEL="skewlab-mac"
RUNNER_DIR="${HOME}/github-runners/skewlab"

fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
note() { printf '\n==> %s\n' "$*"; }

[[ "$(uname -s)" == "Darwin" ]] || fail "Run this script on your Mac."
[[ "$(uname -m)" == "arm64" ]] || fail "This installer is for your ARM64 Mac."
command -v gh >/dev/null || fail "Install GitHub CLI: brew install gh; then gh auth login."
command -v curl >/dev/null || fail "curl is required."
command -v shasum >/dev/null || fail "shasum is required."
command -v xcode-select >/dev/null || fail "Xcode Command Line Tools are required."
xcode-select -p >/dev/null 2>&1 || fail "Run xcode-select --install, complete installation, then retry."
gh auth status -h github.com >/dev/null 2>&1 || fail "Authenticate locally: gh auth login -h github.com -s repo"

HOST_SHORT="$(hostname -s | tr -cd '[:alnum:]_-')"
RUNNER_NAME="skewlab-${HOST_SHORT}-arm64"
mkdir -p "$RUNNER_DIR"
cd "$RUNNER_DIR"

if [[ ! -f .runner ]]; then
  if [[ -e config.sh || -e run.sh ]]; then
    fail "Incomplete install in $RUNNER_DIR. Inspect it rather than overwriting."
  fi

  note "Downloading and verifying the latest official actions/runner for macOS ARM64"
  RELEASE_ASSET="$(gh api repos/actions/runner/releases/latest --jq '.assets[] | select(.name | startswith("actions-runner-osx-arm64-") and endswith(".tar.gz")) | [.browser_download_url, .digest] | @tsv' | head -n 1)"
  [[ -n "$RELEASE_ASSET" ]] || fail "Release unavailable. Use GitHub Settings > Actions > Runners > New self-hosted runner."
  IFS=$'\t' read -r DOWNLOAD_URL DIGEST <<< "$RELEASE_ASSET"
  [[ "$DIGEST" == sha256:* ]] || fail "Official SHA-256 unavailable. Refusing unverified download; use GitHub manual setup."
  EXPECTED_SHA="${DIGEST#sha256:}"
  ARCHIVE="$(basename "$DOWNLOAD_URL")"
  curl --fail --location --retry 3 --output "$ARCHIVE" "$DOWNLOAD_URL"
  ACTUAL_SHA="$(shasum -a 256 "$ARCHIVE" | awk '{print $1}')"
  [[ "$ACTUAL_SHA" == "$EXPECTED_SHA" ]] || fail "SHA-256 mismatch: refusing to install."
  tar -xzf "$ARCHIVE"
  rm -f "$ARCHIVE"

  note "Registering on your Mac; the one-hour GitHub token will not be printed"
  REGISTRATION_TOKEN="$(gh api --method POST "repos/${REPO}/actions/runners/registration-token" --jq '.token')" || fail "Token request failed. Check repository Admin permission."
  [[ -n "$REGISTRATION_TOKEN" ]] || fail "GitHub returned an empty token."
  ./config.sh --unattended \
    --url "https://github.com/${REPO}" \
    --token "$REGISTRATION_TOKEN" \
    --name "$RUNNER_NAME" \
    --labels "$RUNNER_LABEL" \
    --work _work
  unset REGISTRATION_TOKEN
else
  note "Preserving existing runner registration at $RUNNER_DIR"
  grep -q "SkewLab" .runner || fail "This directory is registered to a different repo. Inspect .runner locally."
fi

note "Start SkewLab's dedicated macOS launchd service"
if [[ -f .service ]]; then
  ./svc.sh status || ./svc.sh start
else
  ./svc.sh install
  ./svc.sh start
fi
./svc.sh status
printf '\nRunner directory: %s\nRunner label: %s\n' "$RUNNER_DIR" "$RUNNER_LABEL"
printf 'Verify: https://github.com/%s/settings/actions/runners\n' "$REPO"
printf 'Actions: https://github.com/%s/actions\n' "$REPO"
printf 'Keep this repository private; workflows run with your Mac user privileges.\n'
