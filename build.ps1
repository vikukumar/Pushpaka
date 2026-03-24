<#
.SYNOPSIS
  Pushpaka build script for Windows (equivalent to the Makefile).

.DESCRIPTION
  Usage:
    .\build.ps1 dev           - Run API in dev mode (SQLite, no external deps)
    .\build.ps1 build         - Build all-in-one binary   -> pushpaka.exe
    .\build.ps1 build-api     - Build API-only binary     -> pushpaka-api.exe
    .\build.ps1 build-worker  - Build worker-only binary  -> pushpaka-worker.exe
    .\build.ps1 build-all     - Build all three binaries
    .\build.ps1 front-dev     - Run Next.js dev server (port 3000)
    .\build.ps1 clean         - Remove build artifacts
#>
param([string]$Target = "help")

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot

function Invoke-FrontendBuild {
    Write-Host "Building frontend (static export for Go embedding)..." -ForegroundColor Cyan
    # Patch layout: remove force-dynamic (incompatible with output: export)
    node "$Root\scripts\patch-layout.js" remove
    Push-Location "$Root\frontend"
    try {
        $env:STATIC_EXPORT = '1'
        pnpm build
        $env:STATIC_EXPORT = ''
    } catch {
        $env:STATIC_EXPORT = ''
        Pop-Location
        node "$Root\scripts\patch-layout.js" restore
        throw
    }
    Pop-Location
    node "$Root\scripts\patch-layout.js" restore
    Write-Host "Copying frontend assets to backend/ui/dist..." -ForegroundColor Cyan
    node "$Root\scripts\cpfe.js"
}

$ldflags = "-ldflags=-w -s"

switch ($Target) {
    "dev" {
        Write-Host "Starting API in dev mode (SQLite)..." -ForegroundColor Green
        Set-Location $Root
        & go run ./cmd/pushpaka -dev
    }
    "front-dev" {
        Write-Host "Starting Next.js dev server on :3000..." -ForegroundColor Green
        Set-Location "$Root\frontend"
        pnpm dev
    }
    "front-build" {
        Invoke-FrontendBuild
    }
    "build" {
        Invoke-FrontendBuild
        Write-Host "Building pushpaka.exe..." -ForegroundColor Cyan
        Set-Location $Root
        & go build $ldflags -o pushpaka.exe ./cmd/pushpaka
        Write-Host "Done: .\pushpaka.exe  (set PUSHPAKA_COMPONENT=api|worker|all)" -ForegroundColor Green
    }
    "build-api" {
        Invoke-FrontendBuild
        Write-Host "Building pushpaka-api.exe..." -ForegroundColor Cyan
        Set-Location $Root
        & go build -C backend $ldflags -o ../pushpaka-api.exe ./cmd/server
        Write-Host "Done: .\pushpaka-api.exe" -ForegroundColor Green
    }
    "build-worker" {
        Write-Host "Building pushpaka-worker.exe..." -ForegroundColor Cyan
        Set-Location $Root
        & go build -C worker $ldflags -o ../pushpaka-worker.exe .
        Write-Host "Done: .\pushpaka-worker.exe" -ForegroundColor Green
    }
    "build-all" {
        Invoke-FrontendBuild
        Write-Host "Building all binaries..." -ForegroundColor Cyan
        Set-Location $Root
        & go build $ldflags -o pushpaka.exe ./cmd/pushpaka
        & go build -C backend $ldflags -o ../pushpaka-api.exe ./cmd/server
        & go build -C worker $ldflags -o ../pushpaka-worker.exe .
        Write-Host "Done: pushpaka.exe, pushpaka-api.exe, pushpaka-worker.exe" -ForegroundColor Green
    }
    "clean" {
        Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
        Remove-Item -ErrorAction SilentlyContinue pushpaka.exe, pushpaka-api.exe, pushpaka-worker.exe
        Remove-Item -ErrorAction SilentlyContinue pushpaka-dev.db, pushpaka-dev.db-shm, pushpaka-dev.db-wal
    }
    default {
        Write-Host ""
        Write-Host "  Dev:        .\build.ps1 dev           (API + SQLite)"
        Write-Host "              .\build.ps1 front-dev     (Next.js :3000)"
        Write-Host ""
        Write-Host "  Build:      .\build.ps1 build         -> pushpaka.exe (all-in-one)"
        Write-Host "              .\build.ps1 build-api     -> pushpaka-api.exe"
        Write-Host "              .\build.ps1 build-worker  -> pushpaka-worker.exe"
        Write-Host "              .\build.ps1 build-all     -> all three"
        Write-Host ""
    }
}