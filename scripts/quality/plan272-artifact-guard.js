#!/usr/bin/env node
/**
 * Plan 272 运行产物守卫
 * - 校验 logs/reports/test-results README 是否存在
 * - 限制未压缩日志大小（默认 2MB）
 * - 侦测未压缩的 .html/.json 运行产物（默认 512KB 阈值）
 * - 产出 reports/plan272 与 logs/plan272/guard 的执行记录
 */

const fs = require('fs');
const path = require('path');

const ROOT = process.cwd();
const MAX_LOG_BYTES = parseInt(process.env.PLAN272_MAX_LOG_BYTES || `${2 * 1024 * 1024}`, 10);
const MAX_ARTIFACT_BYTES = parseInt(process.env.PLAN272_MAX_ARTIFACT_BYTES || `${512 * 1024}`, 10);
const TARGET_DIRS = ['logs', 'reports', 'test-results'];
const README_NAME = 'README.md';
const ALLOWLIST_FILE = path.join('scripts', 'todo-temporary-allowlist.txt');
const REPORT_DIR = path.join('reports', 'plan272');
const LOG_DIR = path.join('logs', 'plan272', 'guard');

const timestamp = new Date().toISOString().replace(/\.\d+Z$/, 'Z').replace(/[-:]/g, '');
const reportPath = path.join(REPORT_DIR, `plan272-artifact-guard-${timestamp}.txt`);
const logPath = path.join(LOG_DIR, `plan272-guard-${timestamp}.log`);

fs.mkdirSync(REPORT_DIR, { recursive: true });
fs.mkdirSync(LOG_DIR, { recursive: true });

const outputs = [];
const errors = [];

function log(line) {
  console.log(line);
  outputs.push(line);
}

function escapeRegex(str) {
  return str.replace(/[-/\\^$+?.()|[\]{}]/g, '\\$&');
}

function matchesPattern(pattern, target) {
  const regex = new RegExp('^' + escapeRegex(pattern).replace(/\\\*/g, '.*') + '$');
  return regex.test(target);
}

function loadAllowlist() {
  if (!fs.existsSync(ALLOWLIST_FILE)) {
    return [];
  }
  return fs
    .readFileSync(ALLOWLIST_FILE, 'utf-8')
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#') && line.includes('plan272'))
    .map((line) => line.replace(/^.*plan272[:\s-]*/i, '').trim())
    .filter(Boolean);
}

const allowPatterns = loadAllowlist();

function isAllowed(relPath) {
  return allowPatterns.some((pattern) => matchesPattern(pattern, relPath));
}

function walkFiles(dir) {
  const results = [];
  if (!fs.existsSync(dir)) {
    return results;
  }
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      results.push(...walkFiles(fullPath));
    } else {
      results.push(fullPath);
    }
  }
  return results;
}

for (const dir of TARGET_DIRS) {
  const readmePath = path.join(dir, README_NAME);
  if (!fs.existsSync(readmePath)) {
    errors.push(`[README] 缺少 ${readmePath}，请创建并说明保留策略。`);
  }
}

const logFiles = walkFiles(path.join(ROOT, 'logs')).filter((file) => file.endsWith('.log'));
for (const file of logFiles) {
  const stat = fs.statSync(file);
  if (stat.size > MAX_LOG_BYTES) {
    const rel = path.relative(ROOT, file).replace(/\\/g, '/');
    if (!isAllowed(rel)) {
      errors.push(
        `[LOG] ${rel} 大小 ${Math.round(stat.size / 1024)}KB 超过 ${Math.round(
          MAX_LOG_BYTES / 1024
        )}KB，请执行 make archive-run-artifacts 或添加 README/manifest 后压缩。`
      );
    }
  }
}

const artifactExtensions = ['.html', '.json'];
const artifactDirs = ['logs', 'reports', 'test-results'];
for (const base of artifactDirs) {
  for (const file of walkFiles(path.join(ROOT, base))) {
    const ext = path.extname(file).toLowerCase();
    if (!artifactExtensions.includes(ext)) {
      continue;
    }
    const stat = fs.statSync(file);
    if (stat.size > MAX_ARTIFACT_BYTES) {
      const rel = path.relative(ROOT, file).replace(/\\/g, '/');
      if (!isAllowed(rel)) {
        errors.push(
          `[ARTIFACT] ${rel} (${Math.round(stat.size / 1024)}KB) 仍以 ${ext} 明文存在，请归档到 archive/runtime-artifacts 或将其压缩为 .tar.zst。`
        );
      }
    }
  }
}

if (errors.length === 0) {
  log('✅ Plan 272 artifact guard passed: 未发现超阈日志或未压缩产物。');
} else {
  log('❌ Plan 272 artifact guard 检测到以下问题：');
  errors.forEach((err) => log(`  - ${err}`));
}

fs.writeFileSync(reportPath, outputs.join('\n'), 'utf-8');
fs.writeFileSync(logPath, outputs.join('\n'), 'utf-8');

if (errors.length > 0) {
  log(`📄 报告: ${reportPath}`);
  log(`🗂️  日志: ${logPath}`);
  process.exit(1);
} else {
  log(`📄 报告: ${reportPath}`);
  log(`🗂️  日志: ${logPath}`);
}
