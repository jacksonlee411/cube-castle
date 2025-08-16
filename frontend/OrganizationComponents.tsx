import React, { useState, useEffect } from 'react';

// 7位编码组织单元类型定义
interface OrganizationUnit {
  code: string;
  name: string;
  unit_type: 'COMPANY' | 'DEPARTMENT' | 'PROJECT_TEAM' | 'COST_CENTER';
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  level: number;
  path: string;
  sort_order: number;
  parent_code?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface OrganizationListResponse {
  organizations: OrganizationUnit[];
  total_count: number;
  page: number;
  page_size: number;
}

// API客户端类
class OrganizationAPI {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
  }

  // 验证7位编码格式
  private validateCode(code: string): boolean {
    return /^[0-9]{7}$/.test(code);
  }

  // 获取组织单元列表
  async getAll(params?: {
    unit_type?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<OrganizationListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.unit_type) searchParams.set('unit_type', params.unit_type);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const response = await fetch(`${this.baseURL}/api/v1/organization-units?${searchParams}`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 通过编码获取单个组织单元
  async getByCode(code: string): Promise<OrganizationUnit> {
    if (!this.validateCode(code)) {
      throw new Error(`Invalid organization code: ${code}. Must be 7 digits.`);
    }

    const response = await fetch(`${this.baseURL}/api/v1/organization-units/${code}`);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Organization unit not found: ${code}`);
      }
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 获取统计信息
  async getStats(): Promise<{
    total_count: number;
    by_type: Record<string, number>;
    by_status: Record<string, number>;
  }> {
    const response = await fetch(`${this.baseURL}/api/v1/organization-units/stats`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 健康检查
  async healthCheck(): Promise<{
    status: string;
    timestamp: string;
    version: string;
    service: string;
  }> {
    const response = await fetch(`${this.baseURL}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    return response.json();
  }
}

// React Hook - 组织单元数据管理
export const useOrganizationUnits = (apiBaseURL?: string) => {
  const [api] = useState(() => new OrganizationAPI(apiBaseURL));
  const [organizations, setOrganizations] = useState<OrganizationUnit[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<{
    total_count: number;
    by_type: Record<string, number>;
    by_status: Record<string, number>;
  } | null>(null);

  // 获取组织列表
  const fetchOrganizations = async (params?: {
    unit_type?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.getAll(params);
      setOrganizations(response.organizations);
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 获取单个组织
  const fetchOrganizationByCode = async (code: string) => {
    setLoading(true);
    setError(null);
    try {
      const organization = await api.getByCode(code);
      return organization;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 获取统计信息
  const fetchStats = async () => {
    try {
      const statsData = await api.getStats();
      setStats(statsData);
      return statsData;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    }
  };

  // 健康检查
  const checkHealth = async () => {
    try {
      return await api.healthCheck();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    }
  };

  return {
    organizations,
    loading,
    error,
    stats,
    fetchOrganizations,
    fetchOrganizationByCode,
    fetchStats,
    checkHealth,
    api
  };
};

// React组件 - 组织单元选择器
export const OrganizationSelector: React.FC<{
  onSelect: (organization: OrganizationUnit) => void;
  filter?: { unit_type?: string; status?: string };
  placeholder?: string;
  apiBaseURL?: string;
}> = ({ onSelect, filter = {}, placeholder = "选择组织单元", apiBaseURL }) => {
  const { organizations, loading, error, fetchOrganizations } = useOrganizationUnits(apiBaseURL);
  const [selectedCode, setSelectedCode] = useState<string>('');

  useEffect(() => {
    fetchOrganizations(filter);
  }, [filter, fetchOrganizations]);

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const code = event.target.value;
    setSelectedCode(code);
    
    const selected = organizations.find(org => org.code === code);
    if (selected) {
      onSelect(selected);
    }
  };

  return (
    <div className="organization-selector">
      <select 
        value={selectedCode} 
        onChange={handleChange}
        disabled={loading}
        style={{
          padding: '8px 12px',
          border: '1px solid #ddd',
          borderRadius: '4px',
          fontSize: '14px',
          minWidth: '200px'
        }}
      >
        <option value="">{loading ? '加载中...' : placeholder}</option>
        {organizations.map(org => (
          <option key={org.code} value={org.code}>
            {org.code} - {org.name} ({org.unit_type})
          </option>
        ))}
      </select>
      {error && (
        <div style={{ color: 'red', fontSize: '12px', marginTop: '4px' }}>
          {error}
        </div>
      )}
    </div>
  );
};

// React组件 - 组织单元表格
export const OrganizationTable: React.FC<{
  filter?: { unit_type?: string; status?: string };
  onRowClick?: (organization: OrganizationUnit) => void;
  apiBaseURL?: string;
}> = ({ filter = {}, onRowClick, apiBaseURL }) => {
  const { organizations, loading, error, fetchOrganizations, stats, fetchStats } = useOrganizationUnits(apiBaseURL);

  useEffect(() => {
    fetchOrganizations(filter);
    fetchStats();
  }, [filter, fetchOrganizations, fetchStats]);

  if (loading) {
    return <div style={{ padding: '20px', textAlign: 'center' }}>加载中...</div>;
  }

  if (error) {
    return <div style={{ padding: '20px', color: 'red' }}>错误: {error}</div>;
  }

  return (
    <div className="organization-table">
      {stats && (
        <div style={{ marginBottom: '20px', padding: '10px', backgroundColor: '#f5f5f5', borderRadius: '4px' }}>
          <strong>统计信息:</strong> 总计 {stats.total_count} 个组织单元
        </div>
      )}
      
      <table style={{ width: '100%', borderCollapse: 'collapse', border: '1px solid #ddd' }}>
        <thead>
          <tr style={{ backgroundColor: '#f8f9fa' }}>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>编码</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>名称</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>类型</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>状态</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>层级</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>描述</th>
          </tr>
        </thead>
        <tbody>
          {organizations.map(org => (
            <tr 
              key={org.code}
              onClick={() => onRowClick?.(org)}
              style={{ 
                cursor: onRowClick ? 'pointer' : 'default',
                backgroundColor: onRowClick ? 'transparent' : undefined
              }}
              onMouseEnter={(e) => {
                if (onRowClick) e.currentTarget.style.backgroundColor = '#f8f9fa';
              }}
              onMouseLeave={(e) => {
                if (onRowClick) e.currentTarget.style.backgroundColor = 'transparent';
              }}
            >
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                <code style={{ backgroundColor: '#e9ecef', padding: '2px 4px', borderRadius: '2px' }}>
                  {org.code}
                </code>
              </td>
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>{org.name}</td>
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                <span style={{
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  backgroundColor: org.unit_type === 'COMPANY' ? '#e3f2fd' : 
                               org.unit_type === 'DEPARTMENT' ? '#f3e5f5' : '#fff3e0'
                }}>
                  {org.unit_type}
                </span>
              </td>
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                <span style={{
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  backgroundColor: org.status === 'ACTIVE' ? '#e8f5e8' : '#ffebee',
                  color: org.status === 'ACTIVE' ? '#2e7d32' : '#c62828'
                }}>
                  {org.status}
                </span>
              </td>
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>{org.level}</td>
              <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                {org.description || '-'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {organizations.length === 0 && (
        <div style={{ padding: '20px', textAlign: 'center', color: '#666' }}>
          暂无数据
        </div>
      )}
    </div>
  );
};

// 导出类型和组件
export type { OrganizationUnit, OrganizationListResponse };
export { OrganizationAPI };