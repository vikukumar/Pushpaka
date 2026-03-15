# ── Stage 1: Build ─────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy workspace descriptor first for layer caching
COPY go.work go.work.sum* ./

# Copy module manifests
COPY backend/go.mod backend/go.sum ./backend/
COPY worker/go.mod worker/go.sum ./worker/
COPY cmd/pushpaka/go.mod ./cmd/pushpaka/

# Download all workspace dependencies in one pass
RUN go work sync

# Copy all source
COPY backend/ ./backend/
COPY worker/ ./worker/
COPY cmd/pushpaka/ ./cmd/pushpaka/

# Build the unified binary; -C changes directory before building
RUN go build -C cmd/pushpaka \
    -ldflags="-w -s -X main.version=v1.0.0" \
    -o /pushpaka .

# ── Stage 2: Runtime ───────────────────────────────────────────────────────────
FROM alpine:3.19

# git and docker-cli are needed by the worker component at runtime
RUN apk add --no-cache ca-certificates git docker-cli

WORKDIR /app

COPY --from=builder /pushpaka /usr/local/bin/pushpaka

# Default to running API + worker together. Override with:
#   -e PUSHPAKA_COMPONENT=api
#   -e PUSHPAKA_COMPONENT=worker
#   -e PUSHPAKA_COMPONENT=all   (default)
ENV PUSHPAKA_COMPONENT=all

EXPOSE 8080

ENTRYPOINT ["pushpaka"]
