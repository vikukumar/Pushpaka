#!/usr/bin/env node
// Temporarily removes `export const dynamic = 'force-dynamic'` from the
// dashboard layout before building a static export for Go binary embedding.
//
// `output: 'export'` (Next.js static HTML) cannot coexist with force-dynamic.
// This script patches the file before the build and restores it after, so that
// normal `pnpm build` (Next.js server deployment) still gets force-dynamic.
//
// Usage:
//   node scripts/patch-layout.js remove   # call BEFORE STATIC_EXPORT=1 pnpm build
//   node scripts/patch-layout.js restore  # call AFTER  (success or failure)

const fs = require('fs')
const path = require('path')

const LAYOUT = path.join(__dirname, '../frontend/app/dashboard/layout.tsx')
const BACKUP = LAYOUT + '.bak'

const action = process.argv[2]

if (action === 'remove') {
  const src = fs.readFileSync(LAYOUT, 'utf8')
  fs.writeFileSync(BACKUP, src)
  // Normalise to LF, strip the force-dynamic export line, then write back
  const normalised = src.replace(/\r\n/g, '\n')
  const patched = normalised
    .replace(/\nexport const dynamic = 'force-dynamic'\n/g, '\n')
    // Also strip the preceding explanatory comment block if present
    .replace(/\n\/\/ Force all dashboard pages[\s\S]*?which doesn't support SSR\.\n/g, '\n')
  fs.writeFileSync(LAYOUT, patched)
  console.log('[patch-layout] Removed force-dynamic for static export build')
} else if (action === 'restore') {
  if (fs.existsSync(BACKUP)) {
    fs.copyFileSync(BACKUP, LAYOUT)
    fs.unlinkSync(BACKUP)
    console.log('[patch-layout] Restored dashboard layout')
  } else {
    console.log('[patch-layout] No backup found, nothing to restore')
  }
} else {
  console.error('Usage: patch-layout.js <remove|restore>')
  process.exit(1)
}
