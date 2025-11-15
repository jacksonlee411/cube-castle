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

      // 🚨 日志统一：必须通过 shared/utils/logger.ts 输出
      'no-console': 'error',
      
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
          ],
          patterns: [
            {
              group: ['**/shared/hooks/useOrganizations', '**/shared/hooks/useOrganizations.ts'],
              message: '🚨 兼容封装已废弃：请使用 useEnterpriseOrganizations / useOrganizationDetails。'
            },
            {
              group: [
                '**/features/positions/timelineAdapter',
                '**/features/positions/timelineAdapter.ts',
                '**/features/positions/statusMeta',
                '**/features/positions/statusMeta.ts'
              ],
              message: '🚨 Temporal Entity 命名已统一：请使用 "@/features/temporal/entity/timelineAdapter" 和 "@/features/temporal/entity/statusMeta"。'
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
  
  // 🛡️ 前端源代码额外门禁（禁止硬编码 data-testid）
  {
    files: ['src/**/*.{ts,tsx}'],
    ignores: [
      'src/shared/testids/temporalEntity.ts',
      'src/**/__tests__/**',
      'src/**/*.test.ts',
      'src/**/*.test.tsx',
      'src/**/*.spec.ts',
      'src/**/*.spec.tsx',
    ],
    rules: {
      'no-restricted-syntax': [
        'warn',
        // 保留 alert 禁止
        {
          selector: 'CallExpression[callee.name="alert"]',
          message: '🚨 用户体验违规：禁止使用alert()。请使用统一的showSuccess()或showError()消息系统。'
        },
        // 禁止在组件/源码中直接硬编码 data-testid（统一从 shared/testids/temporalEntity.ts 导入）
        {
          selector: 'JSXAttribute[name.name="data-testid"] > Literal',
          message: '🚨 选择器治理：禁止硬编码 data-testid，请从 shared/testids/temporalEntity.ts 导入并使用 temporalEntitySelectors。'
        },
        {
          selector: 'JSXAttribute[name.name="data-testid"] > TemplateLiteral',
          message: '🚨 选择器治理：禁止硬编码 data-testid，请从 shared/testids/temporalEntity.ts 导入并使用 temporalEntitySelectors。'
        }
      ]
    }
  },
  
  // 🧪 测试文件特殊规则配置 - 允许fetch用于E2E测试和契约测试
  {
    files: ['tests/**/*.{ts,tsx}', 'src/**/*.test.{ts,tsx}', 'src/**/*.spec.{ts,tsx}', 'scripts/**/*.ts', 'playwright.config.ts'],
    rules: {
      // 测试文件允许使用fetch进行API测试
      'no-restricted-globals': 'off',
      // 测试文件允许使用any类型进行模拟数据
      '@typescript-eslint/no-explicit-any': 'off',
      // 测试文件和脚本允许使用console进行调试输出
      'no-console': 'off'
    }
  },
  
  // 🔧 统一客户端文件特殊规则 - 底层实现允许使用fetch
  {
    files: ['src/shared/api/unified-client.ts', 'src/shared/api/auth.ts', 'src/shared/api/client.ts'],
    rules: {
      // 统一客户端实现层允许使用fetch
      'no-restricted-globals': 'off'
    }
  },
  
  // 🔧 组件文件兼容性规则 - 临时允许重新导出以保持向后兼容
  {
    files: [
      'src/features/temporal/components/TemporalDatePicker.tsx',
      'src/features/temporal/components/TemporalStatusSelector.tsx'
    ],
    rules: {
      // 允许重新导出工具函数和常量以保持向后兼容
      'react-refresh/only-export-components': 'warn'
    }
  }
])
