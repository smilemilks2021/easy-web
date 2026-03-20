# easy-web 版本审计日志

记录每个版本的变更摘要、代码审查结论及质量验证结果。

---

## v0.3.0 — 2026-03-20

### 版本概况

| 项目 | 详情 |
|------|------|
| 分支 | `feature/skill-gen` → `main` |
| 合并提交 | `f657605` |
| 代码审查 | 已完成（19 项问题，全部修复） |
| 构建状态 | ✅ `go build ./...` 通过 |
| 测试状态 | ✅ `go test ./...` 全部通过 |
| vet 检查 | ✅ `go vet ./...` 无警告 |

### 新增功能（11 项）

| # | 功能 | 命令 | 对标工具 |
|---|------|------|----------|
| 1 | Auto-generate Claude Code Skill | `capture` / `skill gen` | 独有 |
| 2 | 独立 Skill 生成命令 | `easy-web skill gen` | 独有 |
| 3 | Smart merge（保留手写内容） | `skill gen`（智能追加） | 独有 |
| 4 | 请求重试 | `--retry` / `--retry-delay` | httpie / curlie |
| 5 | 代理支持 | `--proxy` | curl / httpie |
| 6 | JSON 高亮 + jq 过滤 | `--output json` / `--jq` | httpie / jq |
| 7 | Form/文件上传 | `--form` / `--file` | curl / httpie |
| 8 | HAR 1.2 导出 | `--export har` | Charles / Fiddler |
| 9 | 轮询监控模式 | `easy-web watch` | watch + curl |
| 10 | 命名环境管理 | `easy-web env` | postman environments |
| 11 | 请求回放 & 工作流 | `easy-web replay` / `run` | postman collections |

### 代码审查修复记录

#### Critical（必须修复）

| # | 文件 | 问题描述 | 修复方式 | 状态 |
|---|------|---------|----------|------|
| 1 | `internal/request/client.go:83` | retry loop 中 `err == nil && resp.StatusCode < 500` 若 err≠nil 且 resp≠nil 语义模糊 | 拆分为嵌套 if | ✅ |
| 2 | `cmd/request.go:184` | `resp.Status[strings.Index(...)+1:]` 空字符串时越界 | 改用 `strings.TrimPrefix` | ✅ |
| 3 | `internal/request/client.go:47` | `http.DefaultTransport.(*http.Transport)` 不安全断言，DefaultTransport 被替换时 panic | 安全断言 + fallback `&http.Transport{}` | ✅ |

#### Important（重要问题）

| # | 文件 | 问题描述 | 修复方式 | 状态 |
|---|------|---------|----------|------|
| 4 | `cmd/watch.go:52` | `signal.Notify` 后未调用 `signal.Stop`，goroutine 泄漏 | 添加 `defer signal.Stop(sigCh)` | ✅ |
| 6 | `cmd/request.go:219` | `applyJQ` 失败静默丢弃，用户无任何反馈 | 失败时打印 stderr 警告 | ✅ |
| 7 | `internal/request/client.go:54` | proxy URL 解析失败被 `if err == nil` 吞掉 | `NewClient` 返回 `(*Client, error)` | ✅ |
| 8 | `cmd/env.go:49` | `viper.WriteConfigAs` 前未 `ReadInConfig`，覆盖已有配置键 | 先 `ReadInConfig` 再写入 | ✅ |
| 9 | `internal/skill/generator.go:126` | `existingURLs[req.URL]` 仅按 URL 去重，同 URL 不同 method 被误判重复 | 添加 `<!-- easy-web: METHOD URL -->` 结构注释，改为 method+URL 去重 | ✅ |
| 10 | `cmd/request.go` | 缺少 `--timeout` 标志，无法控制单次请求超时 | 新增 `--timeout` 标志并接入 `opts.Timeout` | ✅ |
| 11 | `cmd/run.go:62` | `cookieDomain` 与 `root.go` 中 `parseHost` 功能重复 | 删除 `cookieDomain`，直接使用 `parseHost` | ✅ |
| 12 | `cmd/request.go:57` | 本地 `--url`/`--mode` 覆盖 root persistent flag，行为与其他子命令不一致 | 删除本地标志，统一使用 persistent flag | ✅ |

#### Suggestion（建议改进）

| # | 文件 | 问题描述 | 修复方式 | 状态 |
|---|------|---------|----------|------|
| 14 | `cmd/request.go` | ANSI 颜色无论是否为 TTY 都输出，重定向时污染文件 | 新增 `isTerminal()` 检测 | ✅ |
| 15 | `cmd/har.go:153` | HAR creator version 硬编码 "0.2.0" | 改用 `appVersion`，fallback `"dev"` | ✅ |
| 17 | `cmd/skill.go:65` | `--from-file` 未实现但出现在帮助中 | `MarkHidden("from-file")` | ✅ |

### 变更统计

```
22 files changed, 2048 insertions(+), 58 deletions(-)
新增文件：cmd/env.go cmd/har.go cmd/replay.go cmd/run.go cmd/skill.go cmd/watch.go
         internal/skill/generator.go internal/skill/generator_test.go
         internal/workflow/runner.go
```

### 调用链更新

`NewClient` 签名变更 `(*Client, error)`，已更新全部 6 个调用点：

- `cmd/watch.go`
- `cmd/request.go`
- `cmd/replay.go`
- `internal/auth/multistep.go`
- `internal/workflow/runner.go`
- `internal/request/client_test.go`

---

## v0.2.1 — 2026-03-20

| 项目 | 详情 |
|------|------|
| 构建状态 | ✅ |
| 测试状态 | ✅ |

- `selfupdate` 权限拒绝时显示友好提示 `sudo easy-web selfupdate`

---

## v0.2.0 — 2026-03-20

| 项目 | 详情 |
|------|------|
| 构建状态 | ✅ |
| 测试状态 | ✅ |

- Claude Code Skill (`SKILL.md`) 正式上线
- Generate Site Skill 自动化工作流
- 富文本安装页 `docs/install.html`
- 文档全面升级（GUIDE_zh.md、README 升级指南）

---

## v0.1.0 — 2026-03-20

- 首次发布
- 5 种认证模式：auto / chromedp / browser / chrome / remote
- JWT 过期自动检测
- capture 模式
- 嵌入式 Chromium 管理
- 多步 SSO YAML 配置
- 自更新（selfupdate）
- 跨平台支持：macOS / Linux / Windows
