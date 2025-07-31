// tests/unit/components/ServiceStatus.test.tsx
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import ServiceStatus from '@/components/ServiceStatus';

// 模拟 Ant Design 组件
jest.mock('antd', () => ({
  Alert: ({ children, message, description, type, showIcon, style, action }: any) => (
    <div data-testid="alert" data-type={type} style={style}>
      {message && <div>{message}</div>}
      {description && <div>{description}</div>}
      {action}
      {children}
    </div>
  ),
  Badge: ({ status, text, className, style }: any) => (
    <span data-testid="badge" data-status={status} className={className} style={style}>
      {text}
    </span>
  ),
  Button: ({ children, onClick, loading, size, icon }: any) => (
    <button onClick={onClick} data-testid="button" data-loading={loading} data-size={size}>
      {icon}
      {children}
    </button>
  ),
  Space: ({ children, direction, style }: any) => (
    <div data-testid="space" data-direction={direction} style={style}>
      {children}
    </div>
  ),
  Tooltip: ({ children, title }: any) => (
    <div data-testid="tooltip" title={title}>
      {children}
    </div>
  ),
}));

// 模拟 Ant Design 图标
jest.mock('@ant-design/icons', () => ({
  CheckCircleOutlined: () => <span data-testid="check-circle">✓</span>,
  CloseCircleOutlined: () => <span data-testid="close-circle">✗</span>,
  LoadingOutlined: () => <span data-testid="loading">⟳</span>,
  SyncOutlined: () => <span data-testid="sync">↻</span>,
}));

// 设置模拟
jest.mock('@/lib/graphql-client', () => ({
  apolloClient: {
    query: jest.fn(),
  },
}));

jest.mock('@/lib/rest-api-client', () => ({
  restApiClient: {
    healthCheck: jest.fn(),
  },
}));

describe('ServiceStatus组件', () => {
  const mockApolloClient = require('@/lib/graphql-client').apolloClient;
  const mockRestApiClient = require('@/lib/rest-api-client').restApiClient;

  beforeEach(() => {
    jest.clearAllMocks();
    
    // 设置默认的成功响应
    mockApolloClient.query.mockResolvedValue({
      data: { __typename: 'Query' }
    });
    
    mockRestApiClient.healthCheck.mockResolvedValue({
      success: true,
      data: { status: 'ok' }
    });
  });

  it('正确渲染简单徽章模式', async () => {
    render(<ServiceStatus showDetails={false} />);
    
    // 检查徽章是否存在
    await waitFor(() => {
      expect(screen.getByTestId('tooltip')).toBeInTheDocument();
    });
  });

  it('正确渲染详细模式', async () => {
    render(<ServiceStatus showDetails={true} />);
    
    // 检查详细信息是否显示
    await waitFor(() => {
      expect(screen.getByText('服务状态：')).toBeInTheDocument();
      expect(screen.getByText('刷新')).toBeInTheDocument();
    });
  });

  it('显示GraphQL和REST API状态', async () => {
    render(<ServiceStatus showDetails={true} />);
    
    await waitFor(() => {
      expect(screen.getByText(/GraphQL:/)).toBeInTheDocument();
      expect(screen.getByText(/REST API:/)).toBeInTheDocument();
    });
  });

  it('处理服务检查失败', async () => {
    // 模拟GraphQL失败
    mockApolloClient.query.mockRejectedValue(new Error('GraphQL Error'));
    
    // 模拟REST API失败
    mockRestApiClient.healthCheck.mockRejectedValue(new Error('REST Error'));
    
    render(<ServiceStatus showDetails={true} />);
    
    await waitFor(() => {
      // 应该显示错误警告
      expect(screen.getByTestId('alert')).toBeInTheDocument();
      expect(screen.getByText('服务连接异常')).toBeInTheDocument();
    });
  });

  it('显示上次检查时间', async () => {
    render(<ServiceStatus showDetails={true} />);
    
    await waitFor(() => {
      expect(screen.getByText(/上次检查:/)).toBeInTheDocument();
    });
  });

  it('支持自定义样式和类名', () => {
    const customStyle = { backgroundColor: 'red' };
    const customClassName = 'custom-service-status';
    
    render(
      <ServiceStatus 
        showDetails={false} 
        style={customStyle} 
        className={customClassName} 
      />
    );
    
    const badge = screen.getByTestId('badge');
    expect(badge).toHaveStyle('background-color: red');
    expect(badge).toHaveClass(customClassName);
  });
});