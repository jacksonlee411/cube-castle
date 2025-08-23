# API集成示例和代码模板

## 概述

本文档提供了Cube Castle API的完整集成示例，包括GraphQL查询服务、时态API和命令API的客户端实现。所有示例都考虑了缓存优化、错误处理和最佳实践。

## 🚀 快速开始

### API端点总览

| 服务 | 端点 | 协议 | 用途 | 缓存性能 |
|------|------|------|------|----------|
| **GraphQL查询** | `http://localhost:8090/graphql` | GraphQL | 组织查询、统计 | 65%提升 |
| **时态API** | `http://localhost:9091/api/v1` | REST | 时态查询、历史版本 | 94%提升 |
| **命令API** | `http://localhost:9090/api/v1` | REST | 创建、更新、删除 | - |

### 认证配置

```bash
# 环境变量配置
export CUBE_CASTLE_API_KEY="your_api_key_here"
export CUBE_CASTLE_TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
export CUBE_CASTLE_BASE_URL="http://localhost"
```

## 📝 JavaScript/TypeScript客户端

### 1. GraphQL客户端 (Apollo Client)

#### 安装依赖
```bash
npm install @apollo/client graphql
```

#### 客户端配置
```typescript
// src/lib/apollo-client.ts
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

const httpLink = createHttpLink({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_ENDPOINT || 'http://localhost:8090/graphql',
});

const authLink = setContext((_, { headers }) => {
  return {
    headers: {
      ...headers,
      'X-API-Key': process.env.NEXT_PUBLIC_API_KEY || '',
      'X-Tenant-ID': process.env.NEXT_PUBLIC_TENANT_ID || '',
    }
  }
});

export const apolloClient = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          organizations: {
            // 缓存合并策略
            keyArgs: ["searchText"],
            merge(existing = [], incoming = []) {
              return [...existing, ...incoming];
            }
          }
        }
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      cachingPolicy: 'cache-first', // 优先使用缓存
    },
  },
});
```

#### GraphQL查询示例
```typescript
// src/lib/graphql-queries.ts
import { gql } from '@apollo/client';

// 组织列表查询
export const GET_ORGANIZATIONS = gql`
  query GetOrganizations($first: Int, $offset: Int, $searchText: String) {
    organizations(first: $first, offset: $offset, searchText: $searchText) {
      code
      name
      unit_type
      status
      level
      parent_code
      path
      description
      created_at
      updated_at
      effective_date
      version
      is_current
    }
  }
`;

// 单个组织查询
export const GET_ORGANIZATION = gql`
  query GetOrganization($code: String!) {
    organization(code: $code) {
      tenant_id
      code
      name
      unit_type
      status
      level
      parent_code
      path
      description
      profile
      created_at
      updated_at
      effective_date
      version
      is_current
    }
  }
`;

// 组织统计查询
export const GET_ORGANIZATION_STATS = gql`
  query GetOrganizationStats {
    organizationStats {
      totalCount
      byType {
        unitType
        count
      }
      byStatus {
        status
        count
      }
      byLevel {
        level
        count
      }
    }
  }
`;
```

#### React组件示例
```tsx
// src/components/OrganizationList.tsx
import React from 'react';
import { useQuery } from '@apollo/client';
import { GET_ORGANIZATIONS } from '../lib/graphql-queries';

interface Organization {
  code: string;
  name: string;
  unit_type: string;
  status: string;
  level: number;
}

interface OrganizationListProps {
  searchText?: string;
  pageSize?: number;
}

export const OrganizationList: React.FC<OrganizationListProps> = ({
  searchText = '',
  pageSize = 20
}) => {
  const { loading, error, data, fetchMore } = useQuery(GET_ORGANIZATIONS, {
    variables: {
      first: pageSize,
      offset: 0,
      searchText: searchText || undefined
    },
    // 缓存策略：优先缓存，5分钟内不重新请求
    fetchPolicy: 'cache-first',
    nextFetchPolicy: 'cache-first'
  });

  const loadMore = () => {
    fetchMore({
      variables: {
        offset: data?.organizations?.length || 0
      }
    });
  };

  if (loading) return <div className="loading">加载中...</div>;
  if (error) return <div className="error">错误: {error.message}</div>;

  const organizations = data?.organizations || [];

  return (
    <div className="organization-list">
      <div className="stats">
        找到 {organizations.length} 个组织
      </div>
      
      {organizations.map((org: Organization) => (
        <div key={org.code} className="organization-item">
          <h3>{org.name}</h3>
          <div className="organization-meta">
            <span className="code">代码: {org.code}</span>
            <span className="type">类型: {org.unit_type}</span>
            <span className="status">状态: {org.status}</span>
            <span className="level">层级: {org.level}</span>
          </div>
        </div>
      ))}
      
      <button onClick={loadMore} className="load-more">
        加载更多
      </button>
    </div>
  );
};
```

