#!/usr/bin/env node

/**
 * Cube Castle - 架构验证器
 * 基于静态代码分析的架构一致性验证工具
 * 
 * 用途: 验证CQRS架构、端口配置、API契约一致性
 * 作者: Claude Code Assistant
 * 日期: 2025-09-07
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// 🎨 颜色定义
const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  reset: '\x1b[0m',
  bright: '\x1b[1m'
};

// 🔧 配置
const config = {
  projectRoot: process.cwd(),
  reportDir: path.join(process.cwd(), 'reports', 'architecture'),
  
  // 验证规则配置
  rules: {
    // CQRS架构规则
    cqrsArchitecture: {
      enabled: true,
      frontendPath: 'frontend/src',
      prohibitedRestQueries: ['GET', 'get'],
      // 收敛 GET 直连例外：仅允许认证入口（其余只读场景改为 GraphQL）
      allowedQueryEndpoints: ['/auth'],
      graphqlClientPattern: /graphql|gql/i
    },

    // 禁止端点规则
    forbiddenEndpoints: {
      enabled: true,
      patterns: [
        {
          regex: /\/organization-units\/temporal/gi,
          description: '禁止使用未立项的 /organization-units/temporal REST 路径'
        },
        {
          // 直接连接后端固定端口（应通过前端单基址代理访问）
          // 匹配 http(s)/ws(s) + :9090|:8090 或 localhost/127.0.0.1 + :9090|:8090
          regex: /(https?:\/\/[^\s"'`]+:(9090|8090)\b)|(wss?:\/\/[^\s"'`]+:(9090|8090)\b)|\b(?:localhost|127\.0\.0\.1)\s*:(9090|8090)\b/gi,
          description: '禁止直连后端固定端口 (:9090/:8090)，必须通过单基址代理（/api/v1、/graphql）访问'
        }
      ]
    },
    
    // 端口配置规则
    portConfiguration: {
      enabled: true,
      allowedPorts: [80, 443],
      hardcodedPortPattern: /:\s*(\d{4,5})/g,
      configModulePath: 'frontend/src/shared/config/ports.ts',
      requiredConfigExports: ['SERVICE_PORTS', 'CQRS_ENDPOINTS']
    },
    
    // API契约规则
    apiContracts: {
      enabled: true,
      requiredNamingStyle: 'camelCase',
      deprecatedFields: [
        'parent_unit_id', 'unit_type', 'is_deleted', 'operation_type',
        'created_at', 'updated_at', 'effective_date', 'end_date',
        'record_id', 'tenant_id', 'parent_code', 'is_current'
      ],
      standardFields: {
        identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
        timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'],
        statusFields: ['status', 'isCurrent', 'isFuture', 'isTemporal'],
        operationFields: ['operationType', 'operatedBy', 'operationReason']
      }
    },

    // ESLint例外说明校验规则
    eslintExceptionComment: {
      enabled: true,
      targetPattern: /eslint-disable-next-line\s+camelcase/,
      requireReasonPattern: /eslint-disable-next-line\s+camelcase\s+--\s+\S/
    }
  }
};

// 📊 验证统计
const stats = {
  totalFiles: 0,
  passedFiles: 0,
  failedFiles: 0,
  violations: {
    cqrs: 0,
    ports: 0,
    contracts: 0,
    forbidden: 0,
    eslintExceptions: 0,
    total: 0
  },
  fixedIssues: 0
};

// 📋 日志函数
const log = {
  info: (msg) => console.log(`${colors.blue}ℹ️  ${msg}${colors.reset}`),
  success: (msg) => console.log(`${colors.green}✅ ${msg}${colors.reset}`),
  warning: (msg) => console.log(`${colors.yellow}⚠️  ${msg}${colors.reset}`),
  error: (msg) => console.error(`${colors.red}❌ ${msg}${colors.reset}`),
  verbose: (msg) => process.env.VERBOSE && console.log(`${colors.cyan}🔍 ${msg}${colors.reset}`)
};

// 🔍 文件扫描器
class FileScanner {
  static scanDirectory(dir, extensions = ['.ts', '.tsx', '.js', '.jsx']) {
    const files = [];
    
    function scanRecursive(currentDir) {
      try {
        const entries = fs.readdirSync(currentDir);
        
        for (const entry of entries) {
          const fullPath = path.join(currentDir, entry);
          const stat = fs.statSync(fullPath);
          
          if (stat.isDirectory()) {
            // 跳过外部/产物目录，避免误报
            const ignoreDirs = ['node_modules', 'dist', 'build', '.git', 'third_party', 'playwright-report'];
            if (!ignoreDirs.includes(entry)) {
              scanRecursive(fullPath);
            }
          } else if (stat.isFile()) {
            const ext = path.extname(entry);
            if (extensions.includes(ext)) {
              files.push(fullPath);
            }
          }
        }
      } catch (err) {
        log.warning(`无法扫描目录 ${currentDir}: ${err.message}`);
      }
    }
    
    scanRecursive(dir);
    return files;
  }
}

// 🏗️ CQRS架构验证器
class CQRSArchitectureValidator {
  static validate(filePath, content) {
    const violations = [];
    const lines = content.split('\n');
    
    // 检查REST GET请求（逐行快速规则 + 跨行缺省GET检测）
    const restGetPatterns = [
      /fetch\s*\(\s*['"`][^'"`]*['"`]\s*\)/g,  // fetch without options (default GET)
      /fetch\s*\([^)]*method\s*:\s*['"`]GET['"`]/gi,  // explicit GET method
      /axios\.get\s*\(/gi  // axios.get
    ];
    
    lines.forEach((line, lineNum) => {
      restGetPatterns.forEach(pattern => {
        const matches = line.match(pattern);
        if (matches) {
          // 检查是否为允许的端点
          const isAllowedEndpoint = config.rules.cqrsArchitecture.allowedQueryEndpoints
            .some(endpoint => line.includes(endpoint));
          
          if (!isAllowedEndpoint) {
            violations.push({
              type: 'cqrs',
              line: lineNum + 1,
              column: line.search(pattern),
              message: `禁止使用REST GET请求进行查询，请使用GraphQL客户端`,
              code: 'no-rest-queries',
              severity: 'error'
            });
          }
        }
      });
    });
    
    // 跨行检测：fetch(url, { ... }) 且 options 未包含 method => 仍视为默认GET
    try {
      const fetchWithObjectPattern = /fetch\s*\(\s*([^,]+),\s*\{([\s\S]*?)\}\s*\)/gi;
      let m;
      while ((m = fetchWithObjectPattern.exec(content)) !== null) {
        const optionsBody = m[2] || '';
        const hasMethod = /method\s*:/i.test(optionsBody);
        if (!hasMethod) {
          // 计算位置
          const index = m.index;
          const before = content.substring(0, index);
          const line = before.split('\n').length;
          const col = index - before.lastIndexOf('\n', index - 1);
          const snippet = content.substring(index, Math.min(content.length, index + 200)).replace(/\s+/g, ' ').trim();

          const isAllowedEndpoint = config.rules.cqrsArchitecture.allowedQueryEndpoints
            .some(endpoint => snippet.includes(endpoint));
          if (!isAllowedEndpoint) {
            violations.push({
              type: 'cqrs',
              line,
              column: col,
              message: '禁止使用REST GET请求进行查询，请使用GraphQL客户端（fetch 默认GET且缺少 method）',
              code: 'no-rest-queries',
              severity: 'error',
              context: snippet
            });
          }
        }
      }
    } catch (e) {
      // 忽略解析失败，保持稳健
    }

    // 检查是否使用了GraphQL客户端
    const hasGraphQLClient = config.rules.cqrsArchitecture.graphqlClientPattern.test(content);
    if (violations.length > 0 && !hasGraphQLClient) {
      violations.push({
        type: 'cqrs',
        line: 1,
        column: 1,
        message: '建议导入GraphQL客户端进行查询操作',
        code: 'missing-graphql-client',
        severity: 'warning'
      });
    }
    
    return violations;
  }
}

// 🔧 端口配置验证器
class PortConfigurationValidator {
  static validate(filePath, content) {
    const violations = [];
    const lines = content.split('\n');
    
    // 检查硬编码端口
    const allowedPorts = config.rules.portConfiguration.allowedPorts;

    lines.forEach((rawLine, lineNum) => {
      const line = rawLine;
      const trimmed = line.trim();

      // 跳过注释行，避免将日期/时间等误判为端口（例如 “// 迁移期限: 2025-09-16” 或样式 zIndex: 1000）
      if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
        return;
      }

      // 仅在“URL样式端口”或“port键值对”场景下进行检测，减少误报
      const hasUrlLike = /https?:\/\/|wss?:\/\/|localhost|127\.0\.0\.1/.test(line) && /:\s*\d{2,5}/.test(line);
      const hasPortKey = /\bport\b\s*:\s*\d{2,5}/i.test(line);

      if (!hasUrlLike && !hasPortKey) {
        return;
      }

      // 从可能的场景中抽取数字并校验
      const numMatches = [...line.matchAll(/:\s*(\d{2,5})/g)];
      numMatches.forEach((m) => {
        const port = parseInt(m[1]);

        // 跳过标准端口
        if (allowedPorts.includes(port)) return;

        // 检查是否使用了配置模块
        const usesConfig = /SERVICE_PORTS|CQRS_ENDPOINTS/.test(line);

        if (!usesConfig && port >= 1000 && port <= 65535) {
          const suggestedConfig = this.getSuggestedConfig(port);
          violations.push({
            type: 'ports',
            line: lineNum + 1,
            column: m.index,
            message: `硬编码端口 ${port}，建议使用 ${suggestedConfig}`,
            code: 'no-hardcoded-ports',
            severity: 'error',
            fix: {
              range: [m.index, m.index + m[0].length],
              newText: suggestedConfig
            }
          });
        }
      });
    });
    
    return violations;
  }
  
  static getSuggestedConfig(port) {
    const portMap = {
      3000: 'SERVICE_PORTS.FRONTEND_DEV',
      3001: 'SERVICE_PORTS.FRONTEND_PREVIEW',
      8090: 'SERVICE_PORTS.GRAPHQL_QUERY_SERVICE', 
      9090: 'SERVICE_PORTS.REST_COMMAND_SERVICE',
      5432: 'SERVICE_PORTS.POSTGRESQL',
      6379: 'SERVICE_PORTS.REDIS'
    };
    
    return portMap[port] || 'SERVICE_PORTS.APPROPRIATE_PORT';
  }
}

// 📋 API契约验证器
class APIContractValidator {
  static validate(filePath, content) {
    const violations = [];
    const lines = content.split('\n');
    
    // 检查废弃字段
    const deprecatedFields = config.rules.apiContracts.deprecatedFields;
    
    lines.forEach((rawLine, lineNum) => {
      const line = rawLine;
      const trimmed = line.trim();
      if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
        return;
      }
      deprecatedFields.forEach(field => {
        const fieldPattern = new RegExp(`['"\`]${field}['"\`]|${field}\\s*:|\\b${field}\\b`, 'g');
        const matches = [...line.matchAll(fieldPattern)];
        
        matches.forEach(match => {
          const replacement = this.getReplacementField(field);
          
          violations.push({
            type: 'contracts',
            line: lineNum + 1,
            column: match.index,
            message: `废弃字段 "${field}"，请使用 "${replacement}"`,
            code: 'deprecated-field',
            severity: 'error',
            fix: {
              range: [match.index, match.index + field.length],
              newText: replacement
            }
          });
        });
      });
    });
    
    // 检查snake_case字段名
    const snakeCasePattern = /['"`]([a-z]+_[a-z_]+)['"`]|([a-z]+_[a-z_]+)\s*:/g;
    const allowedSnakeTokens = [
      // OAuth标准值与项目内部键名白名单（不作为契约字段处理）
      'client_credentials',
      'cube_castle_oauth_token'
    ];
    
    lines.forEach((rawLine, lineNum) => {
      const line = rawLine;
      const trimmed = line.trim();
      if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
        return;
      }
      const matches = [...line.matchAll(snakeCasePattern)];
      
      matches.forEach(match => {
        const fieldName = match[1] || match[2];

        // 白名单跳过（值或内部键名，不属于API契约字段）
        if (allowedSnakeTokens.includes(fieldName)) {
          return;
        }
        
        // 跳过OAuth标准字段
        const oauthFields = ['client_id', 'client_secret', 'grant_type', 'refresh_token', 'access_token'];
        if (oauthFields.includes(fieldName)) {
          return;
        }
        
        // 跳过已知废弃字段（已经在上面检查过）
        if (deprecatedFields.includes(fieldName)) {
          return;
        }
        
        const camelCaseField = this.toCamelCase(fieldName);
        
        violations.push({
          type: 'contracts',
          line: lineNum + 1,
          column: match.index,
          message: `使用camelCase字段名 "${camelCaseField}" 替代 snake_case "${fieldName}"`,
          code: 'snake-case-field',
          severity: 'error',
          fix: {
            range: [match.index, match.index + match[0].length],
            newText: match[0].replace(fieldName, camelCaseField)
          }
        });
      });
    });
    
    return violations;
  }
  
  static getReplacementField(deprecatedField) {
    const replacementMap = {
      'parent_unit_id': 'parentCode',
      'unit_type': 'unitType',
      "is_deleted": "status",
      'operation_type': 'operationType',
      'created_at': 'createdAt',
      'updated_at': 'updatedAt',
      'effective_date': 'effectiveDate',
      'end_date': 'endDate',
      'record_id': 'recordId',
      'tenant_id': 'tenantId',
      'parent_code': 'parentCode',
      'is_current': 'isCurrent'
    };
    
    return replacementMap[deprecatedField] || this.toCamelCase(deprecatedField);
  }
  
  static toCamelCase(snakeStr) {
    return snakeStr.replace(/_([a-z])/g, (match, letter) => letter.toUpperCase());
  }
}

// 🛡️ ESLint例外说明验证器
class ESLintExceptionCommentValidator {
  static validate(filePath, content, options) {
    const violations = [];
    if (!options?.targetPattern) {
      return violations;
    }

    const lines = content.split('\n');
    const reasonRegex = options.requireReasonPattern || /eslint-disable-next-line\s+camelcase\s+--\s+\S/;

    lines.forEach((rawLine, lineNum) => {
      const line = rawLine.trim();
      if (!options.targetPattern.test(line)) {
        return;
      }

      if (!reasonRegex.test(line)) {
        violations.push({
          type: 'eslintExceptions',
          line: lineNum + 1,
          column: rawLine.indexOf('eslint-disable-next-line'),
          message: 'eslint-disable-next-line camelcase 必须包含 "-- 原因" 说明',
          code: 'missing-eslintexception-reason',
          severity: 'error'
        });
      }
    });

    return violations;
  }
}

// 🚫 禁止端点验证器
class ForbiddenEndpointValidator {
  static validate(filePath, content, options) {
    const violations = [];
    if (!options?.patterns || options.patterns.length === 0) {
      return violations;
    }

    options.patterns.forEach(patternRule => {
      const { regex, description } = patternRule;
      if (!regex) {
        return;
      }

      let match;
      const pattern = new RegExp(regex.source, regex.flags);
      while ((match = pattern.exec(content)) !== null) {
        const index = match.index;
        const snippet = content
          .substring(Math.max(0, index - 40), Math.min(content.length, index + 80))
          .replace(/\s+/g, ' ');

        violations.push({
          type: 'forbidden',
          line: content.substring(0, index).split('\n').length,
          column: index - content.lastIndexOf('\n', index - 1),
          message: description || '检测到禁止使用的端点模式',
          code: 'forbidden-endpoint',
          severity: 'error',
          context: snippet.trim()
        });

        if (!pattern.global) {
          break;
        }
      }
    });

    return violations;
  }
}

// 🚀 主验证引擎
class ArchitectureValidator {
  constructor(options = {}) {
    const { ruleFilter, ...restOptions } = options;
    this.options = { ...config, ...restOptions };
    this.violations = [];
    this.ruleFilter = Array.isArray(ruleFilter) && ruleFilter.length > 0 ? ruleFilter : null;
  }

  isRuleEnabled(ruleName) {
    if (!this.ruleFilter) {
      return true;
    }
    return this.ruleFilter.includes(ruleName);
  }
  
  async validateFile(filePath) {
    try {
      const content = fs.readFileSync(filePath, 'utf8');
      const fileViolations = [];
      
      stats.totalFiles++;
      
      // CQRS架构验证
      if (this.isRuleEnabled('cqrsArchitecture') &&
          this.options.rules.cqrsArchitecture.enabled && 
          filePath.includes(this.options.rules.cqrsArchitecture.frontendPath)) {
        // 跳过统一客户端底层实现文件，避免将内部 fetch 误判为业务查询
        const relative = path.relative(this.options.projectRoot || process.cwd(), filePath).replace(/\\/g, '/');
        const ignoreCQRSFiles = [
          'frontend/src/shared/api/unified-client.ts',
          // Plan 257: 门面层作为合规入口，避免 CQRS 规则对其内部转发产生误报
          'frontend/src/shared/api/facade/organization.ts'
        ];
        if (!ignoreCQRSFiles.includes(relative)) {
          const cqrsViolations = CQRSArchitectureValidator.validate(filePath, content);
          fileViolations.push(...cqrsViolations);
          stats.violations.cqrs += cqrsViolations.length;
        }
      }
      
      // 端口配置验证
      if (this.isRuleEnabled('portConfiguration') && this.options.rules.portConfiguration.enabled) {
        const portViolations = PortConfigurationValidator.validate(filePath, content);
        fileViolations.push(...portViolations);
        stats.violations.ports += portViolations.length;
      }
      
      // API契约验证
      if (this.isRuleEnabled('apiContracts') && this.options.rules.apiContracts.enabled) {
        const contractViolations = APIContractValidator.validate(filePath, content);
        fileViolations.push(...contractViolations);
        stats.violations.contracts += contractViolations.length;
      }

      // ESLint 例外注释验证
      if (this.isRuleEnabled('eslintExceptionComment') && this.options.rules.eslintExceptionComment?.enabled) {
        const eslintExceptionViolations = ESLintExceptionCommentValidator.validate(
          filePath,
          content,
          this.options.rules.eslintExceptionComment
        );
        fileViolations.push(...eslintExceptionViolations);
        stats.violations.eslintExceptions += eslintExceptionViolations.length;
      }

      // 禁止端点验证（跳过本工具文件自身的模式常量，避免自检误报）
      const isSelfFile = /scripts[\\/]+quality[\\/]+architecture-validator\.js$/.test(filePath);
      if (this.isRuleEnabled('forbiddenEndpoints') && this.options.rules.forbiddenEndpoints?.enabled && !isSelfFile) {
        const forbiddenViolations = ForbiddenEndpointValidator.validate(
          filePath,
          content,
          this.options.rules.forbiddenEndpoints
        );
        fileViolations.push(...forbiddenViolations);
        stats.violations.forbidden += forbiddenViolations.length;
      }
      
      // 统计结果
      if (fileViolations.length > 0) {
        stats.failedFiles++;
        this.violations.push({
          filePath,
          violations: fileViolations
        });
      } else {
        stats.passedFiles++;
      }
      
      stats.violations.total += fileViolations.length;
      
      log.verbose(`验证文件: ${path.relative(config.projectRoot, filePath)} - ${fileViolations.length} 个问题`);
      
      return fileViolations;
      
    } catch (err) {
      log.error(`验证文件失败 ${filePath}: ${err.message}`);
      return [];
    }
  }
  
  async validateDirectory(dirPath) {
    log.info(`扫描目录: ${path.relative(config.projectRoot, dirPath)}`);
    
    const files = FileScanner.scanDirectory(dirPath);
    log.info(`找到 ${files.length} 个文件待验证`);
    
    for (const file of files) {
      await this.validateFile(file);
    }
    
    return this.violations;
  }
  
  generateReport(outPath = null) {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalFiles: stats.totalFiles,
        passedFiles: stats.passedFiles,
        failedFiles: stats.failedFiles,
        totalViolations: stats.violations.total,
        violationsByType: {
          cqrs: stats.violations.cqrs,
          ports: stats.violations.ports,
          contracts: stats.violations.contracts,
          forbidden: stats.violations.forbidden,
          eslintExceptions: stats.violations.eslintExceptions
        }
      },
      violations: this.violations
    };
    
    // 确保报告目录存在
    const reportPath = outPath
      ? outPath
      : path.join(config.reportDir, 'architecture-validation.json');
    const reportDir = path.dirname(reportPath);
    fs.mkdirSync(reportDir, { recursive: true });

    // 保存JSON报告
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    return report;
  }
  
  printSummary() {
    log.info('📊 架构验证统计报告:');
    log.info(`   📁 验证文件: ${stats.totalFiles} 个`);
    log.info(`   ✅ 通过文件: ${stats.passedFiles} 个`);
    
    if (stats.failedFiles > 0) {
      log.warning(`   ❌ 失败文件: ${stats.failedFiles} 个`);
    }
    
    log.info(`   🔍 问题总数: ${stats.violations.total} 个`);
    
    if (stats.violations.cqrs > 0) {
      log.warning(`   🏗️  CQRS违规: ${stats.violations.cqrs} 个`);
    }
    if (stats.violations.ports > 0) {
      log.warning(`   🔧 端口违规: ${stats.violations.ports} 个`);
    }
    if (stats.violations.contracts > 0) {
      log.warning(`   📋 契约违规: ${stats.violations.contracts} 个`);
    }
    if (stats.violations.forbidden > 0) {
      log.error(`   🚫 禁止端点违规: ${stats.violations.forbidden} 个`);
    }
    if (stats.violations.eslintExceptions > 0) {
      log.warning(`   📝 ESLint例外说明缺失: ${stats.violations.eslintExceptions} 个`);
    }

    // 质量门禁判定
    const criticalViolations = stats.violations.cqrs +
      stats.violations.ports +
      stats.violations.forbidden +
      stats.violations.eslintExceptions;
    if (criticalViolations > 0) {
      log.error(`🚫 质量门禁失败: ${criticalViolations} 个关键违规`);
      return false;
    } else {
      log.success('🎉 质量门禁通过: 架构符合企业级标准');
      return true;
    }
  }
}

// 🎯 CLI入口
async function main() {
  const args = process.argv.slice(2);
  const scope = args.includes('--scope') ? args[args.indexOf('--scope') + 1] : 'frontend';
  const ruleArgIndex = args.indexOf('--rule');
  const outArgIndex = args.indexOf('--out');
  const outPath = outArgIndex !== -1 && args[outArgIndex + 1] ? args[outArgIndex + 1] : null;
  const ruleAliases = {
    cqrs: 'cqrsArchitecture',
    ports: 'portConfiguration',
    contracts: 'apiContracts',
    forbidden: 'forbiddenEndpoints',
    'eslint-exception-comment': 'eslintExceptionComment',
    'eslintExceptionComment': 'eslintExceptionComment'
  };
  let ruleFilter = null;
  if (ruleArgIndex !== -1 && args[ruleArgIndex + 1]) {
    ruleFilter = args[ruleArgIndex + 1]
      .split(',')
      .map(rule => ruleAliases[rule] || rule)
      .filter(Boolean);
  }
  
  log.info('🏗️ Cube Castle - 架构验证器启动');
  log.info(`验证范围: ${scope}`);
  if (ruleFilter && ruleFilter.length > 0) {
    log.info(`验证规则: ${ruleFilter.join(', ')}`);
  }
  
  const validator = new ArchitectureValidator({ ruleFilter });
  
  // 确定验证路径
  const targetPath = scope === 'frontend' ? 
    path.join(config.projectRoot, 'frontend/src') : 
    config.projectRoot;
  
  try {
    await validator.validateDirectory(targetPath);
    const report = validator.generateReport(outPath);
    const success = validator.printSummary();
    
    log.info(`📂 详细报告: ${outPath ? outPath : path.join(config.reportDir, 'architecture-validation.json')}`);
    
    process.exit(success ? 0 : 1);
    
  } catch (err) {
    log.error(`验证失败: ${err.message}`);
    process.exit(1);
  }
}

// 运行主程序
if (require.main === module) {
  main().catch(err => {
    console.error('Validation failed:', err);
    process.exit(1);
  });
}

module.exports = { ArchitectureValidator, config };
