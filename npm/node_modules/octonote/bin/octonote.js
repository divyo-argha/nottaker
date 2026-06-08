#!/usr/bin/env node

'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs   = require('fs');
const os   = require('os');

const BINARY_DIR = path.join(__dirname, '..', 'bin', 'binaries');

function getPlatformBinary() {
  const platform = process.platform;
  const arch     = process.arch;

  const platformMap = {
    'darwin-arm64': 'octonote-darwin-arm64',
    'darwin-x64':   'octonote-darwin-amd64',
    'linux-arm64':  'octonote-linux-arm64',
    'linux-x64':    'octonote-linux-amd64',
    'win32-x64':    'octonote-windows-amd64.exe',
    'win32-arm64':  'octonote-windows-arm64.exe',
  };

  const key    = `${platform}-${arch}`;
  const binary = platformMap[key];

  if (!binary) {
    console.error(`[octonote] Unsupported platform: ${key}`);
    console.error('  Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-x64');
    process.exit(1);
  }

  return path.join(BINARY_DIR, binary);
}

const binaryPath = getPlatformBinary();

if (!fs.existsSync(binaryPath)) {
  console.error(`[octonote] Binary not found: ${binaryPath}`);
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
  console.error(`[octonote] Failed to launch: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 0);
