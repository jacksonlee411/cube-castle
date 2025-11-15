'use strict';
/**
 * Preflight guard for Field Standard
 * - Check Goose migration add/drop symmetry for newly added fields
 * - Validate bcId in suggested patches (if provided)
 * - Warn if there are pending contract patches
 */
const fs = require('fs');
const path = require('path');

function scanMigrations(root) {
  const dir = path.join(root, 'database', 'migrations');
  if (!fs.existsSync(dir)) return [];
  return fs
    .readdirSync(dir)
    .filter((f) => f.endsWith('.sql'))
    .map((f) => path.join(dir, f));
}

function read(p) {
  try { return fs.readFileSync(p, 'utf8'); } catch { return ''; }
}

function ok(msg) { console.log('✅', msg); }
function warn(msg) { console.warn('⚠️ ', msg); }
function fail(msg) { console.error('❌', msg); }

function loadBcIds(root) {
  const p = path.join(root, 'scripts', 'fields', 'presets.json');
  if (!fs.existsSync(p)) return [];
  try {
    const j = JSON.parse(fs.readFileSync(p, 'utf8'));
    return Array.isArray(j.bcIds) ? j.bcIds : [];
  } catch { return []; }
}

function validateBcIdInPatches(root, bcIds) {
  const outDir = path.join(root, 'scripts', 'fields', 'out');
  if (!fs.existsSync(outDir)) return { ok: true, message: 'no suggested patches' };
  const files = fs.readdirSync(outDir).filter((f) => f.endsWith('.README.txt'));
  let allOk = true;
  for (const f of files) {
    const t = read(path.join(outDir, f));
    const m = t.match(/^bcId:\s*(.*)$/m);
    if (m && m[1] && m[1] !== '(not provided)') {
      const v = m[1].trim();
      if (!bcIds.includes(v)) {
        allOk = false;
        fail(`bcId "${v}" in ${f} not in allowed list: ${bcIds.join(', ')}`);
      }
    }
  }
  return { ok: allOk };
}

function main() {
  const root = path.resolve(__dirname, '..', '..');
  let errorCount = 0;

  // 1) Check migration symmetry
  const mfiles = scanMigrations(root);
  let symmetryOk = true;
  for (const f of mfiles) {
    const t = read(f);
    if (t.includes('ADD COLUMN') && !t.includes('DROP COLUMN')) {
      symmetryOk = false;
      fail(`Migration add/drop asymmetry: ${path.relative(root, f)}`);
    }
  }
  if (symmetryOk) ok('Migration add/drop symmetry OK');
  else errorCount++;

  // 2) Validate bcId in suggested patches
  const bcIds = loadBcIds(root);
  const v = validateBcIdInPatches(root, bcIds);
  if (v.ok) ok('bcId validation OK (suggested patches)');
  else errorCount++;

  // 3) Warn on pending patches
  const outDir = path.join(root, 'scripts', 'fields', 'out');
  if (fs.existsSync(outDir)) {
    const pending = fs.readdirSync(outDir).filter((f) => f.endsWith('.patch'));
    if (pending.length > 0) {
      warn(`There are ${pending.length} pending contract patches in scripts/fields/out/. Remember to apply/review.`);
    } else {
      ok('No pending contract patches');
    }
  }

  if (errorCount > 0) {
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