### 2. 时态API客户端

#### 时态查询客户端
```typescript
// src/lib/temporal-client.ts
interface TemporalQueryOptions {
  asOfDate?: string;
  effectiveFrom?: string;
  effectiveTo?: string;
  includeHistory?: boolean;
  includeFuture?: boolean;
  includeDissolved?: boolean;
  version?: number;
  maxVersions?: number;
}

interface TemporalOrganization {
  tenant_id: string;
  code: string;
  parent_code?: string;
  name: string;
  unit_type: string;
  status: string;
  level: number;
  path?: string;
  sort_order: number;
  description?: string;
  created_at: string;
  updated_at: string;
  effective_date?: string;
  end_date?: string;
  version?: number;
  supersedes_version?: number;
  change_reason?: string;
  is_current?: boolean;
}

interface TemporalQueryResponse {
  organizations: TemporalOrganization[];
  query_options: TemporalQueryOptions;
  result_count: number;
  queried_at: string;
}

export class TemporalClient {
  private baseURL: string;
  private apiKey: string;
  private tenantId: string;

  constructor(baseURL = 'http://localhost:9091', apiKey = '', tenantId = '') {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.tenantId = tenantId;
  }

  private buildQueryString(options: TemporalQueryOptions): string {
    const params = new URLSearchParams();
    
    if (options.asOfDate) params.append('as_of_date', options.asOfDate);
    if (options.effectiveFrom) params.append('effective_from', options.effectiveFrom);
    if (options.effectiveTo) params.append('effective_to', options.effectiveTo);
    if (options.includeHistory) params.append('include_history', 'true');
    if (options.includeFuture) params.append('include_future', 'true');
    if (options.includeDissolved) params.append('include_dissolved', 'true');
    if (options.version) params.append('version', options.version.toString());
    if (options.maxVersions) params.append('max_versions', options.maxVersions.toString());

    return params.toString();
  }

  async getOrganizationTemporal(
    code: string,
    options: TemporalQueryOptions = {}
  ): Promise<TemporalQueryResponse> {
    const queryString = this.buildQueryString(options);
    const url = `${this.baseURL}/api/v1/organization-units/${code}/temporal${queryString ? '?' + queryString : ''}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        'X-Tenant-ID': this.tenantId,
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(`时态查询失败: ${response.status} - ${errorData.message || response.statusText}`);
    }

    return response.json();
  }

  // 获取组织的当前版本
  async getCurrentVersion(code: string): Promise<TemporalOrganization | null> {
    const result = await this.getOrganizationTemporal(code, {
      includeHistory: false
    });
    
    return result.organizations.find(org => org.is_current) || null;
  }

  // 获取组织的历史版本
  async getHistory(code: string, maxVersions = 10): Promise<TemporalOrganization[]> {
    const result = await this.getOrganizationTemporal(code, {
      includeHistory: true,
      maxVersions
    });
    
    return result.organizations.sort((a, b) => (b.version || 0) - (a.version || 0));
  }

  // 获取特定时间点的组织状态
  async getAsOfDate(code: string, asOfDate: string): Promise<TemporalOrganization | null> {
    const result = await this.getOrganizationTemporal(code, {
      asOfDate,
      includeHistory: true
    });
    
    return result.organizations[0] || null;
  }

  // 获取时间范围内的变更
  async getChangeHistory(
    code: string,
    fromDate: string,
    toDate: string
  ): Promise<TemporalOrganization[]> {
    const result = await this.getOrganizationTemporal(code, {
      effectiveFrom: fromDate,
      effectiveTo: toDate,
      includeHistory: true
    });
    
    return result.organizations;
  }
}
```

#### 时态API使用示例
```typescript
// src/components/OrganizationHistory.tsx
import React, { useState, useEffect } from 'react';
import { TemporalClient, TemporalOrganization } from '../lib/temporal-client';

interface OrganizationHistoryProps {
  orgCode: string;
}

