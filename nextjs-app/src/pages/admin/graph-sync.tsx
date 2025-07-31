import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { format } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { 
  RefreshCw,
  Database,
  Activity,
  CheckCircle,
  XCircle,
  AlertCircle,
  Play,
  Pause,
  RotateCcw,
  Settings,
  Download,
  Upload,
  FileText,
  Users,
  Building2,
  GitBranch,
  Clock,
  TrendingUp,
  Server,
  Zap,
  Shield
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

interface SyncJob {
  id: string;
  name: string;
  type: 'full_sync' | 'incremental_sync' | 'validation' | 'cleanup';
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  progress: number;
  startTime?: string;
  endTime?: string;
  duration?: number;
  recordsProcessed: number;
  recordsTotal: number;
  errorCount: number;
  warnings: string[];
  createdBy: string;
  lastRunAt?: string;
}

interface DataSource {
  id: string;
  name: string;
  type: 'database' | 'api' | 'file' | 'ldap';
  status: 'connected' | 'disconnected' | 'error';
  lastSync?: string;
  recordCount: number;
  healthStatus: 'healthy' | 'warning' | 'critical';
  connectionString: string;
  isActive: boolean;
}

interface SyncMetrics {
  totalRecords: number;
  syncedRecords: number;
  errorRecords: number;
  lastSyncTime: string;
  averageSyncTime: number;
  successRate: number;
  dailySync: number;
}

const GraphSyncPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [activeJobs, setActiveJobs] = useState<SyncJob[]>([]);
  const [jobHistory, setJobHistory] = useState<SyncJob[]>([]);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [metrics, setMetrics] = useState<SyncMetrics | null>(null);
  const [autoSyncEnabled, setAutoSyncEnabled] = useState(true);
  const [selectedJobType, setSelectedJobType] = useState<string>('full_sync');

  // Sample data
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleJobs: SyncJob[] = [
        {
          id: '1',
          name: '组织架构全量同步',
          type: 'full_sync',
          status: 'running',
          progress: 65,
          startTime: '2024-12-03T10:30:00Z',
          recordsProcessed: 2800,
          recordsTotal: 4300,
          errorCount: 12,
          warnings: ['部分员工信息缺少部门映射', '发现3个孤立节点'],
          createdBy: '系统管理员',
          lastRunAt: '2024-12-03T10:30:00Z'
        },
        {
          id: '2',
          name: '人员关系增量同步',
          type: 'incremental_sync',
          status: 'completed',
          progress: 100,
          startTime: '2024-12-03T09:15:00Z',
          endTime: '2024-12-03T09:45:00Z',
          duration: 30,
          recordsProcessed: 156,
          recordsTotal: 156,
          errorCount: 0,
          warnings: [],
          createdBy: 'HR系统',
          lastRunAt: '2024-12-03T09:15:00Z'
        }
      ];

      const sampleHistory: SyncJob[] = [
        {
          id: '3',
          name: '数据完整性验证',
          type: 'validation',
          status: 'completed',
          progress: 100,
          startTime: '2024-12-02T18:00:00Z',
          endTime: '2024-12-02T18:15:00Z',
          duration: 15,
          recordsProcessed: 4288,
          recordsTotal: 4288,
          errorCount: 8,
          warnings: ['发现8个数据不一致问题'],
          createdBy: '系统自动',
          lastRunAt: '2024-12-02T18:00:00Z'
        },
        {
          id: '4',
          name: '历史数据清理',
          type: 'cleanup',
          status: 'failed',
          progress: 45,
          startTime: '2024-12-01T22:00:00Z',
          endTime: '2024-12-01T22:30:00Z',
          duration: 30,
          recordsProcessed: 1200,
          recordsTotal: 2650,
          errorCount: 156,
          warnings: ['权限不足', '外键约束冲突'],
          createdBy: '夜间任务',
          lastRunAt: '2024-12-01T22:00:00Z'
        }
      ];

      const sampleDataSources: DataSource[] = [
        {
          id: '1',
          name: 'HR核心数据库',
          type: 'database',
          status: 'connected',
          lastSync: '2024-12-03T10:30:00Z',
          recordCount: 3456,
          healthStatus: 'healthy',
          connectionString: 'postgresql://hr_db:5432/employees',
          isActive: true
        },
        {
          id: '2',
          name: 'LDAP用户目录',
          type: 'ldap',
          status: 'connected',
          lastSync: '2024-12-03T09:45:00Z',
          recordCount: 4102,
          healthStatus: 'warning',
          connectionString: 'ldap://directory.company.com:389',
          isActive: true
        },
        {
          id: '3',
          name: 'OA系统接口',
          type: 'api',
          status: 'error',
          lastSync: '2024-12-02T14:20:00Z',
          recordCount: 0,
          healthStatus: 'critical',
          connectionString: 'https://oa.company.com/api/v1',
          isActive: false
        }
      ];

      const sampleMetrics: SyncMetrics = {
        totalRecords: 7558,
        syncedRecords: 7398,
        errorRecords: 160,
        lastSyncTime: '2024-12-03T10:30:00Z',
        averageSyncTime: 25,
        successRate: 97.8,
        dailySync: 3
      };

      setActiveJobs(sampleJobs);
      setJobHistory(sampleHistory);
      setDataSources(sampleDataSources);
      setMetrics(sampleMetrics);
      setLoading(false);
    }, 1000);
  }, []);

  const handleStartSync = async (jobType: string) => {
    try {
      setLoading(true);
      
      const newJob: SyncJob = {
        id: Date.now().toString(),
        name: getJobTypeName(jobType),
        type: jobType as SyncJob['type'],
        status: 'running',
        progress: 0,
        startTime: new Date().toISOString(),
        recordsProcessed: 0,
        recordsTotal: 1000,
        errorCount: 0,
        warnings: [],
        createdBy: '当前用户',
        lastRunAt: new Date().toISOString()
      };

      setActiveJobs(prev => [...prev, newJob]);
      toast.success(`开始执行${getJobTypeName(jobType)}`);
      
      // Simulate job progress
      let progress = 0;
      const interval = setInterval(() => {
        progress += Math.random() * 15;
        if (progress >= 100) {
          progress = 100;
          clearInterval(interval);
          
          setActiveJobs(prev => prev.map(job => 
            job.id === newJob.id 
              ? { 
                  ...job, 
                  status: 'completed' as const, 
                  progress: 100,
                  endTime: new Date().toISOString(),
                  recordsProcessed: job.recordsTotal
                }
              : job
          ));
          
          toast.success(`${getJobTypeName(jobType)}执行完成`);
        } else {
          setActiveJobs(prev => prev.map(job => 
            job.id === newJob.id 
              ? { 
                  ...job, 
                  progress,
                  recordsProcessed: Math.floor((progress / 100) * job.recordsTotal)
                }
              : job
          ));
        }
      }, 1000);
      
    } catch (error) {
      toast.error('启动同步任务失败');
    } finally {
      setLoading(false);
    }
  };

  const getJobTypeName = (type: string) => {
    const names = {
      full_sync: '全量同步',
      incremental_sync: '增量同步',
      validation: '数据验证',
      cleanup: '数据清理'
    };
    return names[type as keyof typeof names] || type;
  };

  const getJobTypeColor = (type: string): "default" | "destructive" | "secondary" => {
    const colors = {
      full_sync: 'default' as const,
      incremental_sync: 'default' as const,
      validation: 'secondary' as const,
      cleanup: 'destructive' as const
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getStatusColor = (status: string): "default" | "destructive" | "secondary" => {
    const colors = {
      pending: 'secondary' as const,
      running: 'default' as const,
      completed: 'default' as const,
      failed: 'destructive' as const,
      cancelled: 'secondary' as const
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusIcon = (status: string) => {
    const icons = {
      pending: <Clock className="h-4 w-4" />,
      running: <RefreshCw className="h-4 w-4 animate-spin" />,
      completed: <CheckCircle className="h-4 w-4" />,
      failed: <XCircle className="h-4 w-4" />,
      cancelled: <AlertCircle className="h-4 w-4" />
    };
    return icons[status as keyof typeof icons] || icons.pending;
  };

  const getHealthStatusColor = (status: string): string => {
    const colors = {
      healthy: 'text-green-600',
      warning: 'text-yellow-600',
      critical: 'text-red-600'
    };
    return colors[status as keyof typeof colors] || 'text-gray-600';
  };

  const renderActiveJobs = () => {
    if (activeJobs.length === 0) {
      return (
        <Alert>
          <AlertDescription>
            当前没有运行中的同步任务
          </AlertDescription>
        </Alert>
      );
    }

    return (
      <div className="space-y-4">
        {activeJobs.map((job) => (
          <Card key={job.id}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  {getStatusIcon(job.status)}
                  <div>
                    <h3 className="font-medium">{job.name}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant={getJobTypeColor(job.type)}>
                        {getJobTypeName(job.type)}
                      </Badge>
                      <Badge variant={getStatusColor(job.status)}>
                        {job.status === 'running' ? '运行中' : 
                         job.status === 'completed' ? '已完成' :
                         job.status === 'failed' ? '失败' : job.status}
                      </Badge>
                    </div>
                  </div>
                </div>
                
                <div className="text-right text-sm text-gray-500">
                  <div>进度: {job.progress.toFixed(1)}%</div>
                  <div>{job.recordsProcessed.toLocaleString()} / {job.recordsTotal.toLocaleString()}</div>
                </div>
              </div>
              
              <Progress value={job.progress} className="mb-3" />
              
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div>
                  <p className="text-gray-500">开始时间</p>
                  <p>{job.startTime ? format(new Date(job.startTime), 'HH:mm:ss') : '-'}</p>
                </div>
                <div>
                  <p className="text-gray-500">错误数量</p>
                  <p className={job.errorCount > 0 ? 'text-red-600' : 'text-green-600'}>
                    {job.errorCount}
                  </p>
                </div>
                <div>
                  <p className="text-gray-500">创建者</p>
                  <p>{job.createdBy}</p>
                </div>
              </div>
              
              {job.warnings.length > 0 && (
                <div className="mt-3">
                  <Alert>
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      <div>警告信息:</div>
                      <ul className="mt-1 text-sm">
                        {job.warnings.map((warning, index) => (
                          <li key={index}>• {warning}</li>
                        ))}
                      </ul>
                    </AlertDescription>
                  </Alert>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    );
  };

  const renderDataSources = () => {
    return (
      <div className="space-y-3">
        {dataSources.map((source) => (
          <div key={source.id} className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
            <div className={`w-3 h-3 rounded-full ${
              source.status === 'connected' ? 'bg-green-500' :
              source.status === 'disconnected' ? 'bg-yellow-500' : 'bg-red-500'
            }`} />
            
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-medium">{source.name}</span>
                <Badge variant="outline" className="text-xs">
                  {source.type.toUpperCase()}
                </Badge>
                <Badge variant={source.isActive ? 'default' : 'secondary'} className="text-xs">
                  {source.isActive ? '活跃' : '停用'}
                </Badge>
              </div>
              <div className="flex items-center gap-4 text-sm text-gray-600">
                <span>
                  状态: <span className={getHealthStatusColor(source.healthStatus)}>
                    {source.healthStatus === 'healthy' ? '健康' :
                     source.healthStatus === 'warning' ? '警告' : '严重'}
                  </span>
                </span>
                <span>记录数: {source.recordCount.toLocaleString()}</span>
                {source.lastSync && (
                  <span>
                    最后同步: {format(new Date(source.lastSync), 'MM-dd HH:mm')}
                  </span>
                )}
              </div>
            </div>
            
            <Button variant="outline" size="sm">
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>
    );
  };

  const jobTypes = [
    { value: 'full_sync', label: '全量同步', icon: <Database className="h-4 w-4" /> },
    { value: 'incremental_sync', label: '增量同步', icon: <RefreshCw className="h-4 w-4" /> },
    { value: 'validation', label: '数据验证', icon: <Shield className="h-4 w-4" /> },
    { value: 'cleanup', label: '数据清理', icon: <RotateCcw className="h-4 w-4" /> }
  ];

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">图数据同步管理</h1>
        <p className="text-gray-600 mt-1">
          管理组织架构和人员关系的数据同步任务 - 实时监控同步状态和数据健康度
        </p>
      </div>

      {/* Metrics Cards */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">总记录数</p>
                  <p className="text-2xl font-bold">{metrics.totalRecords.toLocaleString()}</p>
                </div>
                <Database className="h-8 w-8 text-blue-500" />
              </div>
            </CardContent>
          </Card>
          
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">同步成功率</p>
                  <p className="text-2xl font-bold text-green-600">{metrics.successRate}%</p>
                </div>
                <CheckCircle className="h-8 w-8 text-green-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">平均同步时间</p>
                  <p className="text-2xl font-bold text-purple-600">{metrics.averageSyncTime}分钟</p>
                </div>
                <Clock className="h-8 w-8 text-purple-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">今日同步次数</p>
                  <p className="text-2xl font-bold text-orange-600">{metrics.dailySync}</p>
                </div>
                <Activity className="h-8 w-8 text-orange-500" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Job Controls and Active Jobs */}
        <div className="lg:col-span-2 space-y-6">
          {/* Job Controls */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Play className="h-5 w-5" />
                启动同步任务
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">任务类型</label>
                  <Select value={selectedJobType} onValueChange={setSelectedJobType}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {jobTypes.map(type => (
                        <SelectItem key={type.value} value={type.value}>
                          <div className="flex items-center gap-2">
                            {type.icon}
                            {type.label}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                
                <div className="flex items-end">
                  <Button 
                    onClick={() => handleStartSync(selectedJobType)}
                    disabled={loading}
                    className="w-full"
                  >
                    <Play className="mr-2 h-4 w-4" />
                    开始同步
                  </Button>
                </div>
              </div>
              
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center gap-2">
                  <Zap className="h-4 w-4 text-blue-500" />
                  <span className="text-sm font-medium">自动同步</span>
                </div>
                <Button
                  variant={autoSyncEnabled ? "default" : "outline"}
                  size="sm"
                  onClick={() => setAutoSyncEnabled(!autoSyncEnabled)}
                >
                  {autoSyncEnabled ? '已启用' : '已停用'}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Active Jobs */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                运行中的任务
              </CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="text-gray-500">加载中...</div>
                </div>
              ) : (
                renderActiveJobs()
              )}
            </CardContent>
          </Card>
        </div>

        {/* Right Column - Data Sources and Settings */}
        <div className="space-y-6">
          {/* Data Sources */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Server className="h-5 w-5" />
                数据源状态
              </CardTitle>
            </CardHeader>
            <CardContent>
              {renderDataSources()}
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                快捷操作
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button variant="outline" className="w-full justify-start" size="sm">
                <Download className="mr-2 h-4 w-4" />
                导出同步日志
              </Button>
              <Button variant="outline" className="w-full justify-start" size="sm">
                <Upload className="mr-2 h-4 w-4" />
                导入配置文件
              </Button>
              <Button variant="outline" className="w-full justify-start" size="sm">
                <FileText className="mr-2 h-4 w-4" />
                查看系统日志
              </Button>
              <Button variant="outline" className="w-full justify-start" size="sm">
                <Shield className="mr-2 h-4 w-4" />
                数据完整性检查
              </Button>
            </CardContent>
          </Card>

          {/* Recent Activity */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5" />
                最近活动
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3 text-sm">
                {jobHistory.slice(0, 3).map((job) => (
                  <div key={job.id} className="flex items-start gap-3 p-2 bg-gray-50 rounded">
                    {getStatusIcon(job.status)}
                    <div className="flex-1">
                      <div className="font-medium">{job.name}</div>
                      <div className="text-gray-500">
                        {job.endTime ? format(new Date(job.endTime), 'MM-dd HH:mm') : '进行中'}
                      </div>
                    </div>
                    <Badge variant={getStatusColor(job.status)} className="text-xs">
                      {job.status === 'completed' ? '成功' :
                       job.status === 'failed' ? '失败' : job.status}
                    </Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default GraphSyncPage;