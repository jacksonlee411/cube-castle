/**
 * ESLint API合规性配置 - Flat Config 版本
 */

const path = require('path');

module.exports = {
  root: true,
  ignorePatterns: ['**/node_modules/**', '**/dist/**', '**/build/**', '**/coverage/**'],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
    warnOnUnsupportedTypeScriptVersion: true,
    project: [
      path.join(__dirname, 'tsconfig.app.json'),
      path.join(__dirname, 'tsconfig.node.json'),
      path.join(__dirname, 'tsconfig.stories.json'),
    ],
    tsconfigRootDir: __dirname,
  },
  plugins: ['@typescript-eslint', 'react-refresh'],
  extends: [],
  rules: {
    // 🚨 所有日志输出必须通过 shared/utils/logger.ts（桥接层含 eslint-disable 说明）
    'no-console': 'error',
    '@typescript-eslint/no-unused-vars': 'off',
    'react-refresh/only-export-components': 'off',
    // 行级例外需注明原因，详见 Plan 20 桥接清单
    camelcase: ['error', { properties: 'always' }],
    'no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: [
              '**/features/positions/timelineAdapter',
              '**/features/positions/timelineAdapter.ts',
              '**/features/positions/statusMeta',
              '**/features/positions/statusMeta.ts',
            ],
            message:
              '🚨 Temporal Entity 命名已统一：请使用 "@/features/temporal/entity/timelineAdapter" 和 "@/features/temporal/entity/statusMeta"。',
          },
        ],
      },
    ],
  },
  // 例外：选择器唯一事实来源文件允许定义字面量 testid
  overrides: [
    {
      // 应用源码：禁止硬编码 data-testid（测试与工具除外）
      files: ['src/**/*.{ts,tsx}'],
      excludedFiles: ['src/**/*.test.{ts,tsx}', 'src/**/*.spec.{ts,tsx}'],
      rules: {
        'no-restricted-syntax': [
          'error',
          {
            selector: 'JSXAttribute[name.name="data-testid"][value.type="Literal"]',
            message:
              'Do not hard-code data-testid. Import from "@/shared/testids/temporalEntity" (temporalEntitySelectors).',
          },
        ],
      },
    },
    {
      files: ['src/shared/testids/temporalEntity.ts'],
      rules: {
        'no-restricted-syntax': 'off',
      },
    },
    {
      // 测试与 Playwright 目录暂不强制 testid 字面量限制（逐步迁移至 SSoT 选择器）
      files: ['tests/**/*.{ts,tsx}', 'src/**/*.test.{ts,tsx}', 'src/**/*.spec.{ts,tsx}'],
      rules: {
        'no-restricted-syntax': 'off',
      },
    },
  ],
};
