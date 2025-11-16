// ESLint 9 Flat Config - Architecture Guard
// Migrated from .eslintrc.architecture.js
import tsParser from '@typescript-eslint/parser';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import noRestQueries from './scripts/eslint-rules/no-rest-queries.js';
import noHardcodedPorts from './scripts/eslint-rules/no-hardcoded-ports.js';
import enforceApiContracts from './scripts/eslint-rules/enforce-api-contracts.js';

// Load local custom rules under a plugin namespace
const architecturePlugin = {
  rules: {
    'no-rest-queries': noRestQueries,
    'no-hardcoded-ports': noHardcodedPorts,
    'enforce-api-contracts': enforceApiContracts,
  },
};

export default [
  // Ignores
  {
    ignores: ['**/node_modules/**', '**/dist/**', '**/build/**', '**/coverage/**'],
  },
  // Base rules (apply to all targets; guard script narrows paths)
  {
    languageOptions: {
      parser: tsParser,
      ecmaVersion: 2022,
      sourceType: 'module',
      parserOptions: {
        project: ['./frontend/tsconfig.app.json', './frontend/tsconfig.node.json'],
        tsconfigRootDir: process.cwd(),
      },
    },
    plugins: {
      '@typescript-eslint': tsPlugin,
      architecture: architecturePlugin,
    },
    rules: {
      // 🏗️ 架构守护
      'architecture/no-rest-queries': [
        'error',
        {
          allowedRestMethods: ['POST', 'PUT', 'DELETE', 'PATCH'],
          // 收敛 GET 直连例外：仅允许认证入口（其余只读场景改为 GraphQL）
          allowedQueryEndpoints: ['/auth'],
          graphqlClient: 'graphql-client',
        },
      ],
      'architecture/no-hardcoded-ports': [
        'error',
        {
          allowedPorts: [80, 443],
          configModule: '@shared/config/ports',
          allowedPatterns: ['SERVICE_PORTS\\.', 'CQRS_ENDPOINTS\\.', 'TEST_ENDPOINTS\\.'],
        },
      ],
      'architecture/enforce-api-contracts': [
        'error',
        {
          fieldNamingStyle: 'camelCase',
          standardFields: {
            identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
            timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'],
            // 统一状态词表：status/isCurrent/isFuture/isTemporal（不暴露 isDeleted）
            statusFields: ['status', 'isCurrent', 'isFuture', 'isTemporal'],
            operationFields: ['operationType', 'operatedBy', 'operationReason'],
            hierarchyFields: ['level', 'codePath', 'namePath', 'hierarchyDepth'],
            configFields: ['unitType', 'sortOrder', 'description', 'profile'],
          },
          deprecatedFields: [
            'parent_unit_id',
            'unit_type',
            'is_deleted',
            'operation_type',
            'created_at',
            'updated_at',
            'effective_date',
            'end_date',
            'record_id',
            'tenant_id',
            'parent_code',
            'is_current',
          ],
          allowedContexts: ['test', 'mock', 'fixture', 'migration'],
        },
      ],

      // 📋 现有规则强化
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['**/api/**/*query*', '**/services/**/*query*'],
              message: '查询操作请使用GraphQL客户端，不要导入REST查询API',
            },
            {
              group: ['**/config/**', '!**/src/shared/config/**'],
              message: '请使用统一配置模块 @shared/config/*',
            },
            {
              group: ['axios', 'fetch', 'request', 'superagent'],
              importNames: ['get'],
              message: 'HTTP GET请求请使用GraphQL客户端',
            },
          ],
        },
      ],
      '@typescript-eslint/naming-convention': [
        'error',
        { selector: ['interface', 'typeAlias', 'class'], format: ['PascalCase'] },
        { selector: ['variable', 'function', 'method'], format: ['camelCase'], leadingUnderscore: 'allow' },
        {
          selector: 'objectLiteralProperty',
          format: ['camelCase'],
          filter: { regex: '^(client_id|client_secret|grant_type|refresh_token|access_token)$', match: false },
        },
        { selector: 'enumMember', format: ['UPPER_CASE'] },
        { selector: 'variable', modifiers: ['const', 'global'], format: ['UPPER_CASE', 'camelCase'] },
      ],

      // 其他质量规则
      'no-console': ['error', { allow: ['warn', 'error', 'info'] }],
      'prefer-template': 'error',
      'no-var': 'error',
      'prefer-const': 'error',

      // 架构特定禁止项
      'no-restricted-syntax': [
        'error',
        { selector: 'Literal[value="SUSPENDED"]', message: '使用 "INACTIVE" 替代已废弃的 "SUSPENDED" 状态' },
        {
          selector: 'CallExpression[callee.name="fetch"][arguments.1.type!="ObjectExpression"]',
          message: '禁止使用fetch进行GET请求，请使用GraphQL客户端',
        },
        {
          selector: 'VariableDeclarator[id.name=/.*[Pp]ort.*/][init.type="Literal"][init.value>=1024][init.value<=65535]',
          message: '端口定义应在统一配置文件中，不要在业务代码中硬编码',
        },
      ],
      'no-restricted-globals': [
        'error',
        {
          name: 'fetch',
          message: '请使用GraphQL客户端进行查询，fetch仅用于命令操作',
        },
      ],
    },
  },
  // 测试文件宽松规则
  {
    files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
    rules: {
      'architecture/no-rest-queries': 'warn',
      'architecture/enforce-api-contracts': 'off',
      'architecture/no-hardcoded-ports': 'off',
      '@typescript-eslint/naming-convention': 'off',
      'no-console': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      'no-restricted-globals': 'off',
      'no-alert': 'off',
    },
  },
  // 配置文件特殊规则
  {
    files: ['**/config/**/*.ts', '**/config/**/*.js'],
    rules: {
      'architecture/no-hardcoded-ports': 'off',
      '@typescript-eslint/naming-convention': [
        'error',
        { selector: 'objectLiteralProperty', format: ['camelCase', 'UPPER_CASE'] },
      ],
    },
  },
  // 统一客户端实现层：允许底层使用 fetch（架构门禁在业务层生效）
  {
    files: ['frontend/src/shared/api/unified-client.ts'],
    rules: {
      'no-restricted-globals': 'off',
      'architecture/no-rest-queries': 'off',
      'architecture/enforce-api-contracts': 'off',
      '@typescript-eslint/naming-convention': 'off',
      'no-restricted-syntax': 'off',
      'no-restricted-imports': 'off'
    },
  },
  // 迁移和种子文件
  {
    files: ['**/migrations/**/*.ts', '**/seeds/**/*.ts', '**/fixtures/**/*.ts'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      'no-console': 'off',
    },
  },
  // 本地配置与上下文：不作为契约字段检查目标
  {
    files: ['frontend/src/shared/auth/**/*.ts', 'frontend/src/shared/auth/**/*.tsx'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      '@typescript-eslint/naming-convention': 'off',
    },
  },
  {
    files: ['frontend/src/shared/config/**/*.ts', 'frontend/src/shared/config/**/*.tsx'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      'architecture/no-rest-queries': 'off',
      '@typescript-eslint/naming-convention': 'off',
      'no-restricted-imports': 'off',
    },
  },
  // 认证相关文件OAuth字段例外
  {
    files: ['frontend/src/shared/api/auth*.ts', 'frontend/src/shared/api/oauth/**/*.ts'],
    rules: {
      'architecture/enforce-api-contracts': [
        'error',
        {
          fieldNamingStyle: 'camelCase',
          // OAuth 标准字段已在规则内部豁免；此处无需额外 allowedFields 配置
          deprecatedFields: []
        },
      ],
    },
  },
  // 本地配置与上下文：不作为契约字段检查目标（放在末尾以确保覆盖前述规则）
  {
    files: ['frontend/src/shared/auth/**/*.ts', 'frontend/src/shared/auth/**/*.tsx'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      '@typescript-eslint/naming-convention': 'off',
    },
  },
  {
    files: ['frontend/src/shared/config/**/*.ts', 'frontend/src/shared/config/**/*.tsx'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      'architecture/no-rest-queries': 'off',
      '@typescript-eslint/naming-convention': 'off',
      'no-restricted-imports': 'off',
      'prefer-template': 'off'
    },
  },
  // 统一 API - Auth 管理器：作为基础设施，不参与契约字段与架构限制校验（本地迭代期）
  {
    files: ['frontend/src/shared/api/auth.ts'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      'architecture/no-rest-queries': 'off',
      'architecture/no-hardcoded-ports': 'off',
      'no-restricted-imports': 'off',
      'no-restricted-syntax': 'off',
      '@typescript-eslint/naming-convention': 'off',
      'prefer-template': 'off'
    },
  },
];
