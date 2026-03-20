---
name: easy-web
description: |
  Browser authentication CLI — turn login sessions into programmable API calls.
  5 auth modes, automatic JWT expiry detection, works with any login-protected site.

  **When to use easy-web:**
  - Target page requires login; curl/WebFetch returns 401/403/302
  - Need to repeatedly call APIs on internal consoles, dashboards, or Spark UI
  - Want to discover what APIs a page is calling (capture mode)
  - Scripting batch operations on web interfaces that have no official API
  - Want to package a site's auth + APIs into a reusable Claude Code Skill

  **Trigger phrases:**
  - "This page requires login, help me fetch/query data"
  - "Automate operations on this console/workbench"
  - "Record what API calls this page makes"
  - "This Spark/Grafana/Jenkins URL needs authentication"
  - "curl returns 401/403, how do I get auth"
  - "Batch operations on this system"
  - "Turn this site into a skill / generate a skill for this system"
  - "帮我把这个系统封装成一个 Skill"

allowed-tools: Bash(easy-web:*)
user-invocable: true
---

# easy-web

Turn browser login sessions into terminal API calls. Login once, credentials are cached — subsequent requests are automatically authenticated. Like curl, but it handles login for you.

---

## 5 Auth Modes — Choose the Right One

| Mode | When to use |
|------|-------------|
| `auto` (default) | Cache → Chrome DB → headless Chromium, with JWT expiry detection |
| `-m browser` | Page has CAPTCHA / 2FA / SSO redirect requiring human interaction |
| `-m chrome` | Already logged in via Chrome — reads local Cookie DB, no browser launch |
| `-m chromedp` | Force headless Chromium, skip cache check |
| `-m remote` | Take over a running Chrome (`--remote-debugging-port=9222`) |

**auto mode 3-level fallback (most commonly used):**
```
Cache hit → use directly (checks JWT expiry)
     ↓ miss / expired
Read local Chrome Cookie DB
     ↓ miss
Launch headless Chromium to login
```

---

## Basic Usage

```bash
# Login and cache cookies (all subsequent requests to same domain use the cache)
easy-web -u https://dashboard.example.com

# Make requests with cached cookies
easy-web request -u https://dashboard.example.com/api/data
easy-web request -u https://dashboard.example.com/api/tasks -X POST -d '{"name":"foo"}'
easy-web request -u https://dashboard.example.com/api/tasks/123 -X DELETE

# Extract the Authorization header token
easy-web --auth-token -u https://example.com

# Extract token from localStorage / sessionStorage
easy-web --extract-token -u https://example.com
```

---

## Capture Mode — Discover Unknown APIs

Use when you don't know what APIs a page calls. Record first, then call directly.

```bash
# ⚠️ Claude Code MUST use this form — missing either part will block:
yes N | easy-web capture -u https://example.com/app -p /api/ -t 3m --auto-save
```

**Both components are required:**

| Component | Purpose |
|-----------|---------|
| `yes N \|` | Auto-answers the "continue capturing?" idle prompt, runs until timeout |
| `--auto-save` | Skips interactive save confirmation |

```bash
# ❌ Either missing will hang:
easy-web capture -u https://example.com -t 3m --auto-save   # missing yes N |, idle prompt blocks
yes N | easy-web capture -u https://example.com -t 3m        # missing --auto-save, save prompt blocks
```

Quick scan (stop after 10s of no requests):
```bash
yes Y | easy-web capture -u https://example.com -t 5m --auto-save
```

---

## Configuration (~/.easy-web.yaml)

### Custom Token Capture (add CLI flags for frequently used sites)

```yaml
capture_headers:
  grafana-token:               # generates --grafana-token flag
    header: authorization
    cache_keys:
      - grafana.internal.com
```

Usage:
```bash
easy-web --grafana-token -u https://grafana.internal.com/dashboard
```

### Multi-Step SSO Flow

```yaml
multi_step_auth:
  my-sso:
    description: "SSO → token exchange"
    steps:
      - id: browser_login
        type: browser_capture
        url: https://sso.example.com/login
        extract:
          - source: cookie
            key: session_id
            variable: session
            final: true
      - id: exchange
        type: http_request
        method: POST
        url: https://api.example.com/auth/token
        headers:
          Cookie: "session_id=${session}"
        extract:
          - source: json_response
            path: data.access_token
            variable: token
            final: true
```

```bash
easy-web auth --name my-sso
```

---

## Embedded Chromium (no system Chrome required)

```bash
easy-web chromium download              # download Chromium for current platform
easy-web --use-embedded-chromium -u ... # use embedded Chromium for login
easy-web chromium info                  # show version and path
easy-web chromium clean                 # remove downloaded Chromium
```

---

## Cache Management

```bash
easy-web cache list                  # list all cached domains
easy-web cache clear -d example.com  # clear a specific domain
easy-web cache clear --all           # clear all cached credentials
```

---

## Generate a Site Skill

Use `capture` to discover a site's auth and APIs, then let Claude Code package everything into a reusable Skill — so any future automation on that system is one command away.

### Full Workflow

**Step 1 — Capture auth and record API calls**

```bash
# Navigate to the target site, login, and record all API traffic
yes N | easy-web capture -u https://your-system.example.com -p /api/ -t 5m --auto-save
```

This writes the discovered auth strategy and API patterns to `~/.easy-web.yaml`.

**Step 2 — Read captured config**

After capture completes, Claude Code reads `~/.easy-web.yaml` to extract:
- The auth method used (cookie / token / SSO chain)
- Which API endpoints were called
- Request methods, headers, and payload shapes

```bash
cat ~/.easy-web.yaml
```

**Step 3 — Claude Code generates the Skill**

Claude Code creates `~/.claude/skills/<site-name>/SKILL.md` containing:

```markdown
---
name: your-system
description: |
  Automates <site-name> — handles auth and exposes common operations as commands.
  Trigger: "help me ... on <site-name>"
allowed-tools: Bash(easy-web:*)
---

# your-system Skill

## Auth
easy-web -u https://your-system.example.com   # login once, cookies cached

## Common Operations
\`\`\`bash
# List resources
easy-web request -u https://your-system.example.com/api/v1/resources

# Create
easy-web request -u https://your-system.example.com/api/v1/resources \
  -X POST -d '{"name":"value"}'

# Delete
easy-web request -u https://your-system.example.com/api/v1/resources/123 -X DELETE
\`\`\`
```

**Step 4 — Use the generated Skill**

Once saved, Claude Code picks up the Skill automatically. From then on, just say:

> "List all resources on your-system"
> "Create a new task on your-system with name=foo"

Claude Code will call the right `easy-web request` command without any manual setup.

---

### Quick Reference

| Step | Command |
|------|---------|
| 1. Capture | `yes N \| easy-web capture -u <url> -p /api/ -t 5m --auto-save` |
| 2. Review | `cat ~/.easy-web.yaml` |
| 3. Generate | Claude Code reads yaml → writes `~/.claude/skills/<name>/SKILL.md` |
| 4. Use | Tell Claude Code what you want to do on that system |

---

## Troubleshooting

```bash
# Auth failing — debug the flow
easy-web --verbose-auth -u https://example.com

# Clear stale cache and re-authenticate
easy-web cache clear -d example.com && easy-web -u https://example.com

# Update to latest version
easy-web selfupdate

# Reset config
easy-web init
```
