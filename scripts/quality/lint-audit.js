#!/usr/bin/env node

/**
 * lint-audit.js
 *
 * 运行审计模块单元测试，确保关键 DTO / fallback 逻辑未被破坏。
 */

const { execSync } = require('node:child_process');
const path = require('node:path');

const repoRoot = path.resolve(__dirname, '..', '..');

try {
  execSync('go test ./cmd/organization-command-service/internal/audit', {
    cwd: repoRoot,
    stdio: 'inherit',
  });
  console.log('✅ 审计模块校验通过');
} catch (error) {
  const code = typeof error.status === 'number' ? error.status : 1;
  console.error('❌ 审计模块校验失败');
  process.exit(code);
}
