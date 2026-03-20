# easy-web

> Browser authentication CLI — capture login cookies, automate authenticated API calls from your terminal.
>
> 将浏览器登录态变成可编程 API 调用的 CLI 工具。

[![Release](https://img.shields.io/github/v/release/smilemilks2021/easy-web)](https://github.com/smilemilks2021/easy-web/releases)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#platform-support)
[![Install](https://img.shields.io/badge/Install-Guide-blue)](https://smilemilks2021.github.io/easy-web/install.html)

**[Install Guide](https://smilemilks2021.github.io/easy-web/install.html)** · **[中文使用指南](GUIDE_zh.md)** · **[Claude Code Skill](SKILL.md)**

---

## What is easy-web?

Many internal tools and web applications don't provide official APIs — you have to log in through a browser. **easy-web** solves this by:

1. **Capturing** your browser login session (cookies / tokens)
2. **Caching** the session locally
3. **Replaying** those credentials for any HTTP request — like `curl` but fully authenticated

Perfect for automating dashboards, internal consoles, Spark UI, Grafana, and any login-protected page.

---

## Install

### macOS / Linux

```bash
curl -sSL https://smilemilks2021.github.io/easy-web/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://smilemilks2021.github.io/easy-web/install.ps1 | iex
```

### From source

```bash
git clone https://github.com/smilemilks2021/easy-web.git
cd easy-web
go build -o easy-web .
```

### Verify

```bash
easy-web version
```

### Update

```bash
easy-web selfupdate
```

One command updates to the latest release. No re-downloading the install script needed.

---

## Quick Start

```bash
# 1. Initialize config
easy-web init

# 2. Login to a site and capture cookies (auto mode)
easy-web -u https://dashboard.example.com

# 3. Make authenticated API calls
easy-web request -u https://dashboard.example.com/api/data

# 4. Done — cookies are cached, no re-login needed next time
easy-web request -u https://dashboard.example.com/api/data   # uses cache
```

---

## Authentication Modes

| Mode | Flag | Description |
|------|------|-------------|
| **auto** (default) | *(none)* | cache → Chrome DB → headless Chromium (with JWT expiry detection) |
| **chromedp** | `-m chromedp` | Headless Chromium automation |
| **browser** | `-m browser` | Opens a visible browser for manual login |
| **chrome** | `-m chrome` | Reads from your local Chrome cookie database (kooky cross-platform decryption) |
| **remote** | `-m remote` | Connects to an already-running Chrome via CDP WebSocket |

```bash
easy-web -m browser -u https://app.example.com   # visible browser, manual login
easy-web -m chrome  -u https://app.example.com   # read from Chrome DB (no browser launch)
easy-web -m remote  -u https://app.example.com   # connect to chrome --remote-debugging-port=9222
```

---

## Commands

### Login & Capture

```bash
# Login and capture cookies (default: auto mode)
easy-web -u https://example.com

# Specify auth mode
easy-web -m chromedp -u https://example.com
easy-web -m browser  -u https://example.com

# Extract Authorization header token
easy-web --auth-token -u https://example.com

# Extract token from localStorage/sessionStorage
easy-web --extract-token -u https://example.com

# Use embedded Chromium (auto-downloaded)
easy-web --use-embedded-chromium -u https://example.com

# Debug auth flow
easy-web --verbose-auth -u https://example.com
```

### API Request

```bash
# GET request (uses cached cookies automatically)
easy-web request -u https://example.com/api/data

# POST with JSON body
easy-web request -u https://example.com/api/create -X POST -d '{"name":"test"}'

# PUT / DELETE
easy-web request -u https://example.com/api/item/123 -X PUT  -d '{"status":"done"}'
easy-web request -u https://example.com/api/item/123 -X DELETE

# Extra headers
easy-web request -u https://example.com/api/data -H "X-Custom: value"
```

### API Capture (Record Mode)

Record all API requests a page makes — perfect for discovering undocumented APIs.

```bash
# Capture all requests (10-minute timeout)
easy-web capture -u https://example.com/app -t 10m --auto-save

# Filter by URL pattern
easy-web capture -u https://example.com/app -p /api/ -p /graphql --auto-save

# Interactive selection (choose which APIs to save)
easy-web capture -u https://example.com/app --interactive
```

### Cookie Cache

```bash
easy-web cache list              # list all cached domains
easy-web cache clear -d example.com   # clear specific domain
easy-web cache clear --all       # clear all cached cookies
```

### Chromium Management

easy-web can download and manage its own embedded Chromium — no system Chrome required.

```bash
easy-web chromium download   # download Chromium for current platform
easy-web chromium info       # show downloaded version and path
easy-web chromium clean      # remove downloaded Chromium
```

### Multi-Step Auth

Configure complex SSO flows in YAML — chain browser login, token exchange, and final auth.

```bash
easy-web auth --name my-sso-flow   # run a named auth flow from config
```

### Other

```bash
easy-web init          # generate ~/.easy-web.yaml with defaults
easy-web config edit   # open config in $EDITOR
easy-web selfupdate    # update to latest version
easy-web version       # show version, platform, git commit
```

---

## Configuration

Config file: `~/.easy-web.yaml` (created by `easy-web init`)

```yaml
# Default auth mode: auto | chromedp | browser | chrome | remote
mode: "auto"

# Port for browser-mode OAuth callback
port: 8080

# Remote Chrome CDP port (for -m remote)
debug_port: 9222

# Auto-close browser after auth
auto_close: true

# Custom header capture: adds --<name> flag to CLI
capture_headers:
  my-token:
    header: authorization
    cache_keys:
      - api.example.com

# Multi-step SSO auth flows
multi_step_auth:
  my-sso-flow:
    description: "Company SSO login"
    steps:
      - id: get_sso_cookie
        type: browser_capture
        url: https://sso.example.com/login
        extract:
          - source: cookie
            key: session_id
            variable: session
            final: true

      - id: exchange_token
        type: http_request
        method: POST
        url: https://api.example.com/auth/exchange
        headers:
          Cookie: "session_id=${session}"
        extract:
          - source: json_response
            path: data.access_token
            variable: token
            final: true
```

### Custom Token Flags

The `capture_headers` section adds dynamic CLI flags. Example above adds `--my-token`:

```bash
easy-web --my-token -u https://api.example.com/page
# Captures the Authorization header and caches it for api.example.com
```

---

## Platform Support

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS | arm64 (M1/M2/M3/M4) | ✅ |
| macOS | amd64 (Intel) | ✅ |
| Linux | amd64 | ✅ |
| Linux | arm64 | ✅ |
| Windows | amd64 | ✅ |

---

## Real-World Examples

### Automate a Spark History Server

```bash
# Login once
easy-web -u https://spark-history.internal.com

# Query job details
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/jobs"

# Check failed stages
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/stages/0/0/taskList"
```

### Discover Internal Tool APIs

```bash
# Record all API calls made by the dashboard
easy-web capture -u https://internal-dashboard.company.com -p /api/ -t 5m --auto-save

# Now use the captured APIs
easy-web request -u https://internal-dashboard.company.com/api/metrics
```

### Batch Operations via Script

```bash
#!/bin/bash
# Login once
easy-web -u https://workbench.example.com

# Batch create tasks
for item in task1 task2 task3; do
  easy-web request \
    -u https://workbench.example.com/api/tasks \
    -X POST \
    -d "{\"name\": \"$item\"}"
done
```

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────────┐
│  1. AUTH: Open browser (or read Chrome DB) to capture cookies       │
│     auto → check cache → check Chrome DB → launch Chromium         │
├─────────────────────────────────────────────────────────────────────┤
│  2. CACHE: Store cookies as JSON in ~/.easy-web/cache/<domain>.json │
│     JWT cookies: auto-detected, expiry checked on next use         │
├─────────────────────────────────────────────────────────────────────┤
│  3. REQUEST: Replay cached cookies for any HTTP request             │
│     Injects Cookie header → standard net/http client               │
└─────────────────────────────────────────────────────────────────────┘
```

**Tech stack:**
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [viper](https://github.com/spf13/viper) — YAML config
- [chromedp](https://github.com/chromedp/chromedp) — Chrome DevTools Protocol automation
- [kooky](https://github.com/browserutils/kooky) — Cross-platform Chrome cookie decryption
- [go-selfupdate](https://github.com/creativeprojects/go-selfupdate) — Self-update via GitHub Releases
- `CGO_ENABLED=0` throughout — zero native dependencies, fully static binaries

---

## Building from Source

```bash
git clone https://github.com/smilemilks2021/easy-web.git
cd easy-web

# Build for current platform
go build -o easy-web .

# Run tests
go test ./...

# Cross-compile (requires goreleaser)
goreleaser build --snapshot --clean
```

---

## Changelog

### v0.3.0 — 2026-03-20

**New commands**

- **`easy-web request`** — full rewrite: `--retry` / `--retry-delay` / `--proxy` / `--timeout`, `--output json|raw|headers`, `-v` verbose, `--form` / `--file` multipart, `--jq` filter, ANSI JSON highlighting (TTY-only)
- **`easy-web watch`** — poll any URL at a configurable interval (`-i`), optional `--diff` mode to print only changed lines
- **`easy-web env`** — named environments: `add` / `use` / `list` / `rm` / `show`; active env persisted in `~/.easy-web.yaml`
- **`easy-web replay`** — replay previously captured request snapshots for a domain
- **`easy-web run <workflow.yml>`** — chain HTTP steps with variable extraction, jq assertions, and `{{var}}` substitution
- **`easy-web skill gen`** — generate or smart-merge a Claude Code Skill from captured APIs (standalone command)
- **`easy-web skill list`** — list all generated Skills under `~/.claude/skills/`

**Enhancements**

- `capture` now saves HAR 1.2 export (`--export har -o output.har`), calls `skill.Generate` automatically, and persists request snapshots for `replay`
- `request.NewClient` returns `(*Client, error)` — invalid proxy URL now surfaces immediately instead of silently failing
- Retry loop nil-response guard separated from status-code check to eliminate potential panic
- Smart merge uses structured `<!-- easy-web: METHOD URL -->` comments for method-aware dedup (POST + GET on same path handled independently)
- `viper.ReadInConfig` called before `WriteConfigAs` in `env` commands to preserve existing config keys
- HAR creator version reads `appVersion` instead of hardcoded string
- `defer signal.Stop(sigCh)` added in `watch` to prevent goroutine leak on exit
- jq filter failures now print a `[jq] warning:` to stderr instead of silently dropping output
- `--from-file` flag in `skill gen` is hidden until implemented

**Bug fixes (code review)**

- `resp.Status` slicing replaced with `strings.TrimPrefix` (no panic on empty status string)
- Removed duplicate `cookieDomain` in `run.go` — `parseHost` reused from root
- Local `--url` / `--mode` flags removed from `request` command; persistent root flags used consistently
- ANSI colour output suppressed when stdout is not a TTY (`isTerminal()` check)

### v0.2.1 — 2026-03-20

- **fix**: `selfupdate` on `/usr/local/bin` now shows a clear hint instead of a raw permission error:
  > `Permission denied — try: sudo easy-web selfupdate`

### v0.2.0 — 2026-03-20

- **Claude Code Skill** — `SKILL.md` added; easy-web is now a first-class Claude Code skill ([docs](SKILL.md))
- **Generate Site Skill workflow** — `capture` → read config → Claude Code auto-generates a reusable Skill for any login-protected system
- **Rich install page** — `docs/install.html` with OS auto-detection, terminal mockups, and macOS Gatekeeper guidance
- **Docs overhaul** — GUIDE_zh.md section 10: Claude Code integration guide; upgrade instructions added to README and GUIDE_zh.md

### v0.1.0 — 2026-03-20

- Initial release
- 5 auth modes: `auto` / `chromedp` / `browser` / `chrome` / `remote`
- JWT expiry auto-detection in auto mode
- `capture` mode for recording and replaying API calls
- Embedded Chromium management (`chromium download/info/clean`)
- Multi-step SSO auth via YAML config
- Self-update via GitHub Releases (`easy-web selfupdate`)
- Cross-platform: macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64)

---

## License

MIT — see [LICENSE](LICENSE)
