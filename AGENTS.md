# AGENTS.md

本仓库是浏览器 Web TTY：Go（Gin）+ React（Vite / shadcn/ui / Tailwind）+ `@wterm/ghostty`。每个浏览器 tab 一条 WebSocket、一个 PTY（ttyd 同款），进页即全屏终端。不要预建空的 `repository/`、`service/`、`models/` 等分层目录。不要加鉴权、Docker、应用内会话列表或 tmux 持久化。

## 目录

| 路径 | 职责 |
|------|------|
| `cmd/server/` | 薄入口：加载 `.env`、读配置、启动 HTTP |
| `internal/config/` | 环境变量配置 |
| `internal/handler/` | HTTP / WebSocket handler（`ping`、`/ws`） |
| `internal/session/` | 二进制协议、PTY（`creack/pty`）、会话上限 |
| `internal/httpserver/` | Gin 引擎、路由注册、embed `web/dist` 托管 SPA；`.wasm` 的 `Content-Type` 为 `application/wasm` |
| `web/` | 前端。全屏 `@wterm/react` + Ghostty 核。`web/fs.go` 用 `//go:embed all:dist` 把构建产物打进二进制 |
| `.air.toml` | air 热重载；监控 `cmd/`、`internal/`、`web/`（含 embed 的 `web/dist`） |
| `.env.example` | 配置示例；真实 `.env` 不入库 |

`cmd/` 只接线。新 API 写在 `internal/handler`，在 `internal/httpserver` 注册。`/api/*` 走 JSON；`GET /ws` 升级为终端；其余路径从 embed 的 `web/dist` 取静态文件，否则回退 `index.html`。没有 `DIST_DIR`，不允许自定义前端目录。

一 WS 一 PTY：多 tab 互不影响；关 tab / 刷新 / 断线杀进程。重连是新 PTY，不恢复旧会话。

## 命令

在仓库根目录执行。

| 命令 | 用途 |
|------|------|
| `make dev` | **仅供人类本地开发**：先构建前端，再并行 `vite build --watch` + `air` |
| `make test` | `go test ./...` |
| `make build` | 构建前端并 embed 进 `bin/server` |
| `make release-build` | 同上，`CGO_ENABLED=0` + 剥离符号，供发版 |
| `make web-build` | 只构建前端（并补 `web/dist/.gitkeep`，保证纯 `go test` 能编过 embed） |
| `make tidy` | `go mod tidy` + `pnpm --dir web install` |
| `go run ./cmd/server` | 一次性启动后端（须先有 `web/dist`：embed 是编译期依赖） |
| `pnpm --dir web build` | 一次性前端生产构建 |
| `pnpm --dir web dev` | 仅 Vite HMR（`/api`、`/ws` 代理到 `:8080`，需另开后端） |

`make dev` 监听：

- 前端源码 → `pnpm --dir web run build:watch` 写入 `web/dist`
- Go 源码与 **`web/dist` 产物** → air 重建 Go（重新 embed）并重启 Gin（`:8080`）

## 验证

本地开发时 `make dev` 通常已在运行。前端写入 `web/dist`、air 重建 Go（重新 embed）并重启，改动会实时生效，**直接对已有进程验证**（默认 `:8080`），不要再短跑 `go run ./cmd/server`。

`make dev` 是长驻进程，由人维护。**agent 不要自己启动** `make dev` / air，也不要在后台再开一份。**也不要用浏览器工具点页面**；UI / TUI 由人本地验证。

验证用：

1. `pnpm --dir web build`
2. `go test ./...` 或 `make test`
3. 若本机已有服务，**不要**再 `go run` / 不要停掉它，直接 `curl` `/api/ping` 和 `/`
4. 否则才短跑 `go run ./cmd/server`，打完 `/api/ping` 和 `/` 后关掉该进程

一次性生产构建用 `pnpm --dir web build` 或 `make build`。

## 配置

使用 `github.com/joho/godotenv`。启动时 `godotenv.Load()`：

- `.env` 不存在不失败
- **已存在的进程环境变量优先**，`.env` 不覆盖
- 代码默认值：`ADDR=:8080`、`GIN_MODE=debug`、`SHELL=/bin/bash`、`SHELL_ARGS=-l`、`WRITABLE=true`、`MAX_SESSIONS=8`
- `CWD` 空则继承进程 cwd
- `ALLOW_ORIGIN` 空则 WebSocket 仅同源；Vite `:5173` 时再填 `http://127.0.0.1:5173`

