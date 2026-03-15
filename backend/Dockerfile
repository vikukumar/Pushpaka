FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/pushpaka-api ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/pushpaka-api .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s \
  CMD wget -qO- http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./pushpaka-api"]
