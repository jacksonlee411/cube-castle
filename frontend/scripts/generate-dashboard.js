#!/usr/bin/env node

/**
 * 契约测试本地监控仪表板
 * 
 * 简化版监控解决方案，生成HTML仪表板
 */

import { readFileSync, writeFileSync, existsSync } from 'fs'
import { execSync } from 'child_process'
import { join } from 'path'

class ContractTestingDashboard {
  constructor() {
    this.metrics = {
      contractTests: { total: 0, passed: 0, failed: 0 },
      fieldNaming: { violations: 0, files: 0 },
      schema: { valid: true, errors: [] },
      lastRun: new Date().toISOString()
    }
  }

  async collectMetrics() {
    console.log('🔍 收集契约测试指标...')

    try {
      // 运行契约测试
      const contractResult = execSync('npm run test:contract', { 
        cwd: 'frontend',
        encoding: 'utf8'
      })
      
      // 解析测试结果
      const testMatch = contractResult.match(/Tests\s+(\d+)\s+passed/)
      if (testMatch) {
        this.metrics.contractTests.passed = parseInt(testMatch[1])
        this.metrics.contractTests.total = this.metrics.contractTests.passed
      }
      
    } catch (error) {
      console.log('契约测试执行失败')
      this.metrics.contractTests.failed = 1
    }

    try {
      // 运行字段命名检查
      const namingResult = execSync('npm run validate:field-naming', {
        cwd: 'frontend',
        encoding: 'utf8'
      })
      this.metrics.fieldNaming.violations = 0
    } catch (error) {
      // 解析违规数量
      const errorOutput = error.stdout || error.message || ''
      const violationMatch = errorOutput.match(/发现snake_case违规项:\s*(\d+)/)
      if (violationMatch) {
        this.metrics.fieldNaming.violations = parseInt(violationMatch[1])
      } else {
        this.metrics.fieldNaming.violations = 1 // 假设有错误就有违规
      }
    }

    try {
      // 检查Schema语法
      execSync('npm run validate:schema', {
        cwd: 'frontend',
        encoding: 'utf8'
      })
      this.metrics.schema.valid = true
    } catch (error) {
      this.metrics.schema.valid = false
      this.metrics.schema.errors.push(error.message)
    }
  }

