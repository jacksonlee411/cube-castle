import { unifiedGraphQLClient } from './unified-client';
import { SimpleValidationError } from '../validation/simple-validation';

// 审计查询参数接口
export interface AuditQueryParams {
  startDate?: string;        // YYYY-MM-DD格式
  endDate?: string;         // YYYY-MM-DD格式  
  operation?: OperationType; // CREATE/UPDATE/SUSPEND等
  userId?: string;          // 操作人UUID
  limit?: number;           // 记录数量限制 (默认50)
}

// 操作类型枚举
export type OperationType = 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';

// 风险级别枚举
export type RiskLevel = 'LOW' | 'MEDIUM' | 'HIGH';

// 审计时间线条目 (v4.5.0 - 简化版本，移除复杂风险评估)
export interface AuditTimelineEntry {
  auditId: string;
  operation: OperationType;
  timestamp: string;
  userName: string;
  operationReason?: string;
  riskLevel?: RiskLevel; // 风险级别 (可选)
  changesSummary: {
    operationSummary: string;
    totalChanges: number;
    keyChanges: string[];
  };
}

// 操作人信息
export interface OperatedByInfo {
  id: string;
  name: string;
}

// 审计记录详细信息 (v4.6.0 - 精确到recordId)
export interface AuditRecordDetail {
  auditId: string;
  recordId: string;
  operationType: string;
  timestamp: string;
  operatedBy: {
    id: string;
    name: string;
  };
  changesSummary: string;
  operationReason?: string;
  beforeData?: string;
  afterData?: string;
}

// 组织审计历史响应结构 (兼容性类型)
export interface OrganizationAuditHistory {
  organizationCode: string;
  entityName?: string;
  totalVersions?: number;
  auditTimeline: AuditTimelineEntry[];
  meta: {
    totalAuditRecords: number;
    queryStartDate: string | null;
    queryEndDate: string | null;
    pageSize: number;
    operationsSummary?: {
      create: number;
      update: number;
      suspend: number;
      reactivate: number;
      delete: number;
    };
    dateRange?: {
      earliest: string;
      latest: string;
    };
  };
}


/**
 * 审计日志API客户端类 - v4.6.0
 * 基于recordId的精确审计追踪实现
 */