新增配置项时同步改 `internal/config` 和 `.env.example`。在仓库根目录运行进程，以便找到 `.env`。

本仓库不做登录。`CheckOrigin` 只防跨站乱连 `/ws`。握手 JSON 里的 `cmd` / `cwd` / `token` 一律忽略。

推送 `v*` tag 会跑 `.github/workflows/release.yml`：测过后编前端、打 linux/amd64 单二进制，挂到 GitHub Release（`web-tty_<tag>_linux_amd64`）。

## 前端

页面只有全屏终端，无顶栏、无 ping 卡片。断线弹出 shadcn `AlertDialog`：Reconnect（`location.reload()`，新 PTY）或 Close page（`window.close()`，失败则 `about:blank`）。不要自动重连。不要恢复 `d` 切主题快捷键。

Ghostty wasm 由 Vite 通过 `import.meta.url` 打进 `web/dist/assets/*.wasm`，不要手抄进 git。

必须用官方 CLI 添加 shadcn 组件，禁止手写或从 GitHub 复制组件源码：

```bash
pnpm --dir web exec shadcn add <component> -y
```

（`dlx shadcn@latest` 若因上游 CLI 失败，用仓库里 pinned 的 `shadcn`。）

初始化时已用：

```bash
pnpm dlx shadcn@latest init -t vite -n web --no-monorepo --base radix --preset nova -y
```

不要用 `shadcn init -d/--defaults`（会变成 Next.js 模板）。包管理器固定 **pnpm**。

## 后端约定

- 新接口放 `internal/handler`，路由挂在 `internal/httpserver`
- PTY / 会话生命周期在 `internal/session`；handler 注入接口以便单测
- 不要引入数据库相关依赖，除非任务明确要求
- 不要为「以后可能用到」加空包或空文件（尤其不要 `auth/`）
- 测试用标准库 `httptest` + `go test`；PTY 测试在非 unix 上应 skip

## 验证清单

- `go test ./...` 通过
- `pnpm --dir web build` 通过
- `GET /api/ping` 返回 `{"message":"pong"}`
- `/` 能拿到 embed 的前端 `index.html`
- `.wasm` 响应 `Content-Type: application/wasm`（见 `httpserver` 测试）

## 推荐依赖（尚未安装）

扩展时按需添加，不要一次性全装。

### 后端（无数据库）

| 用途 | 包 |
|------|-----|
| CORS | `github.com/gin-contrib/cors` |
| 请求 ID | `github.com/gin-contrib/requestid` |
| 结构化 env 映射 | `github.com/caarlos0/env/v11`（配置变复杂时再加；当前是 godotenv + 手读） |
| 结构化日志 | 标准库 `log/slog`；需要采样/动态级别再用 `go.uber.org/zap` |
| 测试断言 | `github.com/stretchr/testify` |
| 限流 / 超时 | `github.com/gin-contrib/timeout`、`golang.org/x/time/rate` |
| 并发 | `golang.org/x/sync`（errgroup） |

已用：`github.com/gin-gonic/gin`、`github.com/joho/godotenv`、`github.com/gorilla/websocket`、`github.com/creack/pty`。Gin 已带 `go-playground/validator`。

### 前端

| 用途 | 包 |
|------|-----|
| 服务端状态 | `@tanstack/react-query` |
| 路由 | `react-router` 或 `@tanstack/react-router` |
| 校验 | `zod` |
| 表单 | `react-hook-form` + `@hookform/resolvers` |
| 客户端状态 | `zustand` |
| HTTP | 先用 `fetch`；复杂场景再 `ky` |
| Toast | `sonner`（`pnpm --dir web exec shadcn add sonner -y`） |
| 暗色 | `theme-provider.tsx`（无快捷键；`theme-toggle.tsx` 未挂到页面） |
| 测试 | `vitest`、`@testing-library/react`、Playwright |
| 表格 | `@tanstack/react-table` + `pnpm --dir web exec shadcn add table -y` |

`clsx`、`tailwind-merge`、`lucide-react`、`class-variance-authority` 已随 shadcn 安装。已装：button、dropdown-menu、alert-dialog、skeleton、`@wterm/dom`、`@wterm/react`、`@wterm/ghostty`。
