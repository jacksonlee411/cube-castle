import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  Play,
  Pause,
  RotateCcw,
  Eye,
  Settings,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  User,
  FileText,
  DollarSign,
  Briefcase,
  Building2,
  Calendar,
  ArrowRight,
  Zap,
  Target,
  TrendingUp,
  Activity,
  Users,
  GitBranch,
  RefreshCw
} from 'lucide-react';
import { toast } from 'react-hot-toast';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Progress } from '@/components/ui/progress';

interface WorkflowTemplate {
  id: string;
  name: string;
  type: 'position_change' | 'leave_request' | 'expense_claim' | 'document_approval';
  description: string;
  category: 'hr' | 'finance' | 'operations' | 'admin';
  complexity: 'simple' | 'medium' | 'complex';
  averageTime: number;
  steps: WorkflowStep[];
  isPopular: boolean;
  usageCount: number;
}

interface WorkflowStep {
  id: string;
  name: string;
  type: 'approval' | 'task' | 'notification' | 'condition';
  description: string;
  assigneeRole: string;
  estimatedTime: number;
  isRequired: boolean;
  conditions?: string[];
}

interface DemoExecution {
  templateId: string;
  currentStep: number;
  status: 'running' | 'paused' | 'completed' | 'failed';
  startTime: string;
  progress: number;
  logs: ExecutionLog[];
}

interface ExecutionLog {
  timestamp: string;
  step: string;
  action: string;
  actor: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
}

const WorkflowDemoPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [templates, setTemplates] = useState<WorkflowTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<WorkflowTemplate | null>(null);
  const [demoExecution, setDemoExecution] = useState<DemoExecution | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedComplexity, setSelectedComplexity] = useState<string>('all');

  // Sample data
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleTemplates: WorkflowTemplate[] = [
        {
          id: '1',
          name: '员工职位晋升',
          type: 'position_change',
          description: '员工职位晋升申请的完整审批流程，包括绩效评估、薪资调整和系统更新',
          category: 'hr',
          complexity: 'medium',
          averageTime: 5,
          isPopular: true,
          usageCount: 156,
          steps: [
            {
              id: '1',
              name: '提交申请',
              type: 'task',
              description: 'HR部门提交员工晋升申请',
              assigneeRole: 'HR专员',
              estimatedTime: 30,
              isRequired: true
            },
            {
              id: '2',
              name: '直属经理审批',
              type: 'approval',
              description: '直属经理审核员工表现和晋升合理性',
              assigneeRole: '直属经理',
              estimatedTime: 120,
              isRequired: true
            },
            {
              id: '3',
              name: '人事总监审批',
              type: 'approval',
              description: '人事总监最终审批和薪资调整确认',
              assigneeRole: '人事总监',
              estimatedTime: 180,
              isRequired: true
            },
            {
              id: '4',
              name: '系统更新',
              type: 'task',
              description: '更新员工档案和薪资系统',
              assigneeRole: '系统管理员',
              estimatedTime: 15,
              isRequired: true
            },
            {
              id: '5',
              name: '通知相关人员',
              type: 'notification',
              description: '通知员工和相关部门晋升结果',
              assigneeRole: '系统自动',
              estimatedTime: 5,
              isRequired: true
            }
          ]
        },
        {
          id: '2',
          name: '年假申请',
          type: 'leave_request',
          description: '员工年假申请的标准审批流程，适用于所有员工类型',
          category: 'hr',
          complexity: 'simple',
          averageTime: 2,
          isPopular: true,
          usageCount: 342,
          steps: [
            {
              id: '1',
              name: '提交申请',
              type: 'task',
              description: '员工提交年假申请',
              assigneeRole: '申请人',
              estimatedTime: 10,
              isRequired: true
            },
            {
              id: '2',
              name: '直属经理审批',
              type: 'approval',
              description: '直属经理审核请假时间和工作安排',
              assigneeRole: '直属经理',
              estimatedTime: 60,
              isRequired: true
            },
            {
              id: '3',
              name: 'HR确认',
              type: 'task',
              description: 'HR部门确认年假余额和记录',
              assigneeRole: 'HR专员',
              estimatedTime: 15,
              isRequired: true
            },
            {
              id: '4',
              name: '系统记录',
              type: 'notification',
              description: '系统自动记录请假信息',
              assigneeRole: '系统自动',
              estimatedTime: 2,
              isRequired: true
            }
          ]
        },
        {
          id: '3',
          name: '费用报销',
          type: 'expense_claim',
          description: '员工费用报销申请流程，支持多级审批和财务核算',
          category: 'finance',
          complexity: 'complex',
          averageTime: 7,
          isPopular: false,
          usageCount: 89,
          steps: [
            {
              id: '1',
              name: '提交报销单',
              type: 'task',
              description: '员工提交费用报销单和发票',
              assigneeRole: '申请人',
              estimatedTime: 20,
              isRequired: true
            },
            {
              id: '2',
              name: '部门经理审批',
              type: 'approval',
              description: '部门经理审核报销合理性',
              assigneeRole: '部门经理',
              estimatedTime: 90,
              isRequired: true,
              conditions: ['金额 > 1000元']
            },
            {
              id: '3',
              name: '财务审核',
              type: 'approval',
              description: '财务部门审核发票真实性和合规性',
              assigneeRole: '财务专员',
              estimatedTime: 120,
              isRequired: true
            },
            {
              id: '4',
              name: '财务总监审批',
              type: 'approval',
              description: '财务总监最终审批',
              assigneeRole: '财务总监',
              estimatedTime: 180,
              isRequired: true,
              conditions: ['金额 > 5000元']
            },
            {
              id: '5',
              name: '出纳付款',
              type: 'task',
              description: '出纳执行付款操作',
              assigneeRole: '出纳',
              estimatedTime: 30,
              isRequired: true
            },
            {
              id: '6',
              name: '付款通知',
              type: 'notification',
              description: '通知申请人付款完成',
              assigneeRole: '系统自动',
              estimatedTime: 2,
              isRequired: true
            }
          ]
        },
        {
          id: '4',
          name: '合同审批',
          type: 'document_approval',
          description: '合同文档的法务审批和签署流程',
          category: 'operations',
          complexity: 'complex',
          averageTime: 10,
          isPopular: false,
          usageCount: 45,
          steps: [
            {
              id: '1',
              name: '提交合同',
              type: 'task',
              description: '业务部门提交合同文档',
              assigneeRole: '业务专员',
              estimatedTime: 15,
              isRequired: true
            },
            {
              id: '2',
              name: '法务审核',
              type: 'approval',
              description: '法务部门审核合同条款',
              assigneeRole: '法务专员',
              estimatedTime: 240,
              isRequired: true
            },
            {
              id: '3',
              name: '商务总监审批',
              type: 'approval',
              description: '商务总监审批合同内容',
              assigneeRole: '商务总监',
              estimatedTime: 180,
              isRequired: true
            },
            {
              id: '4',
              name: '总经理签署',
              type: 'approval',
              description: '总经理最终签署合同',
              assigneeRole: '总经理',
              estimatedTime: 120,
              isRequired: true,
              conditions: ['合同金额 > 50万元']
            },
            {
              id: '5',
              name: '归档通知',
              type: 'notification',
              description: '合同签署完成，归档保存',
              assigneeRole: '系统自动',
              estimatedTime: 5,
              isRequired: true
            }
          ]
        }
      ];
      
      setTemplates(sampleTemplates);
      setLoading(false);
    }, 800);
  }, []);

  const startDemo = (template: WorkflowTemplate) => {
    setSelectedTemplate(template);
    const newExecution: DemoExecution = {
      templateId: template.id,
      currentStep: 0,
      status: 'running',
      startTime: new Date().toISOString(),
      progress: 0,
      logs: [{
        timestamp: new Date().toISOString(),
        step: '系统',
        action: '开始',
        actor: '演示系统',
        message: `开始执行工作流演示：${template.name}`,
        type: 'info'
      }]
    };
    setDemoExecution(newExecution);
    
    // Start automatic execution
    executeNextStep(newExecution, template);
    toast.success(`开始演示：${template.name}`);
  };

  const executeNextStep = (execution: DemoExecution, template: WorkflowTemplate) => {
    if (execution.currentStep >= template.steps.length) {
      // Workflow completed
      const completedExecution = {
        ...execution,
        status: 'completed' as const,
        progress: 100,
        logs: [...execution.logs, {
          timestamp: new Date().toISOString(),
          step: '系统',
          action: '完成',
          actor: '演示系统',
          message: '工作流演示执行完成',
          type: 'success' as const
        }]
      };
      setDemoExecution(completedExecution);
      toast.success('工作流演示执行完成！');
      return;
    }

    const currentStep = template.steps[execution.currentStep];
    const delay = Math.min(currentStep.estimatedTime * 10, 3000); // Scale down for demo
    
    setTimeout(() => {
      const progress = Math.round(((execution.currentStep + 1) / template.steps.length) * 100);
      const updatedExecution = {
        ...execution,
        currentStep: execution.currentStep + 1,
        progress,
        logs: [...execution.logs, {
          timestamp: new Date().toISOString(),
          step: currentStep.name,
          action: currentStep.type === 'approval' ? '审批通过' : '执行完成',
          actor: currentStep.assigneeRole,
          message: `${currentStep.name} - ${currentStep.description}`,
          type: 'success' as const
        }]
      };
      
      setDemoExecution(updatedExecution);
      
      if (updatedExecution.status === 'running') {
        executeNextStep(updatedExecution, template);
      }
    }, delay);
  };

  const pauseDemo = () => {
    if (demoExecution) {
      setDemoExecution({
        ...demoExecution,
        status: 'paused',
        logs: [...demoExecution.logs, {
          timestamp: new Date().toISOString(),
          step: '系统',
          action: '暂停',
          actor: '用户',
          message: '演示已暂停',
          type: 'warning'
        }]
      });
      toast('演示已暂停', { icon: 'ℹ️' });
    }
  };

  const resumeDemo = () => {
    if (demoExecution && selectedTemplate) {
      const resumedExecution: DemoExecution = {
        ...demoExecution,
        status: 'running' as const,
        logs: [...demoExecution.logs, {
          timestamp: new Date().toISOString(),
          step: '系统',
          action: '恢复',
          actor: '用户',
          message: '演示已恢复',
          type: 'info' as const
        }]
      };
      setDemoExecution(resumedExecution);
      executeNextStep(resumedExecution, selectedTemplate);
      toast.success('演示已恢复');
    }
  };

  const resetDemo = () => {
    setDemoExecution(null);
    setSelectedTemplate(null);
    toast('演示已重置', { icon: 'ℹ️' });
  };

  const getTypeColor = (type: string): "default" | "destructive" | "secondary" => {
    const colors = {
      position_change: 'default' as const,
      leave_request: 'default' as const,
      expense_claim: 'secondary' as const,
      document_approval: 'destructive' as const
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getTypeLabel = (type: string) => {
    const labels = {
      position_change: '职位变更',
      leave_request: '请假申请',
      expense_claim: '费用报销',
      document_approval: '文档审批'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getCategoryLabel = (category: string) => {
    const labels = {
      hr: '人力资源',
      finance: '财务管理',
      operations: '运营管理',
      admin: '行政管理'
    };
    return labels[category as keyof typeof labels] || category;
  };

  const getComplexityColor = (complexity: string): string => {
    const colors = {
      simple: 'text-green-600',
      medium: 'text-yellow-600',
      complex: 'text-red-600'
    };
    return colors[complexity as keyof typeof colors] || 'text-gray-600';
  };

  const getComplexityLabel = (complexity: string) => {
    const labels = {
      simple: '简单',
      medium: '中等',
      complex: '复杂'
    };
    return labels[complexity as keyof typeof labels] || complexity;
  };

  const getStepTypeIcon = (type: string) => {
    const icons = {
      approval: <CheckCircle className="h-4 w-4 text-blue-500" />,
      task: <Settings className="h-4 w-4 text-purple-500" />,
      notification: <AlertCircle className="h-4 w-4 text-green-500" />,
      condition: <GitBranch className="h-4 w-4 text-orange-500" />
    };
    return icons[type as keyof typeof icons] || icons.task;
  };

  const filteredTemplates = templates.filter(template => {
    const categoryMatch = selectedCategory === 'all' || template.category === selectedCategory;
    const complexityMatch = selectedComplexity === 'all' || template.complexity === selectedComplexity;
    return categoryMatch && complexityMatch;
  });

  const renderTemplateCard = (template: WorkflowTemplate) => (
    <Card key={template.id} className="hover:shadow-lg transition-shadow">
      <CardContent className="p-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <h3 className="text-lg font-semibold">{template.name}</h3>
              {template.isPopular && (
                <Badge variant="default" className="text-xs">
                  热门
                </Badge>
              )}
            </div>
            <p className="text-gray-600 text-sm mb-3">{template.description}</p>
            
            <div className="flex items-center gap-4 text-sm text-gray-500 mb-4">
              <div className="flex items-center gap-1">
                <Building2 className="h-4 w-4" />
                <span>{getCategoryLabel(template.category)}</span>
              </div>
              <div className="flex items-center gap-1">
                <TrendingUp className="h-4 w-4" />
                <span className={getComplexityColor(template.complexity)}>
                  {getComplexityLabel(template.complexity)}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <Clock className="h-4 w-4" />
                <span>{template.averageTime}天</span>
              </div>
              <div className="flex items-center gap-1">
                <Users className="h-4 w-4" />
                <span>{template.usageCount}次</span>
              </div>
            </div>
          </div>
          
          <Badge variant={getTypeColor(template.type)}>
            {getTypeLabel(template.type)}
          </Badge>
        </div>
        
        <div className="mb-4">
          <div className="text-sm font-medium text-gray-700 mb-2">
            流程步骤 ({template.steps.length}个)
          </div>
          <div className="space-y-2">
            {template.steps.slice(0, 3).map((step, index) => (
              <div key={step.id} className="flex items-center gap-2 text-sm">
                <span className="w-5 h-5 rounded-full bg-gray-100 text-gray-600 text-xs flex items-center justify-center">
                  {index + 1}
                </span>
                {getStepTypeIcon(step.type)}
                <span>{step.name}</span>
                <span className="text-gray-400">({step.assigneeRole})</span>
              </div>
            ))}
            {template.steps.length > 3 && (
              <div className="text-xs text-gray-500 ml-7">
                还有 {template.steps.length - 3} 个步骤...
              </div>
            )}
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button 
            onClick={() => startDemo(template)}
            disabled={demoExecution?.status === 'running'}
            className="flex-1"
          >
            <Play className="mr-2 h-4 w-4" />
            开始演示
          </Button>
          <Button variant="outline">
            <Eye className="mr-2 h-4 w-4" />
            查看详情
          </Button>
        </div>
      </CardContent>
    </Card>
  );

  const renderExecutionPanel = () => {
    if (!demoExecution || !selectedTemplate) return null;

    return (
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              演示执行中: {selectedTemplate.name}
            </CardTitle>
            <div className="flex gap-2">
              {demoExecution.status === 'running' && (
                <Button size="sm" variant="outline" onClick={pauseDemo}>
                  <Pause className="mr-2 h-4 w-4" />
                  暂停
                </Button>
              )}
              {demoExecution.status === 'paused' && (
                <Button size="sm" onClick={resumeDemo}>
                  <Play className="mr-2 h-4 w-4" />
                  继续
                </Button>
              )}
              <Button size="sm" variant="outline" onClick={resetDemo}>
                <RotateCcw className="mr-2 h-4 w-4" />
                重置
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Progress Section */}
            <div>
              <div className="mb-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium">执行进度</span>
                  <span className="text-sm text-gray-500">{demoExecution.progress}%</span>
                </div>
                <Progress value={demoExecution.progress} />
              </div>
              
              <div className="space-y-3">
                {selectedTemplate.steps.map((step, index) => (
                  <div 
                    key={step.id} 
                    className={`flex items-center gap-3 p-3 rounded-lg ${
                      index < demoExecution.currentStep ? 'bg-green-50 border border-green-200' :
                      index === demoExecution.currentStep ? 'bg-blue-50 border border-blue-200' :
                      'bg-gray-50'
                    }`}
                  >
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                      index < demoExecution.currentStep ? 'bg-green-500 text-white' :
                      index === demoExecution.currentStep ? 'bg-blue-500 text-white' :
                      'bg-gray-300 text-gray-600'
                    }`}>
                      {index < demoExecution.currentStep ? (
                        <CheckCircle className="h-4 w-4" />
                      ) : (
                        index + 1
                      )}
                    </div>
                    <div className="flex-1">
                      <div className="font-medium text-sm">{step.name}</div>
                      <div className="text-xs text-gray-500">{step.assigneeRole}</div>
                    </div>
                    {index === demoExecution.currentStep && demoExecution.status === 'running' && (
                      <RefreshCw className="h-4 w-4 text-blue-500 animate-spin" />
                    )}
                  </div>
                ))}
              </div>
            </div>
            
            {/* Execution Logs */}
            <div>
              <div className="text-sm font-medium mb-3">执行日志</div>
              <div className="space-y-2 max-h-80 overflow-y-auto">
                {demoExecution.logs.map((log, index) => (
                  <div key={index} className="flex items-start gap-2 p-2 bg-gray-50 rounded text-sm">
                    <div className={`w-2 h-2 rounded-full mt-2 flex-shrink-0 ${
                      log.type === 'success' ? 'bg-green-500' :
                      log.type === 'warning' ? 'bg-yellow-500' :
                      log.type === 'error' ? 'bg-red-500' :
                      'bg-blue-500'
                    }`} />
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium">{log.step}</span>
                        <span className="text-gray-500">-</span>
                        <span className="text-gray-500">{log.actor}</span>
                        <span className="text-xs text-gray-400">
                          {format(new Date(log.timestamp), 'HH:mm:ss')}
                        </span>
                      </div>
                      <div className="text-gray-600">{log.message}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">工作流演示中心</h1>
        <p className="text-gray-600 mt-1">
          交互式工作流程演示 - 体验不同类型的审批流程和自动化任务
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">可用模板</p>
                <p className="text-2xl font-bold">{templates.length}</p>
              </div>
              <Target className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">热门模板</p>
                <p className="text-2xl font-bold text-green-600">
                  {templates.filter(t => t.isPopular).length}
                </p>
              </div>
              <Zap className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">总使用次数</p>
                <p className="text-2xl font-bold text-purple-600">
                  {templates.reduce((sum, t) => sum + t.usageCount, 0)}
                </p>
              </div>
              <Activity className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">平均完成时间</p>
                <p className="text-2xl font-bold text-orange-600">
                  {Math.round(templates.reduce((sum, t) => sum + t.averageTime, 0) / templates.length)}天
                </p>
              </div>
              <Clock className="h-8 w-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Execution Panel */}
      {renderExecutionPanel()}

      {/* Filters */}
      <Card className="mb-6">
        <CardContent className="p-4">
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="text-sm font-medium">业务类别</label>
              <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部类别</SelectItem>
                  <SelectItem value="hr">人力资源</SelectItem>
                  <SelectItem value="finance">财务管理</SelectItem>
                  <SelectItem value="operations">运营管理</SelectItem>
                  <SelectItem value="admin">行政管理</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="flex-1">
              <label className="text-sm font-medium">复杂程度</label>
              <Select value={selectedComplexity} onValueChange={setSelectedComplexity}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部复杂度</SelectItem>
                  <SelectItem value="simple">简单</SelectItem>
                  <SelectItem value="medium">中等</SelectItem>
                  <SelectItem value="complex">复杂</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="flex items-end">
              <Button variant="outline">
                <RefreshCw className="mr-2 h-4 w-4" />
                重置筛选
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Template Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {loading ? (
          <div className="col-span-2 flex items-center justify-center py-12">
            <div className="text-gray-500">加载模板中...</div>
          </div>
        ) : filteredTemplates.length > 0 ? (
          filteredTemplates.map(renderTemplateCard)
        ) : (
          <div className="col-span-2">
            <Alert>
              <AlertDescription>
                没有找到符合条件的工作流模板，请调整筛选条件。
              </AlertDescription>
            </Alert>
          </div>
        )}
      </div>
    </div>
  );
};

export default WorkflowDemoPage;