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
    // 🚫 禁止在组件中用字面量硬编码 data-testid；请从 "@/shared/testids/temporalEntity" 引用常量/构造器
    // 仅匹配 JSX 字面量，允许使用表达式（如 data-testid={selectors.xxx}）
    'no-restricted-syntax': [
      'error',
      {
        selector: 'JSXAttribute[name.name="data-testid"][value.type="Literal"]',
        message:
          'Do not hard-code data-testid. Import from "@/shared/testids/temporalEntity" (temporalEntitySelectors).',
      },
    ],
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
      files: ['src/shared/testids/temporalEntity.ts'],
      rules: {
        'no-restricted-syntax': 'off',
      },
    },
  ],
};
