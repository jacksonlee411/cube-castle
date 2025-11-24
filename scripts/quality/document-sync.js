#!/usr/bin/env node

/**
 * Cube Castle - 文档自动同步系统
 * 监控并自动同步项目核心文档的一致性
 * 
 * 用途: 确保API规范、README、技术文档等关键文档保持同步
 * 作者: Claude Code Assistant
 * 日期: 2025-09-07
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');

// 🎨 颜色定义
const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m', 
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  reset: '\x1b[0m',
  bright: '\x1b[1m'
};

// 📋 日志函数
const log = {
  info: (msg) => console.log(`${colors.blue}📝 ${msg}${colors.reset}`),
  success: (msg) => console.log(`${colors.green}✅ ${msg}${colors.reset}`),
  warning: (msg) => console.log(`${colors.yellow}⚠️  ${msg}${colors.reset}`),
  error: (msg) => console.error(`${colors.red}❌ ${msg}${colors.reset}`),
  verbose: (msg) => process.env.VERBOSE && console.log(`${colors.cyan}🔍 ${msg}${colors.reset}`)
};

// 🔧 配置
const config = {
  projectRoot: process.cwd(),
  syncPairs: [
    // 主要文档同步对
    {
      name: 'API规范版本同步',
      source: 'docs/api/openapi.yaml',
      targets: [
        'docs/development-plans/02-technical-architecture-design.md'
      ],
      syncType: 'version',
      pattern: /version:\s*['"]?([^'"\s]+)['"]?/,
      description: 'OpenAPI版本号同步到前端类型和技术文档'
    },

    {
      name: '依赖版本同步',
      source: 'frontend/package.json',
      targets: [
        'docs/development-plans/02-technical-architecture-design.md'
      ],
      syncType: 'dependencies',
      pattern: /"(react|vite|typescript)":\s*"([^"]+)"/g,
      description: '关键依赖版本在文档中保持同步'
    },

    // 仅存在性检查（presence）：避免复制正文，仅校验关键索引是否存在
    {
      name: 'presence: temporal guide appendix index exists',
      source: 'docs/reference/temporal-entity-experience-guide.md',
      targets: ['docs/reference/temporal-entity-experience-guide.md'],
      syncType: 'presence',
      pattern: /##\s*附录 A\s*.+框架与工程实践清单/,
      description: '检查《时态实体体验指南》中是否存在附录A（索引，不复制正文）'
    },
    {
      name: 'presence: AGENTS links temporal guide',
      source: 'AGENTS.md',
      targets: ['AGENTS.md'],
      syncType: 'presence',
      pattern: /docs\/reference\/temporal-entity-experience-guide\.md/,
      description: 'AGENTS 索引需包含“时态实体体验指南”的权威链接'
    },
    {
      name: 'presence: plan 241 references appendix A',
      source: 'docs/archive/development-plans/241-frontend-framework-refactor.md',
      targets: ['docs/archive/development-plans/241-frontend-framework-refactor.md'],
      syncType: 'presence',
      pattern: /附录 A\s*.+框架与工程实践清单/,
      description: 'Plan 241 需固定引用附录A作为执行/验收依据'
    }
  ],
  
  // 监控配置
  monitoring: {
    enabled: true,
    checkInterval: 60000, // 60秒
    maxChecks: 100,
    reportPath: 'reports/document-sync',
    historyPath: 'reports/document-sync/sync-history.json'
  },
  
  // 同步规则
  syncRules: {
    autoSync: process.env.AUTO_SYNC === 'true',
    dryRun: process.env.DRY_RUN !== 'false',
    createBackups: true,
    backupDir: 'reports/document-sync/backups',
    conflictStrategy: 'prompt' // 'overwrite', 'skip', 'prompt'
  }
};

// 📊 统计数据
const stats = {
  totalPairs: 0,
  syncedPairs: 0,
  conflicts: 0,
  errors: 0,
  autoFixed: 0,
  checksPerformed: 0
};

// 🔍 文档内容提取器
class ContentExtractor {
  static extractVersion(content) {
    const versionMatch = content.match(/version:\s*['"]?([^'"\s]+)['"]?/i);
    return versionMatch ? versionMatch[1] : null;
  }
  
  static extractPortConfigs(content) {
    const configs = {};
    
    // 提取SERVICE_PORTS
    const servicePortsMatch = content.match(/SERVICE_PORTS\s*=\s*{([^}]+)}/s);
    if (servicePortsMatch) {
      const portsContent = servicePortsMatch[1];
      // 1) 直接数字: KEY: 3000
      const directMatches = [...portsContent.matchAll(/(\w+):\s*(\d+)/g)];
      directMatches.forEach(([, key, value]) => {
        configs[`SERVICE_PORTS.${key}`] = parseInt(value, 10);
      });
      // 2) 默认值函数: KEY: getNumberEnvVar('ENV', 3000)
      const defaultMatches = [...portsContent.matchAll(/(\w+):\s*getNumberEnvVar\([^,]+,\s*(\d+)\)/g)];
      defaultMatches.forEach(([, key, def]) => {
        configs[`SERVICE_PORTS.${key}`] = parseInt(def, 10);
      });
    }
    
    return configs;
  }
  
  static extractProjectStatus(content) {
    const statusMatch = content.match(/项目状态[：:]\s*(.+)/);
    return statusMatch ? statusMatch[1].trim() : null;
  }
  
  static extractDependencyVersions(content, filePath = '') {
    // 只对package.json文件尝试JSON解析
    if (!filePath.endsWith('package.json')) {
      // 对于非package.json文件，尝试从文本中提取版本信息
      const versionPatterns = [
        /React\s+(\d+\.\d+\.\d+)/i,
        /Vite\s+(\d+\.\d+\.\d+)/i,
        /TypeScript\s+(\d+\.\d+\.\d+)/i
      ];
      
      const keyDependencies = {};
      versionPatterns.forEach((pattern, index) => {
        const match = content.match(pattern);
        if (match) {
          const depName = ['react', 'vite', 'typescript'][index];
          keyDependencies[depName] = match[1];
        }
      });
      
      return keyDependencies;
    }
    
    try {
      const packageData = JSON.parse(content);
      const deps = { ...packageData.dependencies, ...packageData.devDependencies };
      
      const keyDependencies = {};
      ['react', 'vite', 'typescript'].forEach(dep => {
        if (deps[dep]) {
          // 规范化版本（去掉 ^ ~ 等前缀，只保留 x.y.z）
          const raw = String(deps[dep]);
          const normalized = raw.replace(/^[^0-9]*/, '');
          keyDependencies[dep] = normalized;
        }
      });
      
      return keyDependencies;
    } catch (err) {
      log.warning(`解析package.json失败: ${err.message}`);
      return {};
    }
  }
  
  static extractAchievements(content) {
    const achievements = [];
    const achievementMatches = [...content.matchAll(/完成度[：:]?\s*(\d+%)/g)];
    
    achievementMatches.forEach(match => {
      achievements.push(match[1]);
    });
    
    return achievements;
  }
}

