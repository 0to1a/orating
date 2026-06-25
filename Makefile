.PHONY: run gen build check migrate e2e

FRONTEND_SRCS := $(shell find frontend/src frontend/static -type f 2>/dev/null) \
	frontend/package.json frontend/svelte.config.js frontend/vite.config.ts \
	frontend/tsconfig.json

frontend/.build-stamp: $(FRONTEND_SRCS)
	cd frontend && bun run build
	rm -rf internal/platform/frontend/frontend_dist/*
	cp -r frontend/build/. internal/platform/frontend/frontend_dist/
	@touch $@

run: frontend/.build-stamp
	go run ./cmd/server

# Regenerate everything: sqlc queries + openapi.json + TypeScript SDK
gen:
	sqlc generate
	go run ./cmd/gen-spec openapi.json
	cd frontend && bunx openapi-ts

build: gen
	cd frontend && bun run build
	rm -rf internal/platform/frontend/frontend_dist/*
	cp -r frontend/build/. internal/platform/frontend/frontend_dist/
	go build -o bin/server ./cmd/server

check:
	go vet ./...
	@! grep -rn 'json:"[^"]*_[^"]*"' internal/ 2>/dev/null || (echo "ERROR: snake_case JSON tag found"; exit 1)
	@./scripts/check-no-cross-domain.sh
	cd frontend && bun run tsc --noEmit

# Requires: make run (Go server) running in another terminal
e2e:
	cd e2e && bunx playwright test

# make migrate name=add_foo_to_bar
migrate:
	@[ -n "$(name)" ] || (echo "Usage: make migrate name=<name>"; exit 1)
	migrate create -ext sql -dir internal/platform/migration/sql -seq=false -tz UTC $(name)