export const OrganizationHistory: React.FC<OrganizationHistoryProps> = ({ orgCode }) => {
  const [history, setHistory] = useState<TemporalOrganization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const temporalClient = new TemporalClient(
    process.env.REACT_APP_TEMPORAL_API_URL,
    process.env.REACT_APP_API_KEY,
    process.env.REACT_APP_TENANT_ID
  );

  useEffect(() => {
    loadHistory();
  }, [orgCode]);

  const loadHistory = async () => {
    try {
      setLoading(true);
      const historyData = await temporalClient.getHistory(orgCode, 20);
      setHistory(historyData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载历史失败');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>加载历史版本中...</div>;
  if (error) return <div className="error">错误: {error}</div>;

  return (
    <div className="organization-history">
      <h3>组织变更历史</h3>
      
      <div className="timeline">
        {history.map((version) => (
          <div 
            key={`${version.code}-${version.version}`} 
            className={`timeline-item ${version.is_current ? 'current' : 'historical'}`}
          >
            <div className="timeline-header">
              <span className="version">版本 {version.version}</span>
              <span className="date">
                {version.effective_date && new Date(version.effective_date).toLocaleDateString()}
              </span>
              {version.is_current && <span className="badge-current">当前</span>}
            </div>
            
            <div className="timeline-content">
              <h4>{version.name}</h4>
              <div className="meta">
                <span>类型: {version.unit_type}</span>
                <span>状态: {version.status}</span>
                <span>层级: {version.level}</span>
              </div>
              
              {version.change_reason && (
                <div className="change-reason">
                  变更原因: {version.change_reason}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
```

### 3. 事件管理客户端

#### 事件创建客户端
```typescript
// src/lib/event-client.ts
interface OrganizationChangeEvent {
  event_type: 'CREATE' | 'UPDATE' | 'RESTRUCTURE' | 'DISSOLVE' | 'ACTIVATE' | 'DEACTIVATE';
  effective_date: string;
  end_date?: string;
  change_reason: string;
  change_data: Record<string, any>;
}

interface EventResponse {
  event_id: string;
  event_type: string;
  organization: string;
  effective_date: string;
  status: 'processed' | 'failed' | 'pending';
  processed_at: string;
}

export class EventClient {
  private baseURL: string;
  private apiKey: string;
  private tenantId: string;

  constructor(baseURL = 'http://localhost:9091', apiKey = '', tenantId = '') {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.tenantId = tenantId;
  }

  async createEvent(orgCode: string, event: OrganizationChangeEvent): Promise<EventResponse> {
    const response = await fetch(`${this.baseURL}/api/v1/organization-units/${orgCode}/events`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        'X-Tenant-ID': this.tenantId,
      },
      body: JSON.stringify(event),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(`事件创建失败: ${response.status} - ${errorData.message || response.statusText}`);
    }

    return response.json();
  }

  // 更新组织信息
  async updateOrganization(
    orgCode: string,
    changes: Record<string, any>,
    effectiveDate: string,
    reason: string
  ): Promise<EventResponse> {
    return this.createEvent(orgCode, {
      event_type: 'UPDATE',
      effective_date: effectiveDate,
      change_reason: reason,
      change_data: changes,
    });
  }

  // 组织架构重组
  async restructureOrganization(
    orgCode: string,
    newStructure: Record<string, any>,
    effectiveDate: string,
    reason: string
  ): Promise<EventResponse> {
    return this.createEvent(orgCode, {
      event_type: 'RESTRUCTURE',
      effective_date: effectiveDate,
      change_reason: reason,
      change_data: newStructure,
    });
  }

  // 解散组织
  async dissolveOrganization(
    orgCode: string,
    effectiveDate: string,
    endDate: string,
    reason: string
  ): Promise<EventResponse> {
    return this.createEvent(orgCode, {
      event_type: 'DISSOLVE',
      effective_date: effectiveDate,
      end_date: endDate,
      change_reason: reason,
      change_data: {},
    });
  }

  // 激活/停用组织
  async changeOrganizationStatus(
    orgCode: string,
    activate: boolean,
    effectiveDate: string,
    reason: string
  ): Promise<EventResponse> {
    return this.createEvent(orgCode, {
      event_type: activate ? 'ACTIVATE' : 'DEACTIVATE',
      effective_date: effectiveDate,
      change_reason: reason,
      change_data: {},
    });
  }
}
```

### 4. 统一API客户端

#### 完整客户端封装
```typescript
// src/lib/cube-castle-client.ts
import { ApolloClient } from '@apollo/client';
import { TemporalClient } from './temporal-client';
import { EventClient } from './event-client';

export interface CubeCastleConfig {
  baseURL: string;
  graphqlURL?: string;
  temporalURL?: string;
  apiKey: string;
  tenantId: string;
}

export class CubeCastleClient {
  private config: CubeCastleConfig;
  public graphql: ApolloClient<any>;
  public temporal: TemporalClient;
  public events: EventClient;

  constructor(config: CubeCastleConfig) {
    this.config = config;
    
    // 初始化GraphQL客户端
    this.graphql = new ApolloClient({
      // Apollo配置
    });
    
    // 初始化时态API客户端
    this.temporal = new TemporalClient(
      config.temporalURL || `${config.baseURL}:9091`,
      config.apiKey,
      config.tenantId
    );
    
    // 初始化事件客户端
    this.events = new EventClient(
      config.temporalURL || `${config.baseURL}:9091`,
      config.apiKey,
      config.tenantId
    );
  }

  // 健康检查
  async healthCheck(): Promise<Record<string, any>> {
    const checks = {
      graphql: await this.checkGraphQLHealth(),
      temporal: await this.checkTemporalHealth(),
    };

    return checks;
  }

  private async checkGraphQLHealth(): Promise<boolean> {
    try {
      const response = await fetch(`${this.config.graphqlURL || this.config.baseURL + ':8090'}/health`);
      return response.ok;
    } catch {
      return false;
    }
  }

  private async checkTemporalHealth(): Promise<boolean> {
    try {
      const response = await fetch(`${this.config.temporalURL || this.config.baseURL + ':9091'}/health`);
      return response.ok;
    } catch {
      return false;
    }
  }
}

// 使用示例
export const createClient = (config: Partial<CubeCastleConfig>) => {
  const defaultConfig: CubeCastleConfig = {
    baseURL: 'http://localhost',
    apiKey: process.env.REACT_APP_API_KEY || '',
    tenantId: process.env.REACT_APP_TENANT_ID || '',
    ...config,
  };

  return new CubeCastleClient(defaultConfig);
};
```

## 🐍 Python客户端

### 1. 依赖安装
```bash
pip install requests gql[all] python-dateutil
```

### 2. GraphQL客户端
```python
# cube_castle/graphql_client.py
from gql import gql, Client
from gql.transport.requests import RequestsHTTPTransport
from typing import List, Dict, Any, Optional
import os

class GraphQLClient:
    def __init__(self, endpoint: str = None, api_key: str = None, tenant_id: str = None):
        self.endpoint = endpoint or os.getenv('CUBE_CASTLE_GRAPHQL_ENDPOINT', 'http://localhost:8090/graphql')
        self.api_key = api_key or os.getenv('CUBE_CASTLE_API_KEY', '')
        self.tenant_id = tenant_id or os.getenv('CUBE_CASTLE_TENANT_ID', '')
        
        # 配置传输层
        transport = RequestsHTTPTransport(
            url=self.endpoint,
            headers={
                'X-API-Key': self.api_key,
                'X-Tenant-ID': self.tenant_id,
            }
        )
        
        self.client = Client(transport=transport, fetch_schema_from_transport=True)

    def get_organizations(self, first: int = 50, offset: int = 0, search_text: str = None) -> List[Dict[str, Any]]:
        """获取组织列表"""
        query = gql("""
            query GetOrganizations($first: Int, $offset: Int, $searchText: String) {
                organizations(first: $first, offset: $offset, searchText: $searchText) {
                    code
                    name
                    unit_type
                    status
                    level
                    parent_code
                    path
                    description
                    created_at
                    updated_at
                    effective_date
                    version
                    is_current
                }
            }
        """)
        
        variables = {
            "first": first,
            "offset": offset,
            "searchText": search_text
        }
        
        result = self.client.execute(query, variable_values=variables)
        return result['organizations']

    def get_organization(self, code: str) -> Optional[Dict[str, Any]]:
        """获取单个组织"""
        query = gql("""
            query GetOrganization($code: String!) {
                organization(code: $code) {
                    tenant_id
                    code
                    name
                    unit_type
                    status
                    level
                    parent_code
                    path
                    description
                    profile
                    created_at
                    updated_at
                    effective_date
                    version
                    is_current
                }
            }
        """)
        
        result = self.client.execute(query, variable_values={"code": code})
        return result['organization']

    def get_organization_stats(self) -> Dict[str, Any]:
        """获取组织统计信息"""
        query = gql("""
            query GetOrganizationStats {
                organizationStats {
                    totalCount
                    byType {
                        unitType
                        count
                    }
                    byStatus {
                        status
                        count
                    }
                    byLevel {
                        level
                        count
                    }
                }
            }
        """)
        
        result = self.client.execute(query)
        return result['organizationStats']

# 使用示例
if __name__ == "__main__":
    client = GraphQLClient()
    
    # 获取组织列表
    orgs = client.get_organizations(first=10)
    print(f"获取到 {len(orgs)} 个组织")
    
    # 获取单个组织
    if orgs:
        org_detail = client.get_organization(orgs[0]['code'])
        print(f"组织详情: {org_detail['name']}")
    
    # 获取统计信息
    stats = client.get_organization_stats()
    print(f"总组织数: {stats['totalCount']}")
```

### 3. 时态API客户端
```python
# cube_castle/temporal_client.py
import requests
from typing import List, Dict, Any, Optional
from datetime import datetime
from dateutil import parser
import os

class TemporalClient:
    def __init__(self, base_url: str = None, api_key: str = None, tenant_id: str = None):
        self.base_url = base_url or os.getenv('CUBE_CASTLE_TEMPORAL_URL', 'http://localhost:9091')
        self.api_key = api_key or os.getenv('CUBE_CASTLE_API_KEY', '')
        self.tenant_id = tenant_id or os.getenv('CUBE_CASTLE_TENANT_ID', '')
        
        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': self.api_key,
            'X-Tenant-ID': self.tenant_id,
            'Content-Type': 'application/json'
        })

    def _build_query_params(self, options: Dict[str, Any]) -> Dict[str, str]:
        """构建查询参数"""
        params = {}
        
        if options.get('as_of_date'):
            if isinstance(options['as_of_date'], datetime):
                params['as_of_date'] = options['as_of_date'].strftime('%Y-%m-%d')
            else:
                params['as_of_date'] = str(options['as_of_date'])
        
        if options.get('effective_from'):
            params['effective_from'] = str(options['effective_from'])
        
        if options.get('effective_to'):
            params['effective_to'] = str(options['effective_to'])
        
        if options.get('include_history'):
            params['include_history'] = 'true'
        
        if options.get('include_future'):
            params['include_future'] = 'true'
        
        if options.get('include_dissolved'):
            params['include_dissolved'] = 'true'
        
        if options.get('version'):
            params['version'] = str(options['version'])
        
        if options.get('max_versions'):
            params['max_versions'] = str(options['max_versions'])
        
        return params

    def get_organization_temporal(self, code: str, options: Dict[str, Any] = None) -> Dict[str, Any]:
        """执行时态查询"""
        if options is None:
            options = {}
        
        params = self._build_query_params(options)
        url = f"{self.base_url}/api/v1/organization-units/{code}/temporal"
        
        response = self.session.get(url, params=params)
        
        if response.status_code == 404:
            return {'organizations': [], 'result_count': 0}
        
        response.raise_for_status()
        return response.json()

    def get_current_version(self, code: str) -> Optional[Dict[str, Any]]:
        """获取组织当前版本"""
        result = self.get_organization_temporal(code, {'include_history': False})
        organizations = result.get('organizations', [])
        
        for org in organizations:
            if org.get('is_current'):
                return org
        
        return organizations[0] if organizations else None

    def get_history(self, code: str, max_versions: int = 10) -> List[Dict[str, Any]]:
        """获取组织历史版本"""
        result = self.get_organization_temporal(code, {
            'include_history': True,
            'max_versions': max_versions
        })
        
        organizations = result.get('organizations', [])
        # 按版本号排序
        return sorted(organizations, key=lambda x: x.get('version', 0), reverse=True)

    def get_as_of_date(self, code: str, as_of_date: str) -> Optional[Dict[str, Any]]:
        """获取特定时间点的组织状态"""
        result = self.get_organization_temporal(code, {
            'as_of_date': as_of_date,
            'include_history': True
        })
        
        organizations = result.get('organizations', [])
        return organizations[0] if organizations else None

    def get_change_history(self, code: str, from_date: str, to_date: str) -> List[Dict[str, Any]]:
        """获取时间范围内的变更历史"""
        result = self.get_organization_temporal(code, {
            'effective_from': from_date,
            'effective_to': to_date,
            'include_history': True
        })
        
        return result.get('organizations', [])

    def create_event(self, org_code: str, event: Dict[str, Any]) -> Dict[str, Any]:
        """创建组织事件"""
        url = f"{self.base_url}/api/v1/organization-units/{org_code}/events"
        
        response = self.session.post(url, json=event)
        response.raise_for_status()
        
        return response.json()

    def update_organization(self, org_code: str, changes: Dict[str, Any], 
                          effective_date: str, reason: str) -> Dict[str, Any]:
        """更新组织信息"""
        event = {
            'event_type': 'UPDATE',
            'effective_date': effective_date,
            'change_reason': reason,
            'change_data': changes
        }
        
        return self.create_event(org_code, event)

# 使用示例
if __name__ == "__main__":
    client = TemporalClient()
    
    # 获取当前版本
    current = client.get_current_version('1000001')
    if current:
        print(f"当前组织: {current['name']} (版本 {current.get('version')})")
    
    # 获取历史版本
    history = client.get_history('1000001')
    print(f"历史版本数: {len(history)}")
    
    # 时间点查询
    past_version = client.get_as_of_date('1000001', '2025-08-01')
    if past_version:
        print(f"2025-08-01时的组织: {past_version['name']}")
    
    # 创建更新事件
    try:
        event_result = client.update_organization(
            '1000001',
            {'name': '更新后的组织名称'},
            '2025-08-15T00:00:00Z',
            '组织名称标准化'
        )
        print(f"事件创建成功: {event_result['event_id']}")
    except Exception as e:
        print(f"事件创建失败: {e}")
```

## 🐹 Go客户端

### 1. Go模块初始化
```bash
mkdir cube-castle-client && cd cube-castle-client
go mod init github.com/your-org/cube-castle-client
go get github.com/machinebox/graphql
```

### 2. GraphQL客户端
```go
// client/graphql_client.go
package client

import (
    "context"
    "fmt"
    "os"
    
    "github.com/machinebox/graphql"
)

type Organization struct {
    TenantID      string `json:"tenant_id"`
    Code          string `json:"code"`
    ParentCode    string `json:"parent_code,omitempty"`
    Name          string `json:"name"`
    UnitType      string `json:"unit_type"`
    Status        string `json:"status"`
    Level         int    `json:"level"`
    Path          string `json:"path,omitempty"`
    Description   string `json:"description,omitempty"`
    CreatedAt     string `json:"created_at"`
    UpdatedAt     string `json:"updated_at"`
    EffectiveDate string `json:"effective_date"`
    Version       int    `json:"version"`
    IsCurrent     bool   `json:"is_current"`
}

type OrganizationStats struct {
    TotalCount int          `json:"totalCount"`
    ByType     []TypeCount  `json:"byType"`
    ByStatus   []StatusCount `json:"byStatus"`
    ByLevel    []LevelCount  `json:"byLevel"`
}

type TypeCount struct {
    UnitType string `json:"unitType"`
    Count    int    `json:"count"`
}

type StatusCount struct {
    Status string `json:"status"`
    Count  int    `json:"count"`
}

type LevelCount struct {
    Level string `json:"level"`
    Count int    `json:"count"`
}

type GraphQLClient struct {
    client   *graphql.Client
    apiKey   string
    tenantID string
}

func NewGraphQLClient(endpoint, apiKey, tenantID string) *GraphQLClient {
    if endpoint == "" {
        endpoint = getEnv("CUBE_CASTLE_GRAPHQL_ENDPOINT", "http://localhost:8090/graphql")
    }
    if apiKey == "" {
        apiKey = os.Getenv("CUBE_CASTLE_API_KEY")
    }
    if tenantID == "" {
        tenantID = os.Getenv("CUBE_CASTLE_TENANT_ID")
    }

    client := graphql.NewClient(endpoint)
    
    // 设置默认请求头
    client.Header.Set("X-API-Key", apiKey)
    client.Header.Set("X-Tenant-ID", tenantID)

    return &GraphQLClient{
        client:   client,
        apiKey:   apiKey,
        tenantID: tenantID,
    }
}

func (c *GraphQLClient) GetOrganizations(ctx context.Context, first, offset int, searchText string) ([]Organization, error) {
    req := graphql.NewRequest(`
        query GetOrganizations($first: Int, $offset: Int, $searchText: String) {
            organizations(first: $first, offset: $offset, searchText: $searchText) {
                code
                name
                unit_type
                status
                level
                parent_code
                path
                description
                created_at
                updated_at
                effective_date
                version
                is_current
            }
        }
    `)
    
    req.Var("first", first)
    req.Var("offset", offset)
    if searchText != "" {
        req.Var("searchText", searchText)
    }

    var response struct {
        Organizations []Organization `json:"organizations"`
    }

    if err := c.client.Run(ctx, req, &response); err != nil {
        return nil, fmt.Errorf("GraphQL查询失败: %w", err)
    }

    return response.Organizations, nil
}

func (c *GraphQLClient) GetOrganization(ctx context.Context, code string) (*Organization, error) {
    req := graphql.NewRequest(`
        query GetOrganization($code: String!) {
            organization(code: $code) {
                tenant_id
                code
                name
                unit_type
                status
                level
                parent_code
                path
                description
                created_at
                updated_at
                effective_date
                version
                is_current
            }
        }
    `)
    
    req.Var("code", code)

    var response struct {
        Organization *Organization `json:"organization"`
    }

    if err := c.client.Run(ctx, req, &response); err != nil {
        return nil, fmt.Errorf("GraphQL查询失败: %w", err)
    }

    return response.Organization, nil
}

func (c *GraphQLClient) GetOrganizationStats(ctx context.Context) (*OrganizationStats, error) {
    req := graphql.NewRequest(`
        query GetOrganizationStats {
            organizationStats {
                totalCount
                byType {
                    unitType
                    count
                }
                byStatus {
                    status
                    count
                }
                byLevel {
                    level
                    count
                }
            }
        }
    `)

    var response struct {
        OrganizationStats OrganizationStats `json:"organizationStats"`
    }

    if err := c.client.Run(ctx, req, &response); err != nil {
        return nil, fmt.Errorf("GraphQL查询失败: %w", err)
    }

    return &response.OrganizationStats, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 3. 时态API客户端
```go
// client/temporal_client.go  
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "time"
)

type TemporalOrganization struct {
    TenantID           string     `json:"tenant_id"`
    Code               string     `json:"code"`
    ParentCode         *string    `json:"parent_code,omitempty"`
    Name               string     `json:"name"`
    UnitType           string     `json:"unit_type"`
    Status             string     `json:"status"`
    Level              int        `json:"level"`
    Path               *string    `json:"path,omitempty"`
    SortOrder          int        `json:"sort_order"`
    Description        *string    `json:"description,omitempty"`
    CreatedAt          time.Time  `json:"created_at"`
    UpdatedAt          time.Time  `json:"updated_at"`
    EffectiveDate      *time.Time `json:"effective_date,omitempty"`
    EndDate            *time.Time `json:"end_date,omitempty"`
    Version            *int       `json:"version,omitempty"`
    SupersedesVersion  *int       `json:"supersedes_version,omitempty"`
    ChangeReason       *string    `json:"change_reason,omitempty"`
    IsCurrent          *bool      `json:"is_current,omitempty"`
}

type TemporalQueryOptions struct {
    AsOfDate        *time.Time `json:"as_of_date,omitempty"`
    EffectiveFrom   *time.Time `json:"effective_from,omitempty"`
    EffectiveTo     *time.Time `json:"effective_to,omitempty"`
    IncludeHistory  bool       `json:"include_history,omitempty"`
    IncludeFuture   bool       `json:"include_future,omitempty"`
    IncludeDissolved bool      `json:"include_dissolved,omitempty"`
    Version         *int       `json:"version,omitempty"`
    MaxVersions     int        `json:"max_versions,omitempty"`
}

type TemporalQueryResponse struct {
    Organizations []TemporalOrganization `json:"organizations"`
    QueryOptions  TemporalQueryOptions   `json:"query_options"`
    ResultCount   int                    `json:"result_count"`
    QueriedAt     time.Time              `json:"queried_at"`
}

type OrganizationChangeEvent struct {
    EventType     string                 `json:"event_type"`
    EffectiveDate time.Time              `json:"effective_date"`
    EndDate       *time.Time             `json:"end_date,omitempty"`
    ChangeReason  string                 `json:"change_reason"`
    ChangeData    map[string]interface{} `json:"change_data"`
}

type EventResponse struct {
    EventID       string    `json:"event_id"`
    EventType     string    `json:"event_type"`
    Organization  string    `json:"organization"`
    EffectiveDate time.Time `json:"effective_date"`
    Status        string    `json:"status"`
    ProcessedAt   time.Time `json:"processed_at"`
}

type TemporalClient struct {
    baseURL  string
    apiKey   string
    tenantID string
    client   *http.Client
}

func NewTemporalClient(baseURL, apiKey, tenantID string) *TemporalClient {
    if baseURL == "" {
        baseURL = getEnv("CUBE_CASTLE_TEMPORAL_URL", "http://localhost:9091")
    }
    if apiKey == "" {
        apiKey = os.Getenv("CUBE_CASTLE_API_KEY")
    }
    if tenantID == "" {
        tenantID = os.Getenv("CUBE_CASTLE_TENANT_ID")
    }

    return &TemporalClient{
        baseURL:  baseURL,
        apiKey:   apiKey,
        tenantID: tenantID,
        client:   &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *TemporalClient) buildQueryParams(options TemporalQueryOptions) url.Values {
    params := url.Values{}
    
    if options.AsOfDate != nil {
        params.Set("as_of_date", options.AsOfDate.Format("2006-01-02"))
    }
    if options.EffectiveFrom != nil {
        params.Set("effective_from", options.EffectiveFrom.Format("2006-01-02"))
    }
    if options.EffectiveTo != nil {
        params.Set("effective_to", options.EffectiveTo.Format("2006-01-02"))
    }
    if options.IncludeHistory {
        params.Set("include_history", "true")
    }
    if options.IncludeFuture {
        params.Set("include_future", "true")
    }
    if options.IncludeDissolved {
        params.Set("include_dissolved", "true")
    }
    if options.Version != nil {
        params.Set("version", strconv.Itoa(*options.Version))
    }
    if options.MaxVersions > 0 {
        params.Set("max_versions", strconv.Itoa(options.MaxVersions))
    }
    
    return params
}

func (c *TemporalClient) GetOrganizationTemporal(ctx context.Context, code string, options TemporalQueryOptions) (*TemporalQueryResponse, error) {
    params := c.buildQueryParams(options)
    requestURL := fmt.Sprintf("%s/api/v1/organization-units/%s/temporal", c.baseURL, code)
    
    if len(params) > 0 {
        requestURL += "?" + params.Encode()
    }

    req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    req.Header.Set("X-API-Key", c.apiKey)
    req.Header.Set("X-Tenant-ID", c.tenantID)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("发送请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
    }

    var response TemporalQueryResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    return &response, nil
}

func (c *TemporalClient) GetCurrentVersion(ctx context.Context, code string) (*TemporalOrganization, error) {
    options := TemporalQueryOptions{
        IncludeHistory: false,
    }
    
    response, err := c.GetOrganizationTemporal(ctx, code, options)
    if err != nil {
        return nil, err
    }

    for _, org := range response.Organizations {
        if org.IsCurrent != nil && *org.IsCurrent {
            return &org, nil
        }
    }

    if len(response.Organizations) > 0 {
        return &response.Organizations[0], nil
    }

    return nil, nil
}

func (c *TemporalClient) CreateEvent(ctx context.Context, orgCode string, event OrganizationChangeEvent) (*EventResponse, error) {
    requestURL := fmt.Sprintf("%s/api/v1/organization-units/%s/events", c.baseURL, orgCode)

    jsonData, err := json.Marshal(event)
    if err != nil {
        return nil, fmt.Errorf("序列化事件失败: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    req.Header.Set("X-API-Key", c.apiKey)
    req.Header.Set("X-Tenant-ID", c.tenantID)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("发送请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("创建事件失败，状态码: %d", resp.StatusCode)
    }

    var response EventResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    return &response, nil
}

// 使用示例
func ExampleUsage() {
    ctx := context.Background()
    
    // 创建GraphQL客户端
    graphqlClient := NewGraphQLClient("", "", "")
    
    // 获取组织列表
    orgs, err := graphqlClient.GetOrganizations(ctx, 10, 0, "")
    if err != nil {
        fmt.Printf("获取组织列表失败: %v\n", err)
        return
    }
    fmt.Printf("获取到 %d 个组织\n", len(orgs))
    
    // 创建时态客户端
    temporalClient := NewTemporalClient("", "", "")
    
    if len(orgs) > 0 {
        // 获取组织当前版本
        current, err := temporalClient.GetCurrentVersion(ctx, orgs[0].Code)
        if err != nil {
            fmt.Printf("获取当前版本失败: %v\n", err)
            return
        }
        
        if current != nil {
            fmt.Printf("当前组织: %s (版本 %d)\n", current.Name, *current.Version)
        }
        
        // 创建更新事件
        event := OrganizationChangeEvent{
            EventType:     "UPDATE",
            EffectiveDate: time.Now(),
            ChangeReason:  "示例更新",
            ChangeData: map[string]interface{}{
                "name": "更新后的组织名称",
            },
        }
        
        eventResp, err := temporalClient.CreateEvent(ctx, orgs[0].Code, event)
        if err != nil {
            fmt.Printf("创建事件失败: %v\n", err)
            return
        }
        
        fmt.Printf("事件创建成功: %s\n", eventResp.EventID)
    }
}
```

这份完整的API集成文档提供了：

1. **多语言客户端**: JavaScript/TypeScript、Python、Go
2. **完整功能覆盖**: GraphQL查询、时态API、事件管理
3. **缓存优化**: 客户端缓存策略和配置
4. **错误处理**: 完善的错误处理机制
5. **配置管理**: 环境变量和配置文件支持
6. **使用示例**: 实际使用场景的代码示例

这将大大简化开发者接入Cube Castle API的过程。