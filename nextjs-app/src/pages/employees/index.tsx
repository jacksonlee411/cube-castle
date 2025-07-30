// src/pages/employees/index.tsx - Full CRUD functionality for UAT testing
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
  HistoryOutlined,
  EditOutlined,
  DeleteOutlined
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
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [form] = Form.useForm();

  // Sample data with full CRUD capabilities
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
      
      if (editingEmployee) {
        // Update existing employee
        const updatedEmployee: Employee = {
          ...editingEmployee,
          employeeId: values.employeeId,
          legalName: values.legalName,
          preferredName: values.preferredName,
          email: values.email,
          hireDate: dayjs(values.hireDate).format('YYYY-MM-DD'),
          department: values.department,
          position: values.position,
          managerName: values.managerName
        };

        setEmployees(prev => prev.map(emp => 
          emp.id === editingEmployee.id ? updatedEmployee : emp
        ));

        notification.success({
          message: '员工更新成功',
          description: `员工 ${values.legalName} 信息已更新。`,
        });
      } else {
        // Create new employee
        const newEmployee: Employee = {
          id: Date.now().toString(),
          employeeId: values.employeeId,
          legalName: values.legalName,
          preferredName: values.preferredName,
          email: values.email,
          status: 'ACTIVE',
          hireDate: dayjs(values.hireDate).format('YYYY-MM-DD'),
          department: values.department,
          position: values.position,
          managerName: values.managerName
        };

        setEmployees(prev => [...prev, newEmployee]);
        
        notification.success({
          message: '员工创建成功',
          description: `员工 ${values.legalName} 已成功添加到系统中。`,
        });
      }
      
      handleModalClose();
    } catch (error) {
      notification.error({
        message: editingEmployee ? '员工更新失败' : '员工创建失败',
        description: '操作时发生错误，请重试。',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (employee: Employee) => {
    setEditingEmployee(employee);
    form.setFieldsValue({
      ...employee,
      hireDate: dayjs(employee.hireDate)
    });
    setIsModalVisible(true);
  };

  const handleDelete = (employee: Employee) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除员工 ${employee.legalName} 吗？此操作不可撤销。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => {
        setEmployees(prev => prev.filter(emp => emp.id !== employee.id));
        notification.success({
          message: '员工删除成功',
          description: `员工 ${employee.legalName} 已从系统中删除。`,
        });
      }
    });
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingEmployee(null);
    form.resetFields();
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
        key="edit" 
        icon={<EditOutlined />}
        onClick={() => handleEdit(employee)}
      >
        编辑信息
      </Menu.Item>
      <Menu.Item 
        key="positions" 
        icon={<HistoryOutlined />}
        onClick={() => router.push(`/employees/positions/${employee.id}`)}
      >
        职位历史
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item 
        key="delete" 
        icon={<DeleteOutlined />}
        onClick={() => handleDelete(employee)}
        style={{ color: '#ff4d4f' }}
      >
        删除员工
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
              {record.legalName}
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
          <Tooltip title="编辑员工">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
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
            管理公司员工信息、职位变更和组织结构 - 完整CRUD功能
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
          
          <div style={{ color: '#666', fontSize: '14px' }}>
            共找到 {filteredEmployees.length} 名员工
          </div>
        </div>
      </Card>

      {/* Employee Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={filteredEmployees}
          loading={loading}
          rowKey="id"
          pagination={{
            total: filteredEmployees.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Create/Edit Employee Modal */}
      <Modal
        title={editingEmployee ? '编辑员工信息' : '新增员工'}
        open={isModalVisible}
        onCancel={handleModalClose}
        footer={null}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateEmployee}
          initialValues={{
            status: 'ACTIVE'
          }}
        >
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item
              label="员工工号"
              name="employeeId"
              rules={[{ required: true, message: '请输入员工工号' }]}
            >
              <Input placeholder="如: EMP001" />
            </Form.Item>
            
            <Form.Item
              label="法定姓名"
              name="legalName"
              rules={[{ required: true, message: '请输入法定姓名' }]}
            >
              <Input placeholder="员工的法定姓名" />
            </Form.Item>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item
              label="常用姓名"
              name="preferredName"
            >
              <Input placeholder="常用的英文姓名(可选)" />
            </Form.Item>
            
            <Form.Item
              label="邮箱地址"
              name="email"
              rules={[
                { required: true, message: '请输入邮箱地址' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input placeholder="employee@company.com" />
            </Form.Item>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item
              label="所属部门"
              name="department"
              rules={[{ required: true, message: '请选择所属部门' }]}
            >
              <Select placeholder="选择部门">
                <Option value="技术部">技术部</Option>
                <Option value="产品部">产品部</Option>
                <Option value="人事部">人事部</Option>
                <Option value="财务部">财务部</Option>
                <Option value="市场部">市场部</Option>
                <Option value="运营部">运营部</Option>
              </Select>
            </Form.Item>
            
            <Form.Item
              label="职位"
              name="position"
              rules={[{ required: true, message: '请输入职位' }]}
            >
              <Input placeholder="如: 高级软件工程师" />
            </Form.Item>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item
              label="入职日期"
              name="hireDate"
              rules={[{ required: true, message: '请选择入职日期' }]}
            >
              <DatePicker 
                style={{ width: '100%' }}
                placeholder="选择入职日期"
                format="YYYY-MM-DD"
              />
            </Form.Item>
            
            <Form.Item
              label="直属经理"
              name="managerName"
            >
              <Input placeholder="直属经理姓名(可选)" />
            </Form.Item>
          </div>

          <Form.Item style={{ marginTop: '24px', marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={handleModalClose}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingEmployee ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default EmployeesPage;