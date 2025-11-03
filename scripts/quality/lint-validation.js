#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');

const repoRoot = path.resolve(__dirname, '..', '..');

const checks = [
  {
    file: path.join(repoRoot, 'frontend', 'src', 'shared', 'validation', 'schemas.ts'),
    mustInclude: [
      'OrganizationConstraints.codePattern',
      'OrganizationConstraints.parentCodePattern',
      'OrganizationConstraints.levelMax',
    ],
    description: '前端校验需直接引用契约生成的约束常量',
  },
  {
    file: path.join(repoRoot, 'cmd', 'hrms-server', 'command', 'internal', 'utils', 'validation.go'),
    mustInclude: [
      'organizationCodeRegex',
      'validUnitTypes',
    ],
    description: '后端校验需使用契约导出的正则与枚举',
  },
  {
    file: path.join(repoRoot, 'cmd', 'hrms-server', 'command', 'internal', 'repository', 'organization_hierarchy.go'),
    mustInclude: [
      'types.OrganizationLevelMax',
    ],
    description: '层级计算需遵守契约中的最大层级常量',
  },
];

try {
  for (const check of checks) {
    const content = fs.readFileSync(check.file, 'utf-8');
    for (const token of check.mustInclude) {
      if (!content.includes(token)) {
        throw new Error(`${check.description}: 缺少 ${token} (${check.file})`);
      }
    }
  }
  console.log('✅ 契约校验引用检查通过');
} catch (error) {
  console.error('❌ 契约校验引用检查失败:', error.message);
  process.exit(1);
}
