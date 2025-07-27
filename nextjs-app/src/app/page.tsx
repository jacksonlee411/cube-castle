import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  Users, 
  Building2, 
  Workflow, 
  Brain, 
  Shield, 
  BarChart3,
  ArrowRight,
  CheckCircle,
  Clock,
  Zap
} from 'lucide-react'

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted/20">
      {/* 导航栏 */}
      <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container-responsive flex h-16 items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Castle className="h-5 w-5" />
            </div>
            <span className="text-xl font-bold">Cube Castle</span>
          </div>
          <div className="flex items-center space-x-4">
            <Link href="/login">
              <Button variant="ghost">登录</Button>
            </Link>
            <Link href="/demo">
              <Button>体验演示</Button>
            </Link>
          </div>
        </div>
      </nav>

      {/* 首页内容 */}
      <main>
        {/* Hero 区域 */}
        <section className="container-responsive py-24 text-center">
          <div className="mx-auto max-w-4xl">
            <Badge variant="secondary" className="mb-4">
              🎉 v1.4.0 现已发布
            </Badge>
            <h1 className="mb-6 text-4xl font-bold tracking-tight sm:text-6xl">
              企业级 <span className="text-gradient">HR 管理平台</span>
            </h1>
            <p className="mb-8 text-xl text-muted-foreground">
              基于城堡模型架构的现代化 HR SaaS 平台，集成 AI 智能交互、
              分布式工作流编排、企业级安全架构和全面的系统监控
            </p>
            <div className="flex flex-col gap-4 sm:flex-row sm:justify-center">
              <Link href="/demo">
                <Button size="lg" className="w-full sm:w-auto">
                  立即体验
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </Link>
              <Link href="/features">
                <Button variant="outline" size="lg" className="w-full sm:w-auto">
                  了解功能
                </Button>
              </Link>
            </div>
          </div>
        </section>

        {/* 系统状态展示 */}
        <section className="container-responsive py-16">
          <div className="mb-12 text-center">
            <h2 className="mb-4 text-3xl font-bold">系统实时状态</h2>
            <p className="text-lg text-muted-foreground">
              查看我们的系统运行状况和关键指标
            </p>
          </div>
          
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">系统状态</CardTitle>
                <CheckCircle className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">运行正常</div>
                <div className="flex items-center text-xs text-muted-foreground">
                  <Clock className="mr-1 h-3 w-3" />
                  99.9% 可用性
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">响应时间</CardTitle>
                <Zap className="h-4 w-4 text-yellow-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">45ms</div>
                <p className="text-xs text-muted-foreground">
                  P95 延迟 &lt; 100ms
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
                <Users className="h-4 w-4 text-blue-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">1,234</div>
                <p className="text-xs text-muted-foreground">
                  +12% 比上月
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">安全等级</CardTitle>
                <Shield className="h-4 w-4 text-purple-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-purple-600">企业级</div>
                <p className="text-xs text-muted-foreground">
                  SOC2 合规认证
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* 核心功能展示 */}
        <section className="container-responsive py-16">
          <div className="mb-12 text-center">
            <h2 className="mb-4 text-3xl font-bold">核心功能模块</h2>
            <p className="text-lg text-muted-foreground">
              基于城堡模型架构的六大核心功能模块
            </p>
          </div>

          <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
            <FeatureCard
              icon={Users}
              title="员工管理 (CoreHR)"
              description="完整的员工信息管理、组织架构管理、职位管理和汇报关系管理"
              features={['员工 CRUD 操作', '组织架构可视化', '职位权限映射', '事务性发件箱']}
              status="已完成"
            />

            <FeatureCard
              icon={Brain}
              title="智能交互 (AI Gateway)"
              description="基于自然语言处理的智能对话管理与上下文维护"
              features={['意图识别', '多轮对话', 'Redis 状态管理', '智能推荐']}
              status="已完成"
            />

            <FeatureCard
              icon={Workflow}
              title="工作流引擎 (Temporal)"
              description="分布式工作流编排，支持信号驱动的异步流程"
              features={['员工入职流程', '休假审批流程', '批量处理', '实时状态跟踪']}
              status="已完成"
            />

            <FeatureCard
              icon={Shield}
              title="企业级安全架构"
              description="OPA策略引擎 + PostgreSQL RLS 多层安全防护"
              features={['OPA 授权引擎', 'RLS 数据隔离', '审计跟踪', '威胁检测']}
              status="已完成"
            />

            <FeatureCard
              icon={BarChart3}
              title="系统监控与可观测性"
              description="全方位监控，结构化日志和Prometheus指标收集"
              features={['健康检查', '性能监控', '业务指标', '实时数据流']}
              status="已完成"
            />

            <FeatureCard
              icon={Building2}
              title="前端用户界面"
              description="Next.js 现代化前端，响应式设计和实时数据同步"
              features={['响应式设计', 'TypeScript 支持', '组件化架构', '实时更新']}
              status="开发中"
            />
          </div>
        </section>

        {/* 技术亮点 */}
        <section className="bg-muted/50 py-16">
          <div className="container-responsive">
            <div className="mb-12 text-center">
              <h2 className="mb-4 text-3xl font-bold">技术创新亮点</h2>
              <p className="text-lg text-muted-foreground">
                采用最新技术栈，确保系统的可靠性、安全性和性能
              </p>
            </div>

            <div className="grid gap-8 lg:grid-cols-2">
              <div className="space-y-6">
                <TechHighlight
                  title="城堡模型架构 v3.0"
                  description="独特的城堡模型架构设计，实现模块化、可扩展的系统架构"
                />
                <TechHighlight
                  title="多层安全防护"
                  description="API层 + 业务层 + 数据层三重安全保障，确保数据安全"
                />
                <TechHighlight
                  title="企业级工作流引擎"
                  description="基于Temporal的分布式工作流，支持复杂业务流程编排"
                />
              </div>
              <div className="space-y-6">
                <TechHighlight
                  title="AI驱动的智能交互"
                  description="自然语言处理和意图识别，提供智能化的用户体验"
                />
                <TechHighlight
                  title="微服务通信架构"
                  description="gRPC高效通信，Redis状态共享，事件驱动设计"
                />
                <TechHighlight
                  title="全方位可观测性"
                  description="结构化日志、指标收集、性能跟踪，全方位系统监控"
                />
              </div>
            </div>
          </div>
        </section>
      </main>

      {/* 页脚 */}
      <footer className="border-t bg-background py-12">
        <div className="container-responsive">
          <div className="grid gap-8 lg:grid-cols-4">
            <div>
              <div className="flex items-center space-x-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <Castle className="h-5 w-5" />
                </div>
                <span className="text-xl font-bold">Cube Castle</span>
              </div>
              <p className="mt-4 text-sm text-muted-foreground">
                让企业级 HR 管理变得智能、安全、高效
              </p>
            </div>
            
            <div>
              <h3 className="font-semibold">产品</h3>
              <ul className="mt-4 space-y-2 text-sm">
                <li><Link href="/features" className="text-muted-foreground hover:text-foreground">功能特性</Link></li>
                <li><Link href="/pricing" className="text-muted-foreground hover:text-foreground">定价方案</Link></li>
                <li><Link href="/demo" className="text-muted-foreground hover:text-foreground">产品演示</Link></li>
              </ul>
            </div>
            
            <div>
              <h3 className="font-semibold">支持</h3>
              <ul className="mt-4 space-y-2 text-sm">
                <li><Link href="/docs" className="text-muted-foreground hover:text-foreground">帮助文档</Link></li>
                <li><Link href="/contact" className="text-muted-foreground hover:text-foreground">联系我们</Link></li>
                <li><Link href="/status" className="text-muted-foreground hover:text-foreground">系统状态</Link></li>
              </ul>
            </div>
            
            <div>
              <h3 className="font-semibold">公司</h3>
              <ul className="mt-4 space-y-2 text-sm">
                <li><Link href="/about" className="text-muted-foreground hover:text-foreground">关于我们</Link></li>
                <li><Link href="/blog" className="text-muted-foreground hover:text-foreground">技术博客</Link></li>
                <li><Link href="/careers" className="text-muted-foreground hover:text-foreground">加入我们</Link></li>
              </ul>
            </div>
          </div>
          
          <div className="mt-8 border-t pt-8 text-center text-sm text-muted-foreground">
            <p>&copy; 2025 Cube Castle. 保留所有权利。版本 v1.4.0-beta</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

// 功能卡片组件
interface FeatureCardProps {
  icon: React.ComponentType<{ className?: string }>
  title: string
  description: string
  features: string[]
  status: string
}

function FeatureCard({ icon: Icon, title, description, features, status }: FeatureCardProps) {
  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <Icon className="h-8 w-8 text-primary" />
          <Badge variant={status === '已完成' ? 'default' : 'secondary'}>
            {status}
          </Badge>
        </div>
        <CardTitle className="text-xl">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="space-y-2">
          {features.map((feature, index) => (
            <li key={index} className="flex items-center text-sm">
              <CheckCircle className="mr-2 h-4 w-4 text-green-500" />
              {feature}
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}

// 技术亮点组件
interface TechHighlightProps {
  title: string
  description: string
}

function TechHighlight({ title, description }: TechHighlightProps) {
  return (
    <div className="flex space-x-4">
      <div className="flex-shrink-0">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <CheckCircle className="h-5 w-5" />
        </div>
      </div>
      <div>
        <h3 className="font-semibold">{title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
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