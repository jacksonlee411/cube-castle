/**
 * ESLint架构守护配置
 * 企业级代码架构和API契约一致性规则
 * 
 * 用途：确保项目架构标准和API契约一致性的自动化验证
 * 集成：与现有ESLint配置配合使用，专注架构质量
 */

module.exports = {
  extends: [
    './eslint.config.js' // 继承现有配置
  ],
  
  plugins: [
    // 本地自定义规则插件将在运行时动态加载
  ],

  rules: {
    // ==========================================
    // 🏗️ 架构守护规则
    // ==========================================

    // 1. 禁止前端REST查询，强制GraphQL
    'architecture/no-rest-queries': ['error', {
      allowedRestMethods: ['POST', 'PUT', 'DELETE', 'PATCH'],
      allowedQueryEndpoints: ['/auth', '/health', '/metrics'],
      graphqlClient: 'graphql-client'
    }],

    // 2. 禁止硬编码端口，强制统一配置
    'architecture/no-hardcoded-ports': ['error', {
      allowedPorts: [80, 443], // 标准HTTP/HTTPS端口
      configModule: '@shared/config/ports',
      allowedPatterns: [
        'SERVICE_PORTS\\.',
        'CQRS_ENDPOINTS\\.',
        'TEST_ENDPOINTS\\.'
      ]
    }],

    // 3. 强制API契约字段命名一致性
    'architecture/enforce-api-contracts': ['error', {
      fieldNamingStyle: 'camelCase',
      standardFields: {
        // 核心业务字段 (camelCase)
        identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
        timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'],
        statusFields: ['status', 'isDeleted', 'isCurrent', 'isFuture'],
        operationFields: ['operationType', 'operatedBy', 'operationReason'],
        hierarchyFields: ['level', 'codePath', 'namePath', 'hierarchyDepth'],
        configFields: ['unitType', 'sortOrder', 'description', 'profile']
      },
      deprecatedFields: [
        'parent_unit_id', 'unit_type', 'is_deleted', 'operation_type',
        'created_at', 'updated_at', 'effective_date', 'end_date',
        'record_id', 'tenant_id', 'parent_code', 'is_current'
      ],
      allowedContexts: ['test', 'mock', 'fixture', 'migration']
    }],

    // ==========================================
    // 📋 现有规则强化
    // ==========================================

    // 强化导入规则
    'no-restricted-imports': ['error', {
      patterns: [
        {
          group: ['**/api/**/*query*', '**/services/**/*query*'],
          message: '查询操作请使用GraphQL客户端，不要导入REST查询API'
        },
        {
          group: ['**/config/**', '!**/src/shared/config/**'],
          message: '请使用统一配置模块 @shared/config/*'
        },
        {
          group: ['axios', 'fetch', 'request', 'superagent'],
          importNames: ['get'],
          message: 'HTTP GET请求请使用GraphQL客户端'
        }
      ]
    }],

    // 强化命名约定
    '@typescript-eslint/naming-convention': [
      'error',
      // 接口和类型必须使用PascalCase
      {
        selector: ['interface', 'typeAlias', 'class'],
        format: ['PascalCase']
      },
      // 变量和函数必须使用camelCase
      {
        selector: ['variable', 'function', 'method'],
        format: ['camelCase'],
        leadingUnderscore: 'allow'
      },
      // 对象属性使用camelCase（API字段命名）
      {
        selector: 'objectLiteralProperty',
        format: ['camelCase'],
        filter: {
          // OAuth字段例外
          regex: '^(client_id|client_secret|grant_type|refresh_token|access_token)$',
          match: false
        }
      },
      // 枚举使用UPPER_CASE
      {
        selector: 'enumMember',
        format: ['UPPER_CASE']
      },
      // 常量使用UPPER_CASE
      {
        selector: 'variable',
        modifiers: ['const', 'global'],
        format: ['UPPER_CASE', 'camelCase'] // 允许camelCase的配置对象
      }
    ],

    // 禁止console（生产代码）
    'no-console': ['error', {
      allow: ['warn', 'error', 'info'] // 允许日志级别输出
    }],

    // 强制使用模板字符串而不是字符串拼接
    'prefer-template': 'error',

    // 禁止var，强制let/const
    'no-var': 'error',
    'prefer-const': 'error',

    // ==========================================
    // 🚨 架构特定禁止项
    // ==========================================

    'no-restricted-syntax': [
      'error',
      // 禁止使用旧的组织状态枚举
      {
        selector: 'Literal[value="SUSPENDED"]',
        message: '使用 "INACTIVE" 替代已废弃的 "SUSPENDED" 状态'
      },
      // 禁止直接使用fetch GET
      {
        selector: 'CallExpression[callee.name="fetch"][arguments.1.type!="ObjectExpression"]',
        message: '禁止使用fetch进行GET请求，请使用GraphQL客户端'
      },
      // 禁止在非配置文件中定义端口
      {
        selector: 'VariableDeclarator[id.name=/.*[Pp]ort.*/][init.type="Literal"][init.value>=1024][init.value<=65535]',
        message: '端口定义应在统一配置文件中，不要在业务代码中硬编码'
      }
    ],

    // 禁止特定全局变量
    'no-restricted-globals': [
      'error',
      {
        name: 'fetch',
        message: '请使用GraphQL客户端进行查询，fetch仅用于命令操作'
      }
    ]
  },

  // ==========================================
  // 📁 文件特定配置
  // ==========================================
  
  overrides: [
    // 测试文件宽松规则
    {
      files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
      rules: {
        'architecture/no-rest-queries': 'warn', // 测试中允许但警告
        'architecture/enforce-api-contracts': 'off', // 测试数据可以使用任意字段名
        'no-console': 'off', // 测试中允许console
        '@typescript-eslint/no-explicit-any': 'off' // 测试中允许 any 以简化 Mock
      }
    },

    // 配置文件特殊规则
    {
      files: ['**/config/**/*.ts', '**/config/**/*.js'],
      rules: {
        'architecture/no-hardcoded-ports': 'off', // 配置文件中允许端口定义
        '@typescript-eslint/naming-convention': [
          'error',
          {
            selector: 'objectLiteralProperty',
            format: ['camelCase', 'UPPER_CASE'] // 配置中允许大写常量
          }
        ]
      }
    },

    // 迁移和种子文件
    {
      files: ['**/migrations/**/*.ts', '**/seeds/**/*.ts', '**/fixtures/**/*.ts'],
      rules: {
        'architecture/enforce-api-contracts': 'off', // 迁移文件可能使用数据库字段名
        'no-console': 'off' // 迁移脚本允许console输出
      }
    },

    // 认证相关文件OAuth字段例外
    {
      files: ['**/auth/**/*.ts', '**/oauth/**/*.ts'],
      rules: {
        'architecture/enforce-api-contracts': ['error', {
          fieldNamingStyle: 'camelCase',
          // OAuth文件中允许标准OAuth字段名
          allowedFields: ['client_id', 'client_secret', 'grant_type', 'refresh_token', 'access_token'],
          deprecatedFields: [] // OAuth文件中不检查废弃字段
        }]
      }
    }
  ],

  // ==========================================
  // ⚙️ 解析器和环境配置
  // ==========================================
  
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    project: ['./tsconfig.json', './tsconfig.node.json'],
    tsconfigRootDir: __dirname,
    ecmaFeatures: {
      jsx: true
    }
  },

  env: {
    browser: true,
    es2022: true,
    node: true,
    jest: true
  },

  settings: {
    'import/resolver': {
      typescript: {
        alwaysTryTypes: true,
        project: ['./tsconfig.json', './tsconfig.node.json']
      }
    }
  }
};
