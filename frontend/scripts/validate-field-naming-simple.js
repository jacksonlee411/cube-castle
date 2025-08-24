#!/usr/bin/env node

/**
 * 简化的字段命名合规性验证脚本
 * 
 * 仅检查API相关的关键字段
 * 基于契约测试自动化验证体系文档
 */

import { readFileSync, readdirSync, statSync } from 'fs'
import { join, extname } from 'path'

// 需要检查的snake_case违规模式
const SNAKE_CASE_PATTERNS = [
  /parent_unit_id/g,
  /unit_type/g, 
  /is_deleted/g,
  /operation_type/g,
  /created_at/g,
  /updated_at/g,
  /effective_date/g,
  /end_date/g,
  /record_id/g,
  /tenant_id/g,
  /is_current/g,
  /is_future/g,
  /sort_order/g,
  /event_type/g,
  /change_data/g,
  /change_reason/g,
  /client_id/g,
  /client_secret/g
]

class SimpleFieldValidator {
  constructor() {
    this.violations = []
    this.checkedFiles = 0
  }

  scanFile(filePath) {
    try {
      const content = readFileSync(filePath, 'utf-8')
      const lines = content.split('\n')

      lines.forEach((line, index) => {
        const lineNumber = index + 1
        
        // 检查snake_case模式
        SNAKE_CASE_PATTERNS.forEach(pattern => {
          const matches = line.match(pattern)
          if (matches) {
            matches.forEach(match => {
              // OAuth标准字段名例外处理 - RFC 6749标准要求
              if ((match === 'client_id' || match === 'client_secret') && 
                  filePath.includes('auth.ts') && 
                  (line.includes('JSON.stringify') || line.includes('this.config'))) {
                // OAuth协议标准字段名，不算违规
                return
              }
              
              this.violations.push({
                type: 'SNAKE_CASE_VIOLATION',
                field: match,
                file: filePath,
                line: lineNumber,
                message: `发现snake_case字段: ${match}，应该使用camelCase格式`
              })
            })
          }
        })
      })

      this.checkedFiles++
    } catch (error) {
      console.warn(`Warning: Could not read file ${filePath}: ${error.message}`)
    }
  }

  scanDirectory(dirPath, fileExtensions = ['.ts', '.tsx', '.js', '.jsx']) {
    try {
      const items = readdirSync(dirPath)

      items.forEach(item => {
        const fullPath = join(dirPath, item)
        const stat = statSync(fullPath)

        if (stat.isDirectory()) {
          // 跳过特定目录
          if (!['node_modules', 'dist', '.git', 'coverage'].includes(item)) {
            this.scanDirectory(fullPath, fileExtensions)
          }
        } else if (fileExtensions.includes(extname(item))) {
          this.scanFile(fullPath)
        }
      })
    } catch (error) {
      console.warn(`Warning: Could not scan directory ${dirPath}: ${error.message}`)
    }
  }

  generateReport() {
    console.log('\n=== API字段命名合规性检查报告 ===\n')
    
    console.log(`检查文件数: ${this.checkedFiles}`)
    console.log(`发现snake_case违规项: ${this.violations.length}\n`)

    if (this.violations.length === 0) {
      console.log('✅ 未发现snake_case字段命名违规！')
      return true
    }

    // 按文件分组显示违规项
    const violationsByFile = this.violations.reduce((acc, violation) => {
      if (!acc[violation.file]) {
        acc[violation.file] = []
      }
      acc[violation.file].push(violation)
      return acc
    }, {})

    console.log('❌ 发现以下snake_case字段违规：')
    Object.entries(violationsByFile).forEach(([file, violations]) => {
      console.log(`\n📁 ${file.replace(process.cwd(), '.')}:`)
      violations.forEach(violation => {
        console.log(`   第${violation.line}行: ${violation.field}`)
      })
    })

    console.log('\n💡 修复建议:')
    console.log('1. parent_unit_id → parentCode')
    console.log('2. unit_type → unitType')
    console.log('3. is_deleted → isDeleted')
    console.log('4. created_at → createdAt')
    console.log('5. updated_at → updatedAt')
    console.log('6. effective_date → effectiveDate')
    console.log('7. end_date → endDate')
    console.log('8. sort_order → sortOrder')
    console.log('\n🚨 注意: OAuth协议字段 client_id/client_secret 为RFC 6749标准要求，不算违规')

    return false
  }
}

// 主执行函数
function main() {
  const validator = new SimpleFieldValidator()
  
  console.log('开始API字段命名合规性检查...')
  
  // 仅扫描API相关目录
  const srcPath = join(process.cwd(), 'src')
  
  validator.scanDirectory(srcPath)

  // 生成报告
  const isCompliant = validator.generateReport()
  
  // 返回适当的退出码
  process.exit(isCompliant ? 0 : 1)
}

// 执行脚本
if (import.meta.url === `file://${process.argv[1]}`) {
  main()
}

export { SimpleFieldValidator }