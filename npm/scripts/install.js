#!/usr/bin/env node

'use strict';

const https  = require('https');
const http   = require('http');
const fs     = require('fs');
const path   = require('path');
const { pipeline } = require('stream');
const { promisify } = require('util');
const streamPipeline = promisify(pipeline);

const REPO_OWNER   = 'nottaker';
const REPO_NAME    = 'nottaker';
const VERSION      = require('../package.json').version;
const RELEASE_TAG  = `v${VERSION}`;
const GITHUB_BASE  = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${RELEASE_TAG}`;

const BIN_DIR = path.join(__dirname, '..', 'bin', 'binaries');

const PLATFORM_MAP = {
  'darwin-arm64': 'nottaker-darwin-arm64',
  'darwin-x64':   'nottaker-darwin-amd64',
  'linux-arm64':  'nottaker-linux-arm64',
  'linux-x64':    'nottaker-linux-amd64',
  'win32-x64':    'nottaker-windows-amd64.exe',
  'win32-arm64':  'nottaker-windows-arm64.exe',
};

function log(msg)  { process.stdout.write(`[nottaker] ${msg}\n`); }
function warn(msg) { process.stderr.write(`[nottaker] WARN: ${msg}\n`); }
function fail(msg) { process.stderr.write(`[nottaker] ERROR: ${msg}\n`); process.exit(1); }

function followRedirects(url, maxRedirects = 10) {
  return new Promise((resolve, reject) => {
    const attempt = (currentUrl, remaining) => {
      if (remaining <= 0) return reject(new Error('Too many redirects'));

      const lib = currentUrl.startsWith('https') ? https : http;
      const req = lib.get(currentUrl, { headers: { 'User-Agent': 'nottaker-installer/1.0' } }, res => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          res.resume();
          attempt(res.headers.location, remaining - 1);
        } else {
          resolve(res);
        }
      });
      req.on('error', reject);
      req.setTimeout(30_000, () => { req.destroy(); reject(new Error('Request timed out')); });
    };
    attempt(url, maxRedirects);
  });
}

async function download(url, destPath) {
  log(`Downloading ${url}`);
  log(`         → ${destPath}`);

  const res = await followRedirects(url);

  if (res.statusCode !== 200) {
    res.resume();
    throw new Error(`HTTP ${res.statusCode} for ${url}`);
  }

  const totalBytes = parseInt(res.headers['content-length'] || '0', 10);
  let received = 0;

  res.on('data', chunk => {
    received += chunk.length;
    if (totalBytes > 0) {
      const pct = Math.round((received / totalBytes) * 100);
      process.stdout.write(`\r[nottaker]   ${pct}% (${formatBytes(received)} / ${formatBytes(totalBytes)})   `);
    }
  });

  const tmpPath = destPath + '.tmp';
  const writeStream = fs.createWriteStream(tmpPath);
  await streamPipeline(res, writeStream);

  process.stdout.write('\n');

  fs.renameSync(tmpPath, destPath);
}

function formatBytes(bytes) {
  if (bytes < 1024)         return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
}

async function main() {
  const platform = process.platform;
  const arch     = process.arch;
  const key      = `${platform}-${arch}`;
  const binary   = PLATFORM_MAP[key];

  if (!binary) {
    warn(`Platform "${key}" is not pre-built. Skipping binary download.`);
    warn('You can build from source: go build ./tui/ → rename to nottaker');
    process.exit(0);
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });

  const destPath   = path.join(BIN_DIR, binary);
  const downloadUrl = `${GITHUB_BASE}/${binary}`;

  if (fs.existsSync(destPath)) {
    log(`Binary already present: ${destPath}`);
    ensureExecutable(destPath);
    log('nottaker is ready. Run: nottaker');
    return;
  }

  try {
    await download(downloadUrl, destPath);
    ensureExecutable(destPath);
    log(`✓ Installed nottaker ${VERSION} for ${key}`);
    log('Run: nottaker');
  } catch (err) {
    warn(`Could not download binary: ${err.message}`);
    warn('To build from source:');
    warn('  git clone https://github.com/nottaker/nottaker');
    warn('  cd nottaker && go build -o bin/nottaker ./tui/');
  }
}

function ensureExecutable(filePath) {
  if (process.platform !== 'win32') {
    fs.chmodSync(filePath, 0o755);
  }
}

main().catch(err => {
  warn(err.message);
  process.exit(0);
});
