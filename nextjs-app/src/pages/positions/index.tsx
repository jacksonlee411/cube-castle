// src/pages/positions/index.tsx - Full CRUD functionality for UAT testing
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
  InputNumber,
  notification,
  Dropdown,
  Menu,
  Tooltip,
  Row,
  Col,
  Statistic,
  DatePicker
} from 'antd';
import { 
  PlusOutlined, 
  SearchOutlined, 
  MoreOutlined,
  BranchesOutlined,
  TeamOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  PieChartOutlined,
  UserOutlined,
  HomeOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/router';
import Link from 'next/link';
import dayjs from 'dayjs';

const { Search } = Input;
const { Option } = Select;

interface Position {
  id: string;
  positionType: 'FULL_TIME' | 'PART_TIME' | 'CONTINGENT_WORKER' | 'INTERN';
  jobProfileId: string;
  jobTitle: string;
  departmentId: string;
  departmentName: string;
  managerPositionId?: string;
  managerName?: string;
  status: 'OPEN' | 'FILLED' | 'FROZEN' | 'PENDING_ELIMINATION';
  budgetedFte: number;
  actualFte?: number;
  description?: string;
  requirements?: string;
  createdAt: string;
  updatedAt: string;
}

const PositionsPage: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [positions, setPositions] = useState<Position[]>([]);
  const [filteredPositions, setFilteredPositions] = useState<Position[]>([]);
  const [searchText, setSearchText] = useState('');
  const [departmentFilter, setDepartmentFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingPosition, setEditingPosition] = useState<Position | null>(null);
  const [form] = Form.useForm();

  // Sample position data with full CRUD capabilities
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      const samplePositions: Position[] = [
        {
          id: '1',
          positionType: 'FULL_TIME',
          jobProfileId: 'JP-001',
          jobTitle: '高级前端开发工程师',
          departmentId: 'dept-tech',
          departmentName: '技术研发部',
          managerPositionId: 'pos-manager-001',
          managerName: '技术总监',
          status: 'OPEN',
          budgetedFte: 1.0,
          actualFte: 0,
          description: '负责前端应用开发和架构设计',
          requirements: 'React/Vue.js经验3年以上，熟悉TypeScript',
          createdAt: '2024-01-15',
          updatedAt: '2024-01-15'
        },
        {
          id: '2',
          positionType: 'FULL_TIME',
          jobProfileId: 'JP-002',
          jobTitle: '后端开发工程师',
          departmentId: 'dept-tech',
          departmentName: '技术研发部',
          managerPositionId: 'pos-manager-001',
          managerName: '技术总监',
          status: 'FILLED',
          budgetedFte: 1.0,
          actualFte: 1.0,
          description: '负责后端服务开发和API设计',
          requirements: 'Go/Java经验2年以上，熟悉微服务架构',
          createdAt: '2024-01-10',
          updatedAt: '2024-01-20'
        },
        {
          id: '3',
          positionType: 'FULL_TIME',
          jobProfileId: 'JP-003',
          jobTitle: '产品经理',
          departmentId: 'dept-product',
          departmentName: '产品运营部',
          managerPositionId: 'pos-manager-002',
          managerName: '产品总监',
          status: 'FILLED',
          budgetedFte: 1.0,
          actualFte: 1.0,
          description: '负责产品规划和需求管理',
          requirements: '产品管理经验3年以上，有B端产品经验',
          createdAt: '2024-01-08',
          updatedAt: '2024-01-25'
        },
        {
          id: '4',
          positionType: 'PART_TIME',
          jobProfileId: 'JP-004',
          jobTitle: 'UI设计师',
          departmentId: 'dept-design',
          departmentName: '设计部',
          status: 'FROZEN',
          budgetedFte: 0.5,
          actualFte: 0,
          description: '负责用户界面设计和交互设计',
          requirements: '设计经验2年以上，熟悉Figma/Sketch',
          createdAt: '2024-01-12',
          updatedAt: '2024-01-28'
        },
        {
          id: '5',
          positionType: 'INTERN',
          jobProfileId: 'JP-005',
          jobTitle: '前端开发实习生',
          departmentId: 'dept-tech',
          departmentName: '技术研发部',
          managerPositionId: 'pos-manager-001',
          managerName: '技术总监',
          status: 'OPEN',
          budgetedFte: 1.0,
          actualFte: 0,
          description: '参与前端项目开发，学习最新技术',
          requirements: '计算机相关专业在读，有一定前端基础',
          createdAt: '2024-02-01',
          updatedAt: '2024-02-01'
        },
        {
          id: '6',
          positionType: 'CONTINGENT_WORKER',
          jobProfileId: 'JP-006',
          jobTitle: 'DevOps工程师',
          departmentId: 'dept-tech',
          departmentName: '技术研发部',
          status: 'PENDING_ELIMINATION',
          budgetedFte: 1.0,
          actualFte: 0,
          description: '负责CI/CD流水线和基础设施管理',
          requirements: 'DevOps经验2年以上，熟悉K8s/Docker',
          createdAt: '2023-12-15',
          updatedAt: '2024-01-30'
        }
      ];
      
      setPositions(samplePositions);
      setFilteredPositions(samplePositions);
      setLoading(false);
    }, 1000);
  }, []);

  // Filter positions
  useEffect(() => {
    let filtered = positions;

    if (searchText) {
      filtered = filtered.filter(pos => 
        pos.jobTitle.toLowerCase().includes(searchText.toLowerCase()) ||
        pos.jobProfileId.toLowerCase().includes(searchText.toLowerCase()) ||
        pos.departmentName.toLowerCase().includes(searchText.toLowerCase()) ||
        (pos.description && pos.description.toLowerCase().includes(searchText.toLowerCase()))
      );
    }

    if (departmentFilter) {
      filtered = filtered.filter(pos => pos.departmentId === departmentFilter);
    }

    if (statusFilter) {
      filtered = filtered.filter(pos => pos.status === statusFilter);
    }

    if (typeFilter) {
      filtered = filtered.filter(pos => pos.positionType === typeFilter);
    }

    setFilteredPositions(filtered);
  }, [positions, searchText, departmentFilter, statusFilter, typeFilter]);

  const handleSavePosition = async (values: any) => {
    try {
      setLoading(true);
      
      if (editingPosition) {
        // Update existing position
        const updatedPosition: Position = {
          ...editingPosition,
          jobProfileId: values.jobProfileId,
          jobTitle: values.jobTitle,
          departmentId: values.departmentId,
          departmentName: getDepartmentName(values.departmentId),
          positionType: values.positionType,
          status: values.status,
          budgetedFte: values.budgetedFte,
          managerName: values.managerName,
          description: values.description,
          requirements: values.requirements,
          updatedAt: new Date().toISOString().split('T')[0]
        };

        setPositions(prev => prev.map(pos => 
          pos.id === editingPosition.id ? updatedPosition : pos
        ));

        notification.success({
          message: '职位更新成功',
          description: `职位 ${values.jobTitle} 信息已更新。`,
        });
      } else {
        // Create new position
        const newPosition: Position = {
          id: Date.now().toString(),
          jobProfileId: values.jobProfileId,
          jobTitle: values.jobTitle,
          departmentId: values.departmentId,
          departmentName: getDepartmentName(values.departmentId),
          positionType: values.positionType,
          status: values.status || 'OPEN',
          budgetedFte: values.budgetedFte,
          actualFte: 0,
          managerName: values.managerName,
          description: values.description,
          requirements: values.requirements,
          createdAt: new Date().toISOString().split('T')[0],
          updatedAt: new Date().toISOString().split('T')[0]
        };

        setPositions(prev => [...prev, newPosition]);
        
        notification.success({
          message: '职位创建成功',
          description: `职位 ${values.jobTitle} 已成功添加到系统中。`,
        });
      }
      
      handleModalClose();
    } catch (error) {
      notification.error({
        message: editingPosition ? '职位更新失败' : '职位创建失败',
        description: '操作时发生错误，请重试。',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleEditPosition = (position: Position) => {
    setEditingPosition(position);
    form.setFieldsValue({
      jobProfileId: position.jobProfileId,
      jobTitle: position.jobTitle,
      departmentId: position.departmentId,
      positionType: position.positionType,
      status: position.status,
      budgetedFte: position.budgetedFte,
      managerName: position.managerName,
      description: position.description,
      requirements: position.requirements
    });
    setIsModalVisible(true);
  };

  const handleDeletePosition = (position: Position) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除职位 ${position.jobTitle} 吗？此操作不可撤销。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => {
        setPositions(prev => prev.filter(pos => pos.id !== position.id));
        notification.success({
          message: '职位删除成功',
          description: `职位 ${position.jobTitle} 已从系统中删除。`,
        });
      }
    });
  };

  const handleModalClose = () => {
    setIsModalVisible(false);
    setEditingPosition(null);
    form.resetFields();
  };

  const getDepartmentName = (departmentId: string) => {
    const departments = {
      'dept-tech': '技术研发部',
      'dept-product': '产品运营部',
      'dept-design': '设计部',
      'dept-hr': '人力资源部',
      'dept-finance': '财务部'
    };
    return departments[departmentId as keyof typeof departments] || departmentId;
  };

  const getStatusColor = (status: string) => {
    const colors = {
      OPEN: 'blue',
      FILLED: 'green',
      FROZEN: 'orange',
      PENDING_ELIMINATION: 'red'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusLabel = (status: string) => {
    const labels = {
      OPEN: '空缺',
      FILLED: '已填补',
      FROZEN: '冻结',
      PENDING_ELIMINATION: '待裁撤'
    };
    return labels[status as keyof typeof labels] || status;
  };

  const getTypeLabel = (type: string) => {
    const labels = {
      FULL_TIME: '全职',
      PART_TIME: '兼职',
      CONTINGENT_WORKER: '临时工',
      INTERN: '实习生'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getTypeColor = (type: string) => {
    const colors = {
      FULL_TIME: 'blue',
      PART_TIME: 'green',
      CONTINGENT_WORKER: 'orange',
      INTERN: 'purple'
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getActionMenu = (position: Position) => (
    <Menu>
      <Menu.Item 
        key="view" 
        icon={<EyeOutlined />}
        onClick={() => router.push(`/positions/${position.id}`)}
      >
        查看详情
      </Menu.Item>
      <Menu.Item 
        key="edit" 
        icon={<EditOutlined />}
        onClick={() => handleEditPosition(position)}
      >
        编辑职位
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item 
        key="delete" 
        icon={<DeleteOutlined />}
        onClick={() => handleDeletePosition(position)}
        style={{ color: '#ff4d4f' }}
      >
        删除职位
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: '职位信息',
      key: 'position',
      render: (record: Position) => (
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
            {record.positionType === 'FULL_TIME' ? '💼' : 
             record.positionType === 'PART_TIME' ? '⏰' :
             record.positionType === 'INTERN' ? '🎓' : '🔧'}
          </div>
          <div>
            <div style={{ fontWeight: 'bold', marginBottom: '2px' }}>
              <Link href={`/positions/${record.id}`} style={{ color: 'inherit' }}>
                {record.jobTitle}
              </Link>
            </div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              <Space size="small">
                <span>{record.jobProfileId}</span>
                <span>•</span>
                <span>FTE: {record.budgetedFte}</span>
              </Space>
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'positionType',
      key: 'positionType',
      render: (type: string) => (
        <Tag color={getTypeColor(type)}>{getTypeLabel(type)}</Tag>
      ),
      filters: [
        { text: '全职', value: 'FULL_TIME' },
        { text: '兼职', value: 'PART_TIME' },
        { text: '临时工', value: 'CONTINGENT_WORKER' },
        { text: '实习生', value: 'INTERN' },
      ],
    },
    {
      title: '部门',
      dataIndex: 'departmentName',
      key: 'departmentName',
      render: (departmentName: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <TeamOutlined style={{ color: '#1890ff' }} />
          <span>{departmentName}</span>
        </div>
      ),
    },
    {
      title: '上级职位',
      dataIndex: 'managerName',
      key: 'managerName',
      render: (managerName?: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          {managerName ? (
            <>
              <BranchesOutlined style={{ color: '#52c41a' }} />
              <span>{managerName}</span>
            </>
          ) : (
            <span style={{ color: '#999' }}>无</span>
          )}
        </div>
      ),
    },
    {
      title: 'FTE使用率',
      key: 'fteUtilization',
      render: (record: Position) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <UserOutlined style={{ color: record.actualFte === record.budgetedFte ? '#52c41a' : '#1890ff' }} />
          <span>{record.actualFte || 0} / {record.budgetedFte}</span>
          <span style={{ 
            color: record.actualFte === record.budgetedFte ? '#52c41a' : '#666',
            fontSize: '12px'
          }}>
            ({Math.round(((record.actualFte || 0) / record.budgetedFte) * 100)}%)
          </span>
        </div>
      ),
      sorter: (a: Position, b: Position) => 
        ((a.actualFte || 0) / a.budgetedFte) - ((b.actualFte || 0) / b.budgetedFte),
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
        { text: '空缺', value: 'OPEN' },
        { text: '已填补', value: 'FILLED' },
        { text: '冻结', value: 'FROZEN' },
        { text: '待裁撤', value: 'PENDING_ELIMINATION' },
      ],
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      sorter: (a: Position, b: Position) => 
        new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (record: Position) => (
        <Space>
          <Tooltip title="编辑职位">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEditPosition(record)}
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
  const totalPositions = positions.length;
  const openPositions = positions.filter(pos => pos.status === 'OPEN').length;
  const filledPositions = positions.filter(pos => pos.status === 'FILLED').length;
  const frozenPositions = positions.filter(pos => pos.status === 'FROZEN').length;
  const totalBudgetedFte = positions.reduce((sum, pos) => sum + pos.budgetedFte, 0);
  const totalActualFte = positions.reduce((sum, pos) => sum + (pos.actualFte || 0), 0);
  const utilizationRate = totalBudgetedFte > 0 ? (totalActualFte / totalBudgetedFte) * 100 : 0;

  const departments = Array.from(new Set(positions.map(pos => ({ id: pos.departmentId, name: pos.departmentName }))));

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 'bold' }}>职位管理</h1>
          <p style={{ margin: '4px 0 0 0', color: '#666' }}>
            管理组织职位结构、层级关系和FTE预算 - 完整CRUD功能
          </p>
        </div>
        <Space>
          <Button 
            icon={<ReloadOutlined />}
            onClick={() => window.location.reload()}
            loading={loading}
          >
            刷新
          </Button>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            size="large"
            onClick={() => setIsModalVisible(true)}
          >
            新建职位
          </Button>
        </Space>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={4}>
          <Card>
            <Statistic
              title="总职位数"
              value={totalPositions}
              prefix={<BranchesOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="空缺职位"
              value={openPositions}
              valueStyle={{ color: '#1890ff' }}
              prefix={<PieChartOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="已填补职位"
              value={filledPositions}
              valueStyle={{ color: '#52c41a' }}
              prefix={<TeamOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="冻结职位"
              value={frozenPositions}
              valueStyle={{ color: '#faad14' }}
              prefix={<HomeOutlined style={{ color: '#faad14' }} />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="预算FTE"
              value={totalBudgetedFte}
              precision={1}
              prefix={<UserOutlined style={{ color: '#722ed1' }} />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="利用率"
              value={utilizationRate}
              precision={1}
              suffix="%"
              valueStyle={{ color: utilizationRate >= 80 ? '#52c41a' : '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Card style={{ marginBottom: '24px' }}>
        <div style={{ display: 'flex', gap: '16px', alignItems: 'center', flexWrap: 'wrap' }}>
          <Search
            placeholder="搜索职位名称、工作配置或部门"
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
              <Option key={dept.id} value={dept.id}>{dept.name}</Option>
            ))}
          </Select>
          
          <Select
            placeholder="选择状态"
            style={{ width: '120px' }}
            value={statusFilter}
            onChange={setStatusFilter}
            allowClear
          >
            <Option value="OPEN">空缺</Option>
            <Option value="FILLED">已填补</Option>
            <Option value="FROZEN">冻结</Option>
            <Option value="PENDING_ELIMINATION">待裁撤</Option>
          </Select>

          <Select
            placeholder="选择类型"
            style={{ width: '120px' }}
            value={typeFilter}
            onChange={setTypeFilter}
            allowClear
          >
            <Option value="FULL_TIME">全职</Option>
            <Option value="PART_TIME">兼职</Option>
            <Option value="CONTINGENT_WORKER">临时工</Option>
            <Option value="INTERN">实习生</Option>
          </Select>
          
          <div style={{ color: '#666', fontSize: '14px' }}>
            共找到 {filteredPositions.length} 个职位
          </div>
        </div>
      </Card>

      {/* Positions Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={filteredPositions}
          rowKey="id"
          loading={loading}
          pagination={{
            total: filteredPositions.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Create/Edit Position Modal */}
      <Modal
        title={editingPosition ? '编辑职位' : '新建职位'}
        open={isModalVisible}  
        onCancel={handleModalClose}
        footer={null}
        width={700}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSavePosition}
          initialValues={{
            status: 'OPEN',
            budgetedFte: 1.0
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="jobProfileId"
                label="工作配置ID"
                rules={[{ required: true, message: '请输入工作配置ID' }]}
              >
                <Input placeholder="如: JP-001" />
              </Form.Item>
            </Col>
            
            <Col span={12}>
              <Form.Item
                name="jobTitle"
                label="职位名称"
                rules={[{ required: true, message: '请输入职位名称' }]}
              >
                <Input placeholder="如: 高级前端开发工程师" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="positionType"
                label="职位类型"
                rules={[{ required: true, message: '请选择职位类型' }]}
              >
                <Select placeholder="选择职位类型" disabled={!!editingPosition}>
                  <Option value="FULL_TIME">全职</Option>
                  <Option value="PART_TIME">兼职</Option>
                  <Option value="CONTINGENT_WORKER">临时工</Option>
                  <Option value="INTERN">实习生</Option>
                </Select>
              </Form.Item>
            </Col>

            <Col span={12}>
              <Form.Item
                name="status"
                label="职位状态"
                rules={[{ required: true, message: '请选择职位状态' }]}
              >
                <Select placeholder="选择职位状态">
                  <Option value="OPEN">空缺</Option>
                  <Option value="FILLED">已填补</Option>
                  <Option value="FROZEN">冻结</Option>
                  <Option value="PENDING_ELIMINATION">待裁撤</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="departmentId"
                label="所属部门"
                rules={[{ required: true, message: '请选择所属部门' }]}
              >
                <Select placeholder="选择所属部门">
                  <Option value="dept-tech">技术研发部</Option>
                  <Option value="dept-product">产品运营部</Option>
                  <Option value="dept-design">设计部</Option>
                  <Option value="dept-hr">人力资源部</Option>
                  <Option value="dept-finance">财务部</Option>
                </Select>
              </Form.Item>
            </Col>
            
            <Col span={12}>
              <Form.Item
                name="budgetedFte"
                label="预算FTE"
                rules={[
                  { required: true, message: '请输入预算FTE' },
                  { type: 'number', min: 0.1, max: 5.0, message: 'FTE值必须在0.1-5.0范围内' }
                ]}
              >
                <InputNumber
                  min={0.1}
                  max={5.0}
                  step={0.1}
                  placeholder="1.0"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="managerName"
            label="上级职位/经理"
          >
            <Input placeholder="上级职位或经理姓名(可选)" />
          </Form.Item>

          <Form.Item
            name="description"
            label="职位描述"
          >
            <Input.TextArea 
              rows={3}
              placeholder="描述该职位的主要职责和工作内容"
            />
          </Form.Item>

          <Form.Item
            name="requirements"
            label="任职要求"
          >
            <Input.TextArea 
              rows={3}
              placeholder="描述该职位的技能要求和任职条件"
            />
          </Form.Item>

          <Form.Item style={{ marginTop: '24px', marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={handleModalClose}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingPosition ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PositionsPage;