// src/pages/employees/positions/[id].tsx
import React, { useState } from 'react';
import { useRouter } from 'next/router';
import { Card, Typography, Timeline, Button, Modal, Form, DatePicker, Input, Select, notification, Spin, Tabs } from 'antd';
import { PlusOutlined, EditOutlined, HistoryOutlined, UserOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useEmployee, useCreatePositionChange, usePositionTimeline } from '@/hooks/useEmployees';
import { useWorkflowHistory } from '@/hooks/useWorkflows';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { TabPane } = Tabs;

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

const EmployeePositionHistoryPage: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  // Use custom hooks for data fetching
  const { employee, loading: employeeLoading, refetch: refetchEmployee } = useEmployee(id as string);
  const { create: createPositionChange, loading: createLoading } = useCreatePositionChange();
  const { workflows, loading: workflowsLoading } = useWorkflowHistory(id as string);
  const { positionTimeline: positionHistory, loading: positionLoading, refetch: refetchPositions } = usePositionTimeline(id as string, 50);

  const handleCreatePositionChange = async (values: any) => {
    if (!employee) return;

    const result = await createPositionChange({
      employeeId: employee.id,
      positionData: {
        positionTitle: values.positionTitle,
        department: values.department,
        jobLevel: values.jobLevel,
        location: values.location,
        employmentType: values.employmentType,
        reportsToEmployeeId: values.reportsToEmployeeId,
        minSalary: values.minSalary ? parseFloat(values.minSalary) : undefined,
        maxSalary: values.maxSalary ? parseFloat(values.maxSalary) : undefined,
        currency: values.currency || 'CNY'
      },
      effectiveDate: values.effectiveDate.toISOString(),
      changeReason: values.changeReason,
      isRetroactive: dayjs(values.effectiveDate).isBefore(dayjs())
    });

    if (result.success) {
      setIsModalVisible(false);
      form.resetFields();
      refetchEmployee(); // Refresh employee data
      refetchPositions(); // Refresh position timeline
      
      // Navigate to workflow status if workflow was created
      if (result.workflowId) {
        router.push(`/workflows/${result.workflowId}`);
      }
    }
  };

  const renderPositionTimeline = () => {
    const timelineItems = positionHistory.map((position: any, index: number) => ({
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
                <Text>{position.jobLevel}</Text>
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
    if (workflowsLoading) {
      return <Spin />;
    }

    return (
      <div>
        {workflows.map(workflow => (
          <Card key={workflow.workflowId} size="small" style={{ marginBottom: 12 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <Text strong>{workflow.workflowId}</Text>
                <br />
                <Text type="secondary">
                  {dayjs(workflow.startedAt).format('YYYY-MM-DD HH:mm')}
                </Text>
              </div>
              <div>
                <Button 
                  size="small"
                  onClick={() => router.push(`/workflows/${workflow.workflowId}`)}
                >
                  查看详情
                </Button>
              </div>
            </div>
          </Card>
        ))}
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

  if (employeeLoading || positionLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!employee) {
    return <div>员工未找到</div>;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Employee Header */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Title level={2} style={{ margin: 0 }}>
              {employee.legalName}
            </Title>
            <Text type="secondary" style={{ fontSize: '16px' }}>
              员工编号：{employee.employeeId} | 邮箱：{employee.email}
            </Text>
            <br />
            <Text type="secondary">
              入职日期：{dayjs(employee.hireDate).format('YYYY年MM月DD日')} | 状态：{employee.status}
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
            <Input placeholder="请输入部门" />
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