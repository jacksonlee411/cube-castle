#!/usr/bin/env node

/**
 * Plan 263 helper: merge duplicate quality:preflight scripts and ensure guard chain stays intact.
 *
 * Actions:
 * 1. Read package.json at repo root.
 * 2. Force scripts["quality:preflight"] to the canonical guard pipeline.
 * 3. Write updated package.json if changes occurred.
 * 4. Log actions to logs/plan263/plan263-quality-preflight-<timestamp>.log.
 */

const fs = require('node:fs');
const path = require('node:path');

const REPO_ROOT = process.cwd();
const LOG_DIR = path.join(REPO_ROOT, 'logs', 'plan263');
const pkgPath = path.join(REPO_ROOT, 'package.json');
const timestamp = new Date().toISOString().replace(/[-:]/g, '').replace(/\..+$/, '');
const logPath = path.join(LOG_DIR, `plan263-quality-preflight-${timestamp}.log`);

fs.mkdirSync(LOG_DIR, { recursive: true });
const logStream = fs.createWriteStream(logPath, { flags: 'a' });

const log = (message) => {
  const line = `[${new Date().toISOString()}] ${message}`;
  logStream.write(`${line}\n`);
  console.log(line);
};

const guardChain = [
  'node scripts/quality/document-sync.js',
  'npm run guard:selectors-246',
  'npm --prefix frontend run lint',
  'npm run guard:fields',
  'node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden',
  'npm run lint:docs',
];

const canonicalScript = guardChain.join(' && ');

function main() {
  log('Plan263 quality preflight merge started');
  const pkgRaw = fs.readFileSync(pkgPath, 'utf-8');
  let pkgJson;

  try {
    pkgJson = JSON.parse(pkgRaw);
  } catch (error) {
    log(`❌ Failed to parse package.json: ${error.message}`);
    process.exitCode = 1;
    return;
  }

  pkgJson.scripts ??= {};
  const current = pkgJson.scripts['quality:preflight'];

  if (current === canonicalScript) {
    log('ℹ️ quality:preflight already matches canonical guard chain, no changes required.');
  } else {
    pkgJson.scripts['quality:preflight'] = canonicalScript;
    const newContent = `${JSON.stringify(pkgJson, null, 2)}\n`;
    fs.writeFileSync(pkgPath, newContent, 'utf-8');
    log(`✅ quality:preflight updated. Previous value: ${current ?? '<missing>'}`);
    log('✅ Canonical guard chain applied successfully.');
  }

  log('Plan263 quality preflight merge completed.');
}

main();
