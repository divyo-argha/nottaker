#!/usr/bin/env node

'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs   = require('fs');

const BINARY_DIR = path.join(__dirname, '..', 'bin', 'binaries');

function getPlatformBinary() {
  const platform = process.platform;
  const arch     = process.arch;

  const platformMap = {
    'darwin-arm64': 'octonote-gui-darwin-arm64',
    'darwin-x64':   'octonote-gui-darwin-amd64',
    'linux-arm64':  'octonote-gui-linux-arm64',
    'linux-x64':    'octonote-gui-linux-amd64',
    'win32-x64':    'octonote-gui-windows-amd64.exe',
    'win32-arm64':  'octonote-gui-windows-arm64.exe',
  };

  const key    = `${platform}-${arch}`;
  const binary = platformMap[key];

  if (!binary) {
    console.error(`[octonote] Unsupported platform for GUI: ${key}`);
    process.exit(1);
  }

  return path.join(BINARY_DIR, binary);
}

const binaryPath = getPlatformBinary();

if (!fs.existsSync(binaryPath)) {
  console.error(`[octonote] GUI Binary not found: ${binaryPath}`);
  console.error('  Try reinstalling: npm install -g octonote');
  process.exit(1);
}

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

if (result.error) {
  console.error(`[octonote] Failed to launch GUI: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 0);
