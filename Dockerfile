#  Stage 1: Build 
# Go 1.25 matches go.work / cmd/pushpaka/go.mod
FROM golang:1.25-alpine AS builder

# Build arguments
ARG VERSION=v1.0.0
ARG BUILD_DATE
ARG VCS_REF

# Labels for image metadata
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${VCS_REF}"

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

# Ensure the embedded UI path exists for server-only builds in CI/release.
RUN mkdir -p /app/backend/ui/dist && touch /app/backend/ui/dist/.gitkeep

# Build the unified binary (API + embedded worker) from the workspace root
# Use ARG VERSION for build-time version injection
RUN go build -C cmd/pushpaka \
    -ldflags="-w -s -X main.version=${VERSION}" \
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
