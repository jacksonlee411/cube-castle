// src/pages/demo/error-boundary.tsx
import React, { useState } from 'react';
import { Card, Button, Space, Typography, Divider, Alert } from 'antd';
import { BugOutlined, ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons';
import withGraphQLErrorBoundary from '@/components/withGraphQLErrorBoundary';
import ServiceStatus from '@/components/ServiceStatus';
import { useEmployee } from '@/hooks/useEmployees';
import { GetServerSideProps } from 'next';

const { Title, Paragraph, Text } = Typography;

// Component that intentionally triggers GraphQL errors
const GraphQLErrorDemo: React.FC = () => {
  const [triggerError, setTriggerError] = useState(false);
  const [employeeId, setEmployeeId] = useState('test-employee-id');
  
  // This will trigger a GraphQL error since the backend doesn't have GraphQL endpoint
  const { employee, loading, error, refetch, isUsingRestFallback } = useEmployee(
    triggerError ? employeeId : ''
  );

  const handleTriggerError = () => {
    setTriggerError(true);
  };

  const handleReset = () => {
    setTriggerError(false);
  };

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Card style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BugOutlined style={{ marginRight: '8px' }} />
          GraphQL错误边界和REST API降级演示
        </Title>
        <Paragraph>
          本页面演示当GraphQL服务不可用时，系统如何自动切换到REST API作为备用数据源，
          确保用户界面保持可用状态。
        </Paragraph>
      </Card>

      {/* Service Status */}
      <Card title="服务状态监控" style={{ marginBottom: '24px' }}>
        <ServiceStatus showDetails={true} />
      </Card>

      {/* Error Demonstration */}
      <Card title="错误模拟和恢复" style={{ marginBottom: '24px' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Alert
            message="演示说明"
            description="由于后端GraphQL服务未启动，点击下面的按钮将触发GraphQL查询失败，系统会自动切换到REST API。"
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          
          <Space>
            <Button 
              type="primary" 
              icon={<ThunderboltOutlined />}
              onClick={handleTriggerError}
              disabled={triggerError}
            >
              触发GraphQL错误
            </Button>
            <Button 
              icon={<ReloadOutlined />}
              onClick={handleReset}
              disabled={!triggerError}
            >
              重置演示
            </Button>
          </Space>

          {triggerError && (
            <div style={{ marginTop: '16px' }}>
              <Divider />
              <Title level={4}>查询结果</Title>
              
              {loading && (
                <Alert message="正在尝试GraphQL查询..." type="info" showIcon />
              )}
              
              {error && (
                <Alert
                  message="GraphQL查询失败"
                  description={`错误信息: ${error.message}`}
                  type="error"
                  showIcon
                  style={{ marginBottom: '16px' }}
                />
              )}
              
              {isUsingRestFallback && (
                <Alert
                  message="已切换到REST API"
                  description="GraphQL服务不可用，系统已自动切换到REST API备用服务。"
                  type="success"
                  showIcon
                  style={{ marginBottom: '16px' }}
                />
              )}
              
              <div style={{ padding: '16px', background: '#f5f5f5', borderRadius: '4px' }}>
                <Text strong>查询状态:</Text>
                <ul>
                  <li>Loading: {loading ? '是' : '否'}</li>
                  <li>Has Error: {error ? '是' : '否'}</li>
                  <li>Using REST Fallback: {isUsingRestFallback ? '是' : '否'}</li>
                  <li>Employee Data: {employee ? '已获取' : '未获取'}</li>
                </ul>
              </div>
            </div>
          )}
        </Space>
      </Card>

      {/* Implementation Details */}
      <Card title="实现细节">
        <Space direction="vertical" style={{ width: '100%' }}>
          <Title level={4}>1. GraphQL错误边界 (GraphQLErrorBoundary)</Title>
          <Paragraph>
            使用React Error Boundary捕获GraphQL相关的运行时错误，防止整个应用崩溃。
          </Paragraph>
          
          <Title level={4}>2. REST API降级机制</Title>
          <Paragraph>
            在useEmployee Hook中实现了自动降级逻辑：
          </Paragraph>
          <ul>
            <li>首先尝试GraphQL查询</li>
            <li>如果GraphQL失败，自动切换到REST API</li>
            <li>保持相同的数据结构和用户体验</li>
            <li>提供重试机制</li>
          </ul>
          
          <Title level={4}>3. 服务状态监控</Title>
          <Paragraph>
            实时监控GraphQL和REST API的健康状态，为用户提供透明的服务状态信息。
          </Paragraph>
          
          <Title level={4}>4. 用户体验优化</Title>
          <ul>
            <li>友好的错误提示</li>
            <li>自动重试机制</li>
            <li>服务状态可视化</li>
            <li>开发环境详细错误信息</li>
          </ul>
        </Space>
      </Card>
    </div>
  );
};

// 禁用SSR预渲染，避免Apollo Client上下文错误
export const getServerSideProps: GetServerSideProps = async () => {
  return {
    props: {},
  }
}

export default withGraphQLErrorBoundary(GraphQLErrorDemo);