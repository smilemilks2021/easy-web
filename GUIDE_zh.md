# easy-web 使用指南

> **版本**: 0.2.0
> **最后更新**: 2026-03-20

easy-web 是一款命令行浏览器认证工具，能够将任何需要登录的网页转化为可自动化调用的 API。它通过捕获浏览器的认证信息（Cookies、Tokens），并缓存复用，让你在命令行环境下也能轻松访问需要登录的前端接口。

---

## 目录

1. [快速入门](#快速入门)
2. [核心概念](#核心概念)
3. [基本命令](#基本命令)
4. [认证模式详解](#认证模式详解)
5. [典型使用场景](#典型使用场景)
6. [进阶功能](#进阶功能)
7. [配置详解](#配置详解)
8. [常见问题与解决方案](#常见问题与解决方案)
9. [最佳实践](#最佳实践)
10. [与 Claude Code 集成：自动生成系统 Skill](#与-claude-code-集成自动生成系统-skill)

---

## 快速入门

### 安装

**一键安装（推荐）：**

```bash
# macOS / Linux
curl -sSL https://smilemilks2021.github.io/easy-web/install.sh | sh

# Windows（PowerShell）
irm https://smilemilks2021.github.io/easy-web/install.ps1 | iex
```

### 初始化配置

```bash
# 初始化配置（首次使用必须执行）
easy-web init

# 验证安装
easy-web version
```

### 30 秒上手

```bash
# 场景：访问一个需要登录的内部页面

# 步骤 1：获取认证（自动打开浏览器，复用已登录的 Chrome 会话）
easy-web -u "https://internal.example.com/dashboard"

# 步骤 2：使用缓存的认证发起请求
easy-web request -u "https://internal.example.com/api/data"
```

### 工作原理

```
┌─────────────────────────────────────────────────────────────────┐
│                     easy-web 工作流程                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 获取认证                                                     │
│     └─ 启动浏览器 → 复用 Chrome 登录态 → 提取 Cookies/Tokens      │
│                                                                 │
│  2. 缓存凭证                                                     │
│     └─ 保存到 ~/.easy-web/cache/<domain>.json                   │
│                                                                 │
│  3. 复用请求                                                     │
│     └─ 后续请求自动携带缓存的认证信息                              │
│                                                                 │
│  4. 过期刷新                                                     │
│     └─ JWT Token 过期时自动重新获取                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 核心概念

### 认证模式 (Mode)

| 模式 | 参数 | 说明 | 适用场景 |
|------|------|------|---------|
| **auto**（默认）| *(不填)* | 智能模式：缓存 → Chrome DB → Chromium | 日常使用，自动化脚本 |
| **chromedp** | `-m chromedp` | 无头浏览器自动化 | 需要 JavaScript 渲染的页面 |
| **browser** | `-m browser` | 打开可见浏览器 | 需要手动操作的登录流程 |
| **chrome** | `-m chrome` | 读取本地 Chrome DB | 直接提取系统 Chrome 中的 Cookies |
| **remote** | `-m remote` | 连接已运行的 Chrome | 远程调试、连接 Chrome DevTools |

### 缓存层级

easy-web 使用多层缓存策略（auto 模式按优先级依次尝试）：

1. **Tier 1 - Cookie 缓存**：`~/.easy-web/cache/` 目录（JSON 文件）
2. **Tier 2 - Chrome DB**：读取系统 Chrome Cookie 数据库（kooky 跨平台解密）
3. **Tier 3 - 无头 Chromium**：自动启动 Chromium 完成登录

### JWT 自动过期检测

easy-web 自动识别 JWT 格式的 Cookie，在 `auto` 模式下发现 JWT 过期会自动清除旧缓存并重新获取认证，无需手动干预。

---

## 基本命令

### 1. 获取认证（主命令）

```bash
# 基本用法：访问页面并缓存认证
easy-web -u "https://example.com/page"

# 指定认证模式
easy-web -m browser -u "https://example.com"        # 可见浏览器，手动登录
easy-web -m chrome  -u "https://example.com"        # 读取 Chrome DB
easy-web -m chromedp -u "https://example.com"       # 无头自动化

# 捕获 Authorization Header Token
easy-web --auth-token -u "https://example.com"

# 提取 localStorage / sessionStorage 中的 Token
easy-web --extract-token -u "https://example.com"

# 使用内置 Chromium（自动下载）
easy-web --use-embedded-chromium -u "https://example.com"

# 调试认证流程（显示详细日志）
easy-web --verbose-auth -u "https://example.com"
```

自定义 Token 参数（需先在配置文件中定义 `capture_headers`）：

```bash
# 捕获自定义 Header（例如 dw-token，需配置 capture_headers.dw-token）
easy-web --dw-token -u "https://data-platform.example.com/page"
```

### 2. 发起请求（request）

```bash
# GET 请求（自动携带缓存 Cookie）
easy-web request -u "https://example.com/api/data"

# POST 请求
easy-web request -u "https://example.com/api/create" -X POST -d '{"name":"test"}'

# PUT / DELETE
easy-web request -u "https://example.com/api/item/123" -X PUT  -d '{"status":"done"}'
easy-web request -u "https://example.com/api/item/123" -X DELETE

# 添加自定义请求头
easy-web request -u "https://example.com/api" -H "X-Custom: value" -H "Accept: application/json"
```

### 3. 录制 API（capture）

录制页面发出的所有 API 请求，非常适合发现未公开的内部接口。

```bash
# 录制所有请求（默认 10 分钟超时）
easy-web capture -u "https://example.com/app" --auto-save

# 按 URL 路径过滤
easy-web capture -u "https://example.com" -p "/api/" -p "/graphql" --auto-save

# 设置超时时间
easy-web capture -u "https://example.com" -t 5m --auto-save

# 交互式选择要保存哪些接口
easy-web capture -u "https://example.com" --interactive
```

### 4. Cookie 缓存管理（cache）

```bash
# 列出所有已缓存的域名
easy-web cache list

# 清除指定域名的缓存
easy-web cache clear -d example.com

# 清除全部缓存
easy-web cache clear --all
```

### 5. 多步骤认证（auth）

```bash
# 执行配置文件中定义的认证流程
easy-web auth --name my-sso-flow
```

### 6. Chromium 管理（chromium）

easy-web 可以下载和管理独立的内置 Chromium，无需依赖系统 Chrome。

```bash
# 下载当前平台对应的 Chromium
easy-web chromium download

# 查看已下载的版本和路径
easy-web chromium info

# 删除已下载的 Chromium
easy-web chromium clean
```

### 7. 配置管理（config）

```bash
# 用 $EDITOR 打开配置文件
easy-web config edit
```

### 8. 其他命令

```bash
# 生成 ~/.easy-web.yaml 默认配置
easy-web init

# 升级到最新版本（一条命令，无需重新安装）
easy-web selfupdate

# 查看当前版本号、平台、Git commit
easy-web version
```

**版本升级说明：**

easy-web 内置自升级功能，通过 GitHub Releases 自动下载最新二进制文件。

```bash
# 查看当前版本
easy-web version

# 升级（会自动检测是否有新版本）
easy-web selfupdate

# 升级后验证
easy-web version
```

> 如遇升级失败，也可以重新运行安装脚本覆盖安装：
> ```bash
> curl -sSL https://smilemilks2021.github.io/easy-web/install.sh | sh
> ```

---

## 认证模式详解

### Auto 模式（默认推荐）

智能模式，按优先级依次尝试多种方式，适合日常使用：

```bash
easy-web -u "https://example.com"
# 等价于
easy-web -m auto -u "https://example.com"
```

**尝试顺序**：
1. 检查本地 Cookie 缓存是否有效（含 JWT 过期检测）
2. 尝试从系统 Chrome 数据库直接读取
3. 如果都失败，启动无头 Chromium 自动完成登录

**优点**：无需指定模式，自动适配，最大程度复用已有登录态。

---

### ChromeDP 模式

使用无头 Chromium 浏览器自动完成认证：

```bash
easy-web -m chromedp -u "https://example.com"
```

**工作流程**：
1. 启动无头 Chromium
2. 导航到目标页面
3. 提取 Cookies 和 Tokens
4. 缓存认证信息供后续使用

**适用场景**：需要 JavaScript 渲染的 SPA 应用，自动化脚本。

---

### Browser 模式（可见浏览器）

打开可见浏览器，适合需要人工干预的场景：

```bash
easy-web -m browser -u "https://example.com"
```

**适用场景**：
- 验证码登录
- 手机二次验证（MFA）
- 复杂的 SSO 登录流程
- 需要手动选择账号的页面

---

### Chrome 模式

直接从本机 Chrome 数据库提取 Cookies，无需启动浏览器：

```bash
easy-web -m chrome -u "https://example.com"
```

**注意**：需要已在系统 Chrome 中登录过目标网站。支持 macOS、Linux、Windows 跨平台解密（基于 kooky）。

---

### Remote 模式

连接已运行的 Chrome 实例（通过 CDP WebSocket）：

```bash
# 先启动带调试端口的 Chrome
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222

# 然后连接
easy-web -m remote -u "https://example.com"
# 默认连接 9222 端口，可在配置文件中通过 debug_port 修改
```

---

## 典型使用场景

### 场景 1：访问内部数据平台

```bash
# 配置文件中已定义 capture_headers.dw-token 的情况下
easy-web --dw-token -u "https://data-platform.example.com/dataquery"

# 发起数据查询
easy-web request -u "https://data-platform.example.com/api/query" \
  -X POST -d '{"sql":"SELECT * FROM table LIMIT 10"}'
```

### 场景 2：分析 Spark 任务

```bash
# 1. 登录 Spark History Server
easy-web -u "https://spark-history.internal.com/history/app-123"

# 2. 查询任务列表
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/jobs"

# 3. 分析失败的 Stage
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/stages/0/0/taskList"

# 4. 输出保存到文件
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/jobs" \
  > spark-jobs.json
```

### 场景 3：批量操作工作流

```bash
# 1. 登录并缓存认证
easy-web -u "https://workbench.example.com/tasks"

# 2. 批量创建任务
for i in {1..10}; do
  easy-web request -u "https://workbench.example.com/api/tasks" \
    -X POST -d "{\"name\":\"task-$i\",\"priority\":1}"
  echo "Created task-$i"
done

# 3. 批量更新状态
easy-web request -u "https://workbench.example.com/api/tasks/batch-update" \
  -X PUT -d '{"ids":[1,2,3],"status":"done"}'
```

### 场景 4：探索未知 API 接口

```bash
# 1. 录制页面发出的所有 API 请求
easy-web capture -u "https://dashboard.example.com" -p "/api/" -t 5m --auto-save

# 2. 使用录制到的接口
easy-web request -u "https://dashboard.example.com/api/metrics"
easy-web request -u "https://dashboard.example.com/api/users"
```

### 场景 5：定时健康检查脚本

```bash
#!/bin/bash
# health-check.sh - 定时监控内部服务

set -e

ENDPOINTS=(
  "https://service-a.internal.com/health"
  "https://service-b.internal.com/health"
  "https://service-c.internal.com/health"
)

# 确保认证有效
easy-web -u "https://service-a.internal.com" > /dev/null 2>&1

for url in "${ENDPOINTS[@]}"; do
  response=$(easy-web request -u "$url" 2>/dev/null)
  if echo "$response" | grep -q '"status":"ok"'; then
    echo "✅ $url is healthy"
  else
    echo "❌ $url is DOWN!"
    # 发送告警...
  fi
done
```

### 场景 6：配合 Claude Code 使用

easy-web 与 Claude Code 结合，可以让 AI 自动访问需要登录的内部工具：

```bash
# 登录 Grafana
easy-web -u "https://grafana.internal.com/d/dashboard"

# Claude Code 中调用
easy-web request -u "https://grafana.internal.com/api/dashboards/home"
```

---

## 进阶功能

### 多步骤认证（SSO 流）

对于复杂的 SSO 认证流程，可以在配置文件中定义多步骤认证：

```yaml
# ~/.easy-web.yaml
multi_step_auth:
  my-sso-flow:
    description: "公司 SSO 登录 + Token 交换"
    steps:
      # 步骤 1：浏览器登录获取 Cookie
      - id: get_sso_cookie
        type: browser_capture
        url: "https://sso.example.com/login"
        extract:
          - source: cookie
            key: session_id
            variable: session
            final: true

      # 步骤 2：用 Cookie 换取 API Token
      - id: exchange_token
        type: http_request
        method: POST
        url: "https://api.example.com/auth/exchange"
        headers:
          Cookie: "session_id=${session}"
        extract:
          - source: json_response
            path: data.access_token
            variable: token
            final: true
```

执行：

```bash
easy-web auth --name my-sso-flow
```

### 自定义 Token 捕获

在配置文件中添加自定义 Header 捕获规则，CLI 会自动生成对应参数：

```yaml
# ~/.easy-web.yaml
capture_headers:
  # 简单格式：参数名: header名
  my-token:
    header: authorization
    cache_keys:
      - api.example.com

  # 或者简写
  dw-token:
    header: dw-token
    cache_keys:
      - data-platform.example.com
```

配置后会自动生成 `--my-token`、`--dw-token` 等 CLI 参数：

```bash
easy-web --my-token -u "https://api.example.com/page"
easy-web --dw-token -u "https://data-platform.example.com/home"
```

### 使用内置 Chromium

当系统 Chrome 不可用或版本不兼容时，可以使用内置 Chromium：

```bash
# 先下载
easy-web chromium download

# 临时使用
easy-web --use-embedded-chromium -u "https://example.com"
```

---

## 配置详解

### 配置文件位置

| 路径 | 用途 |
|------|------|
| `~/.easy-web.yaml` | 主配置文件 |
| `~/.easy-web/cache/` | Cookie 缓存（JSON 文件，按域名存储） |
| `~/.easy-web/chromium/` | 内置 Chromium 存储目录 |

### 完整配置示例

```yaml
# ~/.easy-web.yaml

# ============================================================
# 基础设置
# ============================================================

# 默认认证模式：auto | chromedp | browser | chrome | remote
mode: "auto"

# browser 模式 OAuth 回调端口
port: 8080

# remote 模式 Chrome 调试端口
debug_port: 9222

# 认证完成后是否自动关闭浏览器
auto_close: true

# ============================================================
# 自定义 Header 捕获（动态生成 CLI 参数）
# ============================================================
capture_headers:
  # 访问数据平台时捕获 dw-token
  dw-token:
    header: dw-token
    cache_keys:
      - data-platform.example.com

  # 通用 Authorization Header
  auth-token:
    header: authorization
    cache_keys:
      - api.example.com

# ============================================================
# 多步骤 SSO 认证流程
# ============================================================
multi_step_auth:
  company-sso:
    description: "公司 SSO 登录"
    steps:
      - id: login
        type: browser_capture
        url: "https://sso.example.com/login"
        extract:
          - source: cookie
            key: session_id
            variable: session
            final: true

      - id: get_api_token
        type: http_request
        method: POST
        url: "https://api.example.com/auth/token"
        headers:
          Cookie: "session_id=${session}"
        extract:
          - source: json_response
            path: data.token
            variable: token
            final: true
```

---

## 常见问题与解决方案

### 问题 1：Chrome 无法启动

**症状**：提示 Chrome 版本不兼容或无法找到 Chrome

**解决方案**：

```bash
# 使用内置 Chromium
easy-web --use-embedded-chromium -u "https://example.com"

# 或者先下载内置 Chromium
easy-web chromium download
easy-web --use-embedded-chromium -u "https://example.com"
```

### 问题 2：Cookie 已过期，请求返回 401

**症状**：请求返回 401 Unauthorized 或页面跳转到登录页

**解决方案**：

```bash
# 清除过期缓存
easy-web cache clear -d example.com

# 重新获取认证
easy-web -u "https://example.com"
```

### 问题 3：Token 提取失败

**症状**：`--auth-token` 或自定义 Token 参数没有捕获到值

**解决方案**：

```bash
# 1. 使用完整 URL（包含完整路径，确保页面能触发 API 请求）
easy-web --auth-token -u "https://example.com/dashboard/home"

# 2. 开启详细日志调试
easy-web --verbose-auth --auth-token -u "https://example.com"
```

### 问题 4：自动化模式无法通过验证码

**症状**：无头模式被网站检测为机器人

**解决方案**：

```bash
# 使用可见浏览器手动登录一次
easy-web -m browser -u "https://example.com"

# 登录成功后 Cookie 已缓存，后续可切回 auto 模式
easy-web request -u "https://example.com/api/data"
```

### 问题 5：多步骤认证失败

**症状**：`easy-web auth --name` 执行报错

**解决方案**：

```bash
# 检查配置文件语法是否正确
easy-web config edit

# 调试单个步骤
easy-web -m browser -u "https://sso.example.com/login"  # 手动测试第一步
```

### 问题 6：自升级后命令报错

```bash
# 重新初始化配置
easy-web init

# 查看当前版本
easy-web version
```

---

## 最佳实践

### 1. 首次使用检查清单

- [ ] 运行 `easy-web init` 初始化配置
- [ ] 确保已在系统 Chrome 中登录目标网站
- [ ] 测试基本功能：`easy-web -u "https://target.com"`
- [ ] 检查缓存：`easy-web cache list`

### 2. 脚本自动化建议

```bash
#!/bin/bash
# 推荐的脚本模板

set -e

TARGET_URL="https://example.com/api/data"
DOMAIN="example.com"

# 检查缓存是否存在（cache list 里能找到域名则认为有缓存）
if ! easy-web cache list | grep -q "$DOMAIN"; then
  echo "缓存不存在，重新获取认证..."
  easy-web -u "https://$DOMAIN"
fi

# 执行业务请求
easy-web request -u "$TARGET_URL"
```

### 3. 安全建议

- **不要**在配置文件中存储明文密码
- **定期**清理过期的 Cookie 缓存：`easy-web cache clear --all`
- **不要**将 `~/.easy-web/cache/` 目录提交到 Git

### 4. 调试技巧

```bash
# 显示详细认证日志
easy-web --verbose-auth -u "https://example.com"

# 强制重新获取（忽略缓存）
easy-web cache clear -d example.com && easy-web -u "https://example.com"

# 使用可见浏览器观察登录流程
easy-web -m browser -u "https://example.com"
```

---

## 附录

### A. 命令速查表

| 命令 | 说明 |
|------|------|
| `easy-web init` | 初始化配置文件 |
| `easy-web -u <URL>` | 登录获取认证（auto 模式） |
| `easy-web -m browser -u <URL>` | 可见浏览器登录 |
| `easy-web -m chrome -u <URL>` | 读取系统 Chrome Cookies |
| `easy-web request -u <URL>` | 发起 GET 请求 |
| `easy-web request -u <URL> -X POST -d '{}'` | 发起 POST 请求 |
| `easy-web capture -u <URL> --auto-save` | 录制 API 请求 |
| `easy-web cache list` | 查看所有缓存 |
| `easy-web cache clear -d <domain>` | 清除指定域名缓存 |
| `easy-web cache clear --all` | 清除全部缓存 |
| `easy-web auth --name <flow>` | 执行多步骤认证流 |
| `easy-web chromium download` | 下载内置 Chromium |
| `easy-web config edit` | 编辑配置文件 |
| `easy-web selfupdate` | 升级到最新版本 |
| `easy-web version` | 查看版本信息 |

### B. 常用参数

| 参数 | 说明 |
|------|------|
| `-u, --url` | 目标 URL |
| `-m, --mode` | 认证模式（auto/chromedp/browser/chrome/remote） |
| `-X, --method` | HTTP 方法（GET/POST/PUT/DELETE） |
| `-d, --data` | 请求体（JSON 字符串） |
| `-H, --header` | 自定义请求头（可多次使用） |
| `--auth-token` | 捕获 Authorization Header |
| `--extract-token` | 提取 localStorage/sessionStorage Token |
| `--use-embedded-chromium` | 使用内置 Chromium |
| `--verbose-auth` | 显示详细认证日志 |

### C. 平台支持

| 平台 | 架构 | 状态 |
|------|------|------|
| macOS | arm64（M1/M2/M3/M4） | ✅ |
| macOS | amd64（Intel） | ✅ |
| Linux | amd64 | ✅ |
| Linux | arm64 | ✅ |
| Windows | amd64 | ✅ |

---

## 与 Claude Code 集成：自动生成系统 Skill

easy-web 的 `capture` 命令不仅能录制接口，还能与 Claude Code 配合，将整套认证流程和 API 清单**自动封装成可复用的 Skill**。封装完成后，你只需用自然语言告诉 Claude Code "帮我在 XX 系统上查询/创建/删除..."，剩下的全部自动完成。

### 为什么需要这个？

| 没有 Skill | 有了 Skill |
|-----------|-----------|
| 每次都要手动告诉 CC 怎么认证、调哪个接口 | CC 直接知道怎么操作，说需求就行 |
| 重复粘贴 URL、Cookie、请求体 | 一次录制，永久复用 |
| 换个人就要重新交接操作流程 | Skill 文件即文档，即操作手册 |

### 完整流程：四步从录制到 Skill

**第一步：用 capture 录制目标系统**

```bash
# 打开目标系统页面，操作一遍你要自动化的功能
# CC 会记录所有发出的接口
yes N | easy-web capture -u https://your-system.example.com -p /api/ -t 5m --auto-save
```

录制完成后，认证策略和接口列表保存到 `~/.easy-web.yaml`。

**第二步：让 CC 查看录制结果**

```bash
cat ~/.easy-web.yaml
```

CC 会读取配置，分析：
- 用了哪种认证方式（Cookie / Token / SSO）
- 录制到了哪些接口（路径、方法、参数结构）

**第三步：CC 自动生成 Skill 文件**

CC 在 `~/.claude/skills/<系统名>/SKILL.md` 创建 Skill，内容示例：

```markdown
---
name: my-console
description: |
  操作 my-console 控制台。触发词："帮我在 my-console 上..."
allowed-tools: Bash(easy-web:*)
---

## 认证
easy-web -u https://my-console.example.com   # 登录一次，自动缓存

## 常用操作
\`\`\`bash
# 查询资源列表
easy-web request -u https://my-console.example.com/api/v1/resources

# 创建资源
easy-web request -u https://my-console.example.com/api/v1/resources \
  -X POST -d '{"name":"xxx"}'

# 删除资源
easy-web request -u https://my-console.example.com/api/v1/resources/123 -X DELETE
\`\`\`
```

**第四步：直接用自然语言操作**

Skill 写好后，再也不需要记命令：

```
你：帮我查询 my-console 上所有状态为 failed 的任务
CC：（调用 easy-web request，返回结果并分析）

你：把 task-456 重新提交一下
CC：（调用 easy-web request -X POST，完成操作）
```

### 适合哪些系统？

只要满足以下条件，任何系统都可以：

- 需要浏览器登录（SSO / Cookie / JWT）
- 有页面 API（前后端分离的系统几乎都有）
- 操作有规律（查询、创建、更新、删除）

**典型例子：** 内部工单系统、数据看板、Spark History Server、Grafana、内部部署平台、任何没有官方 CLI 的业务控制台。

---

> **GitHub**: https://github.com/smilemilks2021/easy-web
>
> **安装页**: https://smilemilks2021.github.io/easy-web/
>
> 如有问题或建议，欢迎提交 [Issue](https://github.com/smilemilks2021/easy-web/issues)
