# SkewLab GitHub Actions on Mac (ARM64)

SkewLab uses its own dedicated repository-level self-hosted runner on the same Mac as OmniGaming.
Existing OmniGaming at ~/actions-runner remains unchanged. SkewLab installs at ~/github-runners/skewlab.
SkewLab runner labels: self-hosted, macOS, ARM64, skewlab-mac.

## One-time installation (run on your Mac Terminal)

1. If GitHub CLI is not installed: brew install gh
2. Sign in locally: gh auth login -h github.com -s repo
3. Install Xcode Command Line Tools if needed: xcode-select --install
4. Download code: gh repo clone twappleon/SkewLab "$HOME/github-runners/skewlab-source"
5. Run: bash "$HOME/github-runners/skewlab-source/scripts/install-macos-runner.sh"

The installer fetches a short-lived registration token on your Mac via gh api; never paste tokens in ChatGPT. It checks the official runner's release SHA-256, creates the distinct runner, and installs a launchd user service. If GitHub CLI lacks permission (HTTP 403), use repository Settings > Actions > Runners > New self-hosted runner and the official manual installer instructions.

## Verify / manage

Open https://github.com/twappleon/SkewLab/settings/actions/runners . The new runner should be Online / Idle.
Open https://github.com/twappleon/SkewLab/actions and manually run the workflow once after registration.
Check service: cd "$HOME/github-runners/skewlab" && ./svc.sh status
Restart stopped service: cd "$HOME/github-runners/skewlab" && ./svc.sh start
If launchd fails: cd "$HOME/github-runners/skewlab" && ./svc.sh stop && ./run.sh to see foreground errors.

GitHub Actions triggers only trusted pushes to main and workflow_dispatch. It deliberately does not run untrusted pull requests on your personal Mac. Mac must remain powered on, awake and connected to GitHub. The Go CI does not require Docker, but go test -race requires Xcode Command Line Tools. No wallet keys or trading secrets should be installed on the runner.

Official docs: https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/add-runners
