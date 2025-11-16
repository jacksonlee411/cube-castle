#!/usr/bin/env node

/**
 * doc-archive-check.js
 *
 * 确保 docs/development-plans 与 docs/archive/development-plans 之间不存在重复文件，
 * 避免同一计划同时存在于“活跃”和“归档”目录。
 */

const fs = require('node:fs');
const path = require('node:path');

const repoRoot = path.resolve(__dirname, '..', '..');
const activeDir = path.join(repoRoot, 'docs', 'development-plans');
const archivedDir = path.join(repoRoot, 'docs', 'archive', 'development-plans');

const readMarkdownFiles = dir =>
  fs
    .readdirSync(dir)
    .filter(name => name.endsWith('.md'))
    .filter(name => fs.statSync(path.join(dir, name)).isFile());

const activeFiles = new Set(readMarkdownFiles(activeDir));
const archivedFiles = new Set(readMarkdownFiles(archivedDir));

// Treat duplicates as errors, but allow temporary exceptions list.
const exceptions = new Set([
  // TODO-TEMPORARY(2025-11-16): allow co-existence during trunk migration; remove after archival pass
  '06-integrated-teams-progress-log.md',
  '231-outbox-dispatcher-gap.md',
  '240bt-org-detail-blank-page-mitigation.md',
  '252-permission-consistency-and-contract-alignment.md',
]);

const duplicates = [...activeFiles].filter(name => archivedFiles.has(name) && !exceptions.has(name));

if (duplicates.length > 0) {
  console.error('❌ 检测到计划文档同时存在于活跃与归档目录:');
  duplicates.forEach(name => console.error(`  - ${name}`));
  process.exit(1);
}

console.log('✅ 文档计划目录检查通过：活跃/归档无重复文件');
