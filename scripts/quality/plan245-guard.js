#!/usr/bin/env node
/**
 * Plan 245 Guard – 冻结旧命名的新增使用
 * 目的：防止在重构进行中继续引入 PositionDetailQuery / query PositionDetail 等旧命名
 * 策略：对 frontend/src 进行静态计数，建立 baseline，后续若计数上升则失败退出
 *
 * 实施说明：
 * - 首次运行会在 reports/plan245/baseline.json 生成基线并退出 0
 * - 后续运行将比较当前计数与基线；只要出现“上升”则失败
 * - 排除目录：generated/、__tests__/、*.test.*、*.spec.*（避免代码生成与测试干扰）
 * - 严格冻结（error）：GraphQL operation/token（PositionDetailQuery / query PositionDetail）
 * - 软冻结（warn）：类型名（OrganizationUnit / PositionRecord），提示但不阻塞构建
 */

const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..', '..');
const SRC_DIR = path.join(ROOT, 'frontend', 'src');
const BASELINE_DIR = path.join(ROOT, 'reports', 'plan245');
const BASELINE_FILE = path.join(BASELINE_DIR, 'baseline.json');

const COLORS = {
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  green: '\x1b[32m',
  reset: '\x1b[0m',
};

// 需要冻结的标记
const TOKENS = [
  // 严格冻结：GraphQL operation / 代码引用
  { id: 'op.positionDetail.operation', regex: /query\s+PositionDetail\b/g, severity: 'error' },
  { id: 'op.positionDetail.type', regex: /\bPositionDetailQuery\b/g, severity: 'error' },

  // 软冻结：类型名（提示，不阻塞）
  { id: 'type.organizationUnit', regex: /\bOrganizationUnit\b/g, severity: 'warn' },
  { id: 'type.positionRecord', regex: /\bPositionRecord\b/g, severity: 'warn' },
];

// 扫描文件
function collectFiles(dir) {
  const files = [];
  const stack = [dir];
  while (stack.length) {
    const current = stack.pop();
    const entries = fs.readdirSync(current, { withFileTypes: true });
    for (const e of entries) {
      const full = path.join(current, e.name);
      if (e.isDirectory()) {
        // 排除
        if (
          e.name === 'generated' ||
          e.name === '__tests__'
        ) continue;
        stack.push(full);
      } else if (e.isFile()) {
        const ext = path.extname(e.name);
        if (!['.ts', '.tsx', '.js', '.jsx', '.graphql', '.gql'].includes(ext)) continue;
        // 排除测试文件
        if (/\.test\./.test(e.name) || /\.spec\./.test(e.name)) continue;
        files.push(full);
      }
    }
  }
  return files;
}

function countTokens(files) {
  const counts = {};
  for (const t of TOKENS) counts[t.id] = 0;
  for (const f of files) {
    const content = fs.readFileSync(f, 'utf8');
    for (const t of TOKENS) {
      const matches = content.match(t.regex);
      if (matches) counts[t.id] += matches.length;
    }
  }
  return counts;
}

function main() {
  const files = collectFiles(SRC_DIR);
  const current = countTokens(files);

  if (!fs.existsSync(BASELINE_DIR)) fs.mkdirSync(BASELINE_DIR, { recursive: true });

  if (!fs.existsSync(BASELINE_FILE)) {
    const baseline = {
      createdAt: new Date().toISOString(),
      notes: 'Plan 245 baseline counts for legacy tokens. Guard will fail if counts increase.',
      counts: current,
    };
    fs.writeFileSync(BASELINE_FILE, JSON.stringify(baseline, null, 2), 'utf8');
    console.log(`${COLORS.green}✅ Plan 245 baseline created at ${BASELINE_FILE}${COLORS.reset}`);
    console.log('Counts:', baseline.counts);
    process.exit(0);
  }

  const baseline = JSON.parse(fs.readFileSync(BASELINE_FILE, 'utf8'));
  const errors = [];
  const warnings = [];

  for (const t of TOKENS) {
    const before = baseline.counts[t.id] ?? 0;
    const now = current[t.id] ?? 0;
    if (now > before) {
      const msg = `${t.id} increased from ${before} to ${now}`;
      if (t.severity === 'error') errors.push(msg);
      else warnings.push(msg);
    }
  }

  // 输出报告
  if (warnings.length) {
    console.warn(`${COLORS.yellow}⚠️  Plan 245 guard warnings:${COLORS.reset}`);
    for (const w of warnings) console.warn(`  - ${w}`);
  }
  if (errors.length) {
    console.error(`${COLORS.red}❌ Plan 245 guard failed (legacy token count increased):${COLORS.reset}`);
    for (const e of errors) console.error(`  - ${e}`);
    process.exit(2);
  }
  console.log(`${COLORS.green}✅ Plan 245 guard passed (no new legacy tokens).${COLORS.reset}`);
}

main();

