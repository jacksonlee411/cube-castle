// src/components/withGraphQLErrorBoundary.tsx
import React from 'react';
import GraphQLErrorBoundary from './GraphQLErrorBoundary';
import { Alert, Button, Card, Space } from 'antd';
import { ReloadOutlined, HomeOutlined } from '@ant-design/icons';
import { useRouter } from 'next/router';

interface ErrorFallbackProps {
  error: Error;
  retry: () => void;
}

const DefaultErrorFallback: React.FC<ErrorFallbackProps> = ({ error, retry }) => {
  const router = useRouter();

  const handleGoHome = () => {
    router.push('/');
  };

  return (
    <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
      <Card>
        <Alert
          message="GraphQL服务连接失败"
          description="系统正在尝试使用备用数据源，某些功能可能受限。"
          type="warning"
          showIcon
          style={{ marginBottom: '24px' }}
        />
        
        <div style={{ textAlign: 'center' }}>
          <h3>服务暂时不可用</h3>
          <p style={{ color: '#666', marginBottom: '24px' }}>
            我们正在努力恢复服务。您可以尝试刷新页面或返回首页。
          </p>
          
          <Space>
            <Button type="primary" icon={<ReloadOutlined />} onClick={retry}>
              重试连接
            </Button>
            <Button icon={<HomeOutlined />} onClick={handleGoHome}>
              返回首页
            </Button>
          </Space>
        </div>

        {process.env.NODE_ENV === 'development' && (
          <details style={{ marginTop: '24px' }}>
            <summary style={{ cursor: 'pointer', fontWeight: 'bold' }}>
              开发环境错误详情
            </summary>
            <pre style={{ 
              marginTop: '12px', 
              padding: '12px', 
              background: '#f5f5f5', 
              borderRadius: '4px',
              fontSize: '12px',
              overflow: 'auto'
            }}>
              {error.toString()}
            </pre>
          </details>
        )}
      </Card>
    </div>
  );
};

// Higher-order component to wrap pages with GraphQL error boundary
function withGraphQLErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  customFallback?: React.ComponentType<ErrorFallbackProps>
) {
  const WithErrorBoundaryComponent = (props: P) => {
    return (
      <GraphQLErrorBoundary 
        fallback={customFallback || DefaultErrorFallback}
        onError={(error, errorInfo) => {
          // GraphQL Error Boundary error logged - integrating with monitoring services
          
          // You can integrate with error reporting services here
          // Example: Sentry, LogRocket, etc.
          if (typeof window !== 'undefined' && (window as any).gtag) {
            (window as any).gtag('event', 'exception', {
              description: error.toString(),
              fatal: false,
            });
          }
        }}
      >
        <WrappedComponent {...props} />
      </GraphQLErrorBoundary>
    );
  };

  WithErrorBoundaryComponent.displayName = `withGraphQLErrorBoundary(${WrappedComponent.displayName || WrappedComponent.name})`;

  return WithErrorBoundaryComponent;
}

export default withGraphQLErrorBoundary;