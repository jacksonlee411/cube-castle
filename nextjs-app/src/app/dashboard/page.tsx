'use client'

import { useState, useEffect } from 'react'
import { AppNav } from '@/components/business/app-nav'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Users, 
  Building2, 
  Workflow, 
  Brain, 
  TrendingUp,
  TrendingDown,
  Activity,
  Clock,
  CheckCircle,
  AlertCircle,
  ArrowUpRight
} from 'lucide-react'
import Link from 'next/link'

interface DashboardStats {
  employees: {
    total: number
    active: number
    newThisMonth: number
    trend: 'up' | 'down' | 'stable'
  }
  organizations: {
    total: number
    active: number
    newThisMonth: number
  }
  workflows: {
    completed: number
    pending: number
    successRate: number
  }
  aiQueries: {
    processed: number
    todayCount: number
    avgResponseTime: number
  }
}

interface RecentActivity {
  id: string
  type: 'employee' | 'organization' | 'workflow' | 'ai'
  title: string
  description: string
  timestamp: string
  status: 'success' | 'warning' | 'error'
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [recentActivities, setRecentActivities] = useState<RecentActivity[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // 模拟数据获取
    const fetchDashboardData = async () => {
      try {
        setLoading(true)
        
        // 模拟 API 延迟
        await new Promise(resolve => setTimeout(resolve, 1000))
        
        // 模拟统计数据
        const mockStats: DashboardStats = {
          employees: {
            total: 1247,
            active: 1198,
            newThisMonth: 23,
            trend: 'up'
          },
          organizations: {
            total: 28,
            active: 26,
            newThisMonth: 2
          },
          workflows: {
            completed: 342,
            pending: 15,
            successRate: 96.5
          },
          aiQueries: {
            processed: 1567,
            todayCount: 89,
            avgResponseTime: 1.2
          }
        }

        // 模拟最近活动
        const mockActivities: RecentActivity[] = [
          {
            id: '1',
            type: 'employee',
            title: '新员工入职',
            description: '张三已完成入职流程',
            timestamp: '2 分钟前',
            status: 'success'
          },
          {
            id: '2',
            type: 'workflow',
            title: '审批流程完成',
            description: '李四的请假申请已审批通过',
            timestamp: '15 分钟前',
            status: 'success'
          },
          {
            id: '3',
            type: 'organization',
            title: '部门结构调整',
            description: '技术部新增AI研发小组',
            timestamp: '1 小时前',
            status: 'warning'
          },
          {
            id: '4',
            type: 'ai',
            title: 'AI查询高峰',
            description: '今日AI查询数量较昨日增长25%',
            timestamp: '2 小时前',
            status: 'success'
          },
          {
            id: '5',
            type: 'employee',
            title: '员工信息更新',
            description: '王五更新了联系方式和紧急联系人',
            timestamp: '3 小时前',
            status: 'success'
          }
        ]

        setStats(mockStats)
        setRecentActivities(mockActivities)
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchDashboardData()
  }, [])

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'employee':
        return <Users className="h-4 w-4" />
      case 'organization':
        return <Building2 className="h-4 w-4" />
      case 'workflow':
        return <Workflow className="h-4 w-4" />
      case 'ai':
        return <Brain className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'text-green-600'
      case 'warning':
        return 'text-yellow-600'
      case 'error':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <AppNav />
      
      <main className="container-responsive py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold tracking-tight">控制台概览</h1>
          <p className="text-muted-foreground">
            查看系统运行状况和关键业务指标
          </p>
        </div>

        {/* 统计卡片 */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总员工数</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : stats?.employees.total.toLocaleString()}
              </div>
              <div className="flex items-center text-xs text-muted-foreground">
                {!loading && stats && (
                  <>
                    {stats.employees.trend === 'up' ? (
                      <TrendingUp className="mr-1 h-3 w-3 text-green-500" />
                    ) : (
                      <TrendingDown className="mr-1 h-3 w-3 text-red-500" />
                    )}
                    本月新增 {stats.employees.newThisMonth} 人
                  </>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">组织部门</CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : stats?.organizations.total}
              </div>
              <p className="text-xs text-muted-foreground">
                {loading ? '...' : `${stats?.organizations.active} 个活跃部门`}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">工作流</CardTitle>
              <Workflow className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : stats?.workflows.completed}
              </div>
              <p className="text-xs text-muted-foreground">
                {loading ? '...' : `成功率 ${stats?.workflows.successRate}%`}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">AI 查询</CardTitle>
              <Brain className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {loading ? '...' : stats?.aiQueries.todayCount}
              </div>
              <p className="text-xs text-muted-foreground">
                {loading ? '...' : `平均响应 ${stats?.aiQueries.avgResponseTime}s`}
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          {/* 快速操作 */}
          <Card>
            <CardHeader>
              <CardTitle>快速操作</CardTitle>
              <CardDescription>
                常用功能快速入口
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Link href="/employees">
                <Button className="w-full justify-start" variant="outline">
                  <Users className="mr-2 h-4 w-4" />
                  员工管理
                  <ArrowUpRight className="ml-auto h-4 w-4" />
                </Button>
              </Link>
              <Link href="/organizations">
                <Button className="w-full justify-start" variant="outline">
                  <Building2 className="mr-2 h-4 w-4" />
                  组织架构
                  <ArrowUpRight className="ml-auto h-4 w-4" />
                </Button>
              </Link>
              <Link href="/chat">
                <Button className="w-full justify-start" variant="outline">
                  <Brain className="mr-2 h-4 w-4" />
                  AI 助手
                  <ArrowUpRight className="ml-auto h-4 w-4" />
                </Button>
              </Link>
            </CardContent>
          </Card>

          {/* 系统状态 */}
          <Card>
            <CardHeader>
              <CardTitle>系统状态</CardTitle>
              <CardDescription>
                核心服务运行状况
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm">Go API 服务</span>
                </div>
                <Badge variant="secondary">运行中</Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm">Python AI 服务</span>
                </div>
                <Badge variant="secondary">运行中</Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm">数据库</span>
                </div>
                <Badge variant="secondary">运行中</Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm">工作流引擎</span>
                </div>
                <Badge variant="secondary">运行中</Badge>
              </div>
            </CardContent>
          </Card>

          {/* 最近活动 */}
          <Card>
            <CardHeader>
              <CardTitle>最近活动</CardTitle>
              <CardDescription>
                系统最新动态和变更
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {loading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="flex items-start space-x-3">
                      <div className="h-4 w-4 animate-pulse rounded bg-muted mt-0.5" />
                      <div className="flex-1 space-y-1">
                        <div className="h-4 animate-pulse rounded bg-muted" />
                        <div className="h-3 w-2/3 animate-pulse rounded bg-muted" />
                      </div>
                    </div>
                  ))
                ) : (
                  recentActivities.slice(0, 5).map((activity) => (
                    <div key={activity.id} className="flex items-start space-x-3">
                      <div className={`mt-0.5 ${getStatusColor(activity.status)}`}>
                        {getActivityIcon(activity.type)}
                      </div>
                      <div className="flex-1 space-y-1">
                        <p className="text-sm font-medium">{activity.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {activity.description}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {activity.timestamp}
                        </p>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}