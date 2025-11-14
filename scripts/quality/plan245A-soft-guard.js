#!/usr/bin/env node
/**
 * Plan 245A Soft Guard — 组织详情子组件“旧字段直读”提示
 * 作用：在组织详情相关组件中，提示对零散旧路径字段（name/status/effectiveDate/endDate）的新增直读，建议迁移到统一 Hook 或版本优先路径
 * 特性：仅输出警告，不阻塞 CI（exit code 0）
 */

const fs = require('fs');
const path = require('path');

const ROOT = process.cwd();
const TARGET_DIR = path.join(ROOT, 'frontend', 'src', 'features', 'temporal', 'components');

// 关注文件（组织详情相关）
const FILE_GLOBS = [
  'TemporalMasterDetailHeader.tsx',
  'TemporalMasterDetailAlerts.tsx',
  'TemporalEditForm.tsx',
  'inlineNewVersionForm',
  'ParentOrganizationSelector.tsx',
];

// 关注字段（潜在旧路径直读，版本/统一 Hook 优先）
const SUSPICIOUS_TOKENS = [
  '.name',
  '.status',
  '.effectiveDate',
  '.endDate',
];

function listFiles(dir) {
  const out = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const e of entries) {
    const full = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (FILE_GLOBS.some(g => e.name.includes(g))) {
        out.push(...listFiles(full));
      } else if (e.name.startsWith('inlineNewVersionForm')) {
        out.push(...listFiles(full));
      } else {
        // 递归检查其余子目录
        out.push(...listFiles(full));
      }
    } else if (e.isFile()) {
      if (FILE_GLOBS.some(g => full.includes(g))) {
        out.push(full);
      }
    }
  }
  return out;
}

function scan() {
  if (!fs.existsSync(TARGET_DIR)) {
    console.log('ℹ️  245A soft guard: no target directory found, skipping.');
    return;
  }
  const files = listFiles(TARGET_DIR);
  const warnings = [];
  for (const f of files) {
    const content = fs.readFileSync(f, 'utf8');
    const lines = content.split('\n');
    lines.forEach((line, idx) => {
      const hit = SUSPICIOUS_TOKENS.find(t => line.includes(t));
      if (hit) {
        warnings.push({ file: f, line: idx + 1, text: line.trim() });
      }
    });
  }
  if (warnings.length === 0) {
    console.log('✅ 245A soft guard: no suspicious direct reads detected.');
    return;
  }
  console.log('⚠️  245A soft guard — potential direct reads of legacy fields (informational):');
  warnings.slice(0, 200).forEach(w => {
    console.log(`  - ${path.relative(ROOT, w.file)}:${w.line} :: ${w.text}`);
  });
  if (warnings.length > 200) {
    console.log(`  ... and ${warnings.length - 200} more`);
  }
  console.log('ℹ️  Guidance: Prefer version/timeline data > unified hook record > legacy path (temporary only).');
  // 仅警告，不阻塞
  process.exit(0);
}

scan();

