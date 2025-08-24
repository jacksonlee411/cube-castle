/**
 * 组织详情GraphQL查询客户端
 * 专门用于时态查询功能：organizationAsOfDate 和 organizationHistory
 */
import type { 
  GraphQLResponse,
} from '../types';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit,
  TimelineEvent
} from '../types/temporal';
import { 
  convertGraphQLToTemporalOrganizationUnit,
  type GraphQLOrganizationData,
  logTypeSyncReport,
  TEMPORAL_ORGANIZATION_UNIT_FIELDS
} from '../types/converters';
import { authManager } from './auth';

// GraphQL端点 - CQRS查询服务（8090端口）
const TEMPORAL_GRAPHQL_ENDPOINT = 'http://localhost:8090/graphql';

interface TemporalGraphQLClient {
  request<T>(query: string, variables?: Record<string, unknown>): Promise<T>;
}

const temporalGraphQLClient: TemporalGraphQLClient = {
  async request<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    // 获取OAuth访问令牌
    const accessToken = await authManager.getAccessToken();
    
    const response = await fetch(TEMPORAL_GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
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
        tenantId
        code
        parentCode
        name
        unitType
        status
        level
        path
        sortOrder
        description
        profile
        createdAt
        updatedAt
        effectiveDate
        endDate
        version
        isCurrent
        changeReason
        validFrom
        validTo
      }
    }
  `,

  // 历史查询 - 查询时间范围内的所有历史记录
  ORGANIZATION_HISTORY: `
    query OrganizationHistory($code: String!, $fromDate: String!, $toDate: String!) {
      organizationHistory(code: $code, fromDate: $fromDate, toDate: $toDate) {
        tenantId
        code
        parentCode
        name
        unitType
        status
        level
        path
        sortOrder
        description
        profile
        createdAt
        updatedAt
        effectiveDate
        endDate
        version
        isCurrent
        changeReason
        validFrom
        validTo
      }
    }
  `,

  // 传统查询保持兼容 - 当前数据
  ORGANIZATIONS_CURRENT: `
    query OrganizationsCurrent($first: Int, $offset: Int, $searchText: String) {
      organizations(first: $first, offset: $offset, searchText: $searchText) {
        tenantId
        code
        parentCode
        name
        unitType
        status
        level
        path
        sortOrder
        description
        profile
        createdAt
        updatedAt
        effectiveDate
        endDate
        version
        isCurrent
      }
    }
  `,

  // 单个组织查询 - 当前数据
  ORGANIZATION_CURRENT: `
    query OrganizationCurrent($code: String!) {
      organization(code: $code) {
        tenantId
        code
        parentCode
        name
        unitType
        status
        level
        path
        sortOrder
        description
        profile
        createdAt
        updatedAt
        effectiveDate
        endDate
        version
        isCurrent
      }
    }
  `
};

// 使用统一的类型转换器 - 已移动到converters.ts
// 时态数据转换器现在使用标准化的convertGraphQLToTemporalOrganizationUnit函数

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
        organizationAsOfDate: GraphQLOrganizationData | null;
      }>(
        TEMPORAL_QUERIES.ORGANIZATION_AS_OF_DATE,
        { code, asOfDate }
      );

      if (!data.organizationAsOfDate) {
        return null;
      }

      // 开发期间类型同步检查
      if (process.env.NODE_ENV === 'development') {
        logTypeSyncReport(
          'organizationAsOfDate',
          data.organizationAsOfDate,
          TEMPORAL_ORGANIZATION_UNIT_FIELDS
        );
      }

      return convertGraphQLToTemporalOrganizationUnit(data.organizationAsOfDate);
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
        organizationHistory: GraphQLOrganizationData[];
      }>(
        TEMPORAL_QUERIES.ORGANIZATION_HISTORY,
        { code, fromDate, toDate }
      );

      // 开发期间类型同步检查（仅检查第一个记录）
      if (process.env.NODE_ENV === 'development' && data.organizationHistory.length > 0) {
        logTypeSyncReport(
          'organizationHistory[0]',
          data.organizationHistory[0],
          TEMPORAL_ORGANIZATION_UNIT_FIELDS
        );
      }

      return data.organizationHistory.map(convertGraphQLToTemporalOrganizationUnit);
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
        organizationCode: record.code,
        timestamp: new Date(record.effective_date), // 修正：使用timestamp而非eventDate
        type: index === 0 ? 'organization_created' : 'organization_updated', // 修正：使用定义的EventType
        title: `${record.name} - ${record.change_reason || '组织变更'}`, // 修正：使用title而非description
        description: record.change_reason || '组织变更',
        changes: [], // 可以根据需要计算变更字段
        status: record.is_current ? 'active' : 'completed', // 修正：添加status字段
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