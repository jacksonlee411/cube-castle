// src/pages/workflows/[id].tsx
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { 
  Card, 
  Steps, 
  Timeline, 
  Typography, 
  Space, 
  Tag, 
  Button, 
  Progress,
  Alert,
  Descriptions,
  Avatar,
  Tooltip,
  Modal,
  Form,
  Input,
  Radio
} from 'antd';
import {
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  UserOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;
const { TextArea } = Input;

interface WorkflowStatus {
  workflowId: string;
  employeeId: string;
  employeeName: string;
  workflowType: string;
  status: 'PENDING' | 'IN_PROGRESS' | 'APPROVED' | 'REJECTED' | 'COMPLETED' | 'FAILED';
  currentStep: string;
  progress: number;
  startedAt: string;
  updatedAt: string;
  completedAt?: string;
  error?: string;
  positionChange?: PositionChangeDetails;
  approvalSteps?: ApprovalStep[];
  activities?: WorkflowActivity[];
}

interface PositionChangeDetails {
  currentPosition: {
    title: string;
    department: string;
    jobLevel?: string;
  };
  newPosition: {
    title: string;
    department: string;
    jobLevel?: string;
    minSalary?: number;
    maxSalary?: number;
    currency?: string;
  };
  effectiveDate: string;
  changeReason: string;
  isRetroactive: boolean;
}

interface ApprovalStep {
  stepId: string;
  approverName: string;
  approverTitle: string;
  stepType: string;
  isRequired: boolean;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'SKIPPED';
  comments?: string;
  approvedAt?: string;
  timeout: string;
}

interface WorkflowActivity {
  id: string;
  activityType: string;
  status: 'COMPLETED' | 'FAILED' | 'IN_PROGRESS';
  startedAt: string;
  completedAt?: string;
  error?: string;
  result?: any;
}

const WorkflowStatusPage: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  
  const [loading, setLoading] = useState(false);
  const [workflowStatus, setWorkflowStatus] = useState<WorkflowStatus | null>(null);
  const [isApprovalModalVisible, setIsApprovalModalVisible] = useState(false);
  const [approvalForm] = Form.useForm();

  // Sample data - in real app would come from GraphQL subscription
  useEffect(() => {
    if (!id) return;
    
    setLoading(true);
    
    // Simulate WebSocket subscription
    const mockWorkflowStatus: WorkflowStatus = {
        workflowId: id as string,
        employeeId: 'EMP001',
        employeeName: '张三',
        workflowType: 'POSITION_CHANGE',
        status: 'IN_PROGRESS',
        currentStep: 'hr-manager-approval',
        progress: 60,
        startedAt: '2024-07-28T10:30:00Z',
        updatedAt: '2024-07-28T11:15:00Z',
        positionChange: {
          currentPosition: {
            title: '软件工程师',
            department: '技术部',
            jobLevel: 'INTERMEDIATE'
          },
          newPosition: {
            title: '高级软件工程师',
            department: '技术部',
            jobLevel: 'SENIOR',
            minSalary: 25000,
            maxSalary: 35000,
            currency: 'CNY'
          },
          effectiveDate: '2024-08-01T00:00:00Z',
          changeReason: '年度晋升',
          isRetroactive: false
        },
        approvalSteps: [
          {
            stepId: 'direct-manager',
            approverName: '李四',
            approverTitle: '技术经理',
            stepType: 'MANAGER',
            isRequired: true,
            status: 'APPROVED',
            comments: '表现优秀，同意晋升',
            approvedAt: '2024-07-28T10:45:00Z',
            timeout: '24h'
          },
          {
            stepId: 'hr-manager',
            approverName: '王五',
            approverTitle: 'HR经理',
            stepType: 'HR',
            isRequired: true,
            status: 'PENDING',
            timeout: '48h'
          },
          {
            stepId: 'hr-director',
            approverName: '赵六',
            approverTitle: 'HR总监',
            stepType: 'HR',
            isRequired: true,
            status: 'PENDING',
            timeout: '72h'
          }
        ],
        activities: [
          {
            id: '1',
            activityType: 'ValidateTemporalConsistency',
            status: 'COMPLETED',
            startedAt: '2024-07-28T10:30:15Z',
            completedAt: '2024-07-28T10:30:18Z',
            result: { isValid: true }
          },
          {
            id: '2',
            activityType: 'AssessPositionChangeRisk',
            status: 'COMPLETED',
            startedAt: '2024-07-28T10:30:20Z',
            completedAt: '2024-07-28T10:30:25Z',
            result: { riskLevel: 'MEDIUM', requiresApproval: true }
          },
          {
            id: '3',
            activityType: 'StartApprovalWorkflow',
            status: 'IN_PROGRESS',
            startedAt: '2024-07-28T10:30:30Z'
          }
        ]
      };

      setWorkflowStatus(mockWorkflowStatus);
      setLoading(false);

      // Simulate real-time updates
      const interval = setInterval(() => {
        setWorkflowStatus(prev => {
          if (!prev) return prev;
          return {
            ...prev,
            updatedAt: new Date().toISOString()
          };
        });
      }, 5000);

      return () => clearInterval(interval);
    }
  }, [id]);

  const getStatusColor = (status: string) => {
    const colors = {
      PENDING: 'orange',
      IN_PROGRESS: 'blue',
      APPROVED: 'green',
      REJECTED: 'red',
      COMPLETED: 'green',
      FAILED: 'red',
      SKIPPED: 'gray'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusIcon = (status: string) => {
    const icons = {
      PENDING: <ClockCircleOutlined />,
      IN_PROGRESS: <PlayCircleOutlined />,
      APPROVED: <CheckCircleOutlined />,
      REJECTED: <CloseCircleOutlined />,
      COMPLETED: <CheckCircleOutlined />,
      FAILED: <CloseCircleOutlined />,
      SKIPPED: <PauseCircleOutlined />
    };
    return icons[status as keyof typeof icons] || <ClockCircleOutlined />;
  };

  const getCurrentStepIndex = () => {
    if (!workflowStatus?.approvalSteps) return 0;
    return workflowStatus.approvalSteps.findIndex(step => step.status === 'PENDING');
  };

  const handleApproval = async (values: any) => {
    try {
      setLoading(true);
      
      // Simulate GraphQL mutation
      // Approval decision processed - values submitted to backend
      
      // Update workflow status
      setWorkflowStatus(prev => {
        if (!prev) return prev;
        
        const updatedSteps = prev.approvalSteps?.map(step => {
          if (step.status === 'PENDING') {
            return {
              ...step,
              status: values.decision,
              comments: values.comments,
              approvedAt: new Date().toISOString()
            };
          }
          return step;
        });

        return {
          ...prev,
          approvalSteps: updatedSteps,
          updatedAt: new Date().toISOString(),
          progress: values.decision === 'APPROVED' ? prev.progress + 20 : prev.progress
        };
      });
      
      setIsApprovalModalVisible(false);
      approvalForm.resetFields();
    } catch (error) {
      // Approval failed - error handled by notification
    } finally {
      setLoading(false);
    }
  };

  if (loading && !workflowStatus) {
    return <div style={{ padding: '24px' }}>加载中...</div>;
  }

  if (!workflowStatus) {
    return <div style={{ padding: '24px' }}>工作流未找到</div>;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>工作流状态监控</Title>
        <Space>
          <Text strong>工作流ID:</Text>
          <Text code>{workflowStatus.workflowId}</Text>
          <Tag color={getStatusColor(workflowStatus.status)} icon={getStatusIcon(workflowStatus.status)}>
            {workflowStatus.status}
          </Tag>
        </Space>
      </div>

      {/* Progress Overview */}
      <Card style={{ marginBottom: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <Title level={4} style={{ margin: 0 }}>总体进度</Title>
          <Text>{workflowStatus.progress}% 完成</Text>
        </div>
        <Progress 
          percent={workflowStatus.progress} 
          status={workflowStatus.status === 'FAILED' ? 'exception' : 'active'}
          strokeColor={{
            from: '#108ee9',
            to: '#87d068',
          }}
        />
        <div style={{ marginTop: '16px', display: 'flex', justifyContent: 'space-between' }}>
          <Text type="secondary">
            开始时间: {dayjs(workflowStatus.startedAt).format('YYYY-MM-DD HH:mm:ss')}
          </Text>
          <Text type="secondary">
            最后更新: {dayjs(workflowStatus.updatedAt).format('YYYY-MM-DD HH:mm:ss')}
          </Text>
        </div>
      </Card>

      {/* Position Change Details */}
      {workflowStatus.positionChange && (
        <Card title="职位变更详情" style={{ marginBottom: '24px' }}>
          <Descriptions column={2} bordered>
            <Descriptions.Item label="员工">
              <Space>
                <Avatar icon={<UserOutlined />} />
                <span>{workflowStatus.employeeName}</span>
                <Text type="secondary">({workflowStatus.employeeId})</Text>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="生效日期">
              {dayjs(workflowStatus.positionChange.effectiveDate).format('YYYY年MM月DD日')}
              {workflowStatus.positionChange.isRetroactive && 
                <Tag color="orange" style={{ marginLeft: '8px' }}>追溯</Tag>
              }
            </Descriptions.Item>
            <Descriptions.Item label="当前职位">
              <div>
                <div><strong>{workflowStatus.positionChange.currentPosition.title}</strong></div>
                <Text type="secondary">
                  {workflowStatus.positionChange.currentPosition.department} • 
                  {workflowStatus.positionChange.currentPosition.jobLevel}
                </Text>
              </div>
            </Descriptions.Item>
            <Descriptions.Item label="新职位">
              <div>
                <div><strong>{workflowStatus.positionChange.newPosition.title}</strong></div>
                <Text type="secondary">
                  {workflowStatus.positionChange.newPosition.department} • 
                  {workflowStatus.positionChange.newPosition.jobLevel}
                </Text>
                {workflowStatus.positionChange.newPosition.minSalary && (
                  <div style={{ marginTop: '4px' }}>
                    <Text type="secondary">
                      薪资: {workflowStatus.positionChange.newPosition.minSalary?.toLocaleString()} - 
                      {workflowStatus.positionChange.newPosition.maxSalary?.toLocaleString()} 
                      {workflowStatus.positionChange.newPosition.currency}
                    </Text>
                  </div>
                )}
              </div>
            </Descriptions.Item>
            <Descriptions.Item label="变更原因" span={2}>
              {workflowStatus.positionChange.changeReason}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
        {/* Approval Steps */}
        {workflowStatus.approvalSteps && (
          <Card 
            title="审批流程" 
            extra={
              workflowStatus.approvalSteps.some(step => step.status === 'PENDING') && (
                <Button 
                  type="primary" 
                  onClick={() => setIsApprovalModalVisible(true)}
                >
                  处理审批
                </Button>
              )
            }
          >
            <Steps
              direction="vertical"
              current={getCurrentStepIndex()}
              status={workflowStatus.status === 'FAILED' ? 'error' : 'process'}
            >
              {workflowStatus.approvalSteps.map((step, index) => (
                <Step
                  key={step.stepId}
                  title={step.approverName}
                  description={
                    <div>
                      <div>{step.approverTitle} • {step.stepType}</div>
                      {step.status === 'APPROVED' && step.comments && (
                        <div style={{ marginTop: '4px', color: '#52c41a' }}>
                          <Text type="success">{step.comments}</Text>
                        </div>
                      )}
                      {step.status === 'REJECTED' && step.comments && (
                        <div style={{ marginTop: '4px', color: '#ff4d4f' }}>
                          <Text type="danger">{step.comments}</Text>
                        </div>
                      )}
                      {step.status === 'PENDING' && (
                        <div style={{ marginTop: '4px' }}>
                          <Text type="secondary">超时时间: {step.timeout}</Text>
                        </div>
                      )}
                      {step.approvedAt && (
                        <div style={{ marginTop: '4px' }}>
                          <Text type="secondary">
                            {dayjs(step.approvedAt).format('MM-DD HH:mm')}
                          </Text>
                        </div>
                      )}
                    </div>
                  }
                  status={
                    step.status === 'APPROVED' ? 'finish' :
                    step.status === 'REJECTED' ? 'error' :
                    step.status === 'PENDING' ? 'process' : 'wait'
                  }
                  icon={getStatusIcon(step.status)}
                />
              ))}
            </Steps>
          </Card>
        )}

        {/* Activity Timeline */}
        <Card title="活动日志">
          <Timeline>
            {workflowStatus.activities?.map((activity) => (
              <Timeline.Item
                key={activity.id}
                color={getStatusColor(activity.status)}
                dot={getStatusIcon(activity.status)}
              >
                <div>
                  <div style={{ fontWeight: 'bold' }}>{activity.activityType}</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    开始: {dayjs(activity.startedAt).format('HH:mm:ss')}
                    {activity.completedAt && (
                      <span> • 完成: {dayjs(activity.completedAt).format('HH:mm:ss')}</span>
                    )}
                  </div>
                  {activity.error && (
                    <Alert 
                      message={activity.error} 
                      type="error" 
                      style={{ marginTop: '8px', fontSize: '12px' }}
                    />
                  )}
                  {activity.result && (
                    <div style={{ marginTop: '8px', fontSize: '12px' }}>
                      <Text type="secondary">结果: {JSON.stringify(activity.result)}</Text>
                    </div>
                  )}
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        </Card>
      </div>

      {/* Approval Modal */}
      <Modal
        title="处理审批"
        open={isApprovalModalVisible}
        onCancel={() => {
          setIsApprovalModalVisible(false);
          approvalForm.resetFields();
        }}
        onOk={() => approvalForm.submit()}
        confirmLoading={loading}
        width={500}
      >
        <Form
          form={approvalForm}
          layout="vertical"
          onFinish={handleApproval}
          initialValues={{ decision: 'APPROVED' }}
        >
          <Form.Item
            name="decision"
            label="审批决定"
            rules={[{ required: true, message: '请选择审批决定' }]}
          >
            <Radio.Group>
              <Radio value="APPROVED">同意</Radio>
              <Radio value="REJECTED">拒绝</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            name="comments"
            label="审批意见"
            rules={[{ required: true, message: '请输入审批意见' }]}
          >
            <TextArea
              rows={4}
              placeholder="请输入审批意见..."
              showCount
              maxLength={500}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default WorkflowStatusPage;