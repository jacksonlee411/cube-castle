import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  ArrowLeft,
  Play,
  Pause,
  Square,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  User,
  Calendar,
  MessageSquare,
  FileText,
  Settings,
  MoreHorizontal,
  Edit2,
  Eye,
  Download,
  Share2,
  RefreshCw,
  ChevronRight,
  ChevronDown
} from 'lucide-react';
import { toast } from 'react-hot-toast';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Progress } from '@/components/ui/progress';

interface WorkflowStep {
  id: string;
  name: string;
  description: string;
  type: 'approval' | 'task' | 'notification' | 'condition';
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'skipped';
  assignee?: string;
  assigneeName?: string;
  dueDate?: string;
  completedAt?: string;
  comments?: string;
  order: number;
}

interface WorkflowInstance {
  id: string;
  title: string;
  description: string;
  type: 'position_change' | 'leave_request' | 'expense_claim' | 'document_approval';
  status: 'draft' | 'running' | 'completed' | 'failed' | 'cancelled';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  initiator: string;
  initiatorName: string;
  currentStep?: string;
  progress: number;
  createdAt: string;
  updatedAt: string;
  dueDate?: string;
  completedAt?: string;
  steps: WorkflowStep[];
  data?: Record<string, any>;
}

interface ActivityLog {
  id: string;
  workflowId: string;
  stepId?: string;
  action: 'created' | 'started' | 'approved' | 'rejected' | 'completed' | 'commented';
  actor: string;
  actorName: string;
  message: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

const WorkflowDetailPage: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  const [loading, setLoading] = useState(false);
  const [workflow, setWorkflow] = useState<WorkflowInstance | null>(null);
  const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([]);
  const [expandedStep, setExpandedStep] = useState<string | null>(null);
  const [isCommentModalVisible, setIsCommentModalVisible] = useState(false);
  const [selectedStep, setSelectedStep] = useState<WorkflowStep | null>(null);
  const [comment, setComment] = useState('');

  // Sample data
  useEffect(() => {
    if (!id) return;
    
    setLoading(true);
    setTimeout(() => {
      const sampleWorkflow: WorkflowInstance = {
        id: id as string,
        title: '张三 - 职位晋升申请',
        description: '从前端工程师晋升至高级前端工程师',
        type: 'position_change',
        status: 'running',
        priority: 'medium',
        initiator: 'HR001',
        initiatorName: '陈静',
        currentStep: '2',
        progress: 60,
        createdAt: '2024-12-01T09:00:00Z',
        updatedAt: '2024-12-02T14:30:00Z',
        dueDate: '2024-12-10T18:00:00Z',
        data: {
          employeeId: 'EMP001',
          employeeName: '张三',
          currentPosition: '前端工程师',
          newPosition: '高级前端工程师',
          currentSalary: 18000,
          newSalary: 25000,
          effectiveDate: '2024-12-15'
        },
        steps: [
          {
            id: '1',
            name: '申请提交',
            description: '人事部门提交职位变更申请',
            type: 'task',
            status: 'completed',
            assignee: 'HR001',
            assigneeName: '陈静',
            completedAt: '2024-12-01T09:15:00Z',
            comments: '已完成员工绩效评估和薪资调整方案',
            order: 1
          },
          {
            id: '2',
            name: '直属经理审批',
            description: '直属经理审批职位变更申请',
            type: 'approval',
            status: 'in_progress',
            assignee: 'TEC001',
            assigneeName: '李强',
            dueDate: '2024-12-05T18:00:00Z',
            order: 2
          },
          {
            id: '3',
            name: '人事总监审批',
            description: '人事总监最终审批',
            type: 'approval',
            status: 'pending',
            assignee: 'HR_DIR',
            assigneeName: '王总监',
            dueDate: '2024-12-08T18:00:00Z',
            order: 3
          },
          {
            id: '4',
            name: '系统更新',
            description: '更新员工信息和薪资系统',
            type: 'task',
            status: 'pending',
            assignee: 'SYS001',
            assigneeName: '系统自动',
            order: 4
          },
          {
            id: '5',
            name: '完成通知',
            description: '通知相关人员职位变更完成',
            type: 'notification',
            status: 'pending',
            assignee: 'SYS001',
            assigneeName: '系统自动',
            order: 5
          }
        ]
      };

      const sampleLogs: ActivityLog[] = [
        {
          id: '1',
          workflowId: id as string,
          stepId: '1',
          action: 'created',
          actor: 'HR001',
          actorName: '陈静',
          message: '创建了职位晋升工作流',
          timestamp: '2024-12-01T09:00:00Z'
        },
        {
          id: '2',
          workflowId: id as string,
          stepId: '1',
          action: 'completed',
          actor: 'HR001',
          actorName: '陈静',
          message: '完成申请提交，已完成员工绩效评估和薪资调整方案',
          timestamp: '2024-12-01T09:15:00Z'
        },
        {
          id: '3',
          workflowId: id as string,
          stepId: '2',
          action: 'started',
          actor: 'TEC001',
          actorName: '李强',
          message: '开始审批职位变更申请',
          timestamp: '2024-12-01T09:20:00Z'
        },
        {
          id: '4',
          workflowId: id as string,
          stepId: '2',
          action: 'commented',
          actor: 'TEC001',
          actorName: '李强',
          message: '张三在项目中表现出色，技术能力强，支持晋升',
          timestamp: '2024-12-02T14:30:00Z'
        }
      ];
      
      setWorkflow(sampleWorkflow);
      setActivityLogs(sampleLogs);
      setLoading(false);
    }, 1000);
  }, [id]);