export class AuditAPI {
  /**
   * 获取指定组织code的审计历史记录（兼容性方法）
   * @param code 组织代码
   * @param params 查询参数
   * @returns Promise<OrganizationAuditHistory>
   */
  static async getOrganizationAuditHistory(
    code: string, 
    params: AuditQueryParams = {}
  ): Promise<OrganizationAuditHistory> {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Organization code is required' }
        ]);
      }

      // 首先通过code获取当前组织的recordId和基本信息
      const orgQuery = `
        query GetOrganizationRecordId($code: String!) {
          organization(code: $code) {
            recordId
            name
            code
          }
        }
      `;

      const orgData = await unifiedGraphQLClient.request<{
        organization: { recordId: string; name: string; code: string } | null;
      }>(orgQuery, { code });

      if (!orgData.organization?.recordId) {
        throw new Error(`Organization with code "${code}" not found`);
      }

      const recordId = orgData.organization.recordId;
      const organizationName = orgData.organization.name;

      // 调用现有的recordId审计历史方法
      const auditHistory = await this.getRecordAuditHistory(recordId, params);

      // 转换为OrganizationAuditHistory格式
      const auditTimeline: AuditTimelineEntry[] = auditHistory.map(detail => {
        // 简单的风险级别评估逻辑
        let riskLevel: RiskLevel = 'LOW';
        if (detail.operationType === 'DELETE') {
          riskLevel = 'HIGH';
        } else if (detail.operationType === 'SUSPEND') {
          riskLevel = 'MEDIUM';
        }

        return {
          auditId: detail.auditId,
          operation: detail.operationType as OperationType,
          timestamp: detail.timestamp,
          userName: detail.operatedBy.name,
          operationReason: detail.operationReason,
          riskLevel,
          changesSummary: {
            operationSummary: detail.changesSummary,
            totalChanges: 1, // 简化版本，每个审计记录算作1个变更
            keyChanges: [detail.changesSummary]
          }
        };
      });

      // 计算操作统计
      const operationsSummary = {
        create: auditHistory.filter(a => a.operationType === 'CREATE').length,
        update: auditHistory.filter(a => a.operationType === 'UPDATE').length,
        suspend: auditHistory.filter(a => a.operationType === 'SUSPEND').length,
        reactivate: auditHistory.filter(a => a.operationType === 'REACTIVATE').length,
        delete: auditHistory.filter(a => a.operationType === 'DELETE').length
      };

      // 计算日期范围
      const timestamps = auditHistory.map(a => a.timestamp).sort();
      const dateRange = timestamps.length > 0 ? {
        earliest: timestamps[0],
        latest: timestamps[timestamps.length - 1]
      } : {
        earliest: new Date().toISOString(),
        latest: new Date().toISOString()
      };

      return {
        organizationCode: code,
        entityName: organizationName,
        totalVersions: auditHistory.length, // 简化版本：审计记录数作为版本数
        auditTimeline,
        meta: {
          totalAuditRecords: auditHistory.length,
          queryStartDate: params.startDate || null,
          queryEndDate: params.endDate || null,
          pageSize: params.limit || 50,
          operationsSummary,
          dateRange
        }
      };

    } catch (error: unknown) {
      console.error('Error fetching organization audit history:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      if (error instanceof Error) {
        if (error.message.includes('not found')) {
          throw new Error(`组织 ${code} 不存在或无审计记录`);
        }
        if (error.message.includes('Network') || error.message.includes('fetch')) {
          throw new Error(`无法获取组织 ${code} 的审计历史，请检查网络连接`);
        }
        if (error.message.includes('401') || error.message.includes('403')) {
          throw new Error(`没有权限查看组织 ${code} 的审计历史`);
        }
      }
      
      throw new Error(`获取组织 ${code} 审计历史失败，请重试`);
    }
  }
  /**
   * 获取指定recordId的审计历史记录
   * @param recordId 组织记录的唯一标识(UUID)
   * @param params 查询参数
   * @returns Promise<AuditLogDetail[]>
   */
  static async getRecordAuditHistory(
    recordId: string, 
    params: AuditQueryParams = {}
  ): Promise<AuditRecordDetail[]> {
    try {
      if (!recordId || typeof recordId !== 'string') {
        throw new SimpleValidationError('Invalid record ID', [
          { field: 'recordId', message: 'Record ID is required' }
        ]);
      }

      // 构建GraphQL查询 - v4.6.0 基于recordId
      const query = `
        query GetAuditHistory(
          $recordId: String!
          $startDate: String
          $endDate: String  
          $operation: String
          $userId: String
          $limit: Int
        ) {
          auditHistory(
            recordId: $recordId
            startDate: $startDate
            endDate: $endDate
            operation: $operation
            userId: $userId
            limit: $limit
          ) {
            auditId
            recordId
            operationType
            operatedBy {
              id
              name
            }
            changesSummary
            operationReason
            timestamp
            beforeData
            afterData
          }
        }
      `;

      const variables = {
        recordId,
        startDate: params.startDate || null,
        endDate: params.endDate || null,
        operation: params.operation || null,
        userId: params.userId || null,
        limit: params.limit || 50
      };

      const data = await unifiedGraphQLClient.request<{
        auditHistory: AuditRecordDetail[];
      }>(query, variables);

      if (!data.auditHistory) {
        throw new Error(`No audit history found for record ${recordId}`);
      }

      // 直接返回后端数据
      return data.auditHistory;

    } catch (error: unknown) {
      console.error('Error fetching record audit history:', recordId, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      if (error instanceof Error) {
        if (error.message.includes('Network') || error.message.includes('fetch')) {
          throw new Error(`无法获取记录 ${recordId} 的审计历史，请检查网络连接`);
        }
        if (error.message.includes('401') || error.message.includes('403')) {
          throw new Error(`没有权限查看记录 ${recordId} 的审计历史`);
        }
        if (error.message.includes('404')) {
          throw new Error(`记录 ${recordId} 不存在或无审计记录`);
        }
      }
      
      throw new Error(`获取记录 ${recordId} 审计历史失败，请重试`);
    }
  }


  /**
   * 获取单个审计记录详情 (v4.6.0)
   * @param auditId 审计记录ID
   * @returns Promise<AuditRecordDetail>
   */
  static async getAuditLogDetail(auditId: string): Promise<AuditRecordDetail> {
    try {
      if (!auditId || typeof auditId !== 'string') {
        throw new SimpleValidationError('Invalid audit ID', [
          { field: 'auditId', message: 'Audit ID is required' }
        ]);
      }

      const query = `
        query GetAuditLogDetail($auditId: String!) {
          auditLog(auditId: $auditId) {
            auditId
            recordId
            operationType
            operatedBy {
              id
              name
            }
            changesSummary
            operationReason
            timestamp
            beforeData
            afterData
          }
        }
      `;

      const data = await unifiedGraphQLClient.request<{
        auditLog: AuditRecordDetail;
      }>(query, { auditId });

      if (!data.auditLog) {
        throw new Error(`Audit record ${auditId} not found`);
      }

      return data.auditLog;

    } catch (error: unknown) {
      console.error('Error fetching audit log detail:', auditId, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      if (error instanceof Error) {
        if (error.message.includes('404')) {
          throw new Error(`审计记录 ${auditId} 不存在`);
        }
        if (error.message.includes('403')) {
          throw new Error(`没有权限查看审计记录 ${auditId}`);
        }
      }
      
      throw new Error(`获取审计记录 ${auditId} 详情失败，请重试`);
    }
  }
}

// 导出API类和所有类型
export default AuditAPI;