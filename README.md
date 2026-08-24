# web-tty

浏览器 Web TTY：Go（Gin）+ React（Vite, shadcn/ui, Tailwind CSS）+ wterm/Ghostty。每个浏览器 tab 一条 WebSocket、一个真 PTY（对齐 ttyd）。前端构建产物在编译期 embed 进二进制，Gin 从 embed FS 托管。

本仓库不做登录（交给前面的反向代理）也不做 Docker；交付是单二进制。

## 要求

- Go 1.26+
- Node.js 与 [pnpm](https://pnpm.io)
- GNU Make
- Unix PTY（发版目标 linux/amd64）

## 启动

```bash
cp .env.example .env   # 可选，代码里已有同样的默认值
make dev
```

浏览器打开 http://127.0.0.1:8080 ，进页即全屏终端。多个 tab 各自独立 shell。断线会弹出是否重连（重连 = 新 shell）或关闭页面。

`make dev` 会：

1. 构建 `web/dist`
2. 并行监控前端源码（`vite build --watch`）和后端（air 看到 `web/dist` 变化后重建 Go、重新 embed，再重启）

配置见 `.env.example`：`ADDR`、`GIN_MODE`、`SHELL`、`SHELL_ARGS`、`CWD`、`WRITABLE`、`ALLOW_ORIGIN`、`MAX_SESSIONS`。`.env` 不覆盖已有环境变量。UI 固定打进二进制，没有 `DIST_DIR`。

## 常用命令

```bash
make test          # go test ./...
make build         # 前端 embed 进 bin/server
make release-build # 发版用单二进制（linux 静态链接）
make web-build     # 只构建前端
go run ./cmd/server
pnpm --dir web dev # 仅 Vite；/api 与 /ws 代理到 :8080
```

## 发版

```bash
git tag v0.1.0
git push origin v0.1.0
```

推送 `v*` tag 后，GitHub Actions 构建 linux/amd64 单二进制并挂到 Release：`web-tty_<tag>_linux_amd64`。下载后直接运行即可（默认 `:8080`，UI 已在二进制内）。

## 加 UI 组件

```bash
pnpm --dir web exec shadcn add button -y
```

给 agent 的完整约定见 [AGENTS.md](./AGENTS.md)。
