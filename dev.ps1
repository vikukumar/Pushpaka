# dev.ps1 — Single-command development runner for Windows PowerShell
# Starts the API with SQLite embedded database (no Postgres, Redis, or Docker needed).
#
# Usage:
#   .\dev.ps1                      # API only on :8080
#   .\dev.ps1 -Port 9090           # custom port
#   .\dev.ps1 -Component all       # API + worker (requires Redis + Docker)
#
param(
    [string]$Port      = "8080",
    [string]$Component = "api",
    [string]$DbPath    = "pushpaka-dev.db"
)

$env:APP_ENV            = "development"
$env:DATABASE_DRIVER    = "sqlite"
$env:DATABASE_URL       = $DbPath
$env:PUSHPAKA_COMPONENT = $Component
$env:LOG_LEVEL          = "debug"
$env:PORT               = $Port

if (-not $env:JWT_SECRET) {
    $env:JWT_SECRET = "dev-secret-change-in-production"
}

Write-Host "Pushpaka dev mode: component=$Component  db=$DbPath  port=$Port" -ForegroundColor Cyan

go run ./cmd/pushpaka
