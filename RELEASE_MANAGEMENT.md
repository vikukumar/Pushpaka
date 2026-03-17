# Release Management Process

## Release Planning

### Version Numbering
Pushpaka uses **Semantic Versioning** (MAJOR.MINOR.PATCH):

- **MAJOR:** Breaking changes
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes

Example: v1.2.3
- 1 = Major version
- 2 = Minor version (features)
- 3 = Patch version (fixes)

### Release Cycle
- **Major releases:** Annual (v2.0, v3.0, etc.)
- **Minor releases:** Quarterly (v1.1, v1.2, v1.3, etc.)
- **Patch releases:** As needed for critical bugs

### Release Branches
```
main (production-ready)
├── v1.0.x (patch releases)
├── v1.1.x (patch releases)
└── develop (development)
```

## Release Checklist

### Planning Phase (2 weeks before)
- [ ] Define release scope
- [ ] Assign features to release
- [ ] Document breaking changes (if any)
- [ ] Create release branch from develop

### Development Phase (Feature development)
- [ ] Implement features using conventional commits
- [ ] Add tests for all changes
- [ ] Update documentation
- [ ] Review PRs thoroughly
- [ ] Maintain CHANGELOG.md

### Pre-Release Phase (1 week before)
- [ ] Freeze feature development
- [ ] Fix remaining bugs
- [ ] Update version numbers
- [ ] Generate release notes
- [ ] Test release artifacts
- [ ] Security scanning complete
- [ ] Performance testing complete

### Release Phase
- [ ] Update VERSION file
- [ ] Update CHANGELOG.md
- [ ] Tag commit with version (v1.0.0)
- [ ] Build and publish artifacts
- [ ] Publish Docker images
- [ ] Publish Helm charts
- [ ] Create GitHub Release
- [ ] Notify community

### Post-Release Phase
- [ ] Monitor for critical issues
- [ ] Release patch versions if needed
- [ ] Update documentation site
- [ ] Blog post/announcement
- [ ] Collect community feedback

## Changelog Format

Use the format specified in CHANGELOG.md:

```markdown
## [1.0.0] - 2026-03-17

### Added
- New feature 1
- New feature 2

### Changed
- Breaking change 1
- Modified behavior

### Fixed
- Bug fix 1
- Bug fix 2

### Security
- Security fix 1

### Deprecated
- Deprecated feature

### Removed
- Removed feature

### Known Issues
- Issue 1
- Issue 2
```

## Conventional Commits

Use conventional commit messages for better automation:

```
feat: Add new feature
fix: Fix bug
docs: Update documentation
style: Format code
refactor: Refactor code
perf: Improve performance
test: Add tests
chore: Update dependencies
ci: Update CI configuration
```

Example:
```
feat: Add scheduled deployments for projects

- Allow users to schedule deployments for specific times
- Add cron-based scheduling
- Persist schedule in database
```

## Breaking Changes

For breaking changes:

1. **Document clearly in CHANGELOG.md**
   ```
   ### BREAKING CHANGES
   - API endpoint `/api/v1/projects/{id}` changed to `/api/v2/projects/{id}`
   - Database schema changed for deployments table
   - Configuration key `LOG_LEVEL` renamed to `PUSHPAKA_LOG_LEVEL`
   ```

2. **Provide migration guide**
   ```
   ### Migration Guide
   
   **For API consumers:**
   Update your scripts to use the new endpoint.
   
   **For database:**
   Run migration: `migrations/v1.1.0_to_v2.0.0.sql`
   
   **For configuration:**
   Replace `LOG_LEVEL` with `PUSHPAKA_LOG_LEVEL` in your .env
   ```

3. **Deprecate gradually (when possible)**
   - v1.9.0: Deprecate old API
   - v1.10.0: Support both old and new
   - v2.0.0: Remove old API

## Version Bumping

### Manual (for releases)
```bash
# Update VERSION file
echo "1.1.0" > VERSION

# Update Chart.yaml for Helm
sed -i 's/version: 1.0.0/version: 1.1.0/' helm/pushpaka/Chart.yaml

# Commit and tag
git add VERSION helm/pushpaka/Chart.yaml
git commit -m "chore: Bump version to 1.1.0"
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin main --tags
```

### Automated (in workflows)
```yaml
- name: Bump version
  run: |
    VERSION=$(cat VERSION)
    NEW_VERSION=$(echo $VERSION | awk -F. '{print $1"."$2"."$3+1}')
    echo $NEW_VERSION > VERSION
```

## Release Artifacts

### Required Artifacts
- [ ] Docker image: `pushpaka:v1.0.0`
- [ ] Docker image: `pushpaka:latest` (for latest release)
- [ ] Helm chart: `pushpaka-1.0.0.tgz`
- [ ] GitHub Release with notes
- [ ] CHANGELOG.md entry
- [ ] Binary releases (if applicable)

### Storage Locations
- **Docker Images:** Docker Hub, GitHub Container Registry
- **Helm Charts:** GitHub Pages (`/helm` directory)
- **Release Notes:** GitHub Releases
- **Binary:** GitHub Releases

## Rollback Procedure

If critical issues are discovered after release:

### For Docker
```bash
# Revert to previous release
docker pull pushpaka:v1.0.0  # Deploy previous stable
```

### For Kubernetes
```bash
# Rollback Helm release
helm rollback pushpaka -n pushpaka
```

### For Single Binary
```bash
# Revert git tag and rebuild
git checkout v1.0.0
make build
```

### Communicate Rollback
1. GitHub Release: Mark as "Pre-release"
2. Announcement: Post in discussions
3. Blog: Update with incident details
4. Monitor: Watch for ongoing issues

## Communication

### Before Release
- Announce feature freeze (1 week before)
- Preview release notes
- Ask for testing help

### At Release
- GitHub Release announcement
- Update website
- Publish blog post
- Community announcements

### After Release
- Monitor for issues
- Answer community questions
- Gather feedback
- Plan improvements

## Tools & Automation

### VERSION File
Location: `./VERSION`
Content: `1.0.0`
Used by: CI/CD pipelines, release workflows

### GitHub Actions Workflows
- `helm-release.yml` - Publishes Helm charts automatically
- `docker-scan.yml` - Scans Docker images for vulnerabilities
- `security-scanning.yml` - Runs security checks
- `manual-release.yml` - Manual release trigger

### Release Checklist Template
```markdown
# Release v1.1.0 Checklist

- [ ] Features implemented
- [ ] Tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Security scans passed
- [ ] Performance tests passed
- [ ] Helm chart updated
- [ ] Docker image built
- [ ] VERSION file updated
- [ ] Git tag created
- [ ] Release notes published
- [ ] Community notified
```

## Emergency Procedures

### Critical Security Issue
1. Create security branch from main
2. Fix issue
3. Release patch version immediately (v1.0.1)
4. No feature releases during emergency

### Major Bug Discovery
1. Assess severity
2. Decide: Patch release or fix in next release
3. If patch: Follow emergency process
4. If next release: Add to backlog

## Contact & Escalation

- **Release Manager:** vikukumar
- **Security Issues:** security@pushpaka.dev
- **Critical Issues:** GitHub issues with urgent label
- **Communication:** GitHub discussions, announcements

---

**Last Updated:** March 17, 2026  
**Next Review:** After v1.1.0 release
