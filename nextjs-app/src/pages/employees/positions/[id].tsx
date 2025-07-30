// src/pages/employees/positions/[id].tsx - Employee Position History without GraphQL
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { Card, Typography, Timeline, Button, Modal, Form, DatePicker, Input, Select, notification, Spin, Tabs, Space } from 'antd';
import { PlusOutlined, EditOutlined, HistoryOutlined, UserOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string;
  email: string;
  status: string;
  hireDate: string;
  department?: string;
  position?: string;
}

interface PositionHistory {
  id: string;
  positionTitle: string;
  department: string;
  jobLevel?: string;
  location?: string;
  employmentType: string;
  reportsToEmployeeId?: string;
  effectiveDate: string;
  endDate?: string;
  changeReason?: string;
  isRetroactive: boolean;
  minSalary?: number;
  maxSalary?: number;
  currency?: string;
}

interface WorkflowHistory {
  workflowId: string;
  startedAt: string;
  status: string;
  type: string;
}

const EmployeePositionHistoryPage: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  
  const [loading, setLoading] = useState(false);
  const [employee, setEmployee] = useState<Employee | null>(null);
  const [positionHistory, setPositionHistory] = useState<PositionHistory[]>([]);
  const [workflowHistory, setWorkflowHistory] = useState<WorkflowHistory[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [form] = Form.useForm();

  // Load employee data and position history
  useEffect(() => {
    if (id) {
      loadEmployeeData(id as string);
    }
  }, [id]);

  const loadEmployeeData = async (employeeId: string) => {
    setLoading(true);
    try {
      // Simulate API call with sample data
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Sample employee data
      const sampleEmployees = [
        {
          id: '1',
          employeeId: 'EMP001',
          legalName: '张三',
          preferredName: 'Zhang San',
          email: 'zhangsan@company.com',
          status: 'ACTIVE',
          hireDate: '2023-01-15',
          department: '技术部',
          position: '高级软件工程师'
        },
        {
          id: '2',
          employeeId: 'EMP002',
          legalName: '王五',
          email: 'wangwu@company.com',
          status: 'ACTIVE',
          hireDate: '2023-03-20',
          department: '产品部',
          position: '产品经理'
        },
        {
          id: '3',
          employeeId: 'EMP003',
          legalName: '刘七',
          email: 'liuqi@company.com',
          status: 'INACTIVE',
          hireDate: '2022-08-10',
          department: '技术部',
          position: '前端工程师'
        },
        {
          id: '4',
          employeeId: 'EMP004',
          legalName: '陈八',
          email: 'chenba@company.com',
          status: 'ACTIVE',
          hireDate: '2024-01-08',
          department: '人事部',
          position: 'HR专员'
        }
      ];

      const foundEmployee = sampleEmployees.find(emp => emp.id === employeeId);
      
      if (foundEmployee) {
        setEmployee(foundEmployee);
        
        // Sample position history data for this employee
        const samplePositionHistory: PositionHistory[] = [
          {
            id: '1',
            positionTitle: foundEmployee.position || '当前职位',
            department: foundEmployee.department || '未分配',
            jobLevel: 'SENIOR',
            location: '北京',
            employmentType: 'FULL_TIME',
            effectiveDate: foundEmployee.hireDate,
            changeReason: '入职',
            isRetroactive: false,
            minSalary: 15000,
            maxSalary: 25000,
            currency: 'CNY'
          }
        ];

        // Add some historical positions for demonstration
        if (employeeId === '1') {
          samplePositionHistory.unshift({
            id: '2',
            positionTitle: '中级软件工程师',
            department: '技术部',
            jobLevel: 'INTERMEDIATE',
            location: '北京',
            employmentType: 'FULL_TIME',
            effectiveDate: '2022-06-01',
            endDate: '2023-01-14',
            changeReason: '晋升',
            isRetroactive: false,
            minSalary: 12000,
            maxSalary: 18000,
            currency: 'CNY'
          });
        }

        setPositionHistory(samplePositionHistory);

        // Sample workflow history
        const sampleWorkflowHistory: WorkflowHistory[] = [
          {
            workflowId: 'WF-001',
            startedAt: foundEmployee.hireDate,
            status: 'COMPLETED',
            type: '入职流程'
          }
        ];

        setWorkflowHistory(sampleWorkflowHistory);
      }
    } catch (error) {
      notification.error({
        message: '加载失败',
        description: '无法加载员工信息，请重试。',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePositionChange = async (values: any) => {
    if (!employee) return;

    setCreateLoading(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 2000));

      const newPositionChange: PositionHistory = {
        id: Date.now().toString(),
        positionTitle: values.positionTitle,
        department: values.department,
        jobLevel: values.jobLevel,
        location: values.location,
        employmentType: values.employmentType,
        effectiveDate: values.effectiveDate.format('YYYY-MM-DD'),
        changeReason: values.changeReason,
        isRetroactive: dayjs(values.effectiveDate).isBefore(dayjs()),
        minSalary: values.minSalary ? parseFloat(values.minSalary) : undefined,
        maxSalary: values.maxSalary ? parseFloat(values.maxSalary) : undefined,
        currency: values.currency || 'CNY'
      };

      // End the current position if the new one is effective now or in the past
      const isCurrentOrPast = dayjs(values.effectiveDate).isSame(dayjs(), 'day') || dayjs(values.effectiveDate).isBefore(dayjs());
      
      if (isCurrentOrPast) {
        setPositionHistory(prev => prev.map(pos => 
          !pos.endDate ? { ...pos, endDate: dayjs(values.effectiveDate).subtract(1, 'day').format('YYYY-MM-DD') } : pos
        ));
      }

      // Add new position to history
      setPositionHistory(prev => [...prev, newPositionChange]);

      // Add workflow record
      const newWorkflow: WorkflowHistory = {
        workflowId: `WF-${Date.now()}`,
        startedAt: new Date().toISOString(),
        status: isCurrentOrPast ? 'COMPLETED' : 'PENDING',
        type: '职位变更'
      };

      setWorkflowHistory(prev => [...prev, newWorkflow]);

      notification.success({
        message: '职位变更已提交',
        description: isCurrentOrPast ? '职位变更已生效。' : '职位变更已提交，将在生效日期执行。',
      });

      setIsModalVisible(false);
      form.resetFields();
    } catch (error) {
      notification.error({
        message: '职位变更失败',
        description: '创建职位变更时发生错误，请重试。',
      });
    } finally {
      setCreateLoading(false);
    }
  };

  const renderPositionTimeline = () => {
    const sortedHistory = [...positionHistory].sort((a, b) => 
      new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime()
    );

    const timelineItems = sortedHistory.map((position, index) => ({
      dot: position.endDate ? <HistoryOutlined /> : <UserOutlined style={{ color: '#1890ff' }} />,
      color: position.endDate ? 'gray' : 'blue',
      children: (
        <Card 
          size="small" 
          className={`position-card ${!position.endDate ? 'current-position' : ''}`}
          style={{ marginBottom: 16 }}
        >
          <div className="position-header">
            <Title level={5} style={{ margin: 0, color: !position.endDate ? '#1890ff' : undefined }}>
              {position.positionTitle}
              {!position.endDate && <span className="current-badge">当前职位</span>}
            </Title>
            <Text type="secondary">{position.department}</Text>
          </div>
          
          <div className="position-details" style={{ marginTop: 12 }}>
            <div className="detail-row">
              <Text strong>生效日期：</Text>
              <Text>{dayjs(position.effectiveDate).format('YYYY年MM月DD日')}</Text>
              {position.isRetroactive && <span className="retroactive-badge">追溯</span>}
            </div>
            
            {position.endDate && (
              <div className="detail-row">
                <Text strong>结束日期：</Text>
                <Text>{dayjs(position.endDate).format('YYYY年MM月DD日')}</Text>
              </div>
            )}
            
            {position.jobLevel && (
              <div className="detail-row">
                <Text strong>职级：</Text>
                <Text>{getJobLevelLabel(position.jobLevel)}</Text>
              </div>
            )}
            
            {position.location && (
              <div className="detail-row">
                <Text strong>工作地点：</Text>
                <Text>{position.location}</Text>
              </div>
            )}
            
            <div className="detail-row">
              <Text strong>雇佣类型：</Text>
              <Text>{getEmploymentTypeLabel(position.employmentType)}</Text>
            </div>
            
            {position.changeReason && (
              <div className="detail-row">
                <Text strong>变更原因：</Text>
                <Text>{position.changeReason}</Text>
              </div>
            )}
            
            {position.minSalary && position.maxSalary && (
              <div className="detail-row">
                <Text strong>薪资范围：</Text>
                <Text>
                  {position.minSalary.toLocaleString()} - {position.maxSalary.toLocaleString()} {position.currency}
                </Text>
              </div>
            )}
          </div>
        </Card>
      )
    }));

    return <Timeline items={timelineItems} />;
  };

  const renderWorkflowHistory = () => {
    const sortedWorkflows = [...workflowHistory].sort((a, b) => 
      new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime()
    );

    return (
      <div>
        {sortedWorkflows.map(workflow => (
          <Card key={workflow.workflowId} size="small" style={{ marginBottom: 12 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <Text strong>{workflow.type}</Text>
                <br />
                <Text type="secondary">
                  {dayjs(workflow.startedAt).format('YYYY-MM-DD HH:mm')}
                </Text>
                <br />
                <Text>状态: {getWorkflowStatusLabel(workflow.status)}</Text>
              </div>
              <div>
                <Button 
                  size="small"
                  onClick={() => notification.info({ message: '功能开发中', description: '工作流详情页面正在开发中' })}
                >
                  查看详情
                </Button>
              </div>
            </div>
          </Card>
        ))}
        {workflowHistory.length === 0 && (
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <Text type="secondary">暂无工作流历史记录</Text>
          </div>
        )}
      </div>
    );
  };

  const getEmploymentTypeLabel = (type: string) => {
    const labels = {
      FULL_TIME: '全职',
      PART_TIME: '兼职',
      CONTRACT: '合同工',
      INTERN: '实习生'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getJobLevelLabel = (level: string) => {
    const labels = {
      INTERN: '实习生',
      JUNIOR: '初级',
      INTERMEDIATE: '中级',
      SENIOR: '高级',
      LEAD: '技术负责人',
      MANAGER: '经理',
      DIRECTOR: '总监',
      VP: '副总裁',
      'C-LEVEL': 'C级高管'
    };
    return labels[level as keyof typeof labels] || level;
  };

  const getWorkflowStatusLabel = (status: string) => {
    const labels = {
      PENDING: '待处理',
      IN_PROGRESS: '进行中',
      COMPLETED: '已完成',
      FAILED: '失败',
      CANCELLED: '已取消'
    };
    return labels[status as keyof typeof labels] || status;
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!employee) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Title level={3}>员工未找到</Title>
        <Text type="secondary">请检查员工ID是否正确</Text>
        <br />
        <Button 
          type="primary" 
          icon={<ArrowLeftOutlined />}
          onClick={() => router.push('/employees')}
          style={{ marginTop: '16px' }}
        >
          返回员工列表
        </Button>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Navigation */}
      <div style={{ marginBottom: '16px' }}>
        <Button 
          icon={<ArrowLeftOutlined />}
          onClick={() => router.push('/employees')}
        >
          返回员工列表
        </Button>
      </div>

      {/* Employee Header */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Title level={2} style={{ margin: 0 }}>
              {employee.legalName}
              {employee.preferredName && (
                <Text type="secondary" style={{ fontSize: '18px', marginLeft: '8px' }}>
                  ({employee.preferredName})
                </Text>
              )}
            </Title>
            <Text type="secondary" style={{ fontSize: '16px' }}>
              员工编号：{employee.employeeId} | 邮箱：{employee.email}
            </Text>
            <br />
            <Text type="secondary">
              入职日期：{dayjs(employee.hireDate).format('YYYY年MM月DD日')} | 状态：{employee.status === 'ACTIVE' ? '在职' : '离职'}
            </Text>
            <br />
            <Text type="secondary">
              当前部门：{employee.department || '未分配'} | 当前职位：{employee.position || '未设置'}
            </Text>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setIsModalVisible(true)}
            size="large"
            loading={createLoading}
          >
            新增职位变更
          </Button>
        </div>
      </Card>

      {/* Position History Tabs */}
      <Card>
        <Tabs defaultActiveKey="timeline">
          <TabPane tab="职位时间线" key="timeline">
            <div style={{ padding: '16px 0' }}>
              {positionHistory.length > 0 ? renderPositionTimeline() : (
                <div style={{ textAlign: 'center', padding: '48px' }}>
                  <Text type="secondary">暂无职位历史记录</Text>
                </div>
              )}
            </div>
          </TabPane>
          
          <TabPane tab="工作流历史" key="workflows">
            <div style={{ padding: '16px 0' }}>
              {renderWorkflowHistory()}
            </div>
          </TabPane>
        </Tabs>
      </Card>

      {/* Position Change Modal */}
      <Modal
        title="新增职位变更"
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        confirmLoading={createLoading}
        width={600}
        maskClosable={false}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreatePositionChange}
          initialValues={{
            currency: 'CNY',
            employmentType: 'FULL_TIME'
          }}
        >
          <Form.Item
            name="positionTitle"
            label="职位名称"
            rules={[{ required: true, message: '请输入职位名称' }]}
          >
            <Input placeholder="请输入职位名称" />
          </Form.Item>

          <Form.Item
            name="department"
            label="部门"
            rules={[{ required: true, message: '请输入部门' }]}
          >
            <Select placeholder="请选择部门">
              <Select.Option value="技术部">技术部</Select.Option>
              <Select.Option value="产品部">产品部</Select.Option>
              <Select.Option value="人事部">人事部</Select.Option>
              <Select.Option value="财务部">财务部</Select.Option>
              <Select.Option value="市场部">市场部</Select.Option>
              <Select.Option value="运营部">运营部</Select.Option>
            </Select>
          </Form.Item>

          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="jobLevel"
              label="职级"
              style={{ flex: 1 }}
            >
              <Select placeholder="请选择职级">
                <Select.Option value="INTERN">实习生</Select.Option>
                <Select.Option value="JUNIOR">初级</Select.Option>
                <Select.Option value="INTERMEDIATE">中级</Select.Option>
                <Select.Option value="SENIOR">高级</Select.Option>
                <Select.Option value="LEAD">技术负责人</Select.Option>
                <Select.Option value="MANAGER">经理</Select.Option>
                <Select.Option value="DIRECTOR">总监</Select.Option>
                <Select.Option value="VP">副总裁</Select.Option>
                <Select.Option value="C-LEVEL">C级高管</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="location"
              label="工作地点"
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入工作地点" />
            </Form.Item>
          </div>

          <Form.Item
            name="employmentType"
            label="雇佣类型"
            rules={[{ required: true, message: '请选择雇佣类型' }]}
          >
            <Select>
              <Select.Option value="FULL_TIME">全职</Select.Option>
              <Select.Option value="PART_TIME">兼职</Select.Option>
              <Select.Option value="CONTRACT">合同工</Select.Option>
              <Select.Option value="INTERN">实习生</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="effectiveDate"
            label="生效日期"
            rules={[{ required: true, message: '请选择生效日期' }]}
          >
            <DatePicker 
              style={{ width: '100%' }}
              placeholder="请选择生效日期"
              disabledDate={(current) => {
                // Allow dates up to 2 years in the future
                return current && current > dayjs().add(2, 'year');
              }}
            />
          </Form.Item>

          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="minSalary"
              label="最低薪资"
              style={{ flex: 1 }}
            >
              <Input type="number" placeholder="请输入最低薪资" addonAfter="元" />
            </Form.Item>

            <Form.Item
              name="maxSalary"
              label="最高薪资"
              style={{ flex: 1 }}
            >
              <Input type="number" placeholder="请输入最高薪资" addonAfter="元" />
            </Form.Item>

            <Form.Item
              name="currency"
              label="货币"
              style={{ flex: 1 }}
            >
              <Select>
                <Select.Option value="CNY">人民币</Select.Option>
                <Select.Option value="USD">美元</Select.Option>
                <Select.Option value="EUR">欧元</Select.Option>
              </Select>
            </Form.Item>
          </div>

          <Form.Item
            name="changeReason"
            label="变更原因"
          >
            <TextArea 
              rows={3} 
              placeholder="请输入职位变更的原因"
              showCount
              maxLength={500}
            />
          </Form.Item>
        </Form>
      </Modal>

      <style jsx>{`
        .position-card {
          transition: all 0.3s ease;
        }
        
        .current-position {
          border-color: #1890ff;
          box-shadow: 0 2px 8px rgba(24, 144, 255, 0.2);
        }
        
        .position-header {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }
        
        .current-badge {
          display: inline-block;
          background: #1890ff;
          color: white;
          padding: 2px 8px;
          border-radius: 12px;
          font-size: 12px;
          margin-left: 8px;
        }
        
        .retroactive-badge {
          display: inline-block;
          background: #ff7875;
          color: white;
          padding: 2px 6px;
          border-radius: 10px;
          font-size: 11px;
          margin-left: 8px;
        }
        
        .detail-row {
          display: flex;
          align-items: center;
          margin-bottom: 8px;
          gap: 8px;
        }
        
        .detail-row:last-child {
          margin-bottom: 0;
        }
      `}</style>
    </div>
  );
};

export default EmployeePositionHistoryPage;