# easy-web

> 浏览器认证 CLI 工具 — 捕获登录 Cookie，在终端自动化执行需要认证的 API 调用。

[![Release](https://img.shields.io/github/v/release/smilemilks2021/easy-web)](https://github.com/smilemilks2021/easy-web/releases)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#平台支持)

---

## 它是什么？

很多内部系统和 Web 应用没有提供官方 API，必须通过浏览器登录才能操作。**easy-web** 解决了这个问题：

1. **捕获** 浏览器登录会话（Cookie / Token）
2. **缓存** 会话到本地
3. **回放** 这些凭证发起任意 HTTP 请求 — 相当于自带完整认证的 `curl`

非常适合自动化各类内部控制台、数据看板、Spark UI、Grafana 以及任何需要登录的页面。

---

## 安装

### macOS / Linux

```bash
curl -sSL https://smilemilks2021.github.io/easy-web/install.sh | sh
```

### Windows（PowerShell）

```powershell
irm https://smilemilks2021.github.io/easy-web/install.ps1 | iex
```

### 从源码编译

```bash
git clone https://github.com/smilemilks2021/easy-web.git
cd easy-web
go build -o easy-web .
```

### 验证安装

```bash
easy-web version
```

---

## 快速上手

```bash
# 1. 初始化配置文件
easy-web init

# 2. 登录网站并捕获 Cookie（auto 模式）
easy-web -u https://dashboard.example.com

# 3. 发起认证 API 请求
easy-web request -u https://dashboard.example.com/api/data

# 4. Cookie 已缓存，下次无需重新登录
easy-web request -u https://dashboard.example.com/api/data   # 直接使用缓存
```

---

## 五种认证模式

| 模式 | 参数 | 说明 |
|------|------|------|
| **auto**（默认） | *(不填)* | 缓存 → Chrome DB → 无头 Chromium（含 JWT 过期检测） |
| **chromedp** | `-m chromedp` | 无头 Chromium 自动化 |
| **browser** | `-m browser` | 打开可见浏览器，手动登录 |
| **chrome** | `-m chrome` | 读取本机 Chrome Cookie 数据库（kooky 跨平台解密） |
| **remote** | `-m remote` | 通过 CDP WebSocket 连接已运行的 Chrome |

```bash
easy-web -m browser -u https://app.example.com   # 可见浏览器，手动登录
easy-web -m chrome  -u https://app.example.com   # 读取 Chrome DB，无需启动浏览器
easy-web -m remote  -u https://app.example.com   # 连接 chrome --remote-debugging-port=9222
```

---

## 命令参考

### 登录与捕获

```bash
# 登录并捕获 Cookie（默认 auto 模式）
easy-web -u https://example.com

# 指定认证模式
easy-web -m chromedp -u https://example.com
easy-web -m browser  -u https://example.com

# 提取 Authorization 请求头 Token
easy-web --auth-token -u https://example.com

# 提取 localStorage / sessionStorage 中的 Token
easy-web --extract-token -u https://example.com

# 使用内置 Chromium（自动下载）
easy-web --use-embedded-chromium -u https://example.com

# 调试认证流程
easy-web --verbose-auth -u https://example.com
```

### API 请求

```bash
# GET 请求（自动使用缓存 Cookie）
easy-web request -u https://example.com/api/data

# POST 带 JSON Body
easy-web request -u https://example.com/api/create -X POST -d '{"name":"test"}'

# PUT / DELETE
easy-web request -u https://example.com/api/item/123 -X PUT  -d '{"status":"done"}'
easy-web request -u https://example.com/api/item/123 -X DELETE

# 添加自定义请求头
easy-web request -u https://example.com/api/data -H "X-Custom: value"
```

### API 录制模式

录制页面发出的所有 API 请求 — 非常适合发现未公开的内部接口。

```bash
# 录制所有请求（10 分钟超时）
easy-web capture -u https://example.com/app -t 10m --auto-save

# 按 URL 路径过滤
easy-web capture -u https://example.com/app -p /api/ -p /graphql --auto-save

# 交互式选择要保存哪些接口
easy-web capture -u https://example.com/app --interactive
```

### Cookie 缓存管理

```bash
easy-web cache list                    # 列出所有已缓存的域名
easy-web cache clear -d example.com    # 清除指定域名的缓存
easy-web cache clear --all             # 清除全部缓存
```

### Chromium 管理

easy-web 可以下载和管理独立的内置 Chromium，无需依赖系统 Chrome。

```bash
easy-web chromium download   # 下载当前平台对应的 Chromium
easy-web chromium info       # 查看已下载的版本和路径
easy-web chromium clean      # 删除已下载的 Chromium
```

### 多步骤认证（SSO 流）

通过 YAML 配置复杂的 SSO 认证流程 — 支持浏览器登录、Token 交换、最终鉴权的链式操作。

```bash
easy-web auth --name my-sso-flow   # 执行配置中定义的认证流程
```

### 其他

```bash
easy-web init          # 生成 ~/.easy-web.yaml 默认配置
easy-web config edit   # 用 $EDITOR 打开配置文件
easy-web selfupdate    # 更新到最新版本
easy-web version       # 显示版本号、平台、Git commit
```

---

## 配置文件

配置文件路径：`~/.easy-web.yaml`（通过 `easy-web init` 生成）

```yaml
# 默认认证模式：auto | chromedp | browser | chrome | remote
mode: "auto"

# browser 模式 OAuth 回调端口
port: 8080

# remote 模式 CDP 端口（chrome --remote-debugging-port）
debug_port: 9222

# 认证后是否自动关闭浏览器
auto_close: true

# 自定义请求头捕获，会动态添加 --<name> CLI 参数
capture_headers:
  my-token:
    header: authorization
    cache_keys:
      - api.example.com

# 多步骤 SSO 认证流程
multi_step_auth:
  my-sso-flow:
    description: "公司 SSO 登录"
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

### 自定义 Token 参数

`capture_headers` 配置会动态生成 CLI 参数。上例会添加 `--my-token` 参数：

```bash
easy-web --my-token -u https://api.example.com/page
# 捕获 Authorization 请求头并缓存到 api.example.com
```

---

## 平台支持

| 平台 | 架构 | 状态 |
|------|------|------|
| macOS | arm64（M1/M2/M3/M4） | ✅ |
| macOS | amd64（Intel） | ✅ |
| Linux | amd64 | ✅ |
| Linux | arm64 | ✅ |
| Windows | amd64 | ✅ |

---

## 实际使用场景

### 自动化 Spark History Server

```bash
# 登录一次
easy-web -u https://spark-history.internal.com

# 查询任务详情
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/jobs"

# 检查失败的 Stage
easy-web request -u "https://spark-history.internal.com/api/v1/applications/app-123/stages/0/0/taskList"
```

### 发现内部工具的隐藏接口

```bash
# 录制控制台发出的所有 API 请求
easy-web capture -u https://internal-dashboard.company.com -p /api/ -t 5m --auto-save

# 直接调用发现的接口
easy-web request -u https://internal-dashboard.company.com/api/metrics
```

### 脚本批量操作

```bash
#!/bin/bash
# 登录一次
easy-web -u https://workbench.example.com

# 批量创建任务
for item in task1 task2 task3; do
  easy-web request \
    -u https://workbench.example.com/api/tasks \
    -X POST \
    -d "{\"name\": \"$item\"}"
done
```

---

## 工作原理

```
┌─────────────────────────────────────────────────────────────────────┐
│  1. 认证：打开浏览器（或读取 Chrome DB）捕获 Cookie                  │
│     auto → 检查缓存 → 检查 Chrome DB → 启动 Chromium               │
├─────────────────────────────────────────────────────────────────────┤
│  2. 缓存：将 Cookie 以 JSON 格式存储到 ~/.easy-web/cache/<域名>.json │
│     JWT Cookie：自动识别，下次使用时检测过期状态                     │
├─────────────────────────────────────────────────────────────────────┤
│  3. 请求：将缓存的 Cookie 注入任意 HTTP 请求                         │
│     注入 Cookie 请求头 → 标准 net/http 客户端发起请求               │
└─────────────────────────────────────────────────────────────────────┘
```

**技术栈：**
- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [viper](https://github.com/spf13/viper) — YAML 配置解析
- [chromedp](https://github.com/chromedp/chromedp) — Chrome DevTools Protocol 自动化
- [kooky](https://github.com/browserutils/kooky) — 跨平台 Chrome Cookie 解密
- [go-selfupdate](https://github.com/creativeprojects/go-selfupdate) — 通过 GitHub Releases 自动更新
- 全程 `CGO_ENABLED=0` — 无原生依赖，完全静态二进制

---

## 从源码构建

```bash
git clone https://github.com/smilemilks2021/easy-web.git
cd easy-web

# 构建当前平台
go build -o easy-web .

# 运行测试
go test ./...

# 跨平台编译（需要 goreleaser）
goreleaser build --snapshot --clean
```

---

## 开源协议

MIT — 详见 [LICENSE](LICENSE)
