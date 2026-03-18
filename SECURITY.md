# Security Policy

## Supported Scope

This policy applies to the Pushpaka repository, its application code, its release assets, and its deployment workflows.

Pushpaka includes:
- backend API services
- worker runtime
- frontend dashboard
- website and documentation
- Helm chart and container delivery workflows

## Reporting A Security Issue

Do not open public issues for suspected vulnerabilities.

Use one of the following approaches:
- open a private security advisory in GitHub if repository settings allow it
- use the security issue template if a private reporting path is configured
- contact the maintainers through the repository security contact process before public disclosure

When reporting, include:
- affected area or file path
- clear reproduction steps
- expected impact
- whether credentials, tokens, or tenant data could be exposed
- logs or screenshots with secrets removed

## What To Include

A strong report usually contains:
- exact version or commit
- deployment mode used
- whether the issue affects dev mode, all-in-one mode, or split mode
- whether Docker, Kubernetes, or direct runtime is involved
- whether AI, terminal, editor, or webhook flows are involved

## Response Expectations

The maintainers will try to:
- acknowledge the report
- reproduce and validate the issue
- determine severity and blast radius
- prepare remediation
- coordinate disclosure timing if the issue is confirmed

No formal SLA is promised in this repository.

## Security Areas In The Platform

Pushpaka currently includes protections around:
- JWT and API-key-based access
- bcrypt password hashing
- secure headers
- rate limiting
- configurable CORS
- redaction of sensitive Git credentials
- workflow and scan-based repository security checks

## Security Best Practices For Operators

When running Pushpaka:
- set strong values for `JWT_SECRET`
- protect PostgreSQL and Redis credentials
- restrict Docker and Kubernetes host access
- use TLS for public endpoints
- rotate Git tokens and AI provider keys regularly
- limit who can use editor and terminal capabilities
- review webhooks, OAuth, and notification integrations carefully
