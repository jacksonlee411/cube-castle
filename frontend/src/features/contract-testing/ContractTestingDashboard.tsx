import React, { useState } from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { Card } from '@workday/canvas-kit-react/card'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { colors } from '@workday/canvas-kit-react/tokens'
import { Flex } from '@workday/canvas-kit-react/layout'
import { contractTestingAPI } from '../../shared/api/contract-testing'
import { useMessages } from '../../shared/hooks/useMessages'
import { MessageDisplay } from '../../shared/components/MessageDisplay'

interface ContractMetrics {
  contractTestPass: number
  contractTestTotal: number
  fieldNamingCompliance: number
  fieldNamingViolations: number
  schemaValidationStatus: 'success' | 'warning' | 'error'
  schemaValidationMessage: string
  timestamp: string
}

const MetricCard: React.FC<{
  title: string
  value: string | number
  status: 'good' | 'warning' | 'error'
  subtitle?: string
  violationDetails?: string[]
}> = ({ title, value, status, subtitle, violationDetails }) => {
  const getStatusColor = () => {
    switch (status) {
      case 'good': return colors.greenApple500
      case 'warning': return colors.cantaloupe500
      case 'error': return colors.cinnamon500
      default: return colors.licorice500
    }
  }

  return (
    <Card padding="l">
      <Text typeLevel="heading.small" marginBottom="s">{title}</Text>
      <Text 
        typeLevel="heading.large" 
        color={getStatusColor()}
        marginBottom="s"
      >
        {value}
      </Text>
      {subtitle && (
        <Text color="licorice500" marginBottom="s">
          {subtitle}
        </Text>
      )}
      {violationDetails && violationDetails.length > 0 && (
        <Box 
          backgroundColor="soap200" 
          padding="s" 
          borderRadius="s"
          marginTop="s"
        >
          <Text fontWeight="bold" color="cinnamon600" marginBottom="xs">
            ⚠️ 需要修复:
          </Text>
          {violationDetails.map((detail, index) => (
            <Text key={index} fontSize="small" color="licorice600">
              • {detail}
            </Text>
          ))}
        </Box>
      )}
    </Card>
  )
}

const QuickAction: React.FC<{
  title: string
  command: string
  description: string
}> = ({ title, command, description }) => (
  <Box marginBottom="s">
    <Text fontWeight="bold" marginBottom="xs">{title}</Text>
    <Box 
      backgroundColor="soap200" 
      padding="s" 
      borderRadius="s" 
      marginBottom="xs"
    >
      <Text fontFamily="monospace" fontSize="small">
        {command}
      </Text>
    </Box>
    <Text fontSize="small" color="licorice500">{description}</Text>
  </Box>
)

