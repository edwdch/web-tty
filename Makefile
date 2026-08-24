.PHONY: dev dev-web dev-api build web-build test tidy

# Human local development only. Agents must not use `make dev` to test.
dev: web-build
	$(MAKE) -j2 dev-web dev-api

dev-web:
	pnpm --dir web run build:watch

dev-api:
	go tool air -c .air.toml

web-build:
	pnpm --dir web build
	touch web/dist/.gitkeep

build: web-build
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy
	pnpm --dir web install