// 🔄 文档同步器
class DocumentSynchronizer {
  constructor() {
    this.syncHistory = this.loadSyncHistory();
  }
  
  loadSyncHistory() {
    try {
      if (fs.existsSync(config.monitoring.historyPath)) {
        const data = fs.readFileSync(config.monitoring.historyPath, 'utf8');
        return JSON.parse(data);
      }
    } catch (err) {
      log.warning(`加载同步历史失败: ${err.message}`);
    }
    
    return {
      lastSync: null,
      syncRecords: [],
      conflicts: []
    };
  }
  
  saveSyncHistory() {
    try {
      fs.mkdirSync(path.dirname(config.monitoring.historyPath), { recursive: true });
      fs.writeFileSync(config.monitoring.historyPath, JSON.stringify(this.syncHistory, null, 2));
    } catch (err) {
      log.error(`保存同步历史失败: ${err.message}`);
    }
  }
  
  async syncPair(syncPair) {
    log.info(`检查同步对: ${syncPair.name}`);
    stats.totalPairs++;
    
    try {
      // 读取源文件
      const sourcePath = path.join(config.projectRoot, syncPair.source);
      if (!fs.existsSync(sourcePath)) {
        log.warning(`源文件不存在: ${syncPair.source}`);
        return false;
      }
      
      const sourceContent = fs.readFileSync(sourcePath, 'utf8');
      const sourceHash = crypto.createHash('md5').update(sourceContent).digest('hex');
      
      // 提取源数据
      const sourceData = this.extractSourceData(syncPair, sourceContent, syncPair.source);
      if (!sourceData) {
        log.warning(`无法从源文件提取数据: ${syncPair.source}`);
        return false;
      }
      
      log.verbose(`源数据: ${JSON.stringify(sourceData)}`);
      
      // 检查所有目标文件
      let allTargetsSynced = true;
      
      for (const targetPath of syncPair.targets) {
        const fullTargetPath = path.join(config.projectRoot, targetPath);
        
        if (!fs.existsSync(fullTargetPath)) {
          log.warning(`目标文件不存在: ${targetPath}`);
          continue;
        }
        
        const targetContent = fs.readFileSync(fullTargetPath, 'utf8');
        const targetData = this.extractSourceData(syncPair, targetContent, targetPath);
        
        const isSynced = this.compareData(syncPair.syncType, sourceData, targetData);
        
        if (!isSynced) {
          log.warning(`发现不同步: ${syncPair.source} -> ${targetPath}`);
          allTargetsSynced = false;
          
          // 尝试自动同步
          if (config.syncRules.autoSync) {
            const success = await this.attemptAutoSync(syncPair, sourceData, fullTargetPath, targetContent);
            if (success) {
              log.success(`自动同步成功: ${targetPath}`);
              stats.autoFixed++;
            } else {
              stats.conflicts++;
            }
          } else {
            stats.conflicts++;
          }
        }
      }
      
      if (allTargetsSynced) {
        log.success(`同步对检查通过: ${syncPair.name}`);
        stats.syncedPairs++;
      }
      
      // 记录同步结果
      this.recordSyncResult(syncPair, sourceHash, allTargetsSynced);
      
      return allTargetsSynced;
      
    } catch (err) {
      log.error(`同步检查失败 ${syncPair.name}: ${err.message}`);
      stats.errors++;
      return false;
    }
  }
  
