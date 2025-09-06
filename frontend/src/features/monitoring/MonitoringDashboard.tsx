import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { Card } from '@workday/canvas-kit-react/card'
import { Heading } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { SystemIcon } from '@workday/canvas-kit-react/icon'
import { 
  dashboardIcon, 
  activityStreamIcon, 
  notificationsIcon, 
  homeIcon 
} from '@workday/canvas-system-icons-web'

interface MonitoringService {
  name: string
  url: string
  description: string
  icon: React.ComponentType<{ size?: number }>
  status: 'healthy' | 'warning' | 'error'
  credentials?: {
    username: string
    password: string
  }
}

const monitoringServices: MonitoringService[] = [
  {
    name: 'Prometheus',
    url: 'http://localhost:9091',
    description: '指标收集和存储 - 420个指标正在收集',
    icon: dashboardIcon,
    status: 'healthy'
  },
  {
    name: 'Grafana',
    url: 'http://localhost:3001',
    description: '数据可视化仪表板 - 实时监控面板',
    icon: activityStreamIcon,
    status: 'healthy',
    credentials: {
      username: 'admin',
      password: 'cube-castle-2025'
    }
  },
  {
    name: 'AlertManager',
    url: 'http://localhost:9093',
    description: '告警管理 - 8条SLO监控规则已加载',
    icon: notificationsIcon,
    status: 'healthy'
  },
  {
    name: 'Node Exporter',
    url: 'http://localhost:9100',
    description: '系统指标采集 - 服务器资源监控',
    icon: homeIcon,
    status: 'healthy'
  }
]

const getStatusColor = (status: string) => {
  switch (status) {
    case 'healthy':
      return '#00875A' // 绿色
    case 'warning':
      return '#FF991F' // 橙色
    case 'error':
      return '#DE350B' // 红色
    default:
      return '#6B778C' // 灰色
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'healthy':
      return '运行正常'
    case 'warning':
      return '需要关注'
    case 'error':
      return '服务异常'
    default:
      return '状态未知'
  }
}

export const MonitoringDashboard: React.FC = () => {
  const handleServiceClick = (service: MonitoringService) => {
    if (service.credentials) {
      // 对于需要认证的服务，显示登录信息
      const message = `服务: ${service.name}\\n地址: ${service.url}\\n用户名: ${service.credentials.username}\\n密码: ${service.credentials.password}\\n\\n点击"确定"将打开新窗口访问该服务。`
      if (window.confirm(message)) {
        window.open(service.url, '_blank')
      }
    } else {
      // 直接打开服务
      window.open(service.url, '_blank')
    }
  }

  return (
    <Box as="div">
      {/* 页面标题 */}
      <Box as="div" marginBottom="xl">
        <Heading size="large" marginBottom="s">
          🔍 系统监控中心
        </Heading>
        <Box as="div" color="licorice700">
          访问和管理Cube Castle监控系统的各个组件，监控系统健康状态和性能指标。
        </Box>
      </Box>

      {/* 监控服务卡片 */}
      <Box as="div" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))', gap: '24px' }}>
        {monitoringServices.map((service) => (
          <Card key={service.name} padding="l">
            <Box as="div" marginBottom="m">
              {/* 服务标题和状态 */}
              <Box as="div" style={{ display: 'flex', alignItems: 'center' }} marginBottom="s">
                <SystemIcon icon={service.icon} size="medium" />
                <Box as="div" marginLeft="s" style={{ flex: 1 }}>
                  <Heading size="small">{service.name}</Heading>
                </Box>
                <Box
                  as="div"
                  padding="xs"
                  borderRadius="s"
                  style={{ 
                    backgroundColor: getStatusColor(service.status),
                    color: 'white',
                    fontSize: '12px'
                  }}
                >
                  {getStatusText(service.status)}
                </Box>
              </Box>

              {/* 服务描述 */}
              <Box as="div" color="licorice700" marginBottom="m" style={{ fontSize: '14px' }}>
                {service.description}
              </Box>

              {/* 服务URL */}
              <Box 
                as="div"
                color="blueberry500" 
                marginBottom="m"
                style={{
                  fontFamily: 'monospace',
                  fontSize: '12px',
                  backgroundColor: '#F5F5F5',
                  padding: '8px',
                  borderRadius: '4px'
                }}
              >
                {service.url}
              </Box>

              {/* 认证信息（如果有） */}
              {service.credentials && (
                <Box 
                  as="div"
                  marginBottom="m"
                  style={{
                    backgroundColor: '#FFF3E0',
                    padding: '12px',
                    borderRadius: '4px',
                    fontSize: '12px'
                  }}
                >
                  <Box as="div" style={{ fontWeight: 'bold' }} marginBottom="xs">登录信息:</Box>
                  <Box as="div" color="licorice700">
                    用户名: {service.credentials.username}<br/>
                    密码: {service.credentials.password}
                  </Box>
                </Box>
              )}

              {/* 操作按钮 */}
              <Box as="div" style={{ display: 'flex', gap: '8px' }}>
                <PrimaryButton
                  onClick={() => handleServiceClick(service)}
                  size="small"
                >
                  打开服务
                </PrimaryButton>
                <SecondaryButton
                  onClick={() => {
                    navigator.clipboard.writeText(service.url)
                    console.log('服务地址已复制到剪贴板！')
                  }}
                  size="small"
                >
                  复制地址
                </SecondaryButton>
              </Box>
            </Box>
          </Card>
        ))}
      </Box>

      {/* 快速操作区域 */}
      <Card marginTop="xl" padding="l">
        <Heading size="small" marginBottom="m">
          🔧 快速操作
        </Heading>
        
        <Box as="div" style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
          <PrimaryButton
            variant="inverse"
            onClick={() => {
              const urls = [
                'http://localhost:9091/targets',
                'http://localhost:9091/rules', 
                'http://localhost:9091/alerts'
              ]
              urls.forEach(url => window.open(url, '_blank'))
            }}
          >
            查看Prometheus监控状态
          </PrimaryButton>
          
          <PrimaryButton
            variant="inverse" 
            onClick={() => {
              window.open('http://localhost:3001/dashboards', '_blank')
            }}
          >
            浏览Grafana仪表板
          </PrimaryButton>

          <SecondaryButton
            onClick={() => {
              const monitoringInfo = monitoringServices.map(service => 
                `${service.name}: ${service.url}${service.credentials ? ` (${service.credentials.username}/${service.credentials.password})` : ''}`
              ).join('\\n')
              
              navigator.clipboard.writeText(`Cube Castle 监控系统信息:\\n\\n${monitoringInfo}`)
              console.log('所有监控服务信息已复制到剪贴板！')
            }}
          >
            复制所有服务信息
          </SecondaryButton>
        </Box>

        <Box as="div" marginTop="m" style={{ fontSize: '12px' }} color="licorice600">
          💡 提示: 点击"打开服务"将在新窗口中打开监控服务。对于Grafana，会提示登录信息。
        </Box>
      </Card>
    </Box>
  )
}