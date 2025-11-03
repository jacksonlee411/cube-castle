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
  },
};
