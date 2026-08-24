.PHONY: dev dev-web dev-api build web-build release-build test tidy

# Human local development only. Agents must not use `make dev` to test.
dev: web-build
	$(MAKE) -j2 dev-web dev-api

dev-web:
	pnpm --dir web run build:watch

dev-api:
	go tool air -c .air.toml

web-build:
	pnpm --dir web build

build: web-build
	mkdir -p bin
	go build -o bin/server ./cmd/server

release-build: web-build
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

test: web-build
	go test ./...

tidy:
	go mod tidy
	pnpm --dir web install
