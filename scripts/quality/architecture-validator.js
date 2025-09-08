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
      allowedQueryEndpoints: ['/auth', '/health', '/metrics'],
      graphqlClientPattern: /graphql|gql/i
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
        statusFields: ['status', 'isDeleted', 'isCurrent', 'isFuture'],
        operationFields: ['operationType', 'operatedBy', 'operationReason']
      }
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
            // 跳过node_modules等目录
            if (!['node_modules', 'dist', 'build', '.git'].includes(entry)) {
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
    
    // 检查REST GET请求
    const restGetPatterns = [
      /fetch\s*\(\s*['"`][^'"`]*['"`]\s*\)/g,  // fetch without options (default GET)
      /fetch\s*\([^)]*method\s*:\s*['"`]GET['"`]/gi,  // explicit GET method
      /axios\.get\s*\(/gi,  // axios.get
      /\.get\s*\(/gi,  // general .get calls
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
    const portPattern = config.rules.portConfiguration.hardcodedPortPattern;
    const allowedPorts = config.rules.portConfiguration.allowedPorts;
    
    lines.forEach((line, lineNum) => {
      const matches = [...line.matchAll(portPattern)];
      
      matches.forEach(match => {
        const port = parseInt(match[1]);
        
        // 跳过标准端口
        if (allowedPorts.includes(port)) {
          return;
        }
        
        // 检查是否使用了配置模块
        const usesConfig = /SERVICE_PORTS|CQRS_ENDPOINTS/.test(line);
        
        if (!usesConfig && port >= 1000 && port <= 65535) {
          const suggestedConfig = this.getSuggestedConfig(port);
          
          violations.push({
            type: 'ports',
            line: lineNum + 1,
            column: match.index,
            message: `硬编码端口 ${port}，建议使用 ${suggestedConfig}`,
            code: 'no-hardcoded-ports',
            severity: 'error',
            fix: {
              range: [match.index, match.index + match[0].length],
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
    
    lines.forEach((line, lineNum) => {
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
    
    lines.forEach((line, lineNum) => {
      const matches = [...line.matchAll(snakeCasePattern)];
      
      matches.forEach(match => {
        const fieldName = match[1] || match[2];
        
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
      'is_deleted': 'isDeleted',
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

// 🚀 主验证引擎
class ArchitectureValidator {
  constructor(options = {}) {
    this.options = { ...config, ...options };
    this.violations = [];
  }
  
  async validateFile(filePath) {
    try {
      const content = fs.readFileSync(filePath, 'utf8');
      const fileViolations = [];
      
      stats.totalFiles++;
      
      // CQRS架构验证
      if (this.options.rules.cqrsArchitecture.enabled && 
          filePath.includes(this.options.rules.cqrsArchitecture.frontendPath)) {
        const cqrsViolations = CQRSArchitectureValidator.validate(filePath, content);
        fileViolations.push(...cqrsViolations);
        stats.violations.cqrs += cqrsViolations.length;
      }
      
      // 端口配置验证
      if (this.options.rules.portConfiguration.enabled) {
        const portViolations = PortConfigurationValidator.validate(filePath, content);
        fileViolations.push(...portViolations);
        stats.violations.ports += portViolations.length;
      }
      
      // API契约验证
      if (this.options.rules.apiContracts.enabled) {
        const contractViolations = APIContractValidator.validate(filePath, content);
        fileViolations.push(...contractViolations);
        stats.violations.contracts += contractViolations.length;
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
  
  generateReport() {
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
          contracts: stats.violations.contracts
        }
      },
      violations: this.violations
    };
    
    // 确保报告目录存在
    fs.mkdirSync(config.reportDir, { recursive: true });
    
    // 保存JSON报告
    const reportPath = path.join(config.reportDir, 'architecture-validation.json');
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
    
    // 质量门禁判定
    const criticalViolations = stats.violations.cqrs + stats.violations.ports;
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
  
  log.info('🏗️ Cube Castle - 架构验证器启动');
  log.info(`验证范围: ${scope}`);
  
  const validator = new ArchitectureValidator();
  
  // 确定验证路径
  const targetPath = scope === 'frontend' ? 
    path.join(config.projectRoot, 'frontend/src') : 
    config.projectRoot;
  
  try {
    await validator.validateDirectory(targetPath);
    const report = validator.generateReport();
    const success = validator.printSummary();
    
    log.info(`📂 详细报告: ${path.join(config.reportDir, 'architecture-validation.json')}`);
    
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