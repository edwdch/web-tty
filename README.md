# web-tty

浏览器 Web TTY：Go（Gin）+ React（Vite, shadcn/ui, Tailwind CSS）+ wterm/Ghostty。每个浏览器 tab 一条 WebSocket；PTY 在服务进程里继续跑，关页可恢复，多 tab 可共享同一 session。前端构建产物在编译期 embed 进二进制，Gin 从 embed FS 托管。

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

浏览器打开 http://127.0.0.1:8080 ，进页即全屏终端。每个 tab 会记住自己上次的 session：刷新同一标签页会直接 attach，不弹窗。新标签页，或这个 tab 记着的 session 已经没了：后台没有运行中的 session 就直接开新 shell，否则弹出列表选一个继续或新建。关 tab 不会杀掉 shell。右上角半透明按钮可切换或关闭 session（有人连着也能关）。多个页面可以 attach 同一个 session（PTY 只有一个尺寸，后一次 resize 生效）。

断线会弹出是否重连（刷新后再走上面的进页规则，session 可能还在）或关闭页面。不要指望服务重启后还能恢复：session 只活在本进程内存里（`make dev` 的 air 重启也会丢掉）。

空闲回收：最后一个 client 离开超过一周（`SESSION_IDLE=168h`）的 session 会被关掉。同时最多 50 个 PTY（`MAX_SESSIONS`）。

`make dev` 会：

1. 构建 `web/dist`
2. 并行监控前端源码（`vite build --watch`）和后端（air 看到 `web/dist` 变化后重建 Go、重新 embed，再重启）

配置见 `.env.example`：`ADDR`、`GIN_MODE`、`SHELL`、`SHELL_ARGS`、`CWD`、`WRITABLE`、`ALLOW_ORIGIN`、`MAX_SESSIONS`、`SESSION_IDLE`。`.env` 不覆盖已有环境变量。UI 固定打进二进制，没有 `DIST_DIR`。

## 常用命令

```bash
make test          # 先构建前端，再 go test ./...
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
