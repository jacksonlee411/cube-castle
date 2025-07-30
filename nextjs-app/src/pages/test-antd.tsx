// 测试Ant Design组件ES模块兼容性
import React from 'react';
import { 
  Button, 
  Card, 
  Table, 
  Form, 
  Input, 
  DatePicker, 
  Select, 
  notification,
  Space,
  Typography,
  Alert
} from 'antd';
import { 
  UserOutlined, 
  EditOutlined, 
  DeleteOutlined,
  CheckCircleOutlined 
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;

interface TestData {
  key: string;
  name: string;
  age: number;
  address: string;
}

const AntdCompatibilityTest: React.FC = () => {
  const [form] = Form.useForm();

  const testData: TestData[] = [
    {
      key: '1',
      name: '张三',
      age: 32,
      address: '北京市朝阳区'
    },
    {
      key: '2',
      name: '李四',
      age: 28,
      address: '上海市浦东新区'
    }
  ];

  const columns = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <Space>
          <UserOutlined />
          {text}
        </Space>
      )
    },
    {
      title: '年龄',
      dataIndex: 'age',
      key: 'age'
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address'
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />}>编辑</Button>
          <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
        </Space>
      )
    }
  ];

  const handleSubmit = (values: any) => {
    notification.success({
      message: '表单提交成功',
      description: `提交的数据：${JSON.stringify(values)}`,
      icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />
    });
  };

  return (
    <div style={{ padding: '24px', backgroundColor: '#f0f2f5', minHeight: '100vh' }}>
      <Title level={2}>Ant Design ES模块兼容性测试</Title>
      
      <Alert
        message="版本降级测试"
        description={`
          Next.js: 14.1.4 
          | Ant Design: 5.20.6 
          | @ant-design/icons: 5.3.7 
          | rc-util: 5.38.2
        `}
        type="info"
        showIcon
        style={{ marginBottom: '24px' }}
      />

      <Card title="基础组件测试" style={{ marginBottom: '24px' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          
          {/* 按钮测试 */}
          <div>
            <Text strong>按钮组件：</Text>
            <Space style={{ marginLeft: '16px' }}>
              <Button type="primary" icon={<UserOutlined />}>主要按钮</Button>
              <Button icon={<EditOutlined />}>默认按钮</Button>
              <Button type="dashed" icon={<DeleteOutlined />}>虚线按钮</Button>
            </Space>
          </div>

          {/* 表单测试 */}
          <div>
            <Text strong>表单组件：</Text>
            <Form
              form={form}
              layout="inline"
              onFinish={handleSubmit}
              style={{ marginTop: '8px' }}
            >
              <Form.Item
                name="name"
                rules={[{ required: true, message: '请输入姓名' }]}
              >
                <Input placeholder="姓名" prefix={<UserOutlined />} />
              </Form.Item>
              
              <Form.Item
                name="date"
                rules={[{ required: true, message: '请选择日期' }]}
              >
                <DatePicker placeholder="选择日期" />
              </Form.Item>
              
              <Form.Item
                name="status"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="选择状态" style={{ width: '120px' }}>
                  <Option value="active">活跃</Option>
                  <Option value="inactive">非活跃</Option>
                </Select>
              </Form.Item>
              
              <Form.Item>
                <Button type="primary" htmlType="submit">
                  提交测试
                </Button>
              </Form.Item>
            </Form>
          </div>
        </Space>
      </Card>

      {/* 表格测试 */}
      <Card title="表格组件测试">
        <Table 
          dataSource={testData} 
          columns={columns}
          pagination={{ pageSize: 5 }}
          size="middle"
        />
      </Card>

      <div style={{ marginTop: '24px', textAlign: 'center' }}>
        <Text type="success">
          ✅ 如果您能看到此页面且所有组件正常显示，说明ES模块兼容性问题已解决！
        </Text>
      </div>
    </div>
  );
};

export default AntdCompatibilityTest;