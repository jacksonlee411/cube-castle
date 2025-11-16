#!/usr/bin/env node
/**
 * Plan 257 - 领域 Facade 覆盖率扫描
 * 分母：业务代码中直连 unified-client 或 fetch/axios 的调用点（按文件计数）
 * 分子：业务代码中通过 facade（@/shared/api/facade/*）的调用点（按文件计数）
 * 输出：reports/facade/coverage.json；当 --threshold 未达标时退出码非0
 *
 * 说明：
 * - 近似度量：以“文件粒度存在调用点”计数，避免复杂 AST 依赖；后续可平滑升级
 * - 排除：tests、scripts、统一客户端实现与 facade 目录本身
 */
const fs = require('fs');
const path = require('path');

const projectRoot = process.cwd();
const FRONTEND_SRC = path.join(projectRoot, 'frontend', 'src');
const REPORT_DIR = path.join(projectRoot, 'reports', 'facade');
const LOG_DIR = path.join(projectRoot, 'logs', 'plan257');
const REPORT_FILE = path.join(REPORT_DIR, 'coverage.json');

const argv = process.argv.slice(2);
const thresholdArg = argv.find(a => a.startsWith('--threshold='));
const THRESHOLD = thresholdArg ? Number(thresholdArg.split('=')[1]) : 0.8;
const FACADE_EXPORTS = [
  'getOrganizationByCode',
  'listOrganizationVersions',
  'createOrganization',
  'updateOrganization',
  'activateOrganization',
  'suspendOrganization',
];

function scanFiles(dir) {
  const files = [];
  for (const item of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, item.name);
    if (item.isDirectory()) {
      // 排除目录
      const rel = path.relative(FRONTEND_SRC, full).replace(/\\/g, '/');
      if (rel.startsWith('__tests__') || rel.includes('/__tests__/')) continue;
      if (rel.startsWith('tests') || rel.includes('/tests/')) continue;
      if (rel.startsWith('scripts') || rel.includes('/scripts/')) continue;
      if (rel.startsWith('shared/api')) continue; // API 实现层不计入（业务调用面）
      scanFiles(full).forEach(f => files.push(f));
    } else if (item.isFile() && /\.(ts|tsx)$/.test(item.name)) {
      const rel = path.relative(FRONTEND_SRC, full).replace(/\\/g, '/');
      if (rel.startsWith('shared/api/')) continue; // 双保险：文件级别也排除
      files.push(full);
    }
  }
  return files;
}

function read(file) {
  try { return fs.readFileSync(file, 'utf8'); } catch { return ''; }
}

function classifyModule(fileRel) {
  // 以 features/ 下一级目录作为模块名；否则归为 shared/others
  const m = fileRel.match(/^features\/([^/]+)/);
  return m ? m[1] : (fileRel.startsWith('features/') ? 'features' : 'others');
}

function main() {
  if (!fs.existsSync(FRONTEND_SRC)) {
    console.error('frontend/src not found, skip.');
    process.exit(0);
  }
  const files = scanFiles(FRONTEND_SRC);

  let numeratorFiles = new Set();   // 使用了 facade 的文件
  let denominatorFiles = new Set(); // 直连 unified-client 或 fetch/axios 的文件

  const facadeImportRe = /from\s+['"]@?\/?shared\/api\/facade\/[^'"]+['"]/;
  const facadeFromIndexRe = /import\s*{\s*([^}]+)\s*}\s*from\s*['"]@?\/?shared\/api['"]/g;
  const unifiedImportRe = /from\s+['"]@?\/?shared\/api\/unified-client['"]/;
  const unifiedUsageRe = /\bunifiedRESTClient\b|\bunifiedGraphQLClient\b/;
  const fetchRe = /\bfetch\s*\(/;
  const axiosRe = /from\s+['"]axios['"]|axios\s*\(/;

  const perModule = {};

  for (const file of files) {
    const rel = path.relative(path.join(projectRoot, 'frontend', 'src'), file).replace(/\\/g, '/');
    const text = read(file);
    const module = classifyModule(rel);
    if (!perModule[module]) perModule[module] = { numerator: 0, denominator: 0, files: [] };

    const usesFacadeDirect = facadeImportRe.test(text);
    // 识别从 '@/shared/api' 导入的 Facade 成员
    let usesFacadeViaIndex = false;
    for (const m of text.matchAll(facadeFromIndexRe)) {
      const names = (m[1] || '')
        .split(',')
        .map(s => s.trim().split(/\s+as\s+/i)[0].trim())
        .filter(Boolean);
      if (names.some(n => FACADE_EXPORTS.includes(n))) {
        usesFacadeViaIndex = true;
        break;
      }
    }
    const usesFacade = usesFacadeDirect || usesFacadeViaIndex;
    const usesUnified = unifiedImportRe.test(text) || unifiedUsageRe.test(text);
    const usesDirectFetch = fetchRe.test(text);
    const usesAxios = axiosRe.test(text);

    if (usesFacade) {
      numeratorFiles.add(rel);
      perModule[module].numerator++;
    }
    if (usesUnified || usesDirectFetch || usesAxios) {
      denominatorFiles.add(rel);
      perModule[module].denominator++;
    }
    if (usesFacade || usesUnified || usesDirectFetch || usesAxios) {
      perModule[module].files.push(rel);
    }
  }

  const numerator = numeratorFiles.size;
  const denominator = denominatorFiles.size || 1; // avoid div by zero
  const coverage = +(numerator / denominator).toFixed(4);

  const moduleCoverage = {};
  const offenders = [];
  for (const [mod, stat] of Object.entries(perModule)) {
    const denom = stat.denominator || 1;
    moduleCoverage[mod] = +(stat.numerator / denom).toFixed(4);
    // 记录 offenders：在该模块内统计 denominator 且未纳入 numerator 的文件
    const files = (stat.files || []);
    for (const f of files) {
      const isDen = denominatorFiles.has(f);
      const isNum = numeratorFiles.has(f);
      if (isDen && !isNum) {
        offenders.push({ module: mod, file: f });
      }
    }
  }

  // 输出
  fs.mkdirSync(REPORT_DIR, { recursive: true });
  fs.mkdirSync(LOG_DIR, { recursive: true });
  const out = {
    coverage,
    numerator,
    denominator,
    threshold: THRESHOLD,
    byModule: moduleCoverage,
    offenders,
    timestamp: new Date().toISOString()
  };
  fs.writeFileSync(REPORT_FILE, JSON.stringify(out, null, 2));
  // 备份一份到 logs/plan257
  const ts = new Date().toISOString().replace(/[:.]/g, '');
  fs.writeFileSync(path.join(LOG_DIR, `coverage-${ts}.json`), JSON.stringify(out, null, 2));

  // 控制台摘要
  console.log(`Plan 257 Facade Coverage: ${coverage} (numerator=${numerator}, denominator=${denominator}, threshold=${THRESHOLD})`);
  console.log(`Per module: ${JSON.stringify(moduleCoverage)}`);
  if (coverage < THRESHOLD) {
    console.error(`Coverage below threshold (${coverage} < ${THRESHOLD})`);
    process.exit(1);
  }
}

main();
