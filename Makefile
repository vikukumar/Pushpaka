.PHONY: dev dev-all build clean help

# ── Development ────────────────────────────────────────────────────────────────
## Run the API only in dev mode with an embedded SQLite database.
## No Postgres, Redis, or Docker required.
dev:
	go run ./cmd/pushpaka -dev

## Run API + worker in dev mode. Requires local Redis and Docker.
dev-all:
	DATABASE_DRIVER=sqlite DATABASE_URL=pushpaka-dev.db \
	APP_ENV=development LOG_LEVEL=debug \
	go run ./cmd/pushpaka

# ── Build ──────────────────────────────────────────────────────────────────────
## Build the combined pushpaka binary.
build:
	go build -C cmd/pushpaka -ldflags="-w -s" -o ../../pushpaka .

## Run the dev binary directly (requires `make build` first).
run-dev: build
	./pushpaka -dev

# ── Backend & worker individually ─────────────────────────────────────────────
build-api:
	go build -C backend -o ../pushpaka-api ./cmd/server

build-worker:
	go build -C worker -o ../pushpaka-worker .

# ── Housekeeping ───────────────────────────────────────────────────────────────
clean:
	rm -f pushpaka pushpaka.exe pushpaka-api pushpaka-api.exe \
	      pushpaka-worker pushpaka-worker.exe \
	      pushpaka-dev.db

help:
	@echo ""
	@echo "  make dev         Run API only (SQLite – zero external deps)"
	@echo "  make dev-all     Run API + worker (SQLite + Redis + Docker)"
	@echo "  make build       Build the combined binary  →  ./pushpaka"
	@echo "  make build-api   Build backend only         →  ./pushpaka-api"
	@echo "  make build-worker Build worker only         →  ./pushpaka-worker"
	@echo "  make clean       Remove build artifacts and dev database"
	@echo ""
