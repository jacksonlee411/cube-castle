// src/pages/organization/chart.tsx - Full CRUD functionality for UAT testing
import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Input, 
  Select, 
  Space, 
  Tag, 
  Modal,
  Form,
  TreeSelect,
  notification,
  Dropdown,
  Menu,
  Tooltip,
  Row,
  Col,
  Statistic,
  Tree,
  Divider
} from 'antd';
import { 
  PlusOutlined, 
  SearchOutlined, 
  MoreOutlined,
  TeamOutlined,
  UserOutlined,
  EditOutlined,
  DeleteOutlined,
  BranchesOutlined,
  UsergroupAddOutlined,
  HomeOutlined,
  SettingOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import Link from 'next/link';

const { Search } = Input;
const { Option } = Select;
const { TreeNode } = Tree;

interface OrganizationUnit {
  id: string;
  name: string;
  unitType: 'COMPANY' | 'DIVISION' | 'DEPARTMENT' | 'TEAM';
  parentId?: string;
  managerId?: string;
  managerName?: string;
  employeeCount: number;
  description?: string;
  status: 'ACTIVE' | 'INACTIVE';
  createdAt: string;
  children?: OrganizationUnit[];
}

interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  email: string;
  position?: string;
  department?: string;
  status: string;
  managerId?: string;
  organizationUnitId?: string;
}

const OrganizationChartPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [orgUnits, setOrgUnits] = useState<OrganizationUnit[]>([]);
  const [filteredUnits, setFilteredUnits] = useState<OrganizationUnit[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [searchText, setSearchText] = useState('');
  const [unitTypeFilter, setUnitTypeFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingUnit, setEditingUnit] = useState<OrganizationUnit | null>(null);
  const [viewMode, setViewMode] = useState<'table' | 'tree'>('tree');
  const [form] = Form.useForm();

  // Sample organization data with full CRUD capabilities
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const sampleUnits: OrganizationUnit[] = [
        {
          id: '1',
          name: 'Cube Castle 科技有限公司',
          unitType: 'COMPANY',
          employeeCount: 15,
          status: 'ACTIVE',
          createdAt: '2023-01-01',
          description: '总公司'
        },
        {
          id: '2',
          name: '技术研发部',
          unitType: 'DEPARTMENT',
          parentId: '1',
          managerId: 'EMP001',
          managerName: '李技术',
          employeeCount: 8,
          status: 'ACTIVE',
          createdAt: '2023-01-15',
          description: '负责产品研发和技术架构'
        },
        {
          id: '3',
          name: '前端开发团队',
          unitType: 'TEAM',
          parentId: '2',
          managerId: 'EMP002',
          managerName: '王前端',
          employeeCount: 4,
          status: 'ACTIVE',
          createdAt: '2023-02-01',
          description: '负责前端应用开发'
        },
        {
          id: '4',
          name: '后端开发团队',
          unitType: 'TEAM',
          parentId: '2',
          managerId: 'EMP003',
          managerName: '张后端',
          employeeCount: 3,
          status: 'ACTIVE',
          createdAt: '2023-02-01',
          description: '负责后端服务开发'
        },
        {
          id: '5',
          name: '产品运营部',
          unitType: 'DEPARTMENT',
          parentId: '1',
          managerId: 'EMP004',
          managerName: '陈产品',
          employeeCount: 4,
          status: 'ACTIVE',
          createdAt: '2023-01-20',
          description: '负责产品运营和市场推广'
        },
        {
          id: '6',
          name: '人力资源部',
          unitType: 'DEPARTMENT',
          parentId: '1',
          managerId: 'EMP005',
          managerName: '赵人事',
          employeeCount: 2,
          status: 'ACTIVE',
          createdAt: '2023-01-10',
          description: '负责人力资源管理'
        },
        {
          id: '7',
          name: '财务部',
          unitType: 'DEPARTMENT',
          parentId: '1',
          managerId: 'EMP006',
          managerName: '钱财务',
          employeeCount: 1,
          status: 'ACTIVE',
          createdAt: '2023-01-05',
          description: '负责财务管理和会计核算'
        }
      ];
      
      setOrgUnits(sampleUnits);
      setFilteredUnits(sampleUnits);
      setLoading(false);
    }, 1000);
  }, []);

  // Filter organization units
  useEffect(() => {
    let filtered = orgUnits;

    if (searchText) {
      filtered = filtered.filter(unit => 
        unit.name.toLowerCase().includes(searchText.toLowerCase()) ||
        (unit.managerName && unit.managerName.toLowerCase().includes(searchText.toLowerCase())) ||
        (unit.description && unit.description.toLowerCase().includes(searchText.toLowerCase()))
      );
    }

    if (unitTypeFilter) {
      filtered = filtered.filter(unit => unit.unitType === unitTypeFilter);
    }

    if (statusFilter) {
      filtered = filtered.filter(unit => unit.status === statusFilter);
    }

    setFilteredUnits(filtered);
  }, [orgUnits, searchText, unitTypeFilter, statusFilter]);

  const handleCreateUnit = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingUnit) {
        // Update existing unit
        const updatedUnit: OrganizationUnit = {
          ...editingUnit,
          name: values.name,
          unitType: values.unitType,
          parentId: values.parentId,
          managerName: values.managerName,
          description: values.description,
          status: values.status || editingUnit.status
        };

        setOrgUnits(prev => prev.map(unit => 
          unit.id === editingUnit.id ? updatedUnit : unit
        ));

        notification.success({
          message: '组织单元更新成功',
          description: `组织单元 ${values.name} 信息已更新。`,
        });
      } else {
        // Create new unit
        const newUnit: OrganizationUnit = {
          id: Date.now().toString(),
          name: values.name,
          unitType: values.unitType,
          parentId: values.parentId,
          managerName: values.managerName,
          employeeCount: 0,
          status: 'ACTIVE',
          createdAt: new Date().toISOString().split('T')[0],
          description: values.description
        };

        setOrgUnits(prev => [...prev, newUnit]);
        
        notification.success({
          message: '组织单元创建成功',
          description: `组织单元 ${values.name} 已成功添加到系统中。`,
        });
      }
      
      handleModalClose();
    } catch (error) {
      notification.error({
        message: editingUnit ? '组织单元更新失败' : '组织单元创建失败',
        description: '操作时发生错误，请重试。',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (unit: OrganizationUnit) => {
    setEditingUnit(unit);
    form.setFieldsValue({
      name: unit.name,
      unitType: unit.unitType,
      parentId: unit.parentId,
      managerName: unit.managerName,
      description: unit.description,
      status: unit.status
    });
    setIsModalVisible(true);
  };

  const handleDelete = (unit: OrganizationUnit) => {
    // Check if unit has children
    const hasChildren = orgUnits.some(u => u.parentId === unit.id);
    
    if (hasChildren) {
      notification.warning({
        message: '无法删除',
        description: '该组织单元下还有子单元，请先删除或转移子单元。',
      });
      return;
    }

    Modal.confirm({
      title: '确认删除',
      content: `确定要删除组织单元 ${unit.name} 吗？此操作不可撤销。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => {
        setOrgUnits(prev => prev.filter(u => u.id !== unit.id));
        notification.success({
          message: '组织单元删除成功',
          description: `组织单元 ${unit.name} 已从系统中删除。`,
        });
      }
    });
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingUnit(null);
    form.resetFields();
  };

  const getUnitTypeColor = (type: string) => {
    const colors = {
      COMPANY: 'purple',
      DIVISION: 'blue',
      DEPARTMENT: 'green',
      TEAM: 'orange'
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getUnitTypeLabel = (type: string) => {
    const labels = {
      COMPANY: '公司',
      DIVISION: '事业部',
      DEPARTMENT: '部门',
      TEAM: '团队'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getStatusColor = (status: string) => {
    return status === 'ACTIVE' ? 'green' : 'red';
  };

  const getStatusLabel = (status: string) => {
    return status === 'ACTIVE' ? '活跃' : '停用';
  };

  const getActionMenu = (unit: OrganizationUnit) => (
    <Menu>
      <Menu.Item 
        key="edit" 
        icon={<EditOutlined />}
        onClick={() => handleEdit(unit)}
      >
        编辑信息
      </Menu.Item>
      <Menu.Item 
        key="employees" 
        icon={<UserOutlined />}
        onClick={() => router.push(`/organization/units/${unit.id}/employees`)}
      >
        查看员工
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item 
        key="delete" 
        icon={<DeleteOutlined />}
        onClick={() => handleDelete(unit)}
        style={{ color: '#ff4d4f' }}
      >
        删除单元
      </Menu.Item>
    </Menu>
  );

  // Build tree data for tree view
  const buildTreeData = (units: OrganizationUnit[], parentId?: string): any[] => {
    return units
      .filter(unit => unit.parentId === parentId)
      .map(unit => ({
        title: (
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Tag color={getUnitTypeColor(unit.unitType)} size="small">
              {getUnitTypeLabel(unit.unitType)}
            </Tag>
            <span style={{ fontWeight: 'bold' }}>{unit.name}</span>
            <span style={{ color: '#666', fontSize: '12px' }}>
              ({unit.employeeCount} 人)
            </span>
            <Dropdown overlay={getActionMenu(unit)} trigger={['click']}>
              <Button type="text" size="small" icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        ),
        key: unit.id,
        children: buildTreeData(units, unit.id)
      }));
  };

  const columns = [
    {
      title: '组织单元',
      key: 'unit',
      render: (record: OrganizationUnit) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div style={{
            width: '40px',
            height: '40px',
            borderRadius: '8px',
            backgroundColor: '#f0f8ff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            {record.unitType === 'COMPANY' ? '🏢' : 
             record.unitType === 'DIVISION' ? '🏗️' :
             record.unitType === 'DEPARTMENT' ? '🏪' : '👥'}
          </div>
          <div>
            <div style={{ fontWeight: 'bold', marginBottom: '2px' }}>
              {record.name}
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              {record.description || '暂无描述'}
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'unitType',
      key: 'unitType',
      render: (type: string) => (
        <Tag color={getUnitTypeColor(type)}>
          {getUnitTypeLabel(type)}
        </Tag>
      ),
      filters: [
        { text: '公司', value: 'COMPANY' },
        { text: '事业部', value: 'DIVISION' },
        { text: '部门', value: 'DEPARTMENT' },
        { text: '团队', value: 'TEAM' },
      ],
    },
    {
      title: '负责人',
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
            <span style={{ color: '#999' }}>待分配</span>
          )}
        </div>
      ),
    },
    {
      title: '员工数量',
      dataIndex: 'employeeCount',
      key: 'employeeCount',
      render: (count: number) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <UserOutlined style={{ color: '#52c41a' }} />
          <span>{count} 人</span>
        </div>
      ),
      sorter: (a: OrganizationUnit, b: OrganizationUnit) => 
        a.employeeCount - b.employeeCount,
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
        { text: '活跃', value: 'ACTIVE' },
        { text: '停用', value: 'INACTIVE' },
      ],
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      sorter: (a: OrganizationUnit, b: OrganizationUnit) => 
        new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (record: OrganizationUnit) => (
        <Space>
          <Tooltip title="编辑组织单元">
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

  // Calculate statistics
  const totalUnits = orgUnits.length;
  const totalEmployees = orgUnits.reduce((sum, unit) => sum + unit.employeeCount, 0);
  const activeUnits = orgUnits.filter(unit => unit.status === 'ACTIVE').length;
  const departmentCount = orgUnits.filter(unit => unit.unitType === 'DEPARTMENT').length;

  const treeData = buildTreeData(filteredUnits);

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 'bold' }}>组织架构管理</h1>
          <p style={{ margin: '4px 0 0 0', color: '#666' }}>
            管理公司组织架构、部门设置和人员配置 - 完整CRUD功能
          </p>
        </div>
        <Space>
          <Button 
            type={viewMode === 'tree' ? 'primary' : 'default'}
            icon={<BranchesOutlined />}
            onClick={() => setViewMode('tree')}
          >
            树形视图
          </Button>
          <Button 
            type={viewMode === 'table' ? 'primary' : 'default'}
            icon={<SettingOutlined />}
            onClick={() => setViewMode('table')}
          >
            表格视图
          </Button>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            size="large"
            onClick={() => setIsModalVisible(true)}
          >
            新增组织单元
          </Button>
        </Space>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="组织单元总数"
              value={totalUnits}
              prefix={<HomeOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="员工总数"
              value={totalEmployees}
              prefix={<UserOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃单元"
              value={activeUnits}
              prefix={<TeamOutlined style={{ color: '#faad14' }} />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="部门数量"
              value={departmentCount}
              prefix={<BranchesOutlined style={{ color: '#722ed1' }} />}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Card style={{ marginBottom: '24px' }}>
        <div style={{ display: 'flex', gap: '16px', alignItems: 'center', flexWrap: 'wrap' }}>
          <Search
            placeholder="搜索组织单元名称、负责人或描述"
            style={{ width: '300px' }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
          
          <Select
            placeholder="选择单元类型"
            style={{ width: '150px' }}
            value={unitTypeFilter}
            onChange={setUnitTypeFilter}
            allowClear
          >
            <Option value="COMPANY">公司</Option>
            <Option value="DIVISION">事业部</Option>
            <Option value="DEPARTMENT">部门</Option>
            <Option value="TEAM">团队</Option>
          </Select>
          
          <Select
            placeholder="选择状态"
            style={{ width: '120px' }}
            value={statusFilter}
            onChange={setStatusFilter}
            allowClear
          >
            <Option value="ACTIVE">活跃</Option>
            <Option value="INACTIVE">停用</Option>
          </Select>
          
          <div style={{ color: '#666', fontSize: '14px' }}>
            共找到 {filteredUnits.length} 个组织单元
          </div>
        </div>
      </Card>

      {/* Content */}
      <Card>
        {viewMode === 'tree' ? (
          <div>
            <Divider orientation="left">组织架构树</Divider>
            {treeData.length > 0 ? (
              <Tree
                showLine={{ showLeafIcon: false }}
                defaultExpandAll
                treeData={treeData}
                style={{ backgroundColor: '#fafafa', padding: '16px', borderRadius: '6px' }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: '48px', color: '#666' }}>
                <BranchesOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                <div>暂无组织架构数据</div>
              </div>
            )}
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={filteredUnits}
            loading={loading}
            rowKey="id"
            pagination={{
              total: filteredUnits.length,
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => 
                `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
            }}
            scroll={{ x: 1000 }}
          />
        )}
      </Card>

      {/* Create/Edit Unit Modal */}
      <Modal
        title={editingUnit ? '编辑组织单元' : '新增组织单元'}
        open={isModalVisible}
        onCancel={handleModalClose}
        footer={null}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateUnit}
          initialValues={{
            status: 'ACTIVE'
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="单元名称"
                name="name"
                rules={[{ required: true, message: '请输入单元名称' }]}
              >
                <Input placeholder="如: 技术研发部" />
              </Form.Item>
            </Col>
            
            <Col span={12}>
              <Form.Item
                label="单元类型"
                name="unitType"
                rules={[{ required: true, message: '请选择单元类型' }]}
              >
                <Select placeholder="选择类型">
                  <Option value="COMPANY">公司</Option>
                  <Option value="DIVISION">事业部</Option>
                  <Option value="DEPARTMENT">部门</Option>
                  <Option value="TEAM">团队</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="上级单元"
                name="parentId"
              >
                <TreeSelect
                  placeholder="选择上级单元(可选)"
                  allowClear
                  treeDefaultExpandAll
                >
                  {orgUnits.map(unit => (
                    <TreeSelect.TreeNode 
                      key={unit.id} 
                      value={unit.id} 
                      title={unit.name}
                      disabled={editingUnit?.id === unit.id}
                    />
                  ))}
                </TreeSelect>
              </Form.Item>
            </Col>
            
            <Col span={12}>
              <Form.Item
                label="负责人"
                name="managerName"
              >
                <Input placeholder="负责人姓名(可选)" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="单元描述"
            name="description"
          >
            <Input.TextArea 
              rows={3}
              placeholder="描述该组织单元的职责和功能"
            />
          </Form.Item>

          {editingUnit && (
            <Form.Item
              label="状态"
              name="status"
            >
              <Select>
                <Option value="ACTIVE">活跃</Option>
                <Option value="INACTIVE">停用</Option>
              </Select>
            </Form.Item>
          )}

          <Form.Item style={{ marginTop: '24px', marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={handleModalClose}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingUnit ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default OrganizationChartPage;