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
          allowedQueryEndpoints: ['/auth', '/health', '/metrics'],
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
            statusFields: ['status', 'isDeleted', 'isCurrent', 'isFuture'],
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
  // 迁移和种子文件
  {
    files: ['**/migrations/**/*.ts', '**/seeds/**/*.ts', '**/fixtures/**/*.ts'],
    rules: {
      'architecture/enforce-api-contracts': 'off',
      'no-console': 'off',
    },
  },
  // 认证相关文件OAuth字段例外
  {
    files: ['**/auth/**/*.ts', '**/oauth/**/*.ts'],
    rules: {
      'architecture/enforce-api-contracts': [
        'error',
        {
          fieldNamingStyle: 'camelCase',
          allowedFields: ['client_id', 'client_secret', 'grant_type', 'refresh_token', 'access_token'],
          deprecatedFields: [],
        },
      ],
    },
  },
];

