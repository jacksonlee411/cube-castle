/**
 * ESLint API合规性配置 - 基础版本
 * 用于检查API使用的基本规范，不依赖额外插件
 */

module.exports = {
  env: {
    browser: true,
    es6: true,
    node: true
  },
  extends: [
    'eslint:recommended'
  ],
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true
    }
  },
  rules: {
    // 禁止使用 console.log，鼓励使用规范的日志工具
    'no-console': 'warn',
    
    // 禁止使用 alert/confirm/prompt
    'no-alert': 'error',
    
    // 要求使用 === 和 !==
    'eqeqeq': 'error',
    
    // 禁止未声明的变量
    'no-undef': 'error',
    
    // 禁止未使用的变量
    'no-unused-vars': 'warn',
    
    // 强制使用驼峰命名
    'camelcase': ['error', { properties: 'always' }]
  }
};