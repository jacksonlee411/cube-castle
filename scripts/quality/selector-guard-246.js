#!/usr/bin/env node
/**
 * Plan 246 - Selector Guard
 * Freeze new usages of legacy testids prefixed with "organization-" or "position-".
 *
 * Strategy:
 * - Scan frontend sources (src + tests) and count occurrences of legacy selector patterns
 * - On first run, write baseline to reports/plan246/baseline.json and exit(0)
 * - On subsequent runs, if total count increases (any pattern), fail (exit 2)
 * - Allowlist: reports/plan246/allowlist.txt (one substring per line, matched against absolute file path)
 *
 * Covered patterns (to match real-world usage):
 *  - JSX attributes: data-testid="organization-..." | data-testid='position-...'
 *  - Attribute locators: [data-testid="organization-..."]  and prefix: [data-testid^="position-..."]
 *  - Testing APIs: getByTestId('organization-...')  and template literal head: getByTestId(`position-${...}`)
 */

const fs = require('fs');
const path = require('path');

const ROOT = path.resolve(__dirname, '..', '..');
const FRONTEND_DIR = path.join(ROOT, 'frontend');
const SRC_DIR = path.join(FRONTEND_DIR, 'src');
const TESTS_DIR = path.join(FRONTEND_DIR, 'tests');
const REPORT_DIR = path.join(ROOT, 'reports', 'plan246');
const BASELINE_FILE = path.join(REPORT_DIR, 'baseline.json');
const ALLOWLIST_FILE = path.join(REPORT_DIR, 'allowlist.txt');

const COLORS = {
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  green: '\x1b[32m',
  blue: '\x1b[34m',
  reset: '\x1b[0m',
};

const STRICT = process.env.SELECTOR_GUARD_STRICT !== '0';

