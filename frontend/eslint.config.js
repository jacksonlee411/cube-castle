import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { globalIgnores } from 'eslint/config'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs['recommended-latest'],
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          destructuredArrayIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_'
        }
      ],
      
      // 🚨 架构违规防范规则 - 防止FRONTEND-AUTH-BYPASS类问题
      'no-restricted-globals': [
        'error',
        {
          name: 'fetch',
          message: '🚨 架构违规：禁止直接使用fetch调用内部API。请使用unifiedRESTClient或unifiedGraphQLClient以确保JWT认证和CQRS架构合规。'
        }
      ],
      
      // 🚨 强制使用统一客户端
      'no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: 'node-fetch',
              message: '🚨 架构违规：禁止使用node-fetch。请使用unifiedRESTClient或unifiedGraphQLClient。'
            },
            {
              name: 'axios',
              message: '🚨 架构违规：禁止直接使用axios调用内部API。请使用unifiedRESTClient或unifiedGraphQLClient。'
            }
          ]
        }
      ],
      
      // 🚨 Canvas Kit v13企业级标准
      'no-restricted-syntax': [
        'error',
        {
          selector: 'CallExpression[callee.name="alert"]',
          message: '🚨 用户体验违规：禁止使用alert()。请使用统一的showSuccess()或showError()消息系统。'
        }
      ]
    }
  },
])
