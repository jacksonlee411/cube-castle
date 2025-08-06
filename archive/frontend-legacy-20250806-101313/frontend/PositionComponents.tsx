// 职位管理前端组件 - 7位编码优化版
// 版本: v1.0 Optimized
// 创建日期: 2025-08-05
// 基于: 7位编码职位管理API成功实现
// 架构: React + TypeScript + 零转换编码系统

import React, { useState, useEffect } from 'react';

// 7位编码职位类型定义
interface Position {
  code: string;
  organization_code: string;
  manager_position_code?: string;
  position_type: 'FULL_TIME' | 'PART_TIME' | 'CONTINGENT_WORKER' | 'INTERN';
  job_profile_id: string;
  status: 'OPEN' | 'FILLED' | 'FROZEN' | 'PENDING_ELIMINATION';
  budgeted_fte: number;
  details?: string;
  tenant_id: string;
  created_at: string;
  updated_at: string;
}

interface PositionWithRelations extends Position {
  organization?: {
    code: string;
    name: string;
    unit_type: string;
  };
  manager_position?: {
    code: string;
    position_type: string;
    status: string;
  };
  direct_reports?: Array<{
    code: string;
    position_type: string;
    status: string;
  }>;
  incumbents?: Array<{
    code: string;
    first_name: string;
    last_name: string;
    email: string;
  }>;
}

interface PositionListResponse {
  positions: Position[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

interface PositionStats {
  total_positions: number;
  total_budgeted_fte: number;
  by_type: Record<string, number>;
  by_status: Record<string, number>;
}

// API客户端类 - 7位编码专用
class PositionAPI {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8082') {
    this.baseURL = baseURL;
  }

  // 验证7位职位编码格式
  private validatePositionCode(code: string): boolean {
    return /^[0-9]{7}$/.test(code) && 
           parseInt(code) >= 1000000 && 
           parseInt(code) <= 9999999;
  }

  // 验证7位组织编码格式
  private validateOrganizationCode(code: string): boolean {
    return /^[0-9]{7}$/.test(code) && 
           parseInt(code) >= 1000000 && 
           parseInt(code) <= 9999999;
  }

  // 获取职位列表
  async getAll(params?: {
    position_type?: string;
    status?: string;
    organization_code?: string;
    page?: number;
    page_size?: number;
  }): Promise<PositionListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.position_type) searchParams.set('position_type', params.position_type);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.organization_code) searchParams.set('organization_code', params.organization_code);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());

    const response = await fetch(`${this.baseURL}/api/v1/positions?${searchParams}`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 通过7位编码获取职位
  async getByCode(code: string, options?: {
    with_organization?: boolean;
    with_manager?: boolean;
    with_direct_reports?: boolean;
    with_incumbents?: boolean;
  }): Promise<PositionWithRelations> {
    if (!this.validatePositionCode(code)) {
      throw new Error(`Invalid position code: ${code}. Must be 7 digits (1000000-9999999).`);
    }

    const searchParams = new URLSearchParams();
    if (options?.with_organization) searchParams.set('with_organization', 'true');
    if (options?.with_manager) searchParams.set('with_manager', 'true');
    if (options?.with_direct_reports) searchParams.set('with_direct_reports', 'true');
    if (options?.with_incumbents) searchParams.set('with_incumbents', 'true');

    const response = await fetch(`${this.baseURL}/api/v1/positions/${code}?${searchParams}`);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Position not found: ${code}`);
      }
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 创建职位
  async create(position: {
    organization_code: string;
    manager_position_code?: string;
    position_type: string;
    job_profile_id: string;
    status?: string;
    budgeted_fte: number;
    details?: Record<string, any>;
  }): Promise<Position> {
    if (!this.validateOrganizationCode(position.organization_code)) {
      throw new Error('Invalid organization code: must be 7 digits (1000000-9999999)');
    }

    if (position.manager_position_code && !this.validatePositionCode(position.manager_position_code)) {
      throw new Error('Invalid manager position code: must be 7 digits (1000000-9999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/positions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(position),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
    return response.json();
  }

  // 更新职位
  async update(code: string, updates: {
    organization_code?: string;
    manager_position_code?: string;
    status?: string;
    budgeted_fte?: number;
    details?: Record<string, any>;
  }): Promise<Position> {
    if (!this.validatePositionCode(code)) {
      throw new Error('Invalid position code: must be 7 digits (1000000-9999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/positions/${code}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
    return response.json();
  }

  // 删除职位
  async delete(code: string): Promise<void> {
    if (!this.validatePositionCode(code)) {
      throw new Error('Invalid position code: must be 7 digits (1000000-9999999)');
    }

    const response = await fetch(`${this.baseURL}/api/v1/positions/${code}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API error: ${response.status} ${errorText}`);
    }
  }

  // 获取统计信息
  async getStats(): Promise<PositionStats> {
    const response = await fetch(`${this.baseURL}/api/v1/positions/stats`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  // 健康检查
  async healthCheck(): Promise<{
    status: string;
    timestamp: string;
    service: string;
    version: string;
    features: string[];
  }> {
    const response = await fetch(`${this.baseURL}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    return response.json();
  }
}

