// src/pages/workflows/demo.tsx
import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Button, 
  Steps, 
  Timeline, 
  Select, 
  Form, 
  Input, 
  InputNumber,
  DatePicker,
  message,
  Spin,
  Tag,
  Space,
  Row,
  Col,
  Divider,
  Alert,
  Modal
} from 'antd';
import { 
  PlayCircleOutlined, 
  PauseCircleOutlined, 
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
  AuditOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import dayjs from 'dayjs';
import withGraphQLErrorBoundary from '@/components/withGraphQLErrorBoundary';
import ServiceStatus from '@/components/ServiceStatus';

// Force dynamic rendering for this page
export const getServerSideProps = async () => {
  return { props: {} };
};

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Step } = Steps;
const { Item: TimelineItem } = Timeline;

interface WorkflowDemo {
  id: string;
  type: string;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';
  progress: number;
  currentStep: string;
  employee: {
    id: string;
    name: string;
    current_position: string;
  };
  changes: {
    position_title: string;
    department: string;
    salary_change: number;
    effective_date: string;
  };
  approvals: Array<{
    level: number;
    approver: string;
    status: 'PENDING' | 'APPROVED' | 'REJECTED';
    timestamp?: string;
    comments?: string;
  }>;
  created_at: string;
  estimated_completion: string;
}

const WorkflowDemoPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState<boolean>(false);
  const [workflows, setWorkflows] = useState<WorkflowDemo[]>([]);
  const [selectedWorkflow, setSelectedWorkflow] = useState<WorkflowDemo | null>(null);
  const [createModalVisible, setCreateModalVisible] = useState<boolean>(false);
  const [form] = Form.useForm();

  // 模拟工作流数据
  const demoWorkflows: WorkflowDemo[] = [
    {
      id: 'wf-001',
      type: 'POSITION_CHANGE',
      status: 'RUNNING',
      progress: 60,
      currentStep: 'MANAGER_APPROVAL',
      employee: {
        id: 'emp-001',
        name: '张三',
        current_position: '高级开发工程师'
      },
      changes: {
        position_title: '技术主管',
        department: '研发部',
        salary_change: 5000,
        effective_date: '2025-02-01'
      },
      approvals: [
        {
          level: 1,
          approver: '李经理',
          status: 'APPROVED',
          timestamp: '2025-01-25 10:30:00',
          comments: '员工表现优秀，同意晋升'
        },
        {
          level: 2,
          approver: '王总监',
          status: 'PENDING'
        }
      ],
      created_at: '2025-01-25 09:00:00',
      estimated_completion: '2025-01-28 18:00:00'
    },
    {
      id: 'wf-002',
      type: 'BULK_TRANSFER',
      status: 'COMPLETED',
      progress: 100,
      currentStep: 'COMPLETED',
      employee: {
        id: 'bulk-001',
        name: '批量调动 (5人)',
        current_position: '多个职位'
      },
      changes: {
        position_title: '各自新职位',
        department: '新成立部门',
        salary_change: 0,
        effective_date: '2025-01-20'
      },
      approvals: [
        {
          level: 1,
          approver: '人事总监',
          status: 'APPROVED',
          timestamp: '2025-01-18 14:20:00',
          comments: '组织架构调整，批准执行'
        }
      ],
      created_at: '2025-01-18 11:00:00',
      estimated_completion: '2025-01-20 17:00:00'
    },
    {
      id: 'wf-003',
      type: 'POSITION_CHANGE',
      status: 'FAILED',
      progress: 30,
      currentStep: 'VALIDATION_FAILED',
      employee: {
        id: 'emp-003',
        name: '赵六',
        current_position: '产品经理'
      },
      changes: {
        position_title: '高级产品经理',
        department: '产品部',
        salary_change: 3000,
        effective_date: '2025-01-15'
      },
      approvals: [],
      created_at: '2025-01-15 16:00:00',
      estimated_completion: '2025-01-18 18:00:00'
    }
  ];

  useEffect(() => {
    setWorkflows(demoWorkflows);
  }, []);

  const getStatusColor = (status: WorkflowDemo['status']) => {
    switch (status) {
      case 'PENDING': return 'default';
      case 'RUNNING': return 'processing';
      case 'COMPLETED': return 'success';
      case 'FAILED': return 'error';
      case 'CANCELLED': return 'warning';
      default: return 'default';
    }
  };

  const getStatusIcon = (status: WorkflowDemo['status']) => {
    switch (status) {
      case 'PENDING': return <ClockCircleOutlined />;
      case 'RUNNING': return <PlayCircleOutlined />;
      case 'COMPLETED': return <CheckCircleOutlined />;
      case 'FAILED': return <CloseCircleOutlined />;
      case 'CANCELLED': return <PauseCircleOutlined />;
      default: return <ClockCircleOutlined />;
    }
  };

  const getWorkflowSteps = (workflow: WorkflowDemo) => {
    const steps = [
      { title: '创建申请', status: 'finish' },
      { title: '数据验证', status: workflow.progress >= 20 ? 'finish' : 'wait' },
      { title: '经理审批', status: workflow.progress >= 40 ? 'finish' : workflow.progress >= 20 ? 'process' : 'wait' },
      { title: '总监审批', status: workflow.progress >= 60 ? 'finish' : workflow.progress >= 40 ? 'process' : 'wait' },
      { title: 'HR确认', status: workflow.progress >= 80 ? 'finish' : workflow.progress >= 60 ? 'process' : 'wait' },
      { title: '执行变更', status: workflow.progress >= 100 ? 'finish' : workflow.progress >= 80 ? 'process' : 'wait' }
    ];

    if (workflow.status === 'FAILED') {
      const currentIndex = Math.floor(workflow.progress / 20);
      if (currentIndex < steps.length) {
        steps[currentIndex].status = 'error';
      }
    }

    return steps;
  };

  const handleCreateWorkflow = () => {
    setCreateModalVisible(true);
  };

  const handleModalSubmit = async (values: any) => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      const newWorkflow: WorkflowDemo = {
        id: `wf-${Date.now()}`,
        type: 'POSITION_CHANGE',
        status: 'PENDING',
        progress: 0,
        currentStep: 'CREATED',
        employee: {
          id: values.employee_id,
          name: values.employee_name,
          current_position: values.current_position
        },
        changes: {
          position_title: values.new_position,
          department: values.new_department,
          salary_change: values.salary_change || 0,
          effective_date: values.effective_date.format('YYYY-MM-DD')
        },
        approvals: [],
        created_at: dayjs().format('YYYY-MM-DD HH:mm:ss'),
        estimated_completion: dayjs().add(3, 'days').format('YYYY-MM-DD HH:mm:ss')
      };

      setWorkflows(prev => [newWorkflow, ...prev]);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('工作流创建成功！');
    } catch (error) {
      message.error('创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = (workflowId: string) => {
    router.push(`/workflows/${workflowId}`);
  };

  const handleCancelWorkflow = async (workflowId: string) => {
    Modal.confirm({
      title: '确认取消工作流？',
      icon: <ExclamationCircleOutlined />,
      content: '取消后的工作流无法恢复，确定要取消吗？',
      onOk: async () => {
        setLoading(true);
        try {
          await new Promise(resolve => setTimeout(resolve, 1000));
          setWorkflows(prev => 
            prev.map(wf => 
              wf.id === workflowId 
                ? { ...wf, status: 'CANCELLED' as const, progress: 0 }
                : wf
            )
          );
          message.success('工作流已取消');
        } catch (error) {
          message.error('取消失败');
        } finally {
          setLoading(false);
        }
      }
    });
  };

  return (
    <div style={{ padding: '24px', maxWidth: '1400px', margin: '0 auto' }}>
      {/* Header */}
      <Card style={{ marginBottom: 24 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Title level={2} style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 12 }}>
              <AuditOutlined />
              工作流管理演示
            </Title>
            <Text type="secondary">
              Temporal.io 驱动的业务流程自动化演示，支持职位变更、批量操作等复杂工作流
            </Text>
          </Col>
          <Col>
            <Space>
              <ServiceStatus showDetails={false} />
              <Button 
                type="primary" 
                icon={<PlayCircleOutlined />}
                onClick={handleCreateWorkflow}
                size="large"
              >
                创建新工作流
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#1890ff', marginBottom: 8 }}>
                {workflows.filter(w => w.status === 'RUNNING').length}
              </div>
              <Text type="secondary">运行中</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#52c41a', marginBottom: 8 }}>
                {workflows.filter(w => w.status === 'COMPLETED').length}
              </div>
              <Text type="secondary">已完成</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#faad14', marginBottom: 8 }}>
                {workflows.filter(w => w.status === 'PENDING').length}
              </div>
              <Text type="secondary">等待中</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#f5222d', marginBottom: 8 }}>
                {workflows.filter(w => w.status === 'FAILED').length}
              </div>
              <Text type="secondary">失败</Text>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Workflow List */}
      <Row gutter={16}>
        {workflows.map(workflow => (
          <Col span={12} key={workflow.id} style={{ marginBottom: 16 }}>
            <Card
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  {getStatusIcon(workflow.status)}
                  <span>{workflow.employee.name}</span>
                  <Tag color={getStatusColor(workflow.status)}>
                    {workflow.status}
                  </Tag>
                </div>
              }
              extra={
                <Space>
                  <Button 
                    size="small" 
                    onClick={() => handleViewDetails(workflow.id)}
                  >
                    详情
                  </Button>
                  {workflow.status === 'RUNNING' && (
                    <Button 
                      size="small" 
                      danger 
                      onClick={() => handleCancelWorkflow(workflow.id)}
                    >
                      取消
                    </Button>
                  )}
                </Space>
              }
            >
              {/* Employee Info */}
              <div style={{ marginBottom: 16 }}>
                <Row gutter={16}>
                  <Col span={12}>
                    <Text strong>当前职位:</Text>
                    <br />
                    <Text>{workflow.employee.current_position}</Text>
                  </Col>
                  <Col span={12}>
                    <Text strong>目标职位:</Text>
                    <br />
                    <Text>{workflow.changes.position_title}</Text>
                  </Col>
                </Row>
              </div>

              <Divider />

              {/* Progress */}
              <div style={{ marginBottom: 16 }}>
                <Text strong>工作流进度 ({workflow.progress}%)</Text>
                <Steps 
                  current={Math.floor(workflow.progress / 20)} 
                  size="small"
                  style={{ marginTop: 8 }}
                  status={workflow.status === 'FAILED' ? 'error' : 'process'}
                >
                  {getWorkflowSteps(workflow).map((step, index) => (
                    <Step 
                      key={index} 
                      title={step.title}
                      status={step.status as any}
                    />
                  ))}
                </Steps>
              </div>

              {/* Timeline */}
              {workflow.approvals.length > 0 && (
                <div>
                  <Text strong>审批进度:</Text>
                  <Timeline style={{ marginTop: 8 }}>
                    {workflow.approvals.map((approval, index) => (
                      <TimelineItem
                        key={index}
                        dot={
                          approval.status === 'APPROVED' ? 
                            <CheckCircleOutlined style={{ color: '#52c41a' }} /> :
                          approval.status === 'REJECTED' ? 
                            <CloseCircleOutlined style={{ color: '#f5222d' }} /> :
                            <ClockCircleOutlined style={{ color: '#faad14' }} />
                        }
                      >
                        <div>
                          <Text strong>{approval.approver}</Text>
                          <Tag 
                            color={
                              approval.status === 'APPROVED' ? 'green' :
                              approval.status === 'REJECTED' ? 'red' : 'orange'
                            }
                            style={{ marginLeft: 8 }}
                          >
                            {approval.status}
                          </Tag>
                          {approval.timestamp && (
                            <div>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {approval.timestamp}
                              </Text>
                            </div>
                          )}
                          {approval.comments && (
                            <div>
                              <Text style={{ fontSize: '12px' }}>
                                {approval.comments}
                              </Text>
                            </div>
                          )}
                        </div>
                      </TimelineItem>
                    ))}
                  </Timeline>
                </div>
              )}

              {/* Error Alert */}
              {workflow.status === 'FAILED' && (
                <Alert
                  type="error"
                  message="工作流执行失败"
                  description="数据验证失败：薪资变更幅度超过公司政策限制"
                  style={{ marginTop: 16 }}
                />
              )}
            </Card>
          </Col>
        ))}
      </Row>

      {/* Create Workflow Modal */}
      <Modal
        title="创建新的职位变更工作流"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleModalSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="员工ID"
                name="employee_id"
                rules={[{ required: true, message: '请输入员工ID' }]}
              >
                <Input placeholder="例如: EMP001" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="员工姓名"
                name="employee_name"
                rules={[{ required: true, message: '请输入员工姓名' }]}
              >
                <Input placeholder="例如: 张三" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="当前职位"
            name="current_position"
            rules={[{ required: true, message: '请输入当前职位' }]}
          >
            <Input placeholder="例如: 高级开发工程师" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="新职位"
                name="new_position"
                rules={[{ required: true, message: '请输入新职位' }]}
              >
                <Input placeholder="例如: 技术主管" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="新部门"
                name="new_department"
                rules={[{ required: true, message: '请选择新部门' }]}
              >
                <Select placeholder="选择部门">
                  <Option value="研发部">研发部</Option>
                  <Option value="产品部">产品部</Option>
                  <Option value="市场部">市场部</Option>
                  <Option value="人事部">人事部</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="薪资调整 (元)"
                name="salary_change"
              >
                <InputNumber 
                  style={{ width: '100%' }} 
                  placeholder="例如: 5000"
                  min={-10000}
                  max={20000}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="生效日期"
                name="effective_date"
                rules={[{ required: true, message: '请选择生效日期' }]}
              >
                <DatePicker 
                  style={{ width: '100%' }}
                  disabledDate={(current) => current && current < dayjs().endOf('day')}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建工作流
              </Button>
              <Button onClick={() => setCreateModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default withGraphQLErrorBoundary(WorkflowDemoPage);