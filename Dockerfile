#  Stage 1: Build 
# Go 1.25 matches go.work / cmd/pushpaka/go.mod
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy workspace descriptor first for better layer caching
COPY go.work go.work.sum* ./

# Copy module manifests for all three workspace modules
COPY backend/go.mod backend/go.sum ./backend/
COPY worker/go.mod worker/go.sum ./worker/
COPY cmd/pushpaka/go.mod cmd/pushpaka/go.sum* ./cmd/pushpaka/

# Download all workspace dependencies in one pass
RUN go work sync

# Copy all source
COPY backend/ ./backend/
COPY worker/ ./worker/
COPY cmd/pushpaka/ ./cmd/pushpaka/

# Build the unified binary (API + embedded worker) from the workspace root
RUN go build -C cmd/pushpaka \
    -ldflags="-w -s -X main.version=v1.0.0" \
    -o /pushpaka .

#  Stage 2: Runtime 
FROM alpine:3.21

# git    worker clones repositories at deploy time
# docker-cli  worker builds and runs containers (skipped gracefully if absent)
RUN apk add --no-cache ca-certificates curl git docker-cli

WORKDIR /app

COPY --from=builder /pushpaka /usr/local/bin/pushpaka

#  Runtime defaults 
# PUSHPAKA_COMPONENT controls which subsystem(s) start:
#   all    (default)  API server + build workers in one process
#   api               API server only (pair with external worker containers)
#   worker            Build workers only (reads from Redis queue)
ENV PUSHPAKA_COMPONENT=all

# Build worker directories  override via environment or docker-compose volumes
ENV BUILD_CLONE_DIR=/tmp/pushpaka-builds
ENV BUILD_DEPLOY_DIR=/deploy/pushpaka

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/ready || exit 1

ENTRYPOINT ["pushpaka"]