// Regex patterns
const PATTERNS = [
  {
    id: 'jsx-attr-data-testid',
    desc: 'JSX data-testid attribute',
    // data-testid="organization-..."  or  data-testid='position-...'
    regex: /data-testid\s*=\s*["'](organization|position)-/g,
  },
  {
    id: 'locator-attr-eq',
    desc: 'Attribute locator equals',
    // [data-testid="organization-..."]
    regex: /\[data-testid\s*=\s*["'`](organization|position)-/g,
  },
  {
    id: 'locator-attr-prefix',
    desc: 'Attribute locator prefix',
    // [data-testid^="position-..."]
    regex: /\[data-testid\^\s*=\s*["'`](organization|position)-/g,
  },
  {
    id: 'getByTestId-literal',
    desc: 'getByTestId literal',
    // getByTestId('organization-...') or "position-..."
    regex: /getByTestId\s*\(\s*["'`](organization|position)-/g,
  },
  {
    id: 'getByTestId-template-head',
    desc: 'getByTestId template head',
    // getByTestId(`organization-${...}`) or position-...
    regex: /getByTestId\s*\(\s*`(organization|position)-\$\{/g,
  },
];

function readAllowlist() {
  try {
    const text = fs.readFileSync(ALLOWLIST_FILE, 'utf8');
    return text
      .split(/\r?\n/)
      .map((s) => s.trim())
      .filter(Boolean);
  } catch {
    return [];
  }
}

function isAllowed(filePath, allow) {
  return allow.some((fragment) => filePath.includes(fragment));
}

function scanFiles(dir) {
  const files = [];
  const stack = [dir];
  while (stack.length) {
    const cur = stack.pop();
    if (!fs.existsSync(cur)) continue;
    const entries = fs.readdirSync(cur, { withFileTypes: true });
    for (const e of entries) {
      const abs = path.join(cur, e.name);
      if (e.isDirectory()) {
        if (['node_modules', 'dist', 'build', '.git', 'coverage', '.cache'].includes(e.name)) continue;
        stack.push(abs);
      } else if (e.isFile()) {
        const ext = path.extname(e.name);
        if (['.ts', '.tsx', '.js', '.jsx'].includes(ext)) files.push(abs);
      }
    }
  }
  return files;
}

function countLegacyUsages(files, allow) {
  const perPattern = {};
  const perFile = {};
  PATTERNS.forEach((p) => (perPattern[p.id] = 0));

  for (const f of files) {
    if (isAllowed(f, allow)) continue;
    const text = fs.readFileSync(f, 'utf8');
    for (const p of PATTERNS) {
      const matches = text.match(p.regex);
      if (matches && matches.length > 0) {
        perPattern[p.id] += matches.length;
        if (!perFile[f]) perFile[f] = [];
        perFile[f].push({ pattern: p.id, count: matches.length });
      }
    }
  }
  const total = Object.values(perPattern).reduce((a, b) => a + b, 0);
  return { perPattern, perFile, total };
}

function main() {
  const allow = readAllowlist();
  const files = [...scanFiles(SRC_DIR), ...scanFiles(TESTS_DIR)];
  const current = countLegacyUsages(files, allow);

  if (!fs.existsSync(REPORT_DIR)) fs.mkdirSync(REPORT_DIR, { recursive: true });

  // First run: create baseline
  if (!fs.existsSync(BASELINE_FILE)) {
    const baseline = {
      createdAt: new Date().toISOString(),
      notes:
        'Plan 246 selector guard baseline. Fails if legacy selector usage increases. Use reports/plan246/allowlist.txt to silence specific paths.',
      perPattern: current.perPattern,
      total: current.total,
    };
    fs.writeFileSync(BASELINE_FILE, JSON.stringify(baseline, null, 2), 'utf8');
    console.log(`${COLORS.green}✅ Plan 246 selector baseline created at ${BASELINE_FILE}${COLORS.reset}`);
    console.log(`${COLORS.blue}ℹ️  Total legacy usages: ${current.total}${COLORS.reset}`);
    process.exit(0);
  }

  const baseline = JSON.parse(fs.readFileSync(BASELINE_FILE, 'utf8'));
  const deltas = {};
  let increased = false;
  for (const id of Object.keys(current.perPattern)) {
    const before = baseline.perPattern[id] ?? 0;
    const now = current.perPattern[id] ?? 0;
    deltas[id] = { before, now, delta: now - before };
    if (now > before) increased = true;
  }

  // Print summary
  console.log(`${COLORS.blue}ℹ️  Plan 246 selector guard summary${COLORS.reset}`);
  console.log(`   Baseline total: ${baseline.total}  |  Current total: ${current.total}`);
  Object.entries(deltas).forEach(([id, d]) => {
    const sign = d.delta > 0 ? '+' : '';
    const color = d.delta > 0 ? COLORS.red : COLORS.green;
    console.log(`   ${id}: ${d.before} -> ${d.now} (${color}${sign}${d.delta}${COLORS.reset})`);
  });

  if (increased) {
    // Helpful details
    console.error(`${COLORS.red}❌ Legacy selector usage increased. Please migrate to temporalEntitySelectors.${COLORS.reset}`);
    // List top offenders
    const offenders = Object.entries(current.perFile)
      .map(([file, arr]) => ({ file, count: arr.reduce((s, x) => s + x.count, 0) }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 20);
    offenders.forEach((o) => console.error(`   - ${o.file} (${o.count})`));
    if (STRICT) {
      process.exit(2);
    } else {
      console.warn(`${COLORS.yellow}⚠️  STRICT=0: downgrade failure to warning${COLORS.reset}`);
      process.exit(0);
    }
  } else {
    console.log(`${COLORS.green}✅ Plan 246 selector guard passed (no new legacy selector usages).${COLORS.reset}`);
    process.exit(0);
  }
}

try {
  main();
} catch (err) {
  console.error(`${COLORS.red}Selector guard failed: ${err.message}${COLORS.reset}`);
  process.exit(1);
}

