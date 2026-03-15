FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/pushpaka-worker .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates git docker-cli
WORKDIR /app
COPY --from=builder /app/pushpaka-worker .

EXPOSE 9090

ENTRYPOINT ["./pushpaka-worker"]
