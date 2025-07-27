'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Users, 
  Building2, 
  Workflow, 
  Brain, 
  Shield, 
  BarChart3,
  CheckCircle,
  Clock,
  RefreshCw,
  AlertCircle,
  TrendingUp,
  Activity
} from 'lucide-react'
import Link from 'next/link'

interface SystemStatus {
  status: 'healthy' | 'warning' | 'error'
  services: Array<{
    name: string
    status: 'healthy' | 'unhealthy'
    latency: number
    message?: string
  }>
  metrics: {
    memoryUsage: number
    cpuUsage: number
    activeConnections: number
    requestsPerSecond: number
    errorRate: number
  }
}

interface BusinessMetrics {
  totalEmployees: number
  activeEmployees: number
  totalOrganizations: number
  workflowsCompleted: number
  aiQueriesProcessed: number
}

export default function DemoPage() {
  const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null)
  const [businessMetrics, setBusinessMetrics] = useState<BusinessMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 模拟数据获取
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true)
        
        // 模拟系统状态数据
        const mockSystemStatus: SystemStatus = {
          status: 'healthy',
          services: [
            { name: 'Go API Server', status: 'healthy', latency: 45 },
            { name: 'Python AI Service', status: 'healthy', latency: 123 },
            { name: 'PostgreSQL', status: 'healthy', latency: 8 },
            { name: 'Redis', status: 'healthy', latency: 2 },
            { name: 'Temporal', status: 'healthy', latency: 67 },
            { name: 'OPA Authorization', status: 'healthy', latency: 12 }
          ],
          metrics: {
            memoryUsage: 65.4,
            cpuUsage: 23.7,
            activeConnections: 156,
            requestsPerSecond: 87.3,
            errorRate: 0.02
          }
        }

        // 模拟业务指标数据
        const mockBusinessMetrics: BusinessMetrics = {
          totalEmployees: 1247,
          activeEmployees: 1198,
          totalOrganizations: 28,
          workflowsCompleted: 342,
          aiQueriesProcessed: 1567
        }

        // 模拟网络延迟
        await new Promise(resolve => setTimeout(resolve, 1000))

        setSystemStatus(mockSystemStatus)
        setBusinessMetrics(mockBusinessMetrics)
        setError(null)
      } catch (err) {
        setError('无法获取系统数据')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  const refreshData = () => {
    setSystemStatus(null)
    setBusinessMetrics(null)
    setLoading(true)
    // 重新获取数据
    setTimeout(() => {
      setSystemStatus({
        status: 'healthy',
        services: [
          { name: 'Go API Server', status: 'healthy', latency: Math.floor(Math.random() * 100) + 20 },
          { name: 'Python AI Service', status: 'healthy', latency: Math.floor(Math.random() * 200) + 50 },
          { name: 'PostgreSQL', status: 'healthy', latency: Math.floor(Math.random() * 20) + 5 },
          { name: 'Redis', status: 'healthy', latency: Math.floor(Math.random() * 10) + 1 },
          { name: 'Temporal', status: 'healthy', latency: Math.floor(Math.random() * 100) + 30 },
          { name: 'OPA Authorization', status: 'healthy', latency: Math.floor(Math.random() * 30) + 5 }
        ],
        metrics: {
          memoryUsage: Math.random() * 30 + 50,
          cpuUsage: Math.random() * 40 + 15,
          activeConnections: Math.floor(Math.random() * 200) + 100,
          requestsPerSecond: Math.random() * 50 + 50,
          errorRate: Math.random() * 0.1
        }
      })
      setBusinessMetrics({
        totalEmployees: 1247 + Math.floor(Math.random() * 20),
        activeEmployees: 1198 + Math.floor(Math.random() * 10),
        totalOrganizations: 28 + Math.floor(Math.random() * 5),
        workflowsCompleted: 342 + Math.floor(Math.random() * 50),
        aiQueriesProcessed: 1567 + Math.floor(Math.random() * 100)
      })
      setLoading(false)
    }, 1000)
  }

  return (
    <div className="min-h-screen bg-background">
      {/* 导航栏 */}
      <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container-responsive flex h-16 items-center justify-between">
          <Link href="/" className="flex items-center space-x-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Castle className="h-5 w-5" />
            </div>
            <span className="text-xl font-bold">Cube Castle</span>
          </Link>
          <div className="flex items-center space-x-4">
            <Button variant="ghost" onClick={refreshData} disabled={loading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              刷新数据
            </Button>
            <Button variant="outline" asChild>
              <Link href="/">返回首页</Link>
            </Button>
          </div>
        </div>
      </nav>

      <main className="container-responsive py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold">系统演示</h1>
          <p className="mt-2 text-muted-foreground">
            实时查看 Cube Castle 系统的运行状况和业务指标
          </p>
        </div>

        {/* 错误状态 */}
        {error && (
          <Card className="mb-8 border-destructive">
            <CardContent className="flex items-center p-6">
              <AlertCircle className="mr-3 h-5 w-5 text-destructive" />
              <span className="text-destructive">{error}</span>
              <Button variant="outline" size="sm" className="ml-auto" onClick={refreshData}>
                重试
              </Button>
            </CardContent>
          </Card>
        )}

        {/* 系统状态概览 */}
        <div className="mb-8">
          <h2 className="mb-4 text-2xl font-semibold">系统状态概览</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardContent className="flex items-center p-6">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">整体状态</p>
                  <div className="flex items-center mt-2">
                    {loading ? (
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent mr-2" />
                    ) : (
                      <CheckCircle className="mr-2 h-4 w-4 text-green-500" />
                    )}
                    <span className="text-lg font-bold">
                      {loading ? '检查中...' : systemStatus?.status === 'healthy' ? '运行正常' : '异常'}
                    </span>
                  </div>
                </div>
                <Activity className="h-8 w-8 text-muted-foreground" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center p-6">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">平均响应时间</p>
                  <div className="flex items-center mt-2">
                    <Clock className="mr-2 h-4 w-4 text-blue-500" />
                    <span className="text-lg font-bold">
                      {loading ? '计算中...' : `${Math.round((systemStatus?.services.reduce((acc, s) => acc + s.latency, 0) || 0) / (systemStatus?.services.length || 1))}ms`}
                    </span>
                  </div>
                </div>
                <BarChart3 className="h-8 w-8 text-muted-foreground" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center p-6">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">活跃连接</p>
                  <div className="flex items-center mt-2">
                    <TrendingUp className="mr-2 h-4 w-4 text-green-500" />
                    <span className="text-lg font-bold">
                      {loading ? '加载中...' : systemStatus?.metrics.activeConnections || 0}
                    </span>
                  </div>
                </div>
                <Users className="h-8 w-8 text-muted-foreground" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="flex items-center p-6">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">错误率</p>
                  <div className="flex items-center mt-2">
                    <CheckCircle className="mr-2 h-4 w-4 text-green-500" />
                    <span className="text-lg font-bold">
                      {loading ? '计算中...' : `${((systemStatus?.metrics.errorRate || 0) * 100).toFixed(2)}%`}
                    </span>
                  </div>
                </div>
                <Shield className="h-8 w-8 text-muted-foreground" />
              </CardContent>
            </Card>
          </div>
        </div>

        {/* 服务状态 */}
        <div className="mb-8">
          <h2 className="mb-4 text-2xl font-semibold">服务状态</h2>
          <Card>
            <CardHeader>
              <CardTitle>微服务健康检查</CardTitle>
              <CardDescription>
                各个微服务组件的实时健康状况和响应延迟
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {loading ? (
                  Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center space-x-3">
                        <div className="h-4 w-4 animate-pulse rounded-full bg-muted" />
                        <div className="h-4 w-32 animate-pulse rounded bg-muted" />
                      </div>
                      <div className="h-4 w-16 animate-pulse rounded bg-muted" />
                    </div>
                  ))
                ) : (
                  systemStatus?.services.map((service, index) => (
                    <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center space-x-3">
                        <div className={`h-3 w-3 rounded-full ${
                          service.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'
                        }`} />
                        <span className="font-medium">{service.name}</span>
                        <Badge variant={service.status === 'healthy' ? 'default' : 'destructive'}>
                          {service.status === 'healthy' ? '正常' : '异常'}
                        </Badge>
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {service.latency}ms
                      </span>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* 业务指标 */}
        <div className="mb-8">
          <h2 className="mb-4 text-2xl font-semibold">业务指标</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-5">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">总员工数</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {loading ? '...' : businessMetrics?.totalEmployees || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  活跃员工: {loading ? '...' : businessMetrics?.activeEmployees || 0}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">组织架构</CardTitle>
                <Building2 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {loading ? '...' : businessMetrics?.totalOrganizations || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  个部门/团队
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">工作流完成</CardTitle>
                <Workflow className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {loading ? '...' : businessMetrics?.workflowsCompleted || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  本月完成
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">AI查询处理</CardTitle>
                <Brain className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {loading ? '...' : businessMetrics?.aiQueriesProcessed || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  本周处理
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">系统资源</CardTitle>
                <BarChart3 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {loading ? '...' : `${Math.round(systemStatus?.metrics.memoryUsage || 0)}%`}
                </div>
                <p className="text-xs text-muted-foreground">
                  内存使用率
                </p>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* 功能演示 */}
        <div>
          <h2 className="mb-4 text-2xl font-semibold">功能演示</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Users className="mr-2 h-5 w-5" />
                  员工管理
                </CardTitle>
                <CardDescription>
                  完整的员工信息管理系统
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-4">
                  演示员工 CRUD 操作、组织架构管理和职位分配
                </p>
                <Button className="w-full" disabled>
                  即将开放
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Brain className="mr-2 h-5 w-5" />
                  AI 智能助手
                </CardTitle>
                <CardDescription>
                  自然语言处理和智能对话
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-4">
                  体验 AI 驱动的智能查询和自动化建议
                </p>
                <Button className="w-full" disabled>
                  即将开放
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Workflow className="mr-2 h-5 w-5" />
                  工作流管理
                </CardTitle>
                <CardDescription>
                  分布式工作流编排和监控
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-4">
                  查看工作流状态、审批流程和自动化任务
                </p>
                <Button className="w-full" disabled>
                  即将开放
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}

// Castle 图标组件
function Castle({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M3 21V9l2-2h2V5l2-2h6l2 2v2h2l2 2v12H3zm4-4h2v-2H7v2zm6 0h2v-2h-2v2z"/>
    </svg>
  )
}