export const ContractTestingDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<ContractMetrics>({
    contractTestPass: 0,
    contractTestTotal: 0,
    fieldNamingCompliance: 85,
    fieldNamingViolations: 1,
    schemaValidationStatus: 'error',
    schemaValidationMessage: 'spawnSync /bin/sh ENOENT',
    timestamp: new Date().toLocaleString('zh-CN')
  })
  
  const { successMessage, error, showSuccess, showError } = useMessages()

  const [isRefreshing, setIsRefreshing] = useState(false)

  const refreshMetrics = async () => {
    setIsRefreshing(true)
    try {
      // 这里模拟实际的指标获取逻辑
      // 在实际实现中，这应该调用契约测试API
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      setMetrics(prev => ({
        ...prev,
        timestamp: new Date().toLocaleString('zh-CN')
      }))
    } catch (error) {
      console.error('Failed to refresh metrics:', error)
    } finally {
      setIsRefreshing(false)
    }
  }

  const runContractTest = async () => {
    setIsRefreshing(true)
    try {
      const result = await contractTestingAPI.runTests()
      setMetrics(prev => ({
        ...prev,
        contractTestPass: result.passedTests,
        contractTestTotal: result.totalTests,
        timestamp: new Date().toLocaleString('zh-CN')
      }))
      showSuccess(`契约测试完成！通过 ${result.passedTests}/${result.totalTests} 个测试`)
    } catch (error) {
      console.error('Contract test failed:', error)
      showError('契约测试执行失败：' + (error as Error).message)
    } finally {
      setIsRefreshing(false)
    }
  }

  const validateFieldNaming = async () => {
    setIsRefreshing(true)
    try {
      const result = await contractTestingAPI.validateFieldNaming()
      setMetrics(prev => ({
        ...prev,
        fieldNamingViolations: result.violations,
        fieldNamingCompliance: result.complianceRate,
        timestamp: new Date().toLocaleString('zh-CN')
      }))
      showSuccess(`字段命名验证完成！合规率 ${result.complianceRate}%，违规项 ${result.violations} 个`)
    } catch (error) {
      console.error('Field naming validation failed:', error)
      showError('字段命名验证失败：' + (error as Error).message)
    } finally {
      setIsRefreshing(false)
    }
  }

  const validateSchema = async () => {
    setIsRefreshing(true)
    try {
      const result = await contractTestingAPI.validateSchema()
      setMetrics(prev => ({
        ...prev,
        schemaValidationStatus: result.status,
        schemaValidationMessage: result.message,
        timestamp: new Date().toLocaleString('zh-CN')
      }))
      showSuccess(`Schema验证完成！状态：${result.message}`)
    } catch (error) {
      console.error('Schema validation failed:', error)
      showError('Schema验证失败：' + (error as Error).message)
    } finally {
      setIsRefreshing(false)
    }
  }

  const contractPassRate = metrics.contractTestTotal > 0 
    ? Math.round((metrics.contractTestPass / metrics.contractTestTotal) * 100)
    : 0

  return (
    <Box>
      {/* 消息显示区域 */}
      <MessageDisplay 
        successMessage={successMessage}
        errorMessage={error}
      />
      
      {/* 页面标题 */}
      <Flex alignItems="center" marginBottom="l">
        <Text typeLevel="heading.large" marginRight="m">
          🔍 契约测试监控仪表板
        </Text>
        <SecondaryButton 
          onClick={refreshMetrics}
          disabled={isRefreshing}
          size="small"
        >
          {isRefreshing ? '刷新中...' : '刷新数据'}
        </SecondaryButton>
      </Flex>

      <Text color="licorice500" marginBottom="l">
        最后更新: {metrics.timestamp}
      </Text>

      {/* 指标卡片网格 */}
      <Box 
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: '24px'
        }}
        marginBottom="xl"
      >
        <MetricCard
          title="📊 契约测试通过率"
          value={`${contractPassRate}%`}
          status={contractPassRate > 90 ? 'good' : contractPassRate > 70 ? 'warning' : 'error'}
          subtitle={`通过: ${metrics.contractTestPass} / 总数: ${metrics.contractTestTotal}`}
        />

        <MetricCard
          title="📝 字段命名合规率"
          value={`${metrics.fieldNamingCompliance}%`}
          status={metrics.fieldNamingCompliance > 95 ? 'good' : metrics.fieldNamingCompliance > 80 ? 'warning' : 'error'}
          subtitle={`违规项: ${metrics.fieldNamingViolations}`}
          violationDetails={metrics.fieldNamingViolations > 0 ? [
            '将 snake_case 字段改为 camelCase',
            '运行 npm run validate:field-naming 查看详情'
          ] : undefined}
        />

        <MetricCard
          title="🔧 GraphQL Schema状态"
          value={metrics.schemaValidationStatus === 'success' ? '✅ 正常' : 
                 metrics.schemaValidationStatus === 'warning' ? '⚠️ 警告' : '❌ 错误'}
          status={metrics.schemaValidationStatus === 'success' ? 'good' : 
                 metrics.schemaValidationStatus === 'warning' ? 'warning' : 'error'}
          subtitle="Schema v4.2.1 验证"
          violationDetails={metrics.schemaValidationStatus !== 'success' ? [
            `错误详情: ${metrics.schemaValidationMessage}`
          ] : undefined}
        />
      </Box>

      {/* 操作面板 */}
      <Box 
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))',
          gap: '24px'
        }}
      >
        <Card padding="l">
          <Text typeLevel="heading.medium" marginBottom="m">🚀 快速操作</Text>
          <Flex gap="s" marginBottom="l">
            <PrimaryButton onClick={runContractTest}>
              运行契约测试
            </PrimaryButton>
            <SecondaryButton onClick={validateFieldNaming}>
              检查字段命名
            </SecondaryButton>
            <SecondaryButton onClick={validateSchema}>
              验证Schema
            </SecondaryButton>
          </Flex>

          <QuickAction
            title="运行测试:"
            command="cd frontend && npm run test:contract"
            description="执行完整的契约测试套件"
          />
          
          <QuickAction
            title="检查字段命名:"
            command="cd frontend && npm run validate:field-naming"
            description="验证API响应字段使用camelCase命名"
          />
          
          <QuickAction
            title="验证Schema:"
            command="cd frontend && npm run validate:schema"
            description="验证GraphQL Schema语法和一致性"
          />
        </Card>

        <Card padding="l">
          <Text typeLevel="heading.medium" marginBottom="m">📈 趋势分析</Text>
          <Box marginBottom="m">
            <Text fontWeight="bold" marginBottom="s">本次检查发现:</Text>
            <Text as="ul" paddingLeft="m">
              <Text as="li">契约测试: 需要检查</Text>
              <Text as="li">字段命名: {metrics.fieldNamingViolations}个违规</Text>
              <Text as="li">Schema验证: {metrics.schemaValidationStatus === 'success' ? '通过' : '失败'}</Text>
            </Text>
          </Box>
          
          <Box>
            <Text fontWeight="bold" marginBottom="s">建议操作:</Text>
            <Text color="cinnamon600">
              🔧 优先修复字段命名问题，这会阻止代码合并
            </Text>
          </Box>
        </Card>
      </Box>
    </Box>
  )
}