// React Hook - 职位数据管理
export const usePositions = (apiBaseURL?: string) => {
  const [api] = useState(() => new PositionAPI(apiBaseURL));
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<PositionStats | null>(null);

  // 获取职位列表
  const fetchPositions = async (params?: {
    position_type?: string;
    status?: string;
    organization_code?: string;
    page?: number;
    page_size?: number;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.getAll(params);
      setPositions(response.positions);
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 获取单个职位
  const fetchPositionByCode = async (code: string, options?: {
    with_organization?: boolean;
    with_manager?: boolean;
    with_direct_reports?: boolean;
    with_incumbents?: boolean;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const position = await api.getByCode(code, options);
      return position;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 创建职位
  const createPosition = async (position: {
    organization_code: string;
    manager_position_code?: string;
    position_type: string;
    job_profile_id: string;
    status?: string;
    budgeted_fte: number;
    details?: Record<string, any>;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const newPosition = await api.create(position);
      // 刷新列表
      await fetchPositions();
      return newPosition;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 更新职位
  const updatePosition = async (code: string, updates: {
    organization_code?: string;
    manager_position_code?: string;
    status?: string;
    budgeted_fte?: number;
    details?: Record<string, any>;
  }) => {
    setLoading(true);
    setError(null);
    try {
      const updatedPosition = await api.update(code, updates);
      // 刷新列表
      await fetchPositions();
      return updatedPosition;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // 删除职位
  const deletePosition = async (code: string) => {
    setLoading(true);
    setError(null);
    try {
      await api.delete(code);
      // 刷新列表
      await fetchPositions();
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

  return {
    positions,
    loading,
    error,
    stats,
    fetchPositions,
    fetchPositionByCode,
    createPosition,
    updatePosition,
    deletePosition,
    fetchStats,
    api
  };
};

// React组件 - 职位选择器
export const PositionSelector: React.FC<{
  onSelect: (position: Position) => void;
  filter?: { position_type?: string; status?: string; organization_code?: string };
  placeholder?: string;
  apiBaseURL?: string;
}> = ({ onSelect, filter = {}, placeholder = "选择职位", apiBaseURL }) => {
  const { positions, loading, error, fetchPositions } = usePositions(apiBaseURL);
  const [selectedCode, setSelectedCode] = useState<string>('');

  useEffect(() => {
    fetchPositions(filter);
  }, [filter]);

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const code = event.target.value;
    setSelectedCode(code);
    
    const selected = positions.find(pos => pos.code === code);
    if (selected) {
      onSelect(selected);
    }
  };

  const parseDetails = (details?: string) => {
    try {
      return details ? JSON.parse(details) : {};
    } catch {
      return {};
    }
  };

  return (
    <div className="position-selector">
      <select 
        value={selectedCode} 
        onChange={handleChange}
        disabled={loading}
        style={{
          padding: '8px 12px',
          border: '1px solid #ddd',
          borderRadius: '4px',
          fontSize: '14px',
          minWidth: '300px'
        }}
      >
        <option value="">{loading ? '加载中...' : placeholder}</option>
        {positions.map(pos => {
          const details = parseDetails(pos.details);
          return (
            <option key={pos.code} value={pos.code}>
              {pos.code} - {details.title || pos.position_type} ({pos.status})
            </option>
          );
        })}
      </select>
      {error && (
        <div style={{ color: 'red', fontSize: '12px', marginTop: '4px' }}>
          {error}
        </div>
      )}
    </div>
  );
};

// React组件 - 职位表格
export const PositionTable: React.FC<{
  filter?: { position_type?: string; status?: string; organization_code?: string };
  onRowClick?: (position: Position) => void;
  onEdit?: (position: Position) => void;
  onDelete?: (position: Position) => void;
  apiBaseURL?: string;
}> = ({ filter = {}, onRowClick, onEdit, onDelete, apiBaseURL }) => {
  const { positions, loading, error, fetchPositions, stats, fetchStats, deletePosition } = usePositions(apiBaseURL);

  useEffect(() => {
    fetchPositions(filter);
    fetchStats();
  }, [filter]);

  const parseDetails = (details?: string) => {
    try {
      return details ? JSON.parse(details) : {};
    } catch {
      return {};
    }
  };

  const handleDelete = async (position: Position) => {
    if (window.confirm(`确定要删除职位 ${position.code} 吗？`)) {
      try {
        await deletePosition(position.code);
        if (onDelete) onDelete(position);
      } catch (err) {
        alert(`删除失败: ${err}`);
      }
    }
  };

  if (loading) {
    return <div style={{ padding: '20px', textAlign: 'center' }}>加载中...</div>;
  }

  if (error) {
    return <div style={{ padding: '20px', color: 'red' }}>错误: {error}</div>;
  }

  return (
    <div className="position-table">
      {stats && (
        <div style={{ marginBottom: '20px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
          <h4 style={{ margin: '0 0 10px 0' }}>📊 职位统计</h4>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '15px' }}>
            <div>
              <strong>总计:</strong> {stats.total_positions} 个职位<br/>
              <strong>FTE:</strong> {stats.total_budgeted_fte.toFixed(1)}
            </div>
            <div>
              <strong>按类型:</strong><br/>
              全职: {stats.by_type.FULL_TIME || 0}<br/>
              兼职: {stats.by_type.PART_TIME || 0}<br/>
              合同工: {stats.by_type.CONTINGENT_WORKER || 0}<br/>
              实习生: {stats.by_type.INTERN || 0}
            </div>
            <div>
              <strong>按状态:</strong><br/>
              开放: {stats.by_status.OPEN || 0}<br/>
              已填充: {stats.by_status.FILLED || 0}<br/>
              冻结: {stats.by_status.FROZEN || 0}<br/>
              待裁撤: {stats.by_status.PENDING_ELIMINATION || 0}
            </div>
          </div>
        </div>
      )}
      
      <table style={{ width: '100%', borderCollapse: 'collapse', border: '1px solid #ddd', backgroundColor: 'white' }}>
        <thead>
          <tr style={{ backgroundColor: '#f8f9fa' }}>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>编码</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>职位名称</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>类型</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>状态</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>组织</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>管理者</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>FTE</th>
            <th style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'left' }}>操作</th>
          </tr>
        </thead>
        <tbody>
          {positions.map(pos => {
            const details = parseDetails(pos.details);
            return (
              <tr 
                key={pos.code}
                onClick={() => onRowClick?.(pos)}
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
                  <code style={{ 
                    backgroundColor: '#e3f2fd', 
                    padding: '4px 6px', 
                    borderRadius: '4px',
                    color: '#1565c0',
                    fontWeight: 'bold'
                  }}>
                    {pos.code}
                  </code>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd', fontWeight: '500' }}>
                  {details.title || '未设置职位名称'}
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <span style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    fontSize: '11px',
                    fontWeight: '500',
                    backgroundColor: pos.position_type === 'FULL_TIME' ? '#e8f5e8' : 
                                 pos.position_type === 'PART_TIME' ? '#fff3e0' : 
                                 pos.position_type === 'CONTINGENT_WORKER' ? '#f3e5f5' : '#e3f2fd',
                    color: pos.position_type === 'FULL_TIME' ? '#2e7d32' : 
                           pos.position_type === 'PART_TIME' ? '#ef6c00' : 
                           pos.position_type === 'CONTINGENT_WORKER' ? '#7b1fa2' : '#1565c0'
                  }}>
                    {pos.position_type}
                  </span>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <span style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    fontSize: '11px',
                    fontWeight: '500',
                    backgroundColor: pos.status === 'OPEN' ? '#fff3cd' : 
                                 pos.status === 'FILLED' ? '#d4edda' : 
                                 pos.status === 'FROZEN' ? '#f8d7da' : '#e2e3e5',
                    color: pos.status === 'OPEN' ? '#856404' : 
                           pos.status === 'FILLED' ? '#155724' : 
                           pos.status === 'FROZEN' ? '#721c24' : '#495057'
                  }}>
                    {pos.status}
                  </span>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <code style={{ 
                    backgroundColor: '#f3e5f5', 
                    padding: '2px 4px', 
                    borderRadius: '2px', 
                    color: '#7b1fa2',
                    fontSize: '12px'
                  }}>
                    {pos.organization_code}
                  </code>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  {pos.manager_position_code ? (
                    <code style={{ 
                      backgroundColor: '#e8f5e8', 
                      padding: '2px 4px', 
                      borderRadius: '2px', 
                      color: '#2e7d32',
                      fontSize: '12px'
                    }}>
                      {pos.manager_position_code}
                    </code>
                  ) : (
                    <span style={{ color: '#666', fontSize: '12px' }}>无</span>
                  )}
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd', textAlign: 'center' }}>
                  <strong>{pos.budgeted_fte}</strong>
                </td>
                <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    {onEdit && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onEdit(pos);
                        }}
                        style={{
                          padding: '4px 8px',
                          fontSize: '12px',
                          border: '1px solid #007bff',
                          backgroundColor: 'white',
                          color: '#007bff',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        编辑
                      </button>
                    )}
                    {onDelete && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(pos);
                        }}
                        style={{
                          padding: '4px 8px',
                          fontSize: '12px',
                          border: '1px solid #dc3545',
                          backgroundColor: 'white',
                          color: '#dc3545',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        删除
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {positions.length === 0 && (
        <div style={{ 
          padding: '40px', 
          textAlign: 'center', 
          color: '#666',
          backgroundColor: 'white',
          border: '1px solid #ddd',
          borderTop: 'none'
        }}>
          暂无职位数据
        </div>
      )}
    </div>
  );
};

// React组件 - 职位创建表单
export const PositionCreateForm: React.FC<{
  onSuccess?: (position: Position) => void;
  onCancel?: () => void;
  apiBaseURL?: string;
}> = ({ onSuccess, onCancel, apiBaseURL }) => {
  const { createPosition, loading, error } = usePositions(apiBaseURL);
  const [formData, setFormData] = useState({
    organization_code: '',
    manager_position_code: '',
    position_type: 'FULL_TIME',
    job_profile_id: '550e8400-e29b-41d4-a716-446655440000', // 默认UUID
    status: 'OPEN',
    budgeted_fte: 1.0,
    title: '',
    salary_min: '',
    salary_max: '',
    currency: 'CNY'
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const details = {
      title: formData.title,
      salary_range: formData.salary_min && formData.salary_max ? {
        min: parseInt(formData.salary_min),
        max: parseInt(formData.salary_max),
        currency: formData.currency
      } : undefined
    };

    try {
      const position = await createPosition({
        organization_code: formData.organization_code,
        manager_position_code: formData.manager_position_code || undefined,
        position_type: formData.position_type,
        job_profile_id: formData.job_profile_id,
        status: formData.status,
        budgeted_fte: formData.budgeted_fte,
        details
      });
      
      if (onSuccess) onSuccess(position);
      
      // 重置表单
      setFormData({
        organization_code: '',
        manager_position_code: '',
        position_type: 'FULL_TIME',
        job_profile_id: '550e8400-e29b-41d4-a716-446655440000',
        status: 'OPEN',
        budgeted_fte: 1.0,
        title: '',
        salary_min: '',
        salary_max: '',
        currency: 'CNY'
      });
    } catch (err) {
      // 错误已通过hook处理
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ 
      maxWidth: '600px', 
      padding: '20px', 
      border: '1px solid #ddd', 
      borderRadius: '8px',
      backgroundColor: 'white'
    }}>
      <h3 style={{ marginTop: 0 }}>🆕 创建新职位</h3>
      
      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          组织编码 (7位) *
        </label>
        <input
          type="text"
          value={formData.organization_code}
          onChange={(e) => setFormData({...formData, organization_code: e.target.value})}
          placeholder="1000000"
          pattern="[0-9]{7}"
          required
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          管理者职位编码 (7位)
        </label>
        <input
          type="text"
          value={formData.manager_position_code}
          onChange={(e) => setFormData({...formData, manager_position_code: e.target.value})}
          placeholder="1000001"
          pattern="[0-9]{7}"
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          职位名称 *
        </label>
        <input
          type="text"
          value={formData.title}
          onChange={(e) => setFormData({...formData, title: e.target.value})}
          placeholder="高级软件工程师"
          required
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            职位类型 *
          </label>
          <select
            value={formData.position_type}
            onChange={(e) => setFormData({...formData, position_type: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="FULL_TIME">全职</option>
            <option value="PART_TIME">兼职</option>
            <option value="CONTINGENT_WORKER">合同工</option>
            <option value="INTERN">实习生</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            状态 *
          </label>
          <select
            value={formData.status}
            onChange={(e) => setFormData({...formData, status: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="OPEN">开放</option>
            <option value="FILLED">已填充</option>
            <option value="FROZEN">冻结</option>
            <option value="PENDING_ELIMINATION">待裁撤</option>
          </select>
        </div>
      </div>

      <div style={{ marginBottom: '15px' }}>
        <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
          预算FTE *
        </label>
        <input
          type="number"
          step="0.1"
          min="0.1"
          max="5.0"
          value={formData.budgeted_fte}
          onChange={(e) => setFormData({...formData, budgeted_fte: parseFloat(e.target.value)})}
          style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 100px', gap: '10px', marginBottom: '15px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            最低薪资
          </label>
          <input
            type="number"
            value={formData.salary_min}
            onChange={(e) => setFormData({...formData, salary_min: e.target.value})}
            placeholder="20000"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            最高薪资
          </label>
          <input
            type="number"
            value={formData.salary_max}
            onChange={(e) => setFormData({...formData, salary_max: e.target.value})}
            placeholder="35000"
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          />
        </div>
        <div>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: '500' }}>
            币种
          </label>
          <select
            value={formData.currency}
            onChange={(e) => setFormData({...formData, currency: e.target.value})}
            style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
          >
            <option value="CNY">CNY</option>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
          </select>
        </div>
      </div>

      {error && (
        <div style={{ 
          padding: '10px', 
          backgroundColor: '#f8d7da', 
          color: '#721c24', 
          borderRadius: '4px', 
          marginBottom: '15px',
          fontSize: '14px'
        }}>
          {error}
        </div>
      )}

      <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            style={{
              padding: '10px 20px',
              border: '1px solid #6c757d',
              backgroundColor: 'white',
              color: '#6c757d',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            取消
          </button>
        )}
        <button
          type="submit"
          disabled={loading}
          style={{
            padding: '10px 20px',
            border: 'none',
            backgroundColor: loading ? '#6c757d' : '#007bff',
            color: 'white',
            borderRadius: '4px',
            cursor: loading ? 'not-allowed' : 'pointer'
          }}
        >
          {loading ? '创建中...' : '创建职位'}
        </button>
      </div>
    </form>
  );
};

// 导出类型和组件
export type { Position, PositionWithRelations, PositionListResponse, PositionStats };
export { PositionAPI };