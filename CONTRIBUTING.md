# Contributing To Pushpaka

## Overview

Pushpaka is a self-hosted deployment platform with backend, worker, frontend, website, and infrastructure surfaces. Contributions should preserve operational clarity, product coherence, and deployability.

## Before You Start

Read:
- [README.md](README.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/local-dev.md](docs/local-dev.md)

Check for existing work first:
- Issues for bugs and planned work
- Discussions for design and product questions

## Development Setup

### Backend And Worker

```bash
go work sync
go build -C cmd/pushpaka .
go build -C worker .
```

### Dev Mode

```bash
cd cmd/pushpaka
go build -o pushpaka .
./pushpaka -dev
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

### Website

```bash
cd website
npm install
npm run build
```

## Contribution Rules

- keep changes focused and scoped
- preserve existing product behavior unless the change explicitly updates it
- do not revert unrelated user changes in the working tree
- update documentation when changing product behavior
- prefer practical fixes over speculative refactors

## Code Style

### Go
- keep formatting consistent with `gofmt`
- avoid dead code and hidden side effects
- make error paths explicit

### Frontend And Website
- preserve existing visual language unless intentionally redesigning
- keep responsive behavior intact
- avoid placeholder marketing copy that does not match real product behavior

## Testing And Validation

Run what applies to your change:

```bash
go build -C cmd/pushpaka .
go build -C worker .
go vet ./...
```

```bash
cd frontend
pnpm lint
pnpm build
```

```bash
cd website
npm run build
```

If you cannot run a required check, explain why in your PR.

## Pull Requests

A good PR should include:
- what changed
- why it changed
- how it was validated
- any migration or operational impact

Also include:
- screenshots for UI changes
- workflow notes for CI or release changes
- API notes when request or response behavior changes

## Documentation Expectations

Update the relevant files when behavior changes:
- `README.md` for product-level understanding
- `docs/` for platform or operational detail
- website docs for customer-facing usage flows

## Security And Secrets

- never commit real credentials
- redact tokens and secrets from screenshots and logs
- use placeholders in examples
- follow [SECURITY.md](SECURITY.md) for vulnerability reporting
