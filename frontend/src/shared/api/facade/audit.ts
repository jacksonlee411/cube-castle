/**
 * Plan 257 - 领域 Facade（Audit）
 * 提供审计历史查询的 GraphQL 封装。
 */
import { unifiedGraphQLClient } from '@/shared/api/unified-client';

export interface AuditHistoryGraphQLChange {
  field: string;
  oldValue: string | null;
  newValue: string | null;
  dataType: string;
}
export interface AuditHistoryGraphQLEntry {
  auditId: string;
  recordId: string;
  operation: string;
  timestamp: string;
  operationReason?: string | null;
  beforeData?: string | null;
  afterData?: string | null;
  modifiedFields: string[];
  changes: AuditHistoryGraphQLChange[];
}

export interface AuditQueryParams {
  limit?: number;
  startDate?: string | null;
  endDate?: string | null;
  operation?: string | null;
  userId?: string | null;
}

export async function getAuditHistory(recordId: string, params: AuditQueryParams = {}): Promise<AuditHistoryGraphQLEntry[]> {
  const QUERY = /* GraphQL */ `
    query TemporalEntityAuditHistory($recordId: String!, $limit: Int, $startDate: String, $endDate: String, $operation: OperationType, $userId: String) {
      auditHistory(recordId: $recordId, limit: $limit, startDate: $startDate, endDate: $endDate, operation: $operation, userId: $userId) {
        auditId
        recordId
        operation
        timestamp
        operationReason
        beforeData
        afterData
        modifiedFields
        changes {
          field
          oldValue
          newValue
          dataType
        }
      }
    }
  `;
  const res = await unifiedGraphQLClient.request<{ auditHistory: AuditHistoryGraphQLEntry[] }>(QUERY, {
    recordId,
    limit: params.limit ?? 50,
    startDate: params.startDate ?? null,
    endDate: params.endDate ?? null,
    operation: params.operation ?? null,
    userId: params.userId ?? null
  });
  return res.auditHistory ?? [];
}

