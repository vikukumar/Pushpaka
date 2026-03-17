# Issue Fix Template

Use this template as a reference for fixing reported issues systematically.

## Issue Analysis

### Step 1: Reproduce Issue
- [ ] Understand issue description
- [ ] Reproduce the issue locally
- [ ] Confirm with original reporter (if needed)
- [ ] Document reproduction steps

```bash
# Example: Steps to reproduce
1. Create new project
2. Connect GitHub repository
3. Trigger deployment
4. Observe: [Actual behavior]
```

### Step 2: Root Cause Analysis
- [ ] Examine relevant code
- [ ] Review recent changes (git log)
- [ ] Check logs for error messages
- [ ] Identify affected components

**Affected Components:**
- API
- Dashboard
- Worker
- Database
- Infrastructure
- Other: ___

**Root Cause:**
```
Describe what causes the issue
```

### Step 3: Determine Fix Scope
- [ ] Is this a breaking change?
- [ ] What components need updates?
- [ ] Are there database migrations?
- [ ] Do dependencies need updating?

## Fix Implementation

### Create Fix Branch
```bash
# From develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b fix/issue-123-description
# or for bugs
git checkout -b bugfix/issue-123-description
```

### Code Changes

**File 1:** `backend/handler/projects.go`
```go
// Before
func GetProject(c *gin.Context) {
    // Old implementation
}

// After
func GetProject(c *gin.Context) {
    // Fixed implementation
}
```

**File 2:** `frontend/components/ProjectList.tsx`
```typescript
// Changes made to fix issue
```

### Update Tests
```go
// Test case for the fix
func TestProjectFix(t *testing.T) {
    // Test that issue is fixed
    result := GetProject()
    assert.Equal(t, expected, result)
}
```

### Database Migrations (if needed)
```sql
-- migrations/20260317_fix_issue_123.sql
ALTER TABLE deployments ADD COLUMN status_updated_at TIMESTAMP;
CREATE INDEX idx_deployment_status ON deployments(status);
```

### Documentation Updates
- [ ] Update API documentation
- [ ] Update README if needed
- [ ] Add troubleshooting note (if applicable)
- [ ] Update configuration guide (if needed)

## Quality Assurance

### Local Testing
```bash
# Run tests
make test

# Run specific test
go test -v ./backend/handlers/...

# Test the fix manually
./pushpaka -dev
# Test in browser: http://localhost:3000
```

### Checklist
- [ ] Issue is reproducible before fix
- [ ] Issue is NOT reproducible after fix
- [ ] No new errors in logs
- [ ] No performance regression
- [ ] All tests pass

## PR & Review

### Create Pull Request
```markdown
Title: fix: Resolve issue #123 - description

## Description
Fixes #123

### Changes
- Change 1
- Change 2

### Testing
Steps to test the fix

### Checklist
- [x] Tests pass
- [x] Documentation updated
- [x] No breaking changes
```

### Code Review
- [ ] Code follows project style
- [ ] Solution is minimal and focused
- [ ] Comments explain non-obvious changes
- [ ] No temporary debug code

## Deployment

### Update CHANGELOG.md
```markdown
## [Unreleased]

### Fixed
- Fix for issue #123: description of fix
  - Details about what was fixed
  - Impact on users
```

### Version Bump (if patch release)
```bash
# For patch releases (1.0.0 -> 1.0.1)
echo "1.0.1" > VERSION
git add VERSION CHANGELOG.md
git commit -m "fix: Bump version to 1.0.1"
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin main --tags
```

## After Merge

### Verification
- [ ] Fix deployed to production
- [ ] Monitor for related issues
- [ ] Collect user feedback
- [ ] Check error tracking system

### Close Issue
Add comment:
```
This has been fixed in v1.0.1

The fix: [brief description]
Merged in: [PR #123]

Please update and test. Let us know if you encounter any further issues!
```

### Related Issues
- [ ] Are there similar issues?
- [ ] Should other code have similar fix?
- [ ] Create follow-up issues if needed

## Issue Fix Categories

### Bug Fix
```
Issue: Bug behavior
Fix: Correct the behavior
Test: Verify original behavior is fixed
```

### Performance Issue
```
Issue: Slow response time
Fix: Optimize code/database query
Test: Benchmark shows improvement
```

### Security Issue
```
Issue: Vulnerability found
Fix: Patch the vulnerability
Test: Exploit no longer works
Audit: Security scan passes
```

### Documentation Issue
```
Issue: Unclear/wrong documentation
Fix: Update documentation
Test: Documentation is accurate
```

## Common Fix Patterns

### Database Query Optimization
```go
// Before (N+1 query problem)
projects := getProjects()
for _, project := range projects {
    project.Deployments = getDeployments(project.ID)  // Multiple queries
}

// After (Optimized)
projects := getProjectsWithDeployments()  // Single query with JOIN
```

### API Response Fix
```go
// Before (Missing field)
type ProjectResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // Missing: Status field
}

// After (Field added)
type ProjectResponse struct {
    ID     string    `json:"id"`
    Name   string    `json:"name"`
    Status string    `json:"status"`  // Added
}
```

### Frontend Bug
```typescript
// Before (State not updated)
const handleClick = () => {
    deployProject(projectId);
    // Missing: Update UI state
}

// After (State updated)
const handleClick = async () => {
    await deployProject(projectId);
    setDeployments(updated);  // Update UI
    setIsLoading(false);      // Clear loading state
}
```

## Regression Testing

After fixing an issue, check:
- [ ] Related features still work
- [ ] No new error messages
- [ ] API responses are correct
- [ ] Dashboard displays correctly
- [ ] Logs are clean
- [ ] Performance hasn't degraded

## Documentation

Create follow-up documentation:

### If adding new feature as fix:
- Update API documentation
- Add to feature list
- Create user guide

### If fixing common issue:
- Add to troubleshooting guide
- Document workaround (if applicable)
- Link from FAQ

### If security fix:
- Add to security advisory (if public)
- Document upgrade path
- Note in release notes

---

**Template Version:** 1.0  
**Last Updated:** March 17, 2026
