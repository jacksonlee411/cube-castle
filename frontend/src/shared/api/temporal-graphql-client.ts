/**
 * 时态管理GraphQL查询客户端
 * 专门用于时态查询功能：organizationAsOfDate 和 organizationHistory
 */
import type { 
  OrganizationUnit, 
  GraphQLResponse,
} from '../types';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit,
  TimelineEvent
} from '../types/temporal';

// GraphQL端点 - 直接路由到标准GraphQL服务（8090端口）
const TEMPORAL_GRAPHQL_ENDPOINT = '/graphql';

interface TemporalGraphQLClient {
  request<T>(query: string, variables?: Record<string, unknown>): Promise<T>;
}

const temporalGraphQLClient: TemporalGraphQLClient = {
  async request<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    const response = await fetch(TEMPORAL_GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        variables
      }),
    });

    if (!response.ok) {
      throw new Error(`Temporal GraphQL Error: ${response.status} ${response.statusText}`);
    }

    const result = await response.json() as GraphQLResponse<T>;
    
    if (result.errors && result.errors.length > 0) {
      throw new Error(`Temporal GraphQL Error: ${result.errors[0].message}`);
    }

    if (!result.data) {
      throw new Error('Temporal GraphQL Error: No data returned');
    }

    return result.data;
  }
};

// GraphQL查询定义
const TEMPORAL_QUERIES = {
  // 时间点查询 - 查询特定时间点的组织状态
  ORGANIZATION_AS_OF_DATE: `
    query OrganizationAsOfDate($code: String!, $asOfDate: String!) {
      organizationAsOfDate(code: $code, asOfDate: $asOfDate) {
        tenant_id
        code
        parent_code
        name
        unit_type
        status
        level
        path
        sort_order
        description
        profile
        created_at
        updated_at
        effective_date
        end_date
        version
        is_current
        change_reason
        valid_from
        valid_to
      }
    }
  `,

  // 历史查询 - 查询时间范围内的所有历史记录
  ORGANIZATION_HISTORY: `
    query OrganizationHistory($code: String!, $fromDate: String!, $toDate: String!) {
      organizationHistory(code: $code, fromDate: $fromDate, toDate: $toDate) {
        tenant_id
        code
        parent_code
        name
        unit_type
        status
        level
        path
        sort_order
        description
        profile
        created_at
        updated_at
        effective_date
        end_date
        version
        is_current
        change_reason
        valid_from
        valid_to
      }
    }
  `,

  // 传统查询保持兼容 - 当前数据
  ORGANIZATIONS_CURRENT: `
    query OrganizationsCurrent($first: Int, $offset: Int, $searchText: String) {
      organizations(first: $first, offset: $offset, searchText: $searchText) {
        tenant_id
        code
        parent_code
        name
        unit_type
        status
        level
        path
        sort_order
        description
        profile
        created_at
        updated_at
        effective_date
        end_date
        version
        is_current
      }
    }
  `,

  // 单个组织查询 - 当前数据
  ORGANIZATION_CURRENT: `
    query OrganizationCurrent($code: String!) {
      organization(code: $code) {
        tenant_id
        code
        parent_code
        name
        unit_type
        status
        level
        path
        sort_order
        description
        profile
        created_at
        updated_at
        effective_date
        end_date
        version
        is_current
      }
    }
  `
};

// 时态数据转换器
function transformToTemporalOrganization(data: any): TemporalOrganizationUnit {
  return {
    tenant_id: data.tenant_id || '',
    code: data.code || '',
    parent_code: data.parent_code || '',
    name: data.name || '',
    unit_type: data.unit_type as any || 'DEPARTMENT',
    status: data.status as any || 'ACTIVE',
    level: data.level || 1,
    path: data.path || '',
    sort_order: data.sort_order || 0,
    description: data.description || '',
    profile: data.profile || '',
    created_at: data.created_at || '',
    updated_at: data.updated_at || '',
    effective_date: data.effective_date || '',
    end_date: data.end_date || undefined,
    is_current: data.is_current ?? true,
    change_reason: data.change_reason || undefined,
    approved_by: undefined, // GraphQL中暂无此字段
    approved_at: undefined  // GraphQL中暂无此字段
  };
}

