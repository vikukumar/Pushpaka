# GitHub Actions Workflows for Pushpaka (Windows PowerShell)

param(
    [Parameter(Position = 0)]
    [ValidateSet("version", "bump", "validate", "test", "help")]
    [string]$Command = "help",
    
    [Parameter(Position = 1)]
    [string]$Argument
)

function Write-Info { Write-Host "ℹ️  $args" -ForegroundColor Blue }
function Write-Success { Write-Host "✅ $args" -ForegroundColor Green }
function Write-Warning { Write-Host "⚠️  $args" -ForegroundColor Yellow }
function Write-Error { Write-Host "❌ $args" -ForegroundColor Red }

function Get-VersionInfo {
    Write-Info "Version Information"
    
    $version = Get-Content -Path "VERSION" -Raw
    Write-Host "Current VERSION file: $($version.Trim())"
    
    $lastTag = & git describe --tags --abbrev=0 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Last git tag: $lastTag"
        
        $tagVersion = $lastTag -replace 'v' -replace '\.\d+$'
        $tagCounter = $lastTag -match '\d+$'; $tagCounter = $matches[0]
        Write-Host "Last release: v$tagVersion.$tagCounter"
    }
    else {
        Write-Host "Last git tag: none"
    }
}

function Update-Version {
    param([string]$NewVersion)
    
    if ([string]::IsNullOrEmpty($NewVersion)) {
        Write-Error "Version required"
        exit 1
    }
    
    Set-Content -Path "VERSION" -Value $NewVersion
    Write-Success "Version updated to: $NewVersion"
    
    Write-Info "Commit and push the change:"
    Write-Host "  git add VERSION"
    Write-Host "  git commit -m 'chore: bump version to $NewVersion'"
    Write-Host "  git push origin main"
}

function Validate-Workflows {
    Write-Info "Validating workflows..."
    
    $workflows = Get-ChildItem ".github/workflows/*.yml" -ErrorAction SilentlyContinue
    $errors = 0
    
    foreach ($workflow in $workflows) {
        $content = Get-Content $workflow.FullName
        if ($content -match "^name:") {
            Write-Success "✓ $($workflow.Name)"
        }
        else {
            Write-Error "✗ $($workflow.Name) - missing name"
            $errors++
        }
    }
    
    if ($errors -eq 0) {
        Write-Success "All workflows are valid"
    }
    else {
        Write-Error "$errors workflow(s) have issues"
        exit 1
    }
}

function Test-Build {
    param([string]$Component = "all")
    
    switch ($Component) {
        "backend" {
            Write-Info "Testing backend build..."
            Push-Location backend
            & go build ./...
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Backend builds successfully"
            }
            Pop-Location
        }
        "worker" {
            Write-Info "Testing worker build..."
            Push-Location worker
            & go build ./cmd/worker
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Worker builds successfully"
            }
            Pop-Location
        }
        "frontend" {
            Write-Info "Testing frontend build..."
            Push-Location frontend
            & pnpm install --frozen-lockfile
            & pnpm build
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Frontend builds successfully"
            }
            Pop-Location
        }
        "docker" {
            Write-Info "Testing Docker build..."
            & docker build -t "pushpaka:test" .
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Docker image builds successfully"
            }
        }
        "all" {
            Test-Build backend
            Test-Build worker
            Test-Build frontend
        }
        default {
            Write-Error "Unknown component: $Component"
            exit 1
        }
    }
}

function Show-Help {
    $help = @"
Pushpaka CI/CD Management (Windows PowerShell)

USAGE:
    .\manage-workflows.ps1 [COMMAND] [OPTIONS]

COMMANDS:
    version                 Show current version information
    bump <VERSION>          Update VERSION file (e.g., "1.0.1")
    validate                Validate all workflow files
    test [COMPONENT]        Test local builds
                           Components: backend, worker, frontend, docker, all
    help                    Show this help message

EXAMPLES:
    .\manage-workflows.ps1 version              # Show version info
    .\manage-workflows.ps1 bump 1.0.1          # Bump to version 1.0.1
    .\manage-workflows.ps1 validate            # Check all workflows
    .\manage-workflows.ps1 test backend        # Test backend build
    .\manage-workflows.ps1 test docker         # Test Docker build

PATH: .github\workflows\
    release.yml            - auto-trigger on push to main
    ci.yml                 - runs on PR/push for testing
    manual-release.yml     - manual dispatch for selective releases
"@
    Write-Host $help
}

# Main
switch ($Command) {
    "version" {
        Get-VersionInfo
    }
    "bump" {
        Update-Version $Argument
    }
    "validate" {
        Validate-Workflows
    }
    "test" {
        Test-Build $Argument
    }
    "help" {
        Show-Help
    }
    default {
        Show-Help
    }
}
