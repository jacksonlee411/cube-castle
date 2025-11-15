import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { unifiedRESTClient } from '@/shared/api';
import { createQueryError } from '@/shared/api/queryClient';
import { logger } from '@/shared/utils/logger';
import type { APIResponse } from '@/shared/types/api';
import {
  POSITIONS_QUERY_ROOT_KEY,
  VACANT_POSITIONS_QUERY_ROOT_KEY,
} from './useEnterprisePositions';
import { invalidateTemporalDetail } from '@/shared/api/invalidation';
import type {
  CreatePositionRequest,
  UpdatePositionRequest,
  CreatePositionVersionRequest,
  PositionResource,
} from '@/shared/types/positions';

interface PositionCommandPayload {
  code: string;
  status: string;
  recordId?: string;
  effectiveDate?: string;
  operationReason?: string;
  [key: string]: unknown;
}

export interface TransferPositionVariables {
  code: string;
  targetOrganizationCode: string;
  effectiveDate: string;
  operationReason: string;
  reassignReports?: boolean;
}

export interface PositionCommandResult {
  payload: PositionCommandPayload;
  requestId?: string;
  timestamp: string;
}

const ensurePositionSuccess = <T>(
  response: APIResponse<T>,
  fallbackMessage: string,
): T => {
  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? fallbackMessage, {
      code: response.error?.code,
      details: response.error?.details,
      requestId: response.requestId,
    });
  }
  return response.data;
};

/**
 * 统一入口：转发到 SSoT 失效工具，避免在此处分散维护键名
 */
const invalidatePositionCaches = (client: QueryClient, code?: string) => {
  invalidateTemporalDetail(client, 'position', code);
};

export const useCreatePosition = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: CreatePositionRequest): Promise<PositionResource> => {
      logger.mutation('[Mutation] Create position', variables);
      const response = await unifiedRESTClient.request<APIResponse<PositionResource>>('/positions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(variables),
      });

      logger.mutation('[Mutation] Create position response', response);
      return ensurePositionSuccess(response, '创建职位失败');
    },
    onSuccess: (resource) => {
      logger.mutation('[Mutation] Create position settled, refreshing caches', resource.code);
      invalidatePositionCaches(queryClient, resource.code);
    },
    onError: (error) => {
      logger.error('[Mutation] Create position failed', error);
    },
  });
};

export const useUpdatePosition = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: UpdatePositionRequest): Promise<PositionResource> => {
      const { code, ...payload } = variables;
      logger.mutation('[Mutation] Update position', { code, payload });
      const response = await unifiedRESTClient.request<APIResponse<PositionResource>>(`/positions/${code}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      logger.mutation('[Mutation] Update position response', response);
      return ensurePositionSuccess(response, '更新职位失败');
    },
    onSuccess: (_, variables) => {
      logger.mutation('[Mutation] Update position settled', variables.code);
      invalidatePositionCaches(queryClient, variables.code);
    },
    onError: (error, variables) => {
      logger.error('[Mutation] Update position failed', {
        code: variables.code,
        error,
      });
    },
  });
};

export const useCreatePositionVersion = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: CreatePositionVersionRequest): Promise<PositionResource> => {
      const { code, ...payload } = variables;
      logger.mutation('[Mutation] Create position version', { code, payload });
      const response = await unifiedRESTClient.request<APIResponse<PositionResource>>(`/positions/${code}/versions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      logger.mutation('[Mutation] Create position version response', response);
      return ensurePositionSuccess(response, '创建职位时态版本失败');
    },
    onSuccess: (_, variables) => {
      logger.mutation('[Mutation] Create position version settled', variables.code);
      invalidatePositionCaches(queryClient, variables.code);
    },
    onError: (error, variables) => {
      logger.error('[Mutation] Create position version failed', {
        code: variables.code,
        error,
      });
    },
  });
};

export const useTransferPosition = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: TransferPositionVariables): Promise<PositionCommandResult> => {
      const { code, targetOrganizationCode, effectiveDate, operationReason, reassignReports } = variables;
      logger.mutation('[Mutation] Transfer position', {
        code,
        targetOrganizationCode,
        effectiveDate,
        reassignReports,
      });

      const response = await unifiedRESTClient.request<APIResponse<PositionCommandPayload>>(
        `/positions/${code}/transfer`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            targetOrganizationCode,
            effectiveDate,
            operationReason,
            reassignReports,
          }),
        },
      );

      logger.mutation('[Mutation] Transfer response', response);
      const payload = ensurePositionSuccess(response, '转移职位失败');

      return {
        payload,
        requestId: response.requestId,
        timestamp: response.timestamp,
      };
    },
    onSuccess: (_, variables) => {
      logger.mutation('[Mutation] Transfer settled, refreshing caches', variables.code);
      // 列表与统计类仍按根键失效；详情键统一交由 SSoT 工具处理
      queryClient.invalidateQueries({ queryKey: POSITIONS_QUERY_ROOT_KEY, exact: false });
      queryClient.invalidateQueries({ queryKey: VACANT_POSITIONS_QUERY_ROOT_KEY, exact: false });
      invalidatePositionCaches(queryClient, variables.code);
    },
    onError: (error, variables) => {
      logger.error('[Mutation] Transfer position failed', {
        code: variables.code,
        error,
      });
    },
  });
};
