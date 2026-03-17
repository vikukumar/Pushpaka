#!/usr/bin/env bash
# GitHub Actions Workflows for Pushpaka

# This script helps manage version and test the CI/CD pipeline locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
info() { echo -e "${BLUE}ℹ️  $*${NC}"; }
success() { echo -e "${GREEN}✅ $*${NC}"; }
warning() { echo -e "${YELLOW}⚠️  $*${NC}"; }
error() { echo -e "${RED}❌ $*${NC}"; }

# Get current version info
get_version_info() {
    local version=$(cat VERSION)
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    echo "Current VERSION file: ${version}"
    echo "Last git tag: ${last_tag:-"none"}"
    
    if [ -n "$last_tag" ]; then
        local tag_version=$(echo "$last_tag" | sed 's/v//' | rev | cut -d. -f2- | rev)
        local tag_counter=$(echo "$last_tag" | rev | cut -d. -f1 | rev)
        echo "Last release: v${tag_version}.${tag_counter}"
    fi
}

# Show version management
show_version() {
    info "Version Information"
    get_version_info
}

# Update VERSION file
update_version() {
    local new_version=$1
    
    if [ -z "$new_version" ]; then
        error "Version required"
        exit 1
    fi
    
    echo "$new_version" > VERSION
    success "Version updated to: $new_version"
    
    info "Commit and push the change:"
    echo "  git add VERSION"
    echo "  git commit -m 'chore: bump version to $new_version'"
    echo "  git push origin main"
}

# Validate workflows
validate_workflows() {
    info "Validating workflows..."
    
    local workflows=(.github/workflows/*.yml)
    local errors=0
    
    for workflow in "${workflows[@]}"; do
        if [ -f "$workflow" ]; then
            if grep -q "^name:" "$workflow"; then
                success "✓ $(basename $workflow)"
            else
                error "✗ $(basename $workflow) - missing name"
                ((errors++))
            fi
        fi
    done
    
    if [ $errors -eq 0 ]; then
        success "All workflows are valid"
    else
        error "$errors workflow(s) have issues"
        exit 1
    fi
}

# Test builds locally (requires Docker)
test_build() {
    local component=$1
    
    case "$component" in
        backend)
            info "Testing backend build..."
            cd backend
            go build ./...
            success "Backend builds successfully"
            cd ..
            ;;
        worker)
            info "Testing worker build..."
            cd worker
            go build ./cmd/worker
            success "Worker builds successfully"
            cd ..
            ;;
        frontend)
            info "Testing frontend build..."
            cd frontend
            pnpm install --frozen-lockfile
            pnpm build
            success "Frontend builds successfully"
            cd ..
            ;;
        docker)
            info "Testing Docker build..."
            docker build -t pushpaka:test .
            success "Docker image builds successfully"
            ;;
        all)
            test_build backend
            test_build worker
            test_build frontend
            ;;
        *)
            error "Unknown component: $component"
            exit 1
            ;;
    esac
}

# Show help
show_help() {
    cat << EOF
Pushpaka CI/CD Management

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    version                 Show current version information
    bump <VERSION>          Update VERSION file (e.g., "1.0.1")
    validate                Validate all workflow files
    test [COMPONENT]        Test local builds
                           Components: backend, worker, frontend, docker, all
    help                    Show this help message

EXAMPLES:
    $0 version              # Show version info
    $0 bump 1.0.1          # Bump to version 1.0.1
    $0 validate            # Check all workflows
    $0 test backend        # Test backend build
    $0 test docker         # Test Docker build

PATH: .github/workflows/
    release.yml            - auto-trigger on push to main
    ci.yml                 - runs on PR/push for testing
    manual-release.yml     - manual dispatch for selective releases
EOF
}

# Main
case "$1" in
    version)
        show_version
        ;;
    bump)
        update_version "$2"
        ;;
    validate)
        validate_workflows
        ;;
    test)
        test_build "${2:-all}"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        show_help
        ;;
esac
