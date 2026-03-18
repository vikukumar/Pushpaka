# Container Image Security

This document covers Docker image security scanning, vulnerability management, and best practices for Pushpaka.

## Docker Image Security Scanning

### Scan Tools

Pushpaka uses multiple scanning tools:

1. **Trivy** - Fast vulnerability scanning
2. **Grype** - Generate SBOM (Software Bill of Materials)
3. **Docker Scout** - Supply chain security
4. **Snyk** - Dependency vulnerability scanning

### Running Local Scans

```bash
# Install Trivy
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin

# Scan local Docker image
trivy image pushpaka:latest
trivy image pushpaka:latest --severity CRITICAL,HIGH
trivy image pushpaka:latest --format json > scan-report.json

# Scan for config issues
trivy config .

# Generate SBOM
trivy image pushpaka:latest --format cyclonedx > sbom.json
```

### GitHub Actions Scanning

Automated scanning runs:

**On every push to main:**
```yaml
- Trivy filesystem scan
- Trivy Docker image scan
- GoSec (Go code security)
- CodeQL analysis
- Gitleaks (secrets detection)
```

**On schedule (daily at 2 AM UTC):**
```yaml
- Full vulnerability audit
- Dependency check
- License compliance
```

Results uploaded to GitHub Security tab automatically.

## Understanding Scan Results

### Severity Levels

| Level | Action | Timeline |
|-------|--------|----------|
| CRITICAL | Block deployment, immediate fix | < 24 hours |
| HIGH | Fix before release | < 1 week |
| MEDIUM | Plan for next release | < 1 month |
| LOW | Track and monitor | Next review |

### Interpreting Reports

**Critical vulnerability example:**
```
Package: openssl
Vulnerability: CVE-2024-0001
Severity: CRITICAL
Description: Remote code execution in TLS handshake
Action: Update image base layer immediately
```

## Vulnerability Management

### Policy

1. **Known vulnerabilities:** Use when absolutely necessary
2. **Review:** Document reason for allowing
3. **Timeline:** Set remediation date
4. **Monitor:** Track for patches

### Updating Dependencies

```yaml
# .trivyignore - Temporarily accept known vulnerabilities
# Format: <vulnerability_id> <expiry_date>

CVE-2024-0001 2026-06-17  # Pending patch
CVE-2024-0002 2026-03-31  # Needs investigation

# Document reason for each
# CVE-2024-0001: No patch for go version 1.25, upgrading in v1.2.0
```

### Version Updates

```bash
# Update base image in Dockerfile
FROM golang:1.25-alpine3.19 AS builder  # Updated from 1.24

# Update dependencies
go get -u ./...
go mod tidy

# Re-scan after updates
trivy image pushpaka:test
```

## Best Practices

### Dockerfile Security

✅ **DO:**
```dockerfile
# Use specific versions (not :latest)
FROM golang:1.25-alpine3.19 AS builder
FROM alpine:3.19

# Run as non-root
RUN addgroup -g 1000 pushpaka && adduser -u 1000 -G pushpaka pushpaka
USER pushpaka

# Minimal base image
FROM alpine:3.19  # Much smaller than ubuntu

# Multi-stage build
FROM golang:1.25 AS builder
# Build step
FROM alpine:3.19
COPY --from=builder /app/binary /app/
```

❌ **DON'T:**
```dockerfile
# Use :latest tag
FROM golang:latest

# Run as root
RUN app install

# Large base images
FROM ubuntu:22.04
```

### Image Scanning in CI/CD

```yaml
- name: Build and scan image
  run: |
    docker build -t pushpaka:${{ github.sha }} .
    trivy image pushpaka:${{ github.sha }} --severity CRITICAL
    
    if [ $? -ne 0 ]; then
      echo "❌ Critical vulnerabilities found!"
      exit 1
    fi
```

## Vulnerability Response

### When CVE is discovered

1. **Assess impact**
   - Does it affect Pushpaka?
   - Is it in production?
   - What's the severity?

