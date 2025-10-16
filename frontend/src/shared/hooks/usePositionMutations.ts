import { useMutation, useQueryClient } from '@tanstack/react-query';
import { unifiedRESTClient } from '@/shared/api';
import { createQueryError } from '@/shared/api/queryClient';
import { logger } from '@/shared/utils/logger';
import type { APIResponse } from '@/shared/types/api';
import {
  POSITIONS_QUERY_ROOT_KEY,
  POSITION_DETAIL_QUERY_ROOT_KEY,
  VACANT_POSITIONS_QUERY_ROOT_KEY,
  positionDetailQueryKey,
} from './useEnterprisePositions';

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
      queryClient.invalidateQueries({ queryKey: POSITIONS_QUERY_ROOT_KEY, exact: false });
      queryClient.invalidateQueries({ queryKey: VACANT_POSITIONS_QUERY_ROOT_KEY, exact: false });
      queryClient.invalidateQueries({ queryKey: POSITION_DETAIL_QUERY_ROOT_KEY, exact: false });

      queryClient.invalidateQueries({
        queryKey: positionDetailQueryKey(variables.code),
        exact: false,
      });
    },
    onError: (error, variables) => {
      logger.error('[Mutation] Transfer position failed', {
        code: variables.code,
        error,
      });
    },
  });
};
