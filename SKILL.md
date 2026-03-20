---
name: easy-web
description: |
  浏览器认证 CLI 工具，将登录态变成可编程 API 调用。
  支持 5 种认证模式，自动检测 JWT 过期，适合自动化任何需要登录的网站。

  **何时使用 easy-web：**
  - 目标页面需要登录，curl/WebFetch 返回 401/403/302
  - 需要重复调用某个内部控制台、看板、Spark UI 的接口
  - 想发现页面背后有哪些 API（录制模式）
  - 编写脚本批量操作 Web 界面（没有官方 API 的系统）

  **触发词：**
  - "这个页面需要登录，帮我获取/查询数据"
  - "帮我自动化这个工作台的操作"
  - "录制下这个页面发出了哪些接口"
  - "这个 Spark/Grafana/Jenkins URL 需要认证"
  - "curl 返回 401/403，如何获取认证"
  - "帮我批量操作这个系统"

allowed-tools: Bash(easy-web:*)
user-invocable: true
---

# easy-web

将浏览器登录态变成终端里的 API 调用。登录一次，缓存凭证，后续请求自动带上认证——就像会自动登录的 curl。

---

## 5 种认证模式，按需选择

| 模式 | 何时用 |
|------|--------|
| `auto`（默认） | 优先读缓存 → 读 Chrome DB → 启动无头 Chromium，含 JWT 过期检测 |
| `-m browser` | 页面有验证码/二步验证/SSO 跳转，需要人工操作 |
| `-m chrome` | 你已在 Chrome 里登录过，直接读本机 Cookie DB，无需启动浏览器 |
| `-m chromedp` | 强制无头 Chromium，跳过缓存检查 |
| `-m remote` | 接管已运行的 Chrome（`--remote-debugging-port=9222`） |

**auto 模式的三级 fallback（最常用）：**
```
缓存命中 → 直接使用（检测 JWT 是否过期）
     ↓ miss / expired
读本机 Chrome Cookie DB
     ↓ miss
启动无头 Chromium 登录
```

---

## 基本用法

```bash
# 登录并缓存 Cookie（之后同域名请求自动使用）
easy-web -u https://dashboard.example.com

# 用缓存 Cookie 发请求
easy-web request -u https://dashboard.example.com/api/data
easy-web request -u https://dashboard.example.com/api/tasks -X POST -d '{"name":"foo"}'
easy-web request -u https://dashboard.example.com/api/tasks/123 -X DELETE

# 提取 Authorization 头里的 Token
easy-web --auth-token -u https://example.com

# 提取 localStorage / sessionStorage 里的 Token
easy-web --extract-token -u https://example.com
```

---

## 录制 API（发现未知接口）

适合：不知道页面背后有哪些接口时，先录制再调用。

```bash
# ⚠️ Claude Code 必须这样调用，否则会阻塞：
yes N | easy-web capture -u https://example.com/app -p /api/ -t 3m --auto-save
```

**两个必须同时存在的组件：**

| 组件 | 作用 |
|------|------|
| `yes N \|` | 自动回答"继续捕获"提示，持续到超时 |
| `--auto-save` | 跳过交互式保存确认 |

```bash
# ❌ 少了任何一个都会卡住：
easy-web capture -u https://example.com -t 3m --auto-save   # 缺 yes N |，idle 提示阻塞
yes N | easy-web capture -u https://example.com -t 3m        # 缺 --auto-save，保存确认阻塞
```

快速扫描（10s 无请求就停）：
```bash
yes Y | easy-web capture -u https://example.com -t 5m --auto-save
```

---

## 配置文件 (~/.easy-web.yaml)

### 自定义 Token 抓取（为常用站点加 CLI 参数）

```yaml
capture_headers:
  grafana-token:               # 生成 --grafana-token 参数
    header: authorization
    cache_keys:
      - grafana.internal.com
```

使用：
```bash
easy-web --grafana-token -u https://grafana.internal.com/dashboard
```

### 多步骤 SSO 流程

```yaml
multi_step_auth:
  my-sso:
    description: "SSO → Token 交换"
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

## 内置 Chromium（不依赖系统 Chrome）

```bash
easy-web chromium download              # 下载当前平台 Chromium
easy-web --use-embedded-chromium -u ... # 使用内置 Chromium 登录
easy-web chromium info                  # 查看版本和路径
easy-web chromium clean                 # 删除已下载的 Chromium
```

---

## 缓存管理

```bash
easy-web cache list                  # 查看已缓存的域名
easy-web cache clear -d example.com  # 清除指定域名
easy-web cache clear --all           # 清空所有缓存
```

---

## 排错

```bash
# 认证失败，调试
easy-web --verbose-auth -u https://example.com

# 重置该域名缓存再试
easy-web cache clear -d example.com && easy-web -u https://example.com

# 升级工具
easy-web selfupdate

# 重置配置
easy-web init
```