2. **Patch or mitigate**
   ```bash
   # Quick patch
   git stash
   docker build -t pushpaka:patched .
   trivy image pushpaka:patched  # Verify fix
   
   # Deploy patch
   docker push pushpaka:v1.0.1-patched
   ```

3. **Update and release**
   ```bash
   # Update dependency
   go get package@latest
   go mod tidy
   
   # Test thoroughly
   make test
   
   # Release patch
   git tag -a v1.0.1 -m "Security patch"
   ```

4. **Communicate**
   - Create security advisory (if public)
   - Notify users
   - Blog post if needed
   - Update website

## SBOM (Software Bill of Materials)

### Generate SBOM

```bash
# Generate SBOM in CycloneDX format
trivy image pushpaka:latest --format cyclonedx --output sbom.json

# Generate SBOM in SPDX format
trivy image pushpaka:latest --format spdx --output sbom.spdx

# JSON format for easy parsing
cat sbom.json | jq '.components[] | {name, version}'
```

### SBOM Contents

```json
{
  "bomFormat": "CycloneDX",
  "components": [
    {
      "name": "postgresql",
      "version": "17.0",
      "purl": "pkg:deb/debian/postgresql",
      "licenses": [{"id": "PostgreSQL"}]
    },
    {
      "name": "golang",
      "version": "1.25",
      "purl": "pkg:golang"
    }
  ]
}
```

### Uses

- **Inventory:** Know what's in your image
- **Compliance:** Regulatory reporting
- **Security:** Identify vulnerable dependencies
- **License:** License compliance checking

## Supply Chain Security

### Image Signing

```bash
# Enable image signing in docker build
echo '{"auths":{},"signingPreferences":{"signingAlgorithm":"pgp"}}' > ~/.docker/config.json

# Build with signing
docker build -t pushpaka:signed .
docker push pushpaka:signed  # System will sign
```

### Attestation

```bash
# Build with SLSA provenance
docker buildx build \
  --provenance mode=max \
  -t pushpaka:latest .

# Verify attestation
docker buildx imagetools inspect pushpaka:latest
```

## Reporting

### Security Reports

Monthly security report template:

```markdown
# Pushpaka Security Report - March 2026

## Summary
- Total scans: 30
- Vulnerabilities found: 2 (1 HIGH, 1 MEDIUM)
- Critical issues: 0
- Compliance: ✅ Pass

## Vulnerabilities

### HIGH - OpenSSL
- CVE-XXXX-XXXXX
- Status: Fixed in v1.0.1
- Action: Update recommended

### MEDIUM - Node.js
- CVE-YYYY-YYYYY
- Status: Monitoring
- Action: Will be fixed in v1.1.0

## Dependency Updates
- golang: 1.24 → 1.25
- postgresql: 16 → 17
- npm packages: 8 updates

## Upcoming Actions
- Security audit scheduled for Q2 2026
- SBOM generation automated
- License compliance tool added

## Recommendations
- Enable image signing
- Monthly SBOM review
- Quarterly security audit
```

## Tools Installation

### Trivy
```bash
# macOS
brew install trivy

# Linux
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin

# Docker
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy
```

### Grype
```bash
# macOS
brew install grype

# Linux
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin
```

## Continuous Monitoring

### Running Scheduled Scans

```bash
# Add to cron (daily at 2 AM)
0 2 * * * /usr/local/bin/trivy image pushpaka:latest --severity CRITICAL | mail -s "Trivy scan" security@pushpaka.dev
```

### Integration with GitHub Security

1. **Code scanning:** CodeQL, Gitleaks
2. **Dependency scanning:** Trivy, Grype
3. **Container scanning:** Docker Scout
4. **All results:** GitHub Security tab

## Resources

- **Trivy Documentation:** https://github.com/aquasecurity/trivy
- **CVSS Calculator:** https://www.first.org/cvss/calculator/3.1
- **CVE Details:** https://www.cvedetails.com/
- **Docker Security:** https://docs.docker.com/engine/security/

---

**Last Updated:** March 17, 2026  
**Review Schedule:** Monthly