// 时态API客户端
export const temporalAPI = {
  /**
   * 查询指定时间点的组织状态
   * @param code 组织代码
   * @param asOfDate 查询时间点 (YYYY-MM-DD格式)
   * @returns 该时间点有效的组织记录
   */
  async getOrganizationAsOfDate(
    code: string, 
    asOfDate: string
  ): Promise<TemporalOrganizationUnit | null> {
    try {
      const data = await temporalGraphQLClient.request<{
        organizationAsOfDate: any | null;
      }>(
        TEMPORAL_QUERIES.ORGANIZATION_AS_OF_DATE,
        { code, asOfDate }
      );

      if (!data.organizationAsOfDate) {
        return null;
      }

      return transformToTemporalOrganization(data.organizationAsOfDate);
    } catch (error) {
      console.error(`时间点查询失败 [code=${code}, asOfDate=${asOfDate}]:`, error);
      throw new Error(`无法查询指定时间点的组织数据: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  },

  /**
   * 查询组织的完整历史记录
   * @param code 组织代码
   * @param params 查询参数
   * @returns 时间范围内的所有历史记录，按时间倒序
   */
  async getOrganizationHistory(
    code: string,
    params?: {
      fromDate?: string;
      toDate?: string;
    }
  ): Promise<TemporalOrganizationUnit[]> {
    try {
      // 设置默认时间范围 - 查询所有历史记录
      const fromDate = params?.fromDate || '2020-01-01';
      const toDate = params?.toDate || '2050-01-01';

      const data = await temporalGraphQLClient.request<{
        organizationHistory: any[];
      }>(
        TEMPORAL_QUERIES.ORGANIZATION_HISTORY,
        { code, fromDate, toDate }
      );

      return data.organizationHistory.map(transformToTemporalOrganization);
    } catch (error) {
      console.error(`历史查询失败 [code=${code}]:`, error);
      throw new Error(`无法查询组织历史数据: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  },

  /**
   * 获取组织的时间线事件（从历史记录转换）
   * @param code 组织代码
   * @param params 查询参数
   * @returns 时间线事件列表
   */
  async getOrganizationTimeline(
    code: string,
    params?: TemporalQueryParams
  ): Promise<TimelineEvent[]> {
    try {
      const history = await this.getOrganizationHistory(code, params);
      
      // 将历史记录转换为时间线事件
      const timeline: TimelineEvent[] = history.map((record, index) => ({
        id: `${record.code}-${record.effective_date}`,
        eventType: index === 0 ? 'CREATE' : 'UPDATE',
        eventDate: record.effective_date,
        description: `${record.name} - ${record.change_reason || '组织变更'}`,
        organizationCode: record.code,
        organizationName: record.name,
        changedFields: [], // 可以根据需要计算变更字段
        metadata: {
          status: record.status,
          unit_type: record.unit_type,
          effective_date: record.effective_date,
          end_date: record.end_date,
          is_current: record.is_current
        }
      }));

      return timeline;
    } catch (error) {
      console.error(`时间线查询失败 [code=${code}]:`, error);
      throw new Error(`无法生成组织时间线: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  },

  /**
   * 批量时间点查询 - 支持多个组织的同一时间点查询
   * @param codes 组织代码数组
   * @param asOfDate 查询时间点
   * @returns 时间点查询结果映射
   */
  async getBatchOrganizationsAsOfDate(
    codes: string[],
    asOfDate: string
  ): Promise<Record<string, TemporalOrganizationUnit | null>> {
    try {
      const results = await Promise.allSettled(
        codes.map(code => this.getOrganizationAsOfDate(code, asOfDate))
      );

      const resultMap: Record<string, TemporalOrganizationUnit | null> = {};
      
      codes.forEach((code, index) => {
        const result = results[index];
        if (result.status === 'fulfilled') {
          resultMap[code] = result.value;
        } else {
          console.warn(`批量查询失败 [code=${code}]:`, result.reason);
          resultMap[code] = null;
        }
      });

      return resultMap;
    } catch (error) {
      console.error(`批量时间点查询失败:`, error);
      throw new Error(`批量查询失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  },

  /**
   * 获取组织在特定时间范围内的变更统计
   * @param code 组织代码
   * @param params 查询参数
   * @returns 变更统计信息
   */
  async getOrganizationChangeStats(
    code: string,
    params?: {
      fromDate?: string;
      toDate?: string;
    }
  ): Promise<{
    totalChanges: number;
    changesByType: Record<string, number>;
    firstChange: string;
    lastChange: string;
    averageChangeInterval: number; // 天数
  }> {
    try {
      const history = await this.getOrganizationHistory(code, params);
      
      if (history.length <= 1) {
        return {
          totalChanges: 0,
          changesByType: {},
          firstChange: '',
          lastChange: '',
          averageChangeInterval: 0
        };
      }

      const changes = history.slice(1); // 排除当前记录
      const changesByType: Record<string, number> = {};
      
      changes.forEach(change => {
        const reason = change.change_reason || 'UNKNOWN';
        changesByType[reason] = (changesByType[reason] || 0) + 1;
      });

      // 计算平均变更间隔
      const dates = history.map(h => new Date(h.effective_date)).sort((a, b) => a.getTime() - b.getTime());
      const intervals = dates.slice(1).map((date, index) => 
        (date.getTime() - dates[index].getTime()) / (1000 * 60 * 60 * 24) // 转换为天数
      );
      const averageChangeInterval = intervals.length > 0 
        ? intervals.reduce((sum, interval) => sum + interval, 0) / intervals.length 
        : 0;

      return {
        totalChanges: changes.length,
        changesByType,
        firstChange: dates[0]?.toISOString().split('T')[0] || '',
        lastChange: dates[dates.length - 1]?.toISOString().split('T')[0] || '',
        averageChangeInterval: Math.round(averageChangeInterval)
      };
    } catch (error) {
      console.error(`变更统计查询失败 [code=${code}]:`, error);
      throw new Error(`无法获取变更统计: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  }
};

// 导出默认API和时态API
export default temporalAPI;