// Package ui embeds the compiled Next.js static export so it can be served
// directly from the Go binary without any external files.
//
// The dist/ directory is populated by `make build` (pnpm build -> copy out/).
// In dev mode (make dev) the directory contains only a .gitkeep placeholder,
// so no frontend is served from the binary during local development.
package ui

import "embed"

//go:embed all:dist
var FS embed.FS
