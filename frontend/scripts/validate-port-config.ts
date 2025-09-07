#!/usr/bin/env tsx
/**
 * 端口配置验证脚本
 * 🎯 确保端口配置的一致性和无冲突
 * 🔍 扫描代码库中的硬编码端口
 */

import { readFileSync, readdirSync, statSync } from 'fs';
import { join, extname } from 'path';
import { SERVICE_PORTS, validatePortConfiguration, generatePortConfigReport } from '../src/shared/config/ports';

// 🎯 扫描硬编码端口的文件类型
const SCAN_EXTENSIONS = ['.ts', '.tsx', '.js', '.jsx', '.json', '.md'];

// 🎯 排除目录
const EXCLUDE_DIRS = ['node_modules', '.git', 'dist', 'build', '.next', 'coverage'];

// 🎯 硬编码端口模式
const PORT_PATTERNS = [
  /localhost:\d{4}/g,
  /127\.0\.0\.1:\d{4}/g,
  /:\s*\d{4}/g,
  /port.*=.*\d{4}/gi
];

interface HardcodedPort {
  file: string;
  line: number;
  content: string;
  port: string;
  isLegitimate: boolean;
}

// 🎯 合法端口使用（允许的硬编码）
const LEGITIMATE_PORTS = [
  '3000', '3001', '8090', '9090', '5432', '6379', '9091', '3002', '9093', '9100'
];

/**
 * 递归扫描目录
 */
function scanDirectory(dir: string, results: HardcodedPort[] = []): HardcodedPort[] {
  const items = readdirSync(dir);
  
  for (const item of items) {
    const fullPath = join(dir, item);
    const stat = statSync(fullPath);
    
    if (stat.isDirectory()) {
      if (!EXCLUDE_DIRS.includes(item)) {
        scanDirectory(fullPath, results);
      }
    } else if (stat.isFile()) {
      if (SCAN_EXTENSIONS.includes(extname(fullPath))) {
        scanFile(fullPath, results);
      }
    }
  }
  
  return results;
}

/**
 * 扫描单个文件
 */
function scanFile(filePath: string, results: HardcodedPort[]): void {
  try {
    const content = readFileSync(filePath, 'utf-8');
    const lines = content.split('\n');
    
    lines.forEach((line, index) => {
      // 跳过注释行和导入语句中的端口配置
      if (line.trim().startsWith('//') || 
          line.trim().startsWith('*') ||
          line.includes('from ') && line.includes('ports')) {
        return;
      }
      
      PORT_PATTERNS.forEach(pattern => {
        const matches = line.match(pattern);
        if (matches) {
          matches.forEach(match => {
            const port = match.match(/\d{4}/)?.[0];
            if (port) {
              results.push({
                file: filePath.replace(process.cwd(), ''),
                line: index + 1,
                content: line.trim(),
                port,
                isLegitimate: LEGITIMATE_PORTS.includes(port)
              });
            }
          });
        }
      });
    });
  } catch (error) {
    console.warn(`无法扫描文件 ${filePath}:`, error);
  }
}

/**
 * 生成端口使用报告
 */
function generateUsageReport(hardcodedPorts: HardcodedPort[]): string {
  const byPort = hardcodedPorts.reduce((acc, item) => {
    if (!acc[item.port]) {
      acc[item.port] = [];
    }
    acc[item.port].push(item);
    return acc;
  }, {} as Record<string, HardcodedPort[]>);

  const report = [
    '🔍 端口使用分析报告',
    '========================',
    '',
  ];

  Object.keys(byPort).sort().forEach(port => {
    const items = byPort[port];
    const isConfigured = Object.values(SERVICE_PORTS).includes(Number(port));
    const status = isConfigured ? '✅ 已配置' : '❌ 未配置';
    
    report.push(`📍 端口 ${port} (${status}):`);
    
    items.forEach(item => {
      const legitimacy = item.isLegitimate ? '✅' : '⚠️';
      report.push(`  ${legitimacy} ${item.file}:${item.line}`);
      report.push(`     ${item.content}`);
    });
    
    report.push('');
  });

  return report.join('\n');
}

/**
 * 主要验证逻辑
 */
async function main() {
  console.log('🎯 开始端口配置验证...\n');
  
  // 1. 验证端口配置本身
  console.log(generatePortConfigReport());
  console.log('');
  
  const configValidation = validatePortConfiguration();
  if (!configValidation.isValid) {
    console.error('❌ 端口配置验证失败:');
    configValidation.errors.forEach(error => console.error(`  - ${error}`));
    process.exit(1);
  }
  
  // 2. 扫描硬编码端口
  console.log('🔍 扫描硬编码端口...');
  const hardcodedPorts = scanDirectory(process.cwd());
  
  // 3. 生成报告
  console.log(generateUsageReport(hardcodedPorts));
  
  // 4. 检查问题端口
  const problematicPorts = hardcodedPorts.filter(p => !p.isLegitimate);
  
  if (problematicPorts.length > 0) {
    console.warn('⚠️  发现可能问题的端口配置:');
    problematicPorts.forEach(p => {
      console.warn(`  ${p.file}:${p.line} - 端口 ${p.port}`);
    });
  }
  
  // 5. 总结
  const totalHardcoded = hardcodedPorts.length;
  const legitimateCount = hardcodedPorts.filter(p => p.isLegitimate).length;
  const configuredPorts = new Set(Object.values(SERVICE_PORTS)).size;
  
  console.log('');
  console.log('📊 验证汇总:');
  console.log(`  - 配置的端口数量: ${configuredPorts}`);
  console.log(`  - 发现硬编码端口: ${totalHardcoded}`);
  console.log(`  - 合法使用: ${legitimateCount}`);
  console.log(`  - 可能问题: ${problematicPorts.length}`);
  
  if (problematicPorts.length === 0) {
    console.log('✅ 端口配置验证通过！');
    process.exit(0);
  } else {
    console.log('⚠️  端口配置需要检查');
    process.exit(1);
  }
}

// 运行验证
main().catch(error => {
  console.error('验证过程出错:', error);
  process.exit(1);
});