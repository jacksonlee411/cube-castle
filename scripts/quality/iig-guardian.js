#!/usr/bin/env node

/**
 * Cube Castle - IIG护卫系统 (Implementation Inventory Guardian)
 * 实现清单护卫系统：防止重复开发，维护实现唯一性
 * 
 * 核心功能:
 * - 预开发强制检查：运行前分析现有实现，防止重复造轮子
 * - 功能重复检测：深度分析API、组件、服务的重复性
 * - 实现清单管理：维护权威的功能清单索引
 * - P3系统集成：与重复代码检测、架构验证、文档同步深度融合
 * 
 * 作者: Claude Code Assistant (IIG护卫专员)
 * 日期: 2025-09-10
 * 版本: v1.0.0
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// 🎨 颜色配置
const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  shield: '🛡️',
  robot: '🤖',
  search: '🔍',
  warning: '⚠️',
  success: '✅',
  error: '❌'
};

// 🔧 IIG护卫配置
const iigConfig = {
  projectRoot: process.cwd(),
  inventoryScript: 'scripts/generate-implementation-inventory.js',
  reportDir: path.join(process.cwd(), 'reports', 'iig-guardian'),
  
  // 关键实现清单路径
  inventoryDoc: 'docs/reference/02-IMPLEMENTATION-INVENTORY.md',
  
  // P3系统集成
  p3Systems: {
    duplicateDetection: 'scripts/quality/duplicate-detection.sh',
    architectureValidator: 'scripts/quality/architecture-validator.js',
    documentSync: 'scripts/quality/document-sync.js'
  },
  
  // 重复检测规则
  duplicationRules: {
    // API端点重复检测
    apiEndpoints: {
      restPattern: /\/api\/v\d+\/[\w\-\/{}]+/g,
      graphqlPattern: /query|mutation\s+\w+/g,
      threshold: 0.8  // 80%相似度视为重复
    },
    
    // 组件重复检测
    components: {
      hookPattern: /^use[A-Z]\w+/,
      componentPattern: /^[A-Z]\w+Component$/,
      servicePattern: /^[A-Z]\w+Service$/,
      threshold: 0.7
    },
    
    // 功能重复检测
    functionality: {
      crudOperations: ['create', 'read', 'update', 'delete', 'list'],
      businessDomains: ['organization', 'user', 'auth', 'temporal', 'hierarchy'],
      threshold: 0.9
    }
  },
  
  // 风险评估阈值
  riskThresholds: {
    high: 0.9,      // 90%+ 相似度 = 高风险重复
    medium: 0.7,    // 70-89% 相似度 = 中风险重复  
    low: 0.5        // 50-69% 相似度 = 低风险重复
  }
};

// 📊 IIG护卫统计
const guardianStats = {
  scannedFiles: 0,
  analysedImplementations: 0,
  detectedDuplicates: 0,
  preventedDuplicates: 0,
  riskAssessments: {
    high: 0,
    medium: 0,
    low: 0
  },
  p3Integration: {
    duplicateCodeRate: 0,
    architectureViolations: 0,
    documentSyncRate: 0
  }
};

// 📋 日志系统
const guardianLog = {
  shield: (msg) => console.log(`${colors.blue}${colors.shield} [IIG护卫] ${msg}${colors.reset}`),
  search: (msg) => console.log(`${colors.cyan}${colors.search} [扫描] ${msg}${colors.reset}`),
  success: (msg) => console.log(`${colors.green}${colors.success} [成功] ${msg}${colors.reset}`),
  warning: (msg) => console.log(`${colors.yellow}${colors.warning} [警告] ${msg}${colors.reset}`),
  error: (msg) => console.error(`${colors.red}${colors.error} [错误] ${msg}${colors.reset}`),
  info: (msg) => console.log(`${colors.blue}ℹ️  [信息] ${msg}${colors.reset}`),
  robot: (msg) => console.log(`${colors.magenta}${colors.robot} [AI分析] ${msg}${colors.reset}`)
};

// 🔍 实现清单分析器
class ImplementationInventoryAnalyzer {
  constructor() {
    this.currentInventory = null;
    this.implementations = new Map();
  }
  
  // 生成最新实现清单
  async generateCurrentInventory() {
    guardianLog.search('执行实现清单生成...');
    
    try {
      const inventoryOutput = execSync(`node ${iigConfig.inventoryScript}`, { 
        encoding: 'utf8',
        cwd: iigConfig.projectRoot 
      });
      
      this.currentInventory = inventoryOutput;
      guardianStats.scannedFiles++;
      
      guardianLog.success('实现清单生成完成');
      return inventoryOutput;
      
    } catch (error) {
      guardianLog.error(`实现清单生成失败: ${error.message}`);
      throw error;
    }
  }
  
  // 解析实现清单
  parseInventory(inventoryText) {
    const implementations = {
      restAPIs: [],
      graphqlQueries: [],
      goHandlers: [],
      goServices: [],
      frontendExports: [],
      scripts: []
    };
    
    // 解析REST API端点
    const restMatches = inventoryText.match(/- `\/api\/v\d+\/[\w\-\/{}]+`/g) || [];
    implementations.restAPIs = restMatches.map(match => 
      match.replace(/- `|`/g, '').trim()
    );
    
    // 解析GraphQL查询
    const graphqlMatches = inventoryText.match(/- `\w+(\([^)]*\))?(: \w+!?)?`/g) || [];
    implementations.graphqlQueries = graphqlMatches.map(match => 
      match.replace(/- `|`/g, '').split('(')[0].trim()
    );
    
    // 解析Go处理器
    const handlerMatches = inventoryText.match(/- \w+ — [\w\/\-\.]+\.go/g) || [];
    implementations.goHandlers = handlerMatches.map(match => 
      match.replace(/- /, '').split(' — ')[0].trim()
    );
    
    // 解析前端导出
    const frontendMatches = inventoryText.match(/- \[(const|func|class)\] \w+ — [\w\/\-\.]+\.ts/g) || [];
    implementations.frontendExports = frontendMatches.map(match => {
      const parts = match.replace(/- \[(const|func|class)\] /, '').split(' — ');
      return {
        name: parts[0].trim(),
        type: match.match(/\[(const|func|class)\]/)[1],
        file: parts[1].trim()
      };
    });
    
    guardianStats.analysedImplementations = 
      implementations.restAPIs.length + 
      implementations.graphqlQueries.length + 
      implementations.goHandlers.length + 
      implementations.frontendExports.length;
    
    return implementations;
  }
}

// 🔍 重复功能检测器
class DuplicationDetector {
  constructor() {
    this.similarityCache = new Map();
  }
  
  // 检测API端点重复
  detectAPIDuplication(existingAPIs, newAPICandidate) {
    const duplicates = [];
    
    for (const existingAPI of existingAPIs) {
      const similarity = this.calculateStringSimilarity(existingAPI, newAPICandidate);
      
      if (similarity >= iigConfig.duplicationRules.apiEndpoints.threshold) {
        duplicates.push({
          existing: existingAPI,
          candidate: newAPICandidate,
          similarity: similarity,
          risk: this.calculateRiskLevel(similarity),
          recommendation: this.generateRecommendation('api', existingAPI, similarity)
        });
      }
    }
    
    return duplicates;
  }
  
  // 检测组件功能重复
  detectComponentDuplication(existingComponents, newComponentCandidate) {
    const duplicates = [];
    
    for (const component of existingComponents) {
      // 名称相似度检测
      const nameSimilarity = this.calculateStringSimilarity(
        component.name, 
        newComponentCandidate
      );
      
      // 功能相似度检测
      const functionalSimilarity = this.calculateFunctionalSimilarity(
        component, 
        newComponentCandidate
      );
      
      const overallSimilarity = Math.max(nameSimilarity, functionalSimilarity);
      
      if (overallSimilarity >= iigConfig.duplicationRules.components.threshold) {
        duplicates.push({
          existing: component,
          candidate: newComponentCandidate,
          similarity: overallSimilarity,
          risk: this.calculateRiskLevel(overallSimilarity),
          recommendation: this.generateRecommendation('component', component, overallSimilarity)
        });
      }
    }
    
    return duplicates;
  }
  
  // 字符串相似度计算 (Levenshtein距离算法)
  calculateStringSimilarity(str1, str2) {
    const longer = str1.length > str2.length ? str1 : str2;
    const shorter = str1.length > str2.length ? str2 : str1;
    
    if (longer.length === 0) return 1.0;
    
    const editDistance = this.levenshteinDistance(longer, shorter);
    return (longer.length - editDistance) / longer.length;
  }
  
  // Levenshtein距离计算
  levenshteinDistance(str1, str2) {
    const matrix = Array(str2.length + 1).fill(null).map(() => 
      Array(str1.length + 1).fill(null)
    );
    
    for (let i = 0; i <= str1.length; i++) matrix[0][i] = i;
    for (let j = 0; j <= str2.length; j++) matrix[j][0] = j;
    
    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const substitutionCost = str1[i - 1] === str2[j - 1] ? 0 : 1;
        matrix[j][i] = Math.min(
          matrix[j][i - 1] + 1,     // deletion
          matrix[j - 1][i] + 1,     // insertion
          matrix[j - 1][i - 1] + substitutionCost // substitution
        );
      }
    }
    
    return matrix[str2.length][str1.length];
  }
  
  // 功能相似度计算
  calculateFunctionalSimilarity(existing, candidate) {
    // 基于命名模式和业务领域的功能相似度
    const domains = iigConfig.duplicationRules.functionality.businessDomains;
    const operations = iigConfig.duplicationRules.functionality.crudOperations;
    
    let domainMatch = 0;
    let operationMatch = 0;
    
    // 检查业务领域匹配
    for (const domain of domains) {
      if (existing.name?.toLowerCase().includes(domain) && 
          candidate.toLowerCase().includes(domain)) {
        domainMatch = 1;
        break;
      }
    }
    
    // 检查操作类型匹配
    for (const operation of operations) {
      if (existing.name?.toLowerCase().includes(operation) && 
          candidate.toLowerCase().includes(operation)) {
        operationMatch = 1;
        break;
      }
    }
    
    return (domainMatch + operationMatch) / 2;
  }
  
  // 风险等级计算
  calculateRiskLevel(similarity) {
    if (similarity >= iigConfig.riskThresholds.high) return 'HIGH';
    if (similarity >= iigConfig.riskThresholds.medium) return 'MEDIUM';
    if (similarity >= iigConfig.riskThresholds.low) return 'LOW';
    return 'MINIMAL';
  }
  
  // 生成建议
  generateRecommendation(type, existing, similarity) {
    const riskLevel = this.calculateRiskLevel(similarity);
    
    const recommendations = {
      api: {
        HIGH: `强烈建议复用现有API端点: ${existing}`,
        MEDIUM: `考虑扩展现有API端点: ${existing}`,
        LOW: `参考现有API设计模式: ${existing}`
      },
      component: {
        HIGH: `直接复用现有组件: ${existing.name} (${existing.file})`,
        MEDIUM: `考虑扩展现有组件: ${existing.name}`,
        LOW: `参考现有组件实现: ${existing.name}`
      }
    };
    
    return recommendations[type]?.[riskLevel] || `分析现有实现: ${existing}`;
  }
}

// 🔗 P3系统集成器
class P3SystemIntegrator {
  constructor() {
    this.p3Results = {};
  }
  
  // 运行P3.1重复代码检测
  async runDuplicateDetection() {
    guardianLog.search('集成P3.1重复代码检测系统...');
    
    try {
      const result = execSync(`bash ${iigConfig.p3Systems.duplicateDetection}`, {
        encoding: 'utf8',
        cwd: iigConfig.projectRoot
      });
      
      // 解析重复率
      const duplicateRateMatch = result.match(/重复率:\s*([\d.]+)%/);
      if (duplicateRateMatch) {
        guardianStats.p3Integration.duplicateCodeRate = parseFloat(duplicateRateMatch[1]);
      }
      
      this.p3Results.duplicateDetection = result;
      guardianLog.success(`P3.1集成完成 - 重复率: ${guardianStats.p3Integration.duplicateCodeRate}%`);
      
    } catch (error) {
      guardianLog.warning(`P3.1集成失败: ${error.message}`);
    }
  }
  
  // 运行P3.2架构验证
  async runArchitectureValidation() {
    guardianLog.search('集成P3.2架构验证系统...');
    
    try {
      const result = execSync(`node ${iigConfig.p3Systems.architectureValidator}`, {
        encoding: 'utf8',
        cwd: iigConfig.projectRoot
      });
      
      // 解析违规数量
      const violationsMatch = result.match(/问题总数:\s*(\d+)\s*个/);
      if (violationsMatch) {
        guardianStats.p3Integration.architectureViolations = parseInt(violationsMatch[1]);
      }
      
      this.p3Results.architectureValidation = result;
      guardianLog.success(`P3.2集成完成 - 违规: ${guardianStats.p3Integration.architectureViolations}个`);
      
    } catch (error) {
      guardianLog.warning(`P3.2集成失败: ${error.message}`);
    }
  }
  
  // 运行P3.3文档同步检查
  async runDocumentSyncCheck() {
    guardianLog.search('集成P3.3文档同步系统...');
    
    try {
      const result = execSync(`node ${iigConfig.p3Systems.documentSync}`, {
        encoding: 'utf8',
        cwd: iigConfig.projectRoot
      });
      
      // 解析同步率
      const syncRateMatch = result.match(/同步率:\s*([\d.]+)%/);
      if (syncRateMatch) {
        guardianStats.p3Integration.documentSyncRate = parseFloat(syncRateMatch[1]);
      }
      
      this.p3Results.documentSync = result;
      guardianLog.success(`P3.3集成完成 - 同步率: ${guardianStats.p3Integration.documentSyncRate}%`);
      
    } catch (error) {
      guardianLog.warning(`P3.3集成失败: ${error.message}`);
    }
  }
  
  // 综合P3系统结果
  async integrateP3Systems() {
    await Promise.all([
      this.runDuplicateDetection(),
      this.runArchitectureValidation(),
      this.runDocumentSyncCheck()
    ]);
    
    return this.p3Results;
  }
}

// 🛡️ IIG护卫主引擎
class IIGGuardian {
  constructor() {
    this.analyzer = new ImplementationInventoryAnalyzer();
    this.detector = new DuplicationDetector();
    this.p3Integrator = new P3SystemIntegrator();
    this.findings = [];
  }
  
  // 执行预开发检查
  async performPreDevelopmentCheck(proposedFeature) {
    guardianLog.shield('🚀 IIG护卫系统启动 - 执行预开发检查');
    guardianLog.info(`检查提议功能: ${proposedFeature}`);
    
    // 第一步：生成当前实现清单
    const inventory = await this.analyzer.generateCurrentInventory();
    const implementations = this.analyzer.parseInventory(inventory);
    
    // 第二步：重复功能检测
    const duplicates = await this.detectDuplicates(implementations, proposedFeature);
    
    // 第三步：P3系统集成检查
    const p3Results = await this.p3Integrator.integrateP3Systems();
    
    // 第四步：风险评估和建议
    const assessment = this.generateRiskAssessment(duplicates, p3Results);
    
    return {
      inventory: implementations,
      duplicates: duplicates,
      p3Results: p3Results,
      assessment: assessment,
      recommendations: this.generateRecommendations(duplicates, assessment)
    };
  }
  
  // 检测重复功能
  async detectDuplicates(implementations, proposedFeature) {
    guardianLog.search('执行重复功能检测...');
    
    const duplicates = {
      apis: [],
      components: [],
      handlers: []
    };
    
    // API端点重复检测
    if (proposedFeature.includes('api') || proposedFeature.includes('endpoint')) {
      duplicates.apis = this.detector.detectAPIDuplication(
        implementations.restAPIs, 
        proposedFeature
      );
    }
    
    // 组件重复检测
    if (proposedFeature.includes('component') || proposedFeature.includes('hook')) {
      duplicates.components = this.detector.detectComponentDuplication(
        implementations.frontendExports,
        proposedFeature
      );
    }
    
    // 处理器重复检测
    if (proposedFeature.includes('handler') || proposedFeature.includes('service')) {
      duplicates.handlers = this.detector.detectAPIDuplication(
        implementations.goHandlers,
        proposedFeature
      );
    }
    
    const totalDuplicates = duplicates.apis.length + duplicates.components.length + duplicates.handlers.length;
    guardianStats.detectedDuplicates = totalDuplicates;
    
    if (totalDuplicates > 0) {
      guardianLog.warning(`检测到 ${totalDuplicates} 个潜在重复实现`);
    } else {
      guardianLog.success('未检测到重复实现，可以继续开发');
    }
    
    return duplicates;
  }
  
  // 生成风险评估
  generateRiskAssessment(duplicates, p3Results) {
    const assessment = {
      overallRisk: 'LOW',
      factors: [],
      scores: {
        duplication: 0,
        codeQuality: 0,
        architecture: 0,
        documentation: 0
      }
    };
    
    // 重复风险评分
    const highRiskDuplicates = Object.values(duplicates).flat()
      .filter(d => d.risk === 'HIGH').length;
    
    if (highRiskDuplicates > 0) {
      assessment.overallRisk = 'HIGH';
      assessment.factors.push(`发现 ${highRiskDuplicates} 个高风险重复实现`);
      assessment.scores.duplication = 90;
    }
    
    // 代码质量评分
    if (guardianStats.p3Integration.duplicateCodeRate > 5) {
      assessment.factors.push(`代码重复率过高: ${guardianStats.p3Integration.duplicateCodeRate}%`);
      assessment.scores.codeQuality = 80;
      if (assessment.overallRisk === 'LOW') assessment.overallRisk = 'MEDIUM';
    }
    
    // 架构一致性评分
    if (guardianStats.p3Integration.architectureViolations > 20) {
      assessment.factors.push(`架构违规过多: ${guardianStats.p3Integration.architectureViolations}个`);
      assessment.scores.architecture = 70;
      if (assessment.overallRisk === 'LOW') assessment.overallRisk = 'MEDIUM';
    }
    
    // 文档同步评分
    if (guardianStats.p3Integration.documentSyncRate < 80) {
      assessment.factors.push(`文档同步率不足: ${guardianStats.p3Integration.documentSyncRate}%`);
      assessment.scores.documentation = 60;
    }
    
    return assessment;
  }
  
  // 生成建议
  generateRecommendations(duplicates, assessment) {
    const recommendations = [];
    
    // 重复实现建议
    Object.values(duplicates).flat().forEach(duplicate => {
      recommendations.push({
        type: 'duplication',
        priority: duplicate.risk,
        message: duplicate.recommendation,
        action: 'reuse_existing'
      });
    });
    
    // 质量改进建议
    if (assessment.overallRisk === 'HIGH') {
      recommendations.push({
        type: 'quality',
        priority: 'HIGH',
        message: '建议暂缓开发新功能，优先修复现有质量问题',
        action: 'fix_existing_issues'
      });
    }
    
    // P3系统建议
    if (guardianStats.p3Integration.duplicateCodeRate > 5) {
      recommendations.push({
        type: 'code_quality',
        priority: 'MEDIUM',
        message: `运行重复代码清理：bash ${iigConfig.p3Systems.duplicateDetection} --fix`,
        action: 'run_duplicate_cleanup'
      });
    }
    
    return recommendations;
  }
  
  // 生成护卫报告
  generateGuardianReport(checkResults) {
    const report = {
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      guardian: {
        status: 'ACTIVE',
        mode: 'PRE_DEVELOPMENT_CHECK'
      },
      statistics: guardianStats,
      checkResults: checkResults,
      summary: {
        totalImplementations: guardianStats.analysedImplementations,
        duplicatesDetected: guardianStats.detectedDuplicates,
        riskLevel: checkResults.assessment.overallRisk,
        p3Integration: guardianStats.p3Integration
      }
    };
    
    // 确保报告目录存在
    fs.mkdirSync(iigConfig.reportDir, { recursive: true });
    
    // 保存JSON报告
    const reportPath = path.join(iigConfig.reportDir, 'iig-guardian-report.json');
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    guardianLog.success(`IIG护卫报告已生成: ${reportPath}`);
    return report;
  }
  
  // 打印护卫摘要
  printGuardianSummary(checkResults) {
    console.log(`\n${colors.bright}${colors.cyan}======================================${colors.reset}`);
    console.log(`${colors.bright}${colors.cyan}    🛡️ IIG护卫系统检查报告    ${colors.reset}`);
    console.log(`${colors.bright}${colors.cyan}======================================${colors.reset}\n`);
    
    // 实现清单统计
    guardianLog.info(`📊 实现清单统计:`);
    console.log(`   📁 已扫描文件: ${guardianStats.scannedFiles} 个`);
    console.log(`   🔍 已分析实现: ${guardianStats.analysedImplementations} 个`);
    console.log(`   ⚠️  检测重复: ${guardianStats.detectedDuplicates} 个`);
    
    // P3系统集成状态
    guardianLog.info(`🔗 P3系统集成状态:`);
    console.log(`   📊 代码重复率: ${guardianStats.p3Integration.duplicateCodeRate}%`);
    console.log(`   🏗️  架构违规: ${guardianStats.p3Integration.architectureViolations} 个`);
    console.log(`   📚 文档同步率: ${guardianStats.p3Integration.documentSyncRate}%`);
    
    // 风险评估
    const riskColor = {
      'HIGH': colors.red,
      'MEDIUM': colors.yellow,
      'LOW': colors.green
    }[checkResults.assessment.overallRisk] || colors.blue;
    
    console.log(`\n${riskColor}🎯 总体风险评估: ${checkResults.assessment.overallRisk}${colors.reset}`);
    
    // 重复检测结果
    if (guardianStats.detectedDuplicates > 0) {
      guardianLog.warning('🚨 发现潜在重复实现:');
      
      Object.entries(checkResults.duplicates).forEach(([type, duplicates]) => {
        if (duplicates.length > 0) {
          console.log(`   ${type}: ${duplicates.length} 个重复`);
          duplicates.forEach(dup => {
            console.log(`     - ${dup.existing} (相似度: ${(dup.similarity * 100).toFixed(1)}%)`);
          });
        }
      });
    }
    
    // 建议
    if (checkResults.recommendations.length > 0) {
      guardianLog.info('💡 IIG护卫建议:');
      checkResults.recommendations.forEach((rec, index) => {
        const priorityColor = {
          'HIGH': colors.red,
          'MEDIUM': colors.yellow,
          'LOW': colors.green
        }[rec.priority] || colors.blue;
        
        console.log(`   ${index + 1}. ${priorityColor}[${rec.priority}] ${rec.message}${colors.reset}`);
      });
    }
    
    // 最终决策
    if (checkResults.assessment.overallRisk === 'HIGH') {
      guardianLog.error('🛑 IIG护卫决策: 建议暂停开发，优先处理现有问题');
      return false;
    } else if (guardianStats.detectedDuplicates > 0) {
      guardianLog.warning('⚠️  IIG护卫决策: 可以继续开发，但建议优先复用现有实现');
      return true;
    } else {
      guardianLog.success('✅ IIG护卫决策: 可以安全开发新功能');
      return true;
    }
  }
}

// 🎯 CLI入口点
async function main() {
  const args = process.argv.slice(2);
  const proposedFeature = args[0] || 'new-feature';
  const mode = args.includes('--check') ? 'check' : 'guard';
  
  guardianLog.shield('🛡️ Cube Castle IIG护卫系统启动');
  guardianLog.info(`护卫模式: ${mode}`);
  guardianLog.info(`提议功能: ${proposedFeature}`);
  
  const guardian = new IIGGuardian();
  
  try {
    if (mode === 'guard') {
      // 执行完整护卫检查
      const checkResults = await guardian.performPreDevelopmentCheck(proposedFeature);
      const report = guardian.generateGuardianReport(checkResults);
      const shouldProceed = guardian.printGuardianSummary(checkResults);
      
      guardianLog.shield(`🛡️ IIG护卫系统检查完成`);
      process.exit(shouldProceed ? 0 : 1);
      
    } else {
      // 仅执行清单检查
      await guardian.analyzer.generateCurrentInventory();
      guardianLog.success('实现清单检查完成');
      process.exit(0);
    }
    
  } catch (error) {
    guardianLog.error(`IIG护卫系统错误: ${error.message}`);
    process.exit(1);
  }
}

// 运行主程序
if (require.main === module) {
  main().catch(error => {
    console.error('IIG Guardian failed:', error);
    process.exit(1);
  });
}

module.exports = { 
  IIGGuardian, 
  ImplementationInventoryAnalyzer, 
  DuplicationDetector,
  P3SystemIntegrator,
  iigConfig 
};