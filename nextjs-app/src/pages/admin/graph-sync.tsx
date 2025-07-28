// src/pages/admin/graph-sync.tsx
import React, { useState } from 'react';
import { Card, Typography, Button, Row, Col, Progress, Alert, Table, Space, Tag, notification, Modal } from 'antd';
import { SyncOutlined, DatabaseOutlined, CheckCircleOutlined, CloseCircleOutlined, WarningOutlined } from '@ant-design/icons';
import { useMutation } from '@apollo/client';
import { FULL_GRAPH_SYNC, SYNC_DEPARTMENT, SYNC_EMPLOYEE_TO_GRAPH } from '@/lib/graphql-queries';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

interface SyncResult {
  success: boolean;
  syncedEmployees: number;
  syncedPositions: number;
  syncedRelationships: number;
  errors: string[];
}

const GraphSyncAdminPage: React.FC = () => {
  const [syncStatus, setSyncStatus] = useState<'idle' | 'running' | 'completed' | 'error'>('idle');
  const [syncResults, setSyncResults] = useState<SyncResult | null>(null);
  const [lastSyncTime, setLastSyncTime] = useState<string | null>(null);

  // GraphQL mutations
  const [fullGraphSync, { loading: fullSyncLoading }] = useMutation(FULL_GRAPH_SYNC);
  const [syncDepartment, { loading: deptSyncLoading }] = useMutation(SYNC_DEPARTMENT);
  const [syncEmployee, { loading: empSyncLoading }] = useMutation(SYNC_EMPLOYEE_TO_GRAPH);

  const handleFullSync = async () => {
    setSyncStatus('running');
    setSyncResults(null);

    try {
      const result = await fullGraphSync();
      const data = result.data?.fullGraphSync;

      if (data) {
        setSyncResults(data);
        setSyncStatus(data.success ? 'completed' : 'error');
        setLastSyncTime(dayjs().format('YYYY-MM-DD HH:mm:ss'));

        if (data.success) {
          notification.success({
            message: '完整同步成功',
            description: `已同步 ${data.syncedEmployees} 个员工，${data.syncedPositions} 个职位，${data.syncedRelationships} 个关系`,
          });
        } else {
          notification.error({
            message: '同步失败',
            description: `同步过程中遇到 ${data.errors.length} 个错误`,
          });
        }
      }
    } catch (error: any) {
      setSyncStatus('error');
      notification.error({
        message: '同步失败',
        description: error.message || '同步过程中发生未知错误',
      });
    }
  };

  const handleDepartmentSync = async (department: string) => {
    try {
      const result = await syncDepartment({
        variables: { department },
      });

      const data = result.data?.syncDepartment;
      if (data?.success) {
        notification.success({
          message: '部门同步成功',
          description: `${department} 部门已成功同步到图数据库`,
        });
      } else {
        notification.error({
          message: '部门同步失败',
          description: `${department} 部门同步失败: ${data?.errors?.join(', ')}`,
        });
      }
    } catch (error: any) {
      notification.error({
        message: '部门同步失败',
        description: error.message || '同步过程中发生未知错误',
      });
    }
  };

  const handleTestSync = async () => {
    try {
      const result = await syncEmployee({
        variables: { employeeId: 'test-employee-001' },
      });

      if (result.data?.syncEmployeeToGraph) {
        notification.success({
          message: '测试同步成功',
          description: '测试员工数据已成功同步到图数据库',
        });
      }
    } catch (error: any) {
      notification.error({
        message: '测试同步失败',
        description: error.message || '测试过程中发生错误',
      });
    }
  };

  const getSyncStatusIcon = () => {
    switch (syncStatus) {
      case 'running':
        return <SyncOutlined spin style={{ color: '#1890ff' }} />;
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'error':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <DatabaseOutlined style={{ color: '#666' }} />;
    }
  };

  const getSyncStatusText = () => {
    switch (syncStatus) {
      case 'running':
        return '同步进行中...';
      case 'completed':
        return '同步已完成';
      case 'error':
        return '同步失败';
      default:
        return '等待同步';
    }
  };

  const departmentColumns = [
    {
      title: '部门名称',
      dataIndex: 'department',
      key: 'department',
    },
    {
      title: '员工数量',
      dataIndex: 'employeeCount',
      key: 'employeeCount',
    },
    {
      title: '同步状态',
      dataIndex: 'syncStatus',
      key: 'syncStatus',
      render: (status: string) => {
        const config = {
          synced: { color: 'green', text: '已同步' },
          pending: { color: 'orange', text: '待同步' },
          error: { color: 'red', text: '同步失败' },
        };
        const { color, text } = config[status as keyof typeof config] || config.pending;
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '最后同步时间',
      dataIndex: 'lastSync',
      key: 'lastSync',
      render: (time: string | null) => time || '从未同步',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: any) => (
        <Button
          size="small"
          onClick={() => handleDepartmentSync(record.department)}
          loading={deptSyncLoading}
        >
          同步部门
        </Button>
      ),
    },
  ];

  const departmentData = [
    {
      key: '1',
      department: '技术部',
      employeeCount: 45,
      syncStatus: 'synced',
      lastSync: '2024-07-27 10:30:00',
    },
    {
      key: '2',
      department: '产品部',
      employeeCount: 28,
      syncStatus: 'pending',
      lastSync: null,
    },
    {
      key: '3',
      department: '销售部',
      employeeCount: 32,
      syncStatus: 'synced',
      lastSync: '2024-07-27 09:15:00',
    },
    {
      key: '4',
      department: '人力资源部',
      employeeCount: 12,
      syncStatus: 'error',
      lastSync: '2024-07-26 16:45:00',
    },
  ];

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Header */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 12 }}>
          <DatabaseOutlined />
          图数据库同步管理
        </Title>
        <Text type="secondary">
          管理PostgreSQL与Neo4j图数据库之间的数据同步，确保组织结构数据的一致性
        </Text>
      </Card>

      {/* Sync Status Overview */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '48px', marginBottom: 8 }}>
                {getSyncStatusIcon()}
              </div>
              <Text strong>{getSyncStatusText()}</Text>
              {lastSyncTime && (
                <>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    最后同步: {lastSyncTime}
                  </Text>
                </>
              )}
            </div>
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#1890ff', marginBottom: 8 }}>
                {syncResults?.syncedEmployees || 0}
              </div>
              <Text type="secondary">已同步员工</Text>
            </div>
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#52c41a', marginBottom: 8 }}>
                {syncResults?.syncedPositions || 0}
              </div>
              <Text type="secondary">已同步职位</Text>
            </div>
          </Card>
        </Col>

        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '32px', color: '#faad14', marginBottom: 8 }}>
                {syncResults?.syncedRelationships || 0}
              </div>
              <Text type="secondary">已同步关系</Text>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Sync Controls */}
      <Card title="同步控制" style={{ marginBottom: 24 }}>
        <Space size="large">
          <Button
            type="primary"
            size="large"
            icon={<SyncOutlined />}
            onClick={handleFullSync}
            loading={fullSyncLoading || syncStatus === 'running'}
          >
            完整同步
          </Button>

          <Button
            size="large"
            onClick={handleTestSync}
            loading={empSyncLoading}
          >
            测试同步
          </Button>

          <Button
            size="large"
            onClick={() => {
              Modal.confirm({
                title: '确认重建图数据库',
                content: '此操作将清空图数据库并重新同步所有数据，是否继续？',
                okText: '确认',
                cancelText: '取消',
                onOk: handleFullSync,
              });
            }}
            danger
          >
            重建图数据库
          </Button>
        </Space>

        {syncStatus === 'running' && (
          <div style={{ marginTop: 16 }}>
            <Progress percent={75} status="active" showInfo={false} />
            <Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
              正在同步组织数据到图数据库...
            </Text>
          </div>
        )}
      </Card>

      {/* Sync Results */}
      {syncResults && (
        <Card title="同步结果" style={{ marginBottom: 24 }}>
          {syncResults.success ? (
            <Alert
              message="同步成功完成"
              description={
                <div>
                  <p>已成功同步以下数据到图数据库:</p>
                  <ul>
                    <li>员工记录: {syncResults.syncedEmployees} 条</li>
                    <li>职位记录: {syncResults.syncedPositions} 条</li>
                    <li>汇报关系: {syncResults.syncedRelationships} 条</li>
                  </ul>
                </div>
              }
              type="success"
              showIcon
            />
          ) : (
            <Alert
              message="同步过程中发生错误"
              description={
                <div>
                  <p>以下错误导致同步失败:</p>
                  <ul>
                    {syncResults.errors.map((error, index) => (
                      <li key={index}>{error}</li>
                    ))}
                  </ul>
                </div>
              }
              type="error"
              showIcon
            />
          )}
        </Card>
      )}

      {/* Department Sync Status */}
      <Card title="部门同步状态">
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          查看各部门的同步状态，支持单独同步特定部门的数据。
        </Paragraph>

        <Table
          columns={departmentColumns}
          dataSource={departmentData}
          pagination={false}
          size="middle"
        />
      </Card>

      {/* Help Information */}
      <Card title="同步说明" style={{ marginTop: 24 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Title level={4}>同步流程</Title>
            <ol>
              <li>从PostgreSQL读取员工和职位数据</li>
              <li>转换为图数据库节点和关系</li>
              <li>创建或更新Neo4j中的数据</li>
              <li>建立汇报关系和部门归属</li>
              <li>验证数据完整性和一致性</li>
            </ol>
          </Col>

          <Col span={12}>
            <Title level={4}>注意事项</Title>
            <ul>
              <li>完整同步可能需要几分钟时间</li>
              <li>同步期间图查询功能可能受影响</li>
              <li>建议在业务低峰期执行完整同步</li>
              <li>同步失败时会自动回滚变更</li>
              <li>定期检查同步状态确保数据一致性</li>
            </ul>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default GraphSyncAdminPage;