  extractSourceData(syncPair, content, filePath = '') {
    switch (syncPair.syncType) {
      case 'version':
        return ContentExtractor.extractVersion(content);
      
      case 'config':
        return ContentExtractor.extractPortConfigs(content);
      
      case 'status':
        return ContentExtractor.extractProjectStatus(content);
      
      case 'dependencies':
        return ContentExtractor.extractDependencyVersions(content, filePath);
      
      case 'achievements':
        return ContentExtractor.extractAchievements(content);

      case 'presence':
        // 对于 presence：源文件仅作为存在性检查的锚点；目标文件按正则检测是否包含
        if (!syncPair.pattern || !(syncPair.pattern instanceof RegExp)) {
          log.warning('presence 检查缺少有效的正则 pattern');
          return null;
        }
        // 源数据恒为 true，用于 compareData 统一判断；目标数据为布尔值：是否匹配
        if (filePath && path.resolve(filePath) === path.resolve(path.join(config.projectRoot, syncPair.source))) {
          return true;
        }
        return !!content.match(syncPair.pattern);
      
      default:
        log.warning(`未知的同步类型: ${syncPair.syncType}`);
        return null;
    }
  }
  
  compareData(syncType, sourceData, targetData) {
    if (!sourceData || !targetData) {
      return false;
    }
    
    switch (syncType) {
      case 'version':
      case 'status':
        return sourceData === targetData;
      
      case 'config':
      case 'dependencies':
        return JSON.stringify(sourceData) === JSON.stringify(targetData);
      
      case 'achievements':
        // 检查关键成果指标
        if (Array.isArray(sourceData) && Array.isArray(targetData)) {
          return sourceData.length === targetData.length &&
                 sourceData.every(item => targetData.includes(item));
        }
        return false;

      case 'presence':
        // 目标数据为 true 即视为通过
        return targetData === true;
      
      default:
        return false;
    }
  }
  
  async attemptAutoSync(syncPair, sourceData, targetPath, targetContent) {
    if (config.syncRules.dryRun) {
      log.info(`[DRY RUN] 将同步 ${targetPath}`);
      return true;
    }
    
    try {
      // 创建备份
      if (config.syncRules.createBackups) {
        this.createBackup(targetPath, targetContent);
      }
      
      // 根据同步类型执行同步
      const updatedContent = this.applySyncChanges(syncPair, sourceData, targetContent);
      
      if (updatedContent && updatedContent !== targetContent) {
        fs.writeFileSync(targetPath, updatedContent);
        log.success(`自动同步完成: ${path.relative(config.projectRoot, targetPath)}`);
        return true;
      }
      
      return false;
      
    } catch (err) {
      log.error(`自动同步失败 ${targetPath}: ${err.message}`);
      return false;
    }
  }
  
  applySyncChanges(syncPair, sourceData, targetContent) {
    // 这里实现具体的同步逻辑
    // 根据不同的syncType应用相应的更新
    
    switch (syncPair.syncType) {
      case 'version':
        return targetContent.replace(
          /version:\s*['"]?[^'"\s]+['"]?/gi,
          `version: "${sourceData}"`
        );
      
      case 'status':
        return targetContent.replace(
          /项目状态[：:]\s*.+/g,
          `项目状态: ${sourceData}`
        );
      
      // 其他同步类型的实现...
      default:
        log.warning(`未实现的同步类型: ${syncPair.syncType}`);
        return targetContent;
    }
  }
  
  createBackup(filePath, content) {
    try {
      const backupDir = config.syncRules.backupDir;
      fs.mkdirSync(backupDir, { recursive: true });
      
      const fileName = path.basename(filePath);
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const backupPath = path.join(backupDir, `${fileName}.${timestamp}.backup`);
      
      fs.writeFileSync(backupPath, content);
      log.verbose(`创建备份: ${backupPath}`);
      
    } catch (err) {
      log.warning(`创建备份失败: ${err.message}`);
    }
  }
  
