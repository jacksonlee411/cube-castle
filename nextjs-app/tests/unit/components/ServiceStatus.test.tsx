// tests/unit/components/ServiceStatus.test.tsx
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import ServiceStatus from '@/components/ServiceStatus';

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
  });

  test('renders loading state initially', () => {
    render(<ServiceStatus />);
    
    expect(screen.getByText(/服务状态/i)).toBeInTheDocument();
  });

  test('shows healthy status when all services are healthy', async () => {
    mockApolloClient.query.mockResolvedValue({
      data: { __typename: 'Query' }
    });
    
    mockRestApiClient.healthCheck.mockResolvedValue({
      success: true,
      data: { status: 'healthy' }
    });

    render(<ServiceStatus />);

    await waitFor(() => {
      expect(screen.getByText(/服务状态/i)).toBeInTheDocument();
    });
  });

  test('shows error status when GraphQL service fails', async () => {
    mockApolloClient.query.mockRejectedValue(new Error('GraphQL Error'));
    
    mockRestApiClient.healthCheck.mockResolvedValue({
      success: true,
      data: { status: 'healthy' }
    });

    render(<ServiceStatus />);

    await waitFor(() => {
      expect(screen.getByText(/服务状态/i)).toBeInTheDocument();
    });
  });

  test('shows error status when REST API service fails', async () => {
    mockApolloClient.query.mockResolvedValue({
      data: { __typename: 'Query' }
    });
    
    mockRestApiClient.healthCheck.mockRejectedValue(new Error('REST API Error'));

    render(<ServiceStatus />);

    await waitFor(() => {
      expect(screen.getByText(/服务状态/i)).toBeInTheDocument();
    });
  });
});