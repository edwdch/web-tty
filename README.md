# simple-app

Go (Gin) + React (Vite, shadcn/ui, Tailwind CSS) 模板。后端提供 `GET /api/ping`，前端构建产物由 Gin 托管。

## 要求

- Go 1.26+
- Node.js 与 [pnpm](https://pnpm.io)
- GNU Make

## 启动

```bash
cp .env.example .env   # 可选，代码里已有同样的默认值
make dev
```

浏览器打开 http://127.0.0.1:8080 。`make dev` 会：

1. 构建 `web/dist`
2. 并行监控前端源码（`vite build --watch`）和后端（air，含 `web/dist`）

配置见 `.env.example`：`ADDR`、`DIST_DIR`、`GIN_MODE`。`.env` 不覆盖已有环境变量。

## 常用命令

```bash
make test          # go test ./...
make build         # 前端 + bin/server
make web-build     # 只构建前端
go run ./cmd/server
pnpm --dir web dev # 仅 Vite；/api 代理到 :8080
```

## 加 UI 组件

```bash
pnpm --dir web dlx shadcn@latest add button -y
```

给 agent 的完整约定见 [AGENTS.md](./AGENTS.md)。