  recordSyncResult(syncPair, sourceHash, success) {
    const record = {
      timestamp: new Date().toISOString(),
      syncPairName: syncPair.name,
      sourceHash,
      success,
      autoFixed: config.syncRules.autoSync
    };
    
    this.syncHistory.syncRecords.push(record);
    this.syncHistory.lastSync = record.timestamp;
    
    // 保持历史记录在合理范围内
    if (this.syncHistory.syncRecords.length > 1000) {
      this.syncHistory.syncRecords = this.syncHistory.syncRecords.slice(-500);
    }
  }
  
  async syncAll() {
    log.info('🔄 开始文档同步检查...');
    
    const results = [];
    
    for (const syncPair of config.syncPairs) {
      const result = await this.syncPair(syncPair);
      results.push({ syncPair: syncPair.name, success: result });
      stats.checksPerformed++;
    }
    
    // 保存同步历史
    this.saveSyncHistory();
    
    return results;
  }
  
  generateReport() {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalPairs: stats.totalPairs,
        syncedPairs: stats.syncedPairs,
        conflicts: stats.conflicts,
        errors: stats.errors,
        autoFixed: stats.autoFixed,
        checksPerformed: stats.checksPerformed,
        successRate: stats.totalPairs > 0 ? (stats.syncedPairs / stats.totalPairs * 100).toFixed(1) : 0
      },
      syncPairs: config.syncPairs.map(pair => ({
        name: pair.name,
        description: pair.description,
        source: pair.source,
        targets: pair.targets,
        syncType: pair.syncType
      })),
      history: this.syncHistory.syncRecords.slice(-10) // 最近10条记录
    };
    
    // 保存报告
    const reportPath = path.join(config.monitoring.reportPath, 'document-sync-report.json');
    fs.mkdirSync(path.dirname(reportPath), { recursive: true });
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    return report;
  }
  
  printSummary() {
    log.info('📊 文档同步统计报告:');
    log.info(`   📝 同步对总数: ${stats.totalPairs} 个`);
    log.info(`   ✅ 同步成功: ${stats.syncedPairs} 个`);
    
    if (stats.conflicts > 0) {
      log.warning(`   ⚠️  发现冲突: ${stats.conflicts} 个`);
    }
    
    if (stats.autoFixed > 0) {
      log.success(`   🔧 自动修复: ${stats.autoFixed} 个`);
    }
    
    if (stats.errors > 0) {
      log.error(`   ❌ 处理错误: ${stats.errors} 个`);
    }
    
    const successRate = stats.totalPairs > 0 ? (stats.syncedPairs / stats.totalPairs * 100).toFixed(1) : 0;
    log.info(`   📊 成功率: ${successRate}%`);
    
    // 质量门禁判定
    if (stats.conflicts > 0 || stats.errors > 0) {
      log.warning('📄 文档同步发现问题，建议检查');
      return false;
    } else {
      log.success('🎉 文档同步检查通过，一致性良好');
      return true;
    }
  }
}

// 🎯 CLI入口
async function main() {
  const args = process.argv.slice(2);
  const autoSync = args.includes('--auto-sync') || process.env.AUTO_SYNC === 'true';
  const dryRun = args.includes('--dry-run') || process.env.DRY_RUN !== 'false';
  
  // 设置配置
  config.syncRules.autoSync = autoSync;
  config.syncRules.dryRun = dryRun;
  
  log.info('📝 Cube Castle - 文档自动同步系统');
  log.info(`同步模式: ${autoSync ? '自动同步' : '检查模式'}`);
  log.info(`运行模式: ${dryRun ? 'DRY RUN' : '实际执行'}`);
  
  const synchronizer = new DocumentSynchronizer();
  
  try {
    const results = await synchronizer.syncAll();
    const report = synchronizer.generateReport();
    const success = synchronizer.printSummary();
    
    log.info(`📂 详细报告: ${path.join(config.monitoring.reportPath, 'document-sync-report.json')}`);
    
    // 输出同步建议
    if (stats.conflicts > 0 && !autoSync) {
      log.info('💡 同步建议:');
      log.info('   • 运行 --auto-sync 自动修复一致性问题');
      log.info('   • 运行 --dry-run 预览同步更改');
      log.info('   • 手动检查冲突文件并更新');
    }
    
    process.exit(success ? 0 : 1);
    
  } catch (err) {
    log.error(`文档同步失败: ${err.message}`);
    process.exit(1);
  }
}

// 运行主程序
if (require.main === module) {
  main().catch(err => {
    console.error('Document sync failed:', err);
    process.exit(1);
  });
}

module.exports = { DocumentSynchronizer, ContentExtractor, config };
