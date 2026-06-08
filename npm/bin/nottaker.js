#!/usr/bin/env node
/**
 * nottaker — bin/nottaker.js
 *
 * Thin shim: resolves the platform binary that was downloaded by scripts/install.js
 * during `npm install`, then spawns it — forwarding all stdio and process signals.
 *
 * Usage: npx nottaker  OR  nottaker  (when installed globally)
 */

'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs   = require('fs');
const os   = require('os');

const BINARY_DIR = path.join(__dirname, '..', 'bin', 'binaries');

// ── Resolve binary path ────────────────────────────────────────────────────
function getPlatformBinary() {
  const platform = process.platform; // 'darwin' | 'linux' | 'win32'
  const arch     = process.arch;     // 'x64' | 'arm64' | 'ia32'

  const platformMap = {
    'darwin-arm64': 'nottaker-darwin-arm64',
    'darwin-x64':   'nottaker-darwin-amd64',
    'linux-arm64':  'nottaker-linux-arm64',
    'linux-x64':    'nottaker-linux-amd64',
    'win32-x64':    'nottaker-windows-amd64.exe',
    'win32-arm64':  'nottaker-windows-arm64.exe',
  };

  const key    = `${platform}-${arch}`;
  const binary = platformMap[key];

  if (!binary) {
    console.error(`[nottaker] Unsupported platform: ${key}`);
    console.error('  Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-x64');
    process.exit(1);
  }

  return path.join(BINARY_DIR, binary);
}

// ── Spawn binary ───────────────────────────────────────────────────────────
const binaryPath = getPlatformBinary();

if (!fs.existsSync(binaryPath)) {
  console.error(`[nottaker] Binary not found: ${binaryPath}`);
  console.error('  Try reinstalling: npm install -g nottaker');
  process.exit(1);
}

// Ensure the binary is executable (Linux/macOS).
if (process.platform !== 'win32') {
  try {
    fs.accessSync(binaryPath, fs.constants.X_OK);
  } catch (_) {
    fs.chmodSync(binaryPath, 0o755);
  }
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  env: process.env,
});

// Forward exit code.
if (result.error) {
  console.error(`[nottaker] Failed to launch: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 0);
