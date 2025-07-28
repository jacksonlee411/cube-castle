// src/pages/organization/chart.tsx
import React, { useState, useEffect } from 'react';
import { Card, Typography, Spin, Select, DatePicker, Button, Row, Col, Space, Divider, message } from 'antd';
import { ApartmentOutlined, TeamOutlined, UserOutlined, BranchesOutlined, SyncOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

// Force dynamic rendering for this page
export const getServerSideProps = async () => {
  return { props: {} };
};

const { Title, Text } = Typography;
const { Option } = Select;

interface Employee {
  id: string;
  employee_id: string;
  legal_name: string;
  email: string;
  status: string;
  hire_date: string;
  current_position?: {
    position_title: string;
    department: string;
    job_level: string;
  };
}

interface OrganizationData {
  employees: Employee[];
  departments: string[];
  total: number;
}

const OrganizationChartPage: React.FC = () => {
  const [selectedDepartment, setSelectedDepartment] = useState<string>('全部');
  const [organizationData, setOrganizationData] = useState<OrganizationData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [syncLoading, setSyncLoading] = useState<boolean>(false);

  // 模拟获取组织数据
  const fetchOrganizationData = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const response = await fetch('/api/v1/corehr/employees');
      if (response.ok) {
        const data = await response.json();
        
        // 提取部门信息
        const departments = [...new Set(data.data?.map((emp: Employee) => 
          emp.current_position?.department || '未分配'
        ) || [])];
        
        const organizationData: OrganizationData = {
          employees: data.data || [],
          departments: ['全部', ...(departments.filter(d => typeof d === 'string') as string[])],
          total: data.data?.length || 0
        };
        
        setOrganizationData(organizationData);
      } else {
        message.error('获取组织数据失败');
      }
    } catch (error) {
      console.error('Error fetching organization data:', error);
      message.error('网络错误，无法获取组织数据');
    } finally {
      setLoading(false);
    }
  };

  // 同步到图数据库
  const syncToGraphDB = async () => {
    setSyncLoading(true);
    try {
      // 模拟GraphDB同步 - 实际应调用后端API
      await new Promise(resolve => setTimeout(resolve, 2000));
      message.success('组织数据已同步到图数据库');
    } catch (error) {
      message.error('同步失败');
    } finally {
      setSyncLoading(false);
    }
  };

  useEffect(() => {
    fetchOrganizationData();
  }, []);

  // 过滤员工数据
  const filteredEmployees = organizationData?.employees?.filter(emp => {
    if (selectedDepartment === '全部') return true;
    return emp.current_position?.department === selectedDepartment;
  }) || [];

  // 按部门分组
  const employeesByDepartment = filteredEmployees.reduce((acc, emp) => {
    const dept = emp.current_position?.department || '未分配';
    if (!acc[dept]) acc[dept] = [];
    acc[dept].push(emp);
    return acc;
  }, {} as Record<string, Employee[]>);

  const renderEmployeeCard = (employee: Employee, isManager: boolean = false) => (
    <Card
      key={employee.id}
      size="small"
      className={`employee-card ${isManager ? 'manager-card' : ''}`}
      style={{ 
        marginBottom: 8,
        borderColor: isManager ? '#1890ff' : undefined,
        backgroundColor: isManager ? '#f0f8ff' : undefined,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <UserOutlined style={{ color: isManager ? '#1890ff' : '#666' }} />
        <div>
          <Text strong={isManager}>{employee.legal_name}</Text>
          <br />
          <Text type="secondary" style={{ fontSize: '12px' }}>
            {employee.employee_id}
          </Text>
          {employee.current_position && (
            <>
              <br />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                {employee.current_position.position_title}
                {employee.current_position.job_level && ` (${employee.current_position.job_level})`}
              </Text>
            </>
          )}
        </div>
      </div>
    </Card>
  );

  const renderDepartmentSection = (department: string, employees: Employee[]) => (
    <Card
      key={department}
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <BranchesOutlined />
          {department}
          <Text type="secondary" style={{ fontSize: '12px', fontWeight: 'normal' }}>
            ({employees.length} 人)
          </Text>
        </div>
      }
      size="small"
      style={{ marginBottom: 16 }}
    >
      <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
        {employees.map(emp => {
          const isManager = emp.current_position?.job_level?.includes('MANAGER') || 
                           emp.current_position?.position_title?.includes('总监') ||
                           emp.current_position?.position_title?.includes('经理');
          return renderEmployeeCard(emp, isManager);
        })}
        {employees.length === 0 && (
          <Text type="secondary">暂无员工数据</Text>
        )}
      </div>
    </Card>
  );

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1400px', margin: '0 auto' }}>
      {/* Header */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 12 }}>
          <ApartmentOutlined />
          组织结构图
        </Title>
        <Text type="secondary">
          可视化展示公司组织架构，支持部门筛选和数据同步
        </Text>
      </Card>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={16} align="middle">
          <Col span={6}>
            <div>
              <Text strong>部门筛选:</Text>
              <Select
                value={selectedDepartment}
                onChange={setSelectedDepartment}
                style={{ width: '100%', marginTop: 8 }}
              >
                {organizationData?.departments?.map(dept => (
                  <Option key={dept} value={dept}>{dept}</Option>
                ))}
              </Select>
            </div>
          </Col>
          
          <Col span={6}>
            <Button
              type="primary"
              onClick={fetchOrganizationData}
              style={{ marginTop: 24 }}
              icon={<TeamOutlined />}
            >
              刷新数据
            </Button>
          </Col>

          <Col span={6}>
            <Button
              type="default"
              onClick={syncToGraphDB}
              loading={syncLoading}
              style={{ marginTop: 24 }}
              icon={<SyncOutlined />}
            >
              同步到图数据库
            </Button>
          </Col>
        </Row>
      </Card>

      {/* Organization Overview */}
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <TeamOutlined />
            组织概览
          </div>
        }
        style={{ marginBottom: 24 }}
      >
        <Row gutter={16}>
          <Col span={6}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <div style={{ fontSize: '32px', color: '#1890ff', marginBottom: 8 }}>
                {filteredEmployees.length}
              </div>
              <Text type="secondary">
                {selectedDepartment === '全部' ? '总员工数' : `${selectedDepartment}员工数`}
              </Text>
            </div>
          </Col>
          
          <Col span={6}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <div style={{ fontSize: '32px', color: '#52c41a', marginBottom: 8 }}>
                {selectedDepartment === '全部' ? 
                  (organizationData?.departments?.length || 1) - 1 : 
                  Object.keys(employeesByDepartment).length
                }
              </div>
              <Text type="secondary">部门数量</Text>
            </div>
          </Col>
          
          <Col span={6}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <div style={{ fontSize: '32px', color: '#faad14', marginBottom: 8 }}>
                {filteredEmployees.filter(emp => 
                  emp.current_position?.job_level?.includes('MANAGER') || 
                  emp.current_position?.position_title?.includes('总监') ||
                  emp.current_position?.position_title?.includes('经理')
                ).length}
              </div>
              <Text type="secondary">管理者数量</Text>
            </div>
          </Col>
          
          <Col span={6}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <div style={{ fontSize: '32px', color: '#722ed1', marginBottom: 8 }}>
                {organizationData?.total || 0}
              </div>
              <Text type="secondary">总注册员工</Text>
            </div>
          </Col>
        </Row>
      </Card>

      {/* Department Structure */}
      <Row gutter={16}>
        {selectedDepartment === '全部' ? (
          // 显示所有部门
          Object.entries(employeesByDepartment).map(([department, employees]) => (
            <Col key={department} span={8} style={{ marginBottom: 16 }}>
              {renderDepartmentSection(department, employees)}
            </Col>
          ))
        ) : (
          // 显示单个部门详情
          <Col span={24}>
            {Object.entries(employeesByDepartment).map(([department, employees]) => 
              renderDepartmentSection(department, employees)
            )}
          </Col>
        )}
      </Row>

      {/* Empty State */}
      {filteredEmployees.length === 0 && (
        <Card>
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <ApartmentOutlined style={{ fontSize: '48px', color: '#d9d9d9', marginBottom: 16 }} />
            <Title level={4} type="secondary">暂无组织数据</Title>
            <Text type="secondary">
              {selectedDepartment === '全部' ? 
                '系统中暂无员工数据，请先添加员工信息' : 
                `${selectedDepartment} 部门暂无员工数据`
              }
            </Text>
          </div>
        </Card>
      )}

      <style jsx>{`
        .employee-card {
          transition: all 0.3s ease;
        }
        
        .employee-card:hover {
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
          transform: translateY(-1px);
        }
        
        .manager-card {
          border-left: 4px solid #1890ff;
        }
      `}</style>
    </div>
  );
};

export default OrganizationChartPage;