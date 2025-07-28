// src/components/GraphQLErrorBoundary.tsx
import React from 'react';
import { Alert, Button, Space } from 'antd';
import { ReloadOutlined, ExclamationCircleOutlined } from '@ant-design/icons';

interface GraphQLErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

interface GraphQLErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; retry: () => void }>;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

class GraphQLErrorBoundary extends React.Component<
  GraphQLErrorBoundaryProps,
  GraphQLErrorBoundaryState
> {
  constructor(props: GraphQLErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<GraphQLErrorBoundaryState> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // Call optional error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // Log to console for debugging
    console.error('GraphQL Error Boundary caught an error:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      const { fallback: FallbackComponent } = this.props;
      const { error } = this.state;

      // Use custom fallback component if provided
      if (FallbackComponent && error) {
        return <FallbackComponent error={error} retry={this.handleRetry} />;
      }

      // Default error UI
      return (
        <div style={{ padding: '24px' }}>
          <Alert
            message="GraphQL服务连接失败"
            description="无法连接到GraphQL服务器，正在尝试使用本地数据。如果问题持续存在，请联系系统管理员。"
            type="warning"
            icon={<ExclamationCircleOutlined />}
            action={
              <Space>
                <Button size="small" danger onClick={this.handleRetry}>
                  <ReloadOutlined />
                  重试
                </Button>
              </Space>
            }
            style={{ marginBottom: '16px' }}
          />
          
          {process.env.NODE_ENV === 'development' && (
            <Alert
              message="开发环境错误详情"
              description={
                <details>
                  <summary>点击查看错误详情</summary>
                  <pre style={{ marginTop: '8px', fontSize: '12px' }}>
                    {error?.toString()}
                    {this.state.errorInfo?.componentStack}
                  </pre>
                </details>
              }
              type="error"
              style={{ marginBottom: '16px' }}
            />
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

export default GraphQLErrorBoundary;