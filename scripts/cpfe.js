#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const root = path.join(__dirname, "..");
const src = path.join(root, "frontend", "out");
const dst = path.join(root, "backend", "ui", "dist");

if (!fs.existsSync(src)) {
  console.error(`ERROR: ${src} does not exist. Run 'pnpm build' in frontend/ first.`);
  process.exit(1);
}

if (fs.existsSync(dst)) {
  fs.rmSync(dst, { recursive: true });
}
fs.mkdirSync(dst, { recursive: true });

copyDir(src, dst);
console.log(`Copied ${src} → ${dst}`);

function copyDir(from, to) {
  fs.mkdirSync(to, { recursive: true });
  for (const entry of fs.readdirSync(from, { withFileTypes: true })) {
    const s = path.join(from, entry.name);
    const d = path.join(to, entry.name);
    if (entry.isDirectory()) {
      copyDir(s, d);
    } else {
      fs.copyFileSync(s, d);
    }
  }
}
