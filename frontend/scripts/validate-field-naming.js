#!/usr/bin/env node

/**
 * 字段命名合规性验证脚本
 * 
 * 自动化检查项目中的字段命名是否符合camelCase规范
 * 基于契约测试自动化验证体系文档
 */

import { readFileSync, readdirSync, statSync } from 'fs'
import { join, extname } from 'path'

// 标准字段命名词汇表
const STANDARD_FIELDS = {
  // 核心业务字段 (camelCase)
  identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
  timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'], 
  statusFields: ['status', 'isCurrent', 'isFuture'],
  operationFields: ['operationType', 'operatedBy', 'operationReason'],
  hierarchyFields: ['level', 'codePath', 'namePath', 'hierarchyDepth'],
  configFields: ['unitType', 'sortOrder', 'description', 'profile']
}

// 禁止使用的旧字段名 (已废弃)
const PROHIBITED_FIELDS = [
  'parent_unit_id', 'unit_type', 'is_deleted', 'operation_type',
  'created_at', 'updated_at', 'effective_date', 'end_date',
  'record_id', 'tenant_id', 'parent_code', 'is_current'
]

// camelCase 正则表达式
const CAMEL_CASE_REGEX = /^[a-z][a-zA-Z0-9]*$/
const SNAKE_CASE_REGEX = /_/

class FieldNamingValidator {
  constructor() {
    this.violations = []
    this.checkedFiles = 0
  }

  /**
   * 验证字段名是否符合camelCase规范
   */
  validateFieldName(fieldName, filePath, lineNumber) {
    // 检查是否为禁用字段
    if (PROHIBITED_FIELDS.includes(fieldName)) {
      this.violations.push({
        type: 'PROHIBITED_FIELD',
        field: fieldName,
        file: filePath,
        line: lineNumber,
        message: `禁用的snake_case字段: ${fieldName}`
      })
      return false
    }

    // 检查是否包含下划线
    if (SNAKE_CASE_REGEX.test(fieldName)) {
      this.violations.push({
        type: 'SNAKE_CASE_VIOLATION',
        field: fieldName,
        file: filePath,
        line: lineNumber,
        message: `字段名包含下划线: ${fieldName}`
      })
      return false
    }

    // 检查是否符合camelCase格式
    if (!CAMEL_CASE_REGEX.test(fieldName)) {
      this.violations.push({
        type: 'CAMEL_CASE_VIOLATION',
        field: fieldName,
        file: filePath,
        line: lineNumber,
        message: `字段名不符合camelCase格式: ${fieldName}`
      })
      return false
    }

    return true
  }

  /**
   * 扫描TypeScript文件中的字段命名
   */
  scanTypeScriptFile(filePath) {
    try {
      const content = readFileSync(filePath, 'utf-8')
      const lines = content.split('\n')

      lines.forEach((line, index) => {
        const lineNumber = index + 1

        // 匹配接口字段定义
        const interfaceFieldMatch = line.match(/^\s*(\w+)\s*[:?]/);
        if (interfaceFieldMatch) {
          const fieldName = interfaceFieldMatch[1]
          this.validateFieldName(fieldName, filePath, lineNumber)
        }

        // 匹配对象属性
        const objectPropertyMatch = line.match(/['"`]?(\w+)['"`]?\s*:/g);
        if (objectPropertyMatch) {
          objectPropertyMatch.forEach(match => {
            const fieldName = match.replace(/['"`:\s]/g, '')
            if (fieldName && fieldName !== 'type' && fieldName !== 'status') {
              this.validateFieldName(fieldName, filePath, lineNumber)
            }
          })
        }

        // 匹配GraphQL字段定义
        const graphqlFieldMatch = line.match(/^\s*(\w+)\s*[:!]/);
        if (graphqlFieldMatch) {
          const fieldName = graphqlFieldMatch[1]
          this.validateFieldName(fieldName, filePath, lineNumber)
        }
      })

      this.checkedFiles++
    } catch (error) {
      console.warn(`Warning: Could not read file ${filePath}: ${error.message}`)
    }
  }

  /**
   * 递归扫描目录
   */
  scanDirectory(dirPath, fileExtensions = ['.ts', '.tsx', '.js', '.jsx']) {
    try {
      const items = readdirSync(dirPath)

      items.forEach(item => {
        const fullPath = join(dirPath, item)
        const stat = statSync(fullPath)

        if (stat.isDirectory()) {
          // 跳过node_modules和dist目录
          if (!['node_modules', 'dist', '.git'].includes(item)) {
            this.scanDirectory(fullPath, fileExtensions)
          }
        } else if (fileExtensions.includes(extname(item))) {
          this.scanTypeScriptFile(fullPath)
        }
      })
    } catch (error) {
      console.warn(`Warning: Could not scan directory ${dirPath}: ${error.message}`)
    }
  }

  /**
   * 生成验证报告
   */
  generateReport() {
    console.log('\n=== 字段命名合规性验证报告 ===\n')
    
    console.log(`检查文件数: ${this.checkedFiles}`)
    console.log(`发现违规项: ${this.violations.length}\n`)

    if (this.violations.length === 0) {
      console.log('✅ 所有字段命名都符合camelCase规范！')
      return true
    }

    // 按类型分组违规项
    const violationsByType = this.violations.reduce((acc, violation) => {
      if (!acc[violation.type]) {
        acc[violation.type] = []
      }
      acc[violation.type].push(violation)
      return acc
    }, {})

    // 输出违规详情
    Object.entries(violationsByType).forEach(([type, violations]) => {
      console.log(`\n❌ ${type} (${violations.length}项):`)
      violations.forEach(violation => {
        console.log(`   ${violation.file}:${violation.line} - ${violation.message}`)
      })
    })

    console.log('\n💡 修复建议:')
    console.log('1. 将所有snake_case字段改为camelCase格式')
    console.log('2. 移除禁用的旧字段名')
    console.log('3. 确保新字段遵循标准命名词汇表')
    console.log('4. 运行 npm run validate:field-naming 定期检查合规性')

    return false
  }
}

// 主执行函数
function main() {
  const validator = new FieldNamingValidator()
  
  console.log('开始字段命名合规性验证...')
  
  // 扫描源代码目录
  const srcPath = join(process.cwd(), 'src')
  const testsPath = join(process.cwd(), 'tests')
  
  validator.scanDirectory(srcPath)
  validator.scanDirectory(testsPath)

  // 生成报告
  const isCompliant = validator.generateReport()
  
  // 返回适当的退出码
  process.exit(isCompliant ? 0 : 1)
}

// 执行脚本
if (import.meta.url === `file://${process.argv[1]}`) {
  main()
}

export { FieldNamingValidator }