  const handleApproveStep = async (stepId: string) => {
    if (!workflow) return;
    
    try {
      setLoading(true);
      
      const updatedSteps = workflow.steps.map(step => {
        if (step.id === stepId) {
          return {
            ...step,
            status: 'completed' as const,
            completedAt: new Date().toISOString(),
            comments: comment || undefined
          };
        }
        if (step.order === workflow.steps.find(s => s.id === stepId)!.order + 1) {
          return {
            ...step,
            status: 'in_progress' as const
          };
        }
        return step;
      });

      const completedSteps = updatedSteps.filter(s => s.status === 'completed').length;
      const progress = Math.round((completedSteps / updatedSteps.length) * 100);
      const newStatus = progress === 100 ? 'completed' : 'running';

      const updatedWorkflow: WorkflowInstance = {
        ...workflow,
        steps: updatedSteps,
        progress,
        status: newStatus as WorkflowInstance['status'],
        updatedAt: new Date().toISOString(),
        completedAt: progress === 100 ? new Date().toISOString() : undefined
      };

      setWorkflow(updatedWorkflow);
      
      const newLog: ActivityLog = {
        id: Date.now().toString(),
        workflowId: workflow.id,
        stepId,
        action: 'approved',
        actor: 'current_user',
        actorName: '当前用户',
        message: comment || '已审批通过',
        timestamp: new Date().toISOString()
      };
      
      setActivityLogs(prev => [newLog, ...prev]);
      setComment('');
      setIsCommentModalVisible(false);
      
      toast.success('审批成功');
    } catch (error) {
      toast.error('操作失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleRejectStep = async (stepId: string) => {
    if (!workflow || !comment.trim()) {
      toast.error('请填写拒绝原因');
      return;
    }
    
    try {
      setLoading(true);
      
      const updatedSteps = workflow.steps.map(step => {
        if (step.id === stepId) {
          return {
            ...step,
            status: 'failed' as const,
            completedAt: new Date().toISOString(),
            comments: comment
          };
        }
        return step;
      });

      const updatedWorkflow: WorkflowInstance = {
        ...workflow,
        steps: updatedSteps,
        status: 'failed' as const,
        updatedAt: new Date().toISOString()
      };

      setWorkflow(updatedWorkflow);
      
      const newLog: ActivityLog = {
        id: Date.now().toString(),
        workflowId: workflow.id,
        stepId,
        action: 'rejected',
        actor: 'current_user',
        actorName: '当前用户',
        message: comment,
        timestamp: new Date().toISOString()
      };
      
      setActivityLogs(prev => [newLog, ...prev]);
      setComment('');
      setIsCommentModalVisible(false);
      
      toast.success('已拒绝申请');
    } catch (error) {
      toast.error('操作失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string): "default" | "destructive" | "secondary" => {
    const colors = {
      draft: 'secondary' as const,
      running: 'default' as const,
      completed: 'default' as const,
      failed: 'destructive' as const,
      cancelled: 'secondary' as const
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusLabel = (status: string) => {
    const labels = {
      draft: '草稿',
      running: '进行中',
      completed: '已完成',
      failed: '已失败',
      cancelled: '已取消'
    };
    return labels[status as keyof typeof labels] || status;
  };

  const getPriorityColor = (priority: string): "default" | "destructive" | "secondary" => {
    const colors = {
      low: 'secondary' as const,
      medium: 'default' as const,
      high: 'default' as const,
      urgent: 'destructive' as const
    };
    return colors[priority as keyof typeof colors] || 'default';
  };

  const getPriorityLabel = (priority: string) => {
    const labels = {
      low: '低',
      medium: '中',
      high: '高',
      urgent: '紧急'
    };
    return labels[priority as keyof typeof labels] || priority;
  };

  const getStepStatusIcon = (status: string) => {
    const icons = {
      pending: <Clock className="h-4 w-4 text-gray-400" />,
      in_progress: <RefreshCw className="h-4 w-4 text-blue-500 animate-spin" />,
      completed: <CheckCircle className="h-4 w-4 text-green-500" />,
      failed: <XCircle className="h-4 w-4 text-red-500" />,
      skipped: <AlertCircle className="h-4 w-4 text-yellow-500" />
    };
    return icons[status as keyof typeof icons] || icons.pending;
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

  const renderWorkflowSteps = () => {
    if (!workflow) return null;

    return (
      <div className="space-y-4">
        {workflow.steps.map((step, index) => (
          <div key={step.id} className="relative">
            {/* Connection Line */}
            {index < workflow.steps.length - 1 && (
              <div className="absolute left-6 top-12 w-px h-16 bg-gray-200 z-0"></div>
            )}
            
            <Card className={`relative z-10 ${step.status === 'in_progress' ? 'ring-2 ring-blue-500' : ''}`}>
              <CardContent className="p-4">
                <div className="flex items-start gap-4">
                  {/* Step Icon */}
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
                    step.status === 'completed' ? 'bg-green-100' :
                    step.status === 'in_progress' ? 'bg-blue-100' :
                    step.status === 'failed' ? 'bg-red-100' :
                    'bg-gray-100'
                  }`}>
                    {getStepStatusIcon(step.status)}
                  </div>
                  
                  {/* Step Content */}
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-2">
                      <div>
                        <h3 className="font-semibold flex items-center gap-2">
                          {step.name}
                          <Badge variant="outline" className="text-xs">
                            {step.type === 'approval' ? '审批' : 
                             step.type === 'task' ? '任务' :
                             step.type === 'notification' ? '通知' : '条件'}
                          </Badge>
                        </h3>
                        <p className="text-sm text-gray-600">{step.description}</p>
                      </div>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setExpandedStep(expandedStep === step.id ? null : step.id)}
                      >
                        {expandedStep === step.id ? (
                          <ChevronDown className="h-4 w-4" />
                        ) : (
                          <ChevronRight className="h-4 w-4" />
                        )}
                      </Button>
                    </div>

                    <div className="flex items-center gap-4 text-sm text-gray-500 mb-2">
                      {step.assigneeName && (
                        <div className="flex items-center gap-1">
                          <User className="h-3 w-3" />
                          <span>{step.assigneeName}</span>
                        </div>
                      )}
                      
                      {step.dueDate && step.status !== 'completed' && (
                        <div className="flex items-center gap-1">
                          <Calendar className="h-3 w-3" />
                          <span>截止: {format(new Date(step.dueDate), 'MM-dd HH:mm')}</span>
                        </div>
                      )}
                      
                      {step.completedAt && (
                        <div className="flex items-center gap-1">
                          <CheckCircle className="h-3 w-3" />
                          <span>完成: {format(new Date(step.completedAt), 'MM-dd HH:mm')}</span>
                        </div>
                      )}
                    </div>

                    {/* Expanded Content */}
                    {expandedStep === step.id && (
                      <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                        {step.comments && (
                          <div className="mb-4">
                            <p className="text-sm font-medium mb-1">备注：</p>
                            <p className="text-sm text-gray-600 bg-white p-2 rounded border">
                              {step.comments}
                            </p>
                          </div>
                        )}
                        
                        {step.status === 'in_progress' && step.type === 'approval' && (
                          <div className="flex gap-2">
                            <Button
                              size="sm"
                              onClick={() => {
                                setSelectedStep(step);
                                setIsCommentModalVisible(true);
                              }}
                            >
                              <CheckCircle className="mr-2 h-4 w-4" />
                              审批通过
                            </Button>
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={() => {
                                setSelectedStep(step);
                                setIsCommentModalVisible(true);
                              }}
                            >
                              <XCircle className="mr-2 h-4 w-4" />
                              拒绝申请
                            </Button>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        ))}
      </div>
    );
  };

  const renderActivityLog = () => {
    return (
      <div className="space-y-3">
        {activityLogs.map((log) => (
          <div key={log.id} className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
            <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
              <User className="h-4 w-4 text-blue-600" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-medium text-sm">{log.actorName}</span>
                <Badge variant="outline" className="text-xs">
                  {log.action === 'created' ? '创建' :
                   log.action === 'started' ? '开始' :
                   log.action === 'approved' ? '审批' :
                   log.action === 'rejected' ? '拒绝' :
                   log.action === 'completed' ? '完成' :
                   log.action === 'commented' ? '评论' : log.action}
                </Badge>
                <span className="text-xs text-gray-500">
                  {format(new Date(log.timestamp), 'MM-dd HH:mm')}
                </span>
              </div>
              <p className="text-sm text-gray-600">{log.message}</p>
            </div>
          </div>
        ))}
      </div>
    );
  };

  if (!workflow) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">工作流不存在</h1>
          <p className="text-gray-600 mb-4">请检查工作流ID是否正确</p>
          <Button onClick={() => router.push('/workflows')}>
            返回工作流列表
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex justify-between items-center">
        <div className="flex items-center gap-4">
          <Button variant="ghost" onClick={() => router.back()}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回
          </Button>
          <div>
            <h1 className="text-2xl font-bold">{workflow.title}</h1>
            <p className="text-gray-600 mt-1">{workflow.description}</p>
          </div>
        </div>
        
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">
              <MoreHorizontal className="mr-2 h-4 w-4" />
              操作
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>
              <Edit2 className="mr-2 h-4 w-4" />
              编辑工作流
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Download className="mr-2 h-4 w-4" />
              导出报告
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Share2 className="mr-2 h-4 w-4" />
              分享链接
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-destructive">
              <Square className="mr-2 h-4 w-4" />
              取消工作流
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Status Bar */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-4">
              <Badge variant={getStatusColor(workflow.status)} className="text-sm">
                {getStatusLabel(workflow.status)}
              </Badge>
              <Badge variant={getPriorityColor(workflow.priority)} className="text-sm">
                优先级: {getPriorityLabel(workflow.priority)}
              </Badge>
              <Badge variant="outline" className="text-sm">
                {getTypeLabel(workflow.type)}
              </Badge>
            </div>
            
            <div className="text-sm text-gray-500">
              进度: {workflow.progress}%
            </div>
          </div>
          
          <Progress value={workflow.progress} className="mb-4" />
          
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-gray-500 mb-1">发起人</p>
              <p className="font-medium">{workflow.initiatorName}</p>
            </div>
            <div>
              <p className="text-gray-500 mb-1">创建时间</p>
              <p className="font-medium">
                {format(new Date(workflow.createdAt), 'yyyy-MM-dd HH:mm')}
              </p>
            </div>
            <div>
              <p className="text-gray-500 mb-1">最后更新</p>
              <p className="font-medium">
                {format(new Date(workflow.updatedAt), 'yyyy-MM-dd HH:mm')}
              </p>
            </div>
            {workflow.dueDate && (
              <div>
                <p className="text-gray-500 mb-1">截止时间</p>
                <p className="font-medium">
                  {format(new Date(workflow.dueDate), 'yyyy-MM-dd HH:mm')}
                </p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Workflow Steps */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                工作流步骤
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="text-gray-500">加载中...</div>
                </div>
              ) : (
                renderWorkflowSteps()
              )}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Workflow Data */}
          {workflow.data && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-5 w-5" />
                  申请详情
                </CardTitle>
              </CardHeader>
              <CardContent className="p-6">
                <div className="space-y-3 text-sm">
                  <div>
                    <p className="text-gray-500 mb-1">员工姓名</p>
                    <p className="font-medium">{workflow.data.employeeName}</p>
                  </div>
                  <div>
                    <p className="text-gray-500 mb-1">当前职位</p>
                    <p className="font-medium">{workflow.data.currentPosition}</p>
                  </div>
                  <div>
                    <p className="text-gray-500 mb-1">目标职位</p>
                    <p className="font-medium">{workflow.data.newPosition}</p>
                  </div>
                  <div>
                    <p className="text-gray-500 mb-1">薪资调整</p>
                    <p className="font-medium">
                      ¥{workflow.data.currentSalary?.toLocaleString()} → ¥{workflow.data.newSalary?.toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-500 mb-1">生效日期</p>
                    <p className="font-medium">{workflow.data.effectiveDate}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Activity Log */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5" />
                活动日志
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              {renderActivityLog()}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Comment Modal */}
      <Dialog open={isCommentModalVisible} onOpenChange={setIsCommentModalVisible}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>
              {selectedStep ? `处理步骤: ${selectedStep.name}` : '添加评论'}
            </DialogTitle>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">评论或原因</label>
              <Textarea 
                placeholder="请输入评论或处理原因..."
                rows={4}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setIsCommentModalVisible(false)}>
                取消
              </Button>
              {selectedStep && (
                <>
                  <Button
                    variant="destructive"
                    onClick={() => selectedStep && handleRejectStep(selectedStep.id)}
                    disabled={loading}
                  >
                    拒绝
                  </Button>
                  <Button 
                    onClick={() => selectedStep && handleApproveStep(selectedStep.id)}
                    disabled={loading}
                  >
                    通过
                  </Button>
                </>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default WorkflowDetailPage;