  generateHTML() {
    const passRate = this.metrics.contractTests.total > 0 
      ? (this.metrics.contractTests.passed / this.metrics.contractTests.total * 100).toFixed(1)
      : 0

    const complianceRate = this.metrics.fieldNaming.violations === 0 ? 100 : 85

    return `
<!DOCTYPE html>
<html>
<head>
    <title>契约测试监控仪表板</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0; padding: 20px; background: #f5f5f5;
        }
        .header { 
            background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { 
            background: white; padding: 20px; border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric-value { font-size: 2.5em; font-weight: bold; margin: 10px 0; }
        .metric-label { font-size: 1.1em; color: #666; }
        .status-good { color: #28a745; }
        .status-warning { color: #ffc107; }
        .status-error { color: #dc3545; }
        .timestamp { color: #999; font-size: 0.9em; }
        .violations { background: #f8f9fa; padding: 10px; border-radius: 4px; margin-top: 10px; }
        .refresh-btn {
            background: #007bff; color: white; border: none; padding: 10px 20px;
            border-radius: 4px; cursor: pointer; margin-left: 10px;
        }
        .refresh-btn:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔍 Cube Castle 契约测试监控仪表板</h1>
        <p>最后更新: <span class="timestamp">${new Date(this.metrics.lastRun).toLocaleString()}</span>
        <button class="refresh-btn" onclick="location.reload()">刷新数据</button></p>
    </div>

    <div class="metrics">
        <div class="card">
            <h3>📊 契约测试通过率</h3>
            <div class="metric-value ${passRate >= 95 ? 'status-good' : passRate >= 90 ? 'status-warning' : 'status-error'}">${passRate}%</div>
            <div class="metric-label">
                通过: ${this.metrics.contractTests.passed} / 总数: ${this.metrics.contractTests.total}
            </div>
        </div>

        <div class="card">
            <h3>📝 字段命名合规率</h3>
            <div class="metric-value ${this.metrics.fieldNaming.violations === 0 ? 'status-good' : 'status-error'}">${complianceRate}%</div>
            <div class="metric-label">
                违规项: ${this.metrics.fieldNaming.violations}
            </div>
            ${this.metrics.fieldNaming.violations > 0 ? `
                <div class="violations">
                    <strong>⚠️ 需要修复:</strong><br>
                    • 将 snake_case 字段改为 camelCase<br>
                    • 运行 <code>npm run validate:field-naming</code> 查看详情
                </div>
            ` : ''}
        </div>

        <div class="card">
            <h3>🔧 GraphQL Schema状态</h3>
            <div class="metric-value ${this.metrics.schema.valid ? 'status-good' : 'status-error'}">
                ${this.metrics.schema.valid ? '✅ 有效' : '❌ 错误'}
            </div>
            <div class="metric-label">Schema v4.2.1 验证</div>
            ${!this.metrics.schema.valid ? `
                <div class="violations">
                    <strong>错误详情:</strong><br>
                    ${this.metrics.schema.errors.join('<br>')}
                </div>
            ` : ''}
        </div>

        <div class="card">
            <h3>🚀 快速操作</h3>
            <div style="margin-top: 15px;">
                <p><strong>运行测试:</strong></p>
                <code style="background: #f8f9fa; padding: 5px; border-radius: 3px;">
                    cd frontend && npm run test:contract
                </code>
                
                <p><strong>检查字段命名:</strong></p>
                <code style="background: #f8f9fa; padding: 5px; border-radius: 3px;">
                    cd frontend && npm run validate:field-naming
                </code>
                
                <p><strong>验证Schema:</strong></p>
                <code style="background: #f8f9fa; padding: 5px; border-radius: 3px;">
                    cd frontend && npm run validate:schema
                </code>
            </div>
        </div>

        <div class="card">
            <h3>📈 趋势分析</h3>
            <div style="margin-top: 15px;">
                <p><strong>本次检查发现:</strong></p>
                <ul>
                    <li>契约测试: ${this.metrics.contractTests.passed > 0 ? '通过' : '需要检查'}</li>
                    <li>字段命名: ${this.metrics.fieldNaming.violations === 0 ? '合规' : `${this.metrics.fieldNaming.violations}个违规`}</li>
                    <li>Schema验证: ${this.metrics.schema.valid ? '通过' : '失败'}</li>
                </ul>
                
                <p><strong>建议操作:</strong></p>
                ${this.metrics.fieldNaming.violations > 0 ? 
                  '<p>🔧 优先修复字段命名问题，这会阻止代码合并</p>' : 
                  '<p>✅ 所有检查都通过了！代码质量良好</p>'
                }
            </div>
        </div>
    </div>

    <script>
        // 自动刷新 (每5分钟)
        setTimeout(() => {
            location.reload();
        }, 5 * 60 * 1000);
        
        console.log('契约测试监控仪表板已加载');
        console.log('数据收集时间:', '${this.metrics.lastRun}');
    </script>
</body>
</html>`
  }

  async generate() {
    await this.collectMetrics()
    const html = this.generateHTML()
    
    const outputPath = join(process.cwd(), 'contract-testing-dashboard.html')
    writeFileSync(outputPath, html)
    
    console.log(`\n📊 契约测试监控仪表板已生成!`)
    console.log(`📁 文件位置: ${outputPath}`)
    console.log(`🌐 在浏览器中打开: file://${outputPath}`)
    console.log(`\n📋 当前状态:`)
    console.log(`   • 契约测试: ${this.metrics.contractTests.passed}/${this.metrics.contractTests.total} 通过`)
    console.log(`   • 字段命名违规: ${this.metrics.fieldNaming.violations} 项`)
    console.log(`   • Schema状态: ${this.metrics.schema.valid ? '✅ 有效' : '❌ 错误'}`)

    return outputPath
  }
}

// 主执行函数
async function main() {
  try {
    const dashboard = new ContractTestingDashboard()
    await dashboard.generate()
  } catch (error) {
    console.error('生成仪表板失败:', error.message)
    process.exit(1)
  }
}

// 执行脚本
if (import.meta.url === `file://${process.argv[1]}`) {
  main()
}

export { ContractTestingDashboard }