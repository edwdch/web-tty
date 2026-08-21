# AGENTS.md

本仓库是 Go（Gin）+ React（Vite / shadcn/ui / Tailwind）模板。默认只提供 `GET /api/ping`，可在此结构上扩展。不要预建空的 `repository/`、`service/`、`models/` 等分层目录。

## 目录

| 路径 | 职责 |
|------|------|
| `cmd/server/` | 薄入口：加载 `.env`、读配置、启动 HTTP |
| `internal/config/` | 环境变量配置（`ADDR`、`DIST_DIR`、`GIN_MODE`） |
| `internal/handler/` | HTTP handler（目前只有 ping） |
| `internal/httpserver/` | Gin 引擎、路由注册、`web/dist` SPA 托管 |
| `web/` | 前端。由官方 `shadcn init -t vite` 生成 |
| `.air.toml` | air 热重载；监控 `cmd/`、`internal/`、**`web/dist`** |
| `.env.example` | 配置示例；真实 `.env` 不入库 |

`cmd/` 只接线。新 API 写在 `internal/handler`，在 `internal/httpserver` 注册。`/api/*` 走 JSON；其余路径先找 `web/dist` 静态文件，否则回退 `index.html`。

## 命令

在仓库根目录执行。

| 命令 | 用途 |
|------|------|
| `make dev` | **仅供人类本地开发**：先构建前端，再并行 `vite build --watch` + `air` |
| `make test` | `go test ./...` |
| `make build` | 构建 `web/dist` 与 `bin/server` |
| `make web-build` | 只构建前端 |
| `make tidy` | `go mod tidy` + `pnpm --dir web install` |
| `go run ./cmd/server` | 一次性启动后端（需已有 `web/dist` 才能看页面） |
| `pnpm --dir web build` | 一次性前端生产构建 |
| `pnpm --dir web dev` | 仅 Vite HMR（`/api` 代理到 `:8080`，需另开后端） |

`make dev` 监听：

- 前端源码 → `pnpm --dir web run build:watch` 写入 `web/dist`
- Go 源码与 **`web/dist` 产物** → air 重建并重启 Gin（`:8080`）

## 禁止用 `make dev` 测试

`make dev` 是长驻进程，**agent 不得用它做实现、验证或测试**。

改用：

1. `pnpm --dir web build`
2. `go test ./...` 或 `make test`
3. 短跑 `go run ./cmd/server`，用 `curl` 打 `/api/ping` 和 `/`
4. 验证结束后关掉该进程

`make build` 同样可用。不要在后台留着 air / `make dev`。

## 配置

使用 `github.com/joho/godotenv`。启动时 `godotenv.Load()`：

- `.env` 不存在不失败
- **已存在的进程环境变量优先**，`.env` 不覆盖
- 代码默认值：`ADDR=:8080`、`DIST_DIR=web/dist`、`GIN_MODE=debug`

新增配置项时同步改 `internal/config` 和 `.env.example`。在仓库根目录运行进程，以便找到 `.env` 与 `web/dist`。

## 前端组件

必须用官方 CLI 添加 shadcn 组件，禁止手写或从 GitHub 复制组件源码：

```bash
pnpm --dir web dlx shadcn@latest add <component> -y
```

初始化时已用：

```bash
pnpm dlx shadcn@latest init -t vite -n web --no-monorepo --base radix --preset nova -y
```

不要用 `shadcn init -d/--defaults`（会变成 Next.js 模板）。包管理器固定 **pnpm**。

## 后端约定

- 新接口放 `internal/handler`，路由挂在 `internal/httpserver`
- 不要引入数据库相关依赖，除非任务明确要求
- 不要为「以后可能用到」加空包或空文件
- 测试用标准库 `httptest` + `go test`

## 验证清单

- `go test ./...` 通过
- `pnpm --dir web build` 通过
- `GET /api/ping` 返回 `{"message":"pong"}`
- `/` 能拿到前端 `index.html`（需先 `pnpm --dir web build`）

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

Gin 已带 `go-playground/validator`。`.env` 读取已使用 `github.com/joho/godotenv`。

### 前端

| 用途 | 包 |
|------|-----|
| 服务端状态 | `@tanstack/react-query` |
| 路由 | `react-router` 或 `@tanstack/react-router` |
| 校验 | `zod` |
| 表单 | `react-hook-form` + `@hookform/resolvers` |
| 客户端状态 | `zustand` |
| HTTP | 先用 `fetch`；复杂场景再 `ky` |
| Toast | `sonner`（`pnpm --dir web dlx shadcn@latest add sonner -y`） |
| 暗色 | 已有 `web/src/components/theme-provider.tsx`（按 `d` 切换） |
| 测试 | `vitest`、`@testing-library/react`、Playwright |
| 表格 | `@tanstack/react-table` + `pnpm --dir web dlx shadcn@latest add table -y` |

`clsx`、`tailwind-merge`、`lucide-react`、`class-variance-authority` 已随 shadcn 安装。
