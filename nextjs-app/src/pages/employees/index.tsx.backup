// src/pages/employees/index.tsx
import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Input, 
  Select, 
  Space, 
  Tag, 
  Avatar, 
  Modal,
  Form,
  DatePicker,
  notification,
  Dropdown,
  Menu,
  Tooltip
} from 'antd';
import { 
  PlusOutlined, 
  SearchOutlined, 
  MoreOutlined,
  UserOutlined,
  MailOutlined,
  PhoneOutlined,
  CalendarOutlined,
  TeamOutlined,
  HistoryOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import Link from 'next/link';
import dayjs from 'dayjs';

const { Search } = Input;
const { Option } = Select;

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
  managerId?: string;
  managerName?: string;
  avatar?: string;
}

const EmployeesPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [filteredEmployees, setFilteredEmployees] = useState<Employee[]>([]);
  const [searchText, setSearchText] = useState('');
  const [departmentFilter, setDepartmentFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  // Sample data - in real app would come from GraphQL
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleEmployees: Employee[] = [
        {
          id: '1',
          employeeId: 'EMP001',
          legalName: '张三',
          preferredName: 'Zhang San',
          email: 'zhangsan@company.com',
          status: 'ACTIVE',
          hireDate: '2023-01-15',
          department: '技术部',
          position: '高级软件工程师',
          managerName: '李四'
        },
        {
          id: '2',
          employeeId: 'EMP002',
          legalName: '王五',
          email: 'wangwu@company.com',
          status: 'ACTIVE',
          hireDate: '2023-03-20',
          department: '产品部',
          position: '产品经理',
          managerName: '赵六'
        },
        {
          id: '3',
          employeeId: 'EMP003',
          legalName: '刘七',
          email: 'liuqi@company.com',
          status: 'INACTIVE',
          hireDate: '2022-08-10',
          department: '技术部',
          position: '前端工程师',
          managerName: '李四'
        },
        {
          id: '4',
          employeeId: 'EMP004',
          legalName: '陈八',
          email: 'chenba@company.com',
          status: 'ACTIVE',
          hireDate: '2024-01-08',
          department: '人事部',
          position: 'HR专员',
          managerName: '周九'
        }
      ];
      
      setEmployees(sampleEmployees);
      setFilteredEmployees(sampleEmployees);
      setLoading(false);
    }, 1000);
  }, []);

  // Filter employees based on search and filters
  useEffect(() => {
    let filtered = employees;

    if (searchText) {
      filtered = filtered.filter(emp => 
        emp.legalName.toLowerCase().includes(searchText.toLowerCase()) ||
        emp.employeeId.toLowerCase().includes(searchText.toLowerCase()) ||
        emp.email.toLowerCase().includes(searchText.toLowerCase()) ||
        (emp.position && emp.position.toLowerCase().includes(searchText.toLowerCase()))
      );
    }

    if (departmentFilter) {
      filtered = filtered.filter(emp => emp.department === departmentFilter);
    }

    if (statusFilter) {
      filtered = filtered.filter(emp => emp.status === statusFilter);
    }

    setFilteredEmployees(filtered);
  }, [employees, searchText, departmentFilter, statusFilter]);

  const handleCreateEmployee = async (values: any) => {
    try {
      setLoading(true);
      
      // Simulate GraphQL mutation
      const newEmployee: Employee = {
        id: Date.now().toString(),
        employeeId: values.employeeId,
        legalName: values.legalName,
        preferredName: values.preferredName,
        email: values.email,
        status: 'ACTIVE',
        hireDate: values.hireDate.format('YYYY-MM-DD'),
        department: values.department,
        position: values.position,
        managerName: values.managerName
      };

      setEmployees(prev => [...prev, newEmployee]);
      
      notification.success({
        message: '员工创建成功',
        description: `员工 ${values.legalName} 已成功添加到系统中。`,
      });
      
      setIsModalVisible(false);
      form.resetFields();
    } catch (error) {
      notification.error({
        message: '员工创建失败',
        description: '创建员工时发生错误，请重试。',
      });
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    const colors = {
      ACTIVE: 'green',
      INACTIVE: 'red',
      PENDING: 'orange'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusLabel = (status: string) => {
    const labels = {
      ACTIVE: '在职',
      INACTIVE: '离职',
      PENDING: '待入职'
    };
    return labels[status as keyof typeof labels] || status;
  };

  const getActionMenu = (employee: Employee) => (
    <Menu>
      <Menu.Item 
        key="view" 
        icon={<UserOutlined />}
        onClick={() => router.push(`/employees/${employee.id}`)}
      >
        查看详情
      </Menu.Item>
      <Menu.Item 
        key="positions" 
        icon={<HistoryOutlined />}
        onClick={() => router.push(`/employees/positions/${employee.id}`)}
      >
        职位历史
      </Menu.Item>
      <Menu.Item 
        key="edit" 
        icon={<MoreOutlined />}
      >
        编辑信息
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item 
        key="status" 
        disabled={employee.status !== 'ACTIVE'}
      >
        {employee.status === 'ACTIVE' ? '标记为离职' : '激活员工'}
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: '员工信息',
      key: 'employee',
      render: (record: Employee) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Avatar 
            size={40} 
            icon={<UserOutlined />}
            src={record.avatar}
            style={{ backgroundColor: '#1890ff' }}
          >
            {record.legalName.charAt(0)}
          </Avatar>
          <div>
            <div style={{ fontWeight: 'bold', marginBottom: '2px' }}>
              <Link href={`/employees/${record.id}`} style={{ color: 'inherit' }}>
                {record.legalName}
              </Link>
              {record.preferredName && (
                <span style={{ color: '#888', marginLeft: '8px' }}>
                  ({record.preferredName})
                </span>
              )}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              <Space size="small">
                <span>{record.employeeId}</span>
                <span>•</span>
                <span>{record.email}</span>
              </Space>
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '职位信息',
      key: 'position',
      render: (record: Employee) => (
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: '2px' }}>
            {record.position || '未设置'}
          </div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            {record.department || '未设置部门'}
          </div>
        </div>
      ),
    },
    {
      title: '直属经理',
      dataIndex: 'managerName',
      key: 'manager',
      render: (managerName: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          {managerName ? (
            <>
              <TeamOutlined style={{ color: '#1890ff' }} />
              <span>{managerName}</span>
            </>
          ) : (
            <span style={{ color: '#999' }}>无</span>
          )}
        </div>
      ),
    },
    {
      title: '入职日期',
      dataIndex: 'hireDate',
      key: 'hireDate',
      render: (hireDate: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <CalendarOutlined style={{ color: '#52c41a' }} />
          <span>{dayjs(hireDate).format('YYYY年MM月DD日')}</span>
        </div>
      ),
      sorter: (a: Employee, b: Employee) => 
        dayjs(a.hireDate).unix() - dayjs(b.hireDate).unix(),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {getStatusLabel(status)}
        </Tag>
      ),
      filters: [
        { text: '在职', value: 'ACTIVE' },
        { text: '离职', value: 'INACTIVE' },
        { text: '待入职', value: 'PENDING' },
      ],
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (record: Employee) => (
        <Space>
          <Tooltip title="查看职位历史">
            <Button 
              type="text" 
              icon={<HistoryOutlined />}
              onClick={() => router.push(`/employees/positions/${record.id}`)}
            />
          </Tooltip>
          <Dropdown 
            overlay={getActionMenu(record)} 
            trigger={['click']}
            placement="bottomRight"
          >
            <Button type="text" icon={<MoreOutlined />} />
          </Dropdown>
        </Space>
      ),
    },
  ];

  const departments = Array.from(new Set(employees.map(emp => emp.department).filter(Boolean)));

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 'bold' }}>员工管理</h1>
          <p style={{ margin: '4px 0 0 0', color: '#666' }}>
            管理公司员工信息、职位变更和组织结构
          </p>
        </div>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          size="large"
          onClick={() => setIsModalVisible(true)}
        >
          新增员工
        </Button>
      </div>

      {/* Filters */}
      <Card style={{ marginBottom: '24px' }}>
        <div style={{ display: 'flex', gap: '16px', alignItems: 'center', flexWrap: 'wrap' }}>
          <Search
            placeholder="搜索员工姓名、工号、邮箱或职位"
            style={{ width: '300px' }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
          
          <Select
            placeholder="选择部门"
            style={{ width: '150px' }}
            value={departmentFilter}
            onChange={setDepartmentFilter}
            allowClear
          >
            {departments.map(dept => (
              <Option key={dept} value={dept}>{dept}</Option>
            ))}
          </Select>
          
          <Select
            placeholder="选择状态"
            style={{ width: '120px' }}
            value={statusFilter}
            onChange={setStatusFilter}
            allowClear
          >
            <Option value="ACTIVE">在职</Option>
            <Option value="INACTIVE">离职</Option>
            <Option value="PENDING">待入职</Option>
          </Select>
          
          <div style={{ color: '#666' }}>
            总计 {filteredEmployees.length} 名员工
          </div>
        </div>
      </Card>

      {/* Employee Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={filteredEmployees}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      {/* Create Employee Modal */}
      <Modal
        title="新增员工"
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        confirmLoading={loading}
        width={600}
        maskClosable={false}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateEmployee}
        >
          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="employeeId"
              label="员工工号"
              rules={[{ required: true, message: '请输入员工工号' }]}
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入员工工号" />
            </Form.Item>

            <Form.Item
              name="legalName"
              label="姓名"
              rules={[{ required: true, message: '请输入员工姓名' }]}
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入员工姓名" />
            </Form.Item>
          </div>

          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="preferredName"
              label="英文名"
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入英文名（可选）" />
            </Form.Item>

            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱地址' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入邮箱地址" />
            </Form.Item>
          </div>

          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="department"
              label="部门"
              rules={[{ required: true, message: '请输入部门' }]}
              style={{ flex: 1 }}
            >
              <Select placeholder="请选择部门">
                <Option value="技术部">技术部</Option>
                <Option value="产品部">产品部</Option>
                <Option value="人事部">人事部</Option>
                <Option value="市场部">市场部</Option>
                <Option value="财务部">财务部</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="position"
              label="职位"
              rules={[{ required: true, message: '请输入职位' }]}
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入职位" />
            </Form.Item>
          </div>

          <div style={{ display: 'flex', gap: '16px' }}>
            <Form.Item
              name="hireDate"
              label="入职日期"
              rules={[{ required: true, message: '请选择入职日期' }]}
              style={{ flex: 1 }}
            >
              <DatePicker 
                style={{ width: '100%' }} 
                placeholder="请选择入职日期"
                disabledDate={(current) => current && current > dayjs()}
              />
            </Form.Item>

            <Form.Item
              name="managerName"
              label="直属经理"
              style={{ flex: 1 }}
            >
              <Input placeholder="请输入直属经理姓名（可选）" />
            </Form.Item>
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default EmployeesPage;