import { logger } from '@/shared/utils/logger';
import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { unifiedRESTClient } from "../api";
import { createQueryError } from '../api/queryClient';
import {
  organizationByCodeQueryKey,
  ORGANIZATIONS_QUERY_ROOT_KEY,
} from './useEnterpriseOrganizations';
import type {
  OrganizationUnit,
  OrganizationRequest,
  APIResponse,
} from "../types";
import type { JsonValue } from '@/shared/types/json';

const ensureSuccess = <T>(
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

const formatIfMatchHeader = (etag: string): string => {
  const trimmed = etag.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.startsWith('"') || trimmed.startsWith("W/")) {
    return trimmed;
  }
  return `"${trimmed}"`;
};

const DEFAULT_DELETE_ORGANIZATION_REASON = '通过组织详情页删除组织编码';

const invalidateOrganizationsCache = (client: QueryClient) => {
  client.invalidateQueries({
    queryKey: ORGANIZATIONS_QUERY_ROOT_KEY,
    exact: false,
  });
};

const invalidateOrganizationDetailCache = (client: QueryClient, code?: string | null) => {
  if (!code) {
    return;
  }
  client.invalidateQueries({
    queryKey: organizationByCodeQueryKey(code),
    exact: false,
  });
};

const removeOrganizationDetailCache = (client: QueryClient, code?: string | null) => {
  if (!code) {
    return;
  }
  client.removeQueries({
    queryKey: organizationByCodeQueryKey(code),
    exact: false,
  });
};

export interface OrganizationStateMutationVariables {
  code: string;
  effectiveDate: string;
  operationReason?: string;
  currentETag?: string | null;
  idempotencyKey?: string;
}

export interface OrganizationStateMutationResult {
  organization: OrganizationUnit;
  etag: string | null;
  headers: Record<string, string>;
}

export interface DeleteOrganizationVariables {
  code: string;
  effectiveDate: string;
  operationReason?: string;
  currentETag?: string | null;
  idempotencyKey?: string;
}

interface DeleteOrganizationEventData {
  code: string;
  status: string;
  operationType: string;
  recordId: string | null;
  effectiveDate: string;
  operationReason?: string | null;
  timeline?: Array<Record<string, JsonValue>>;
}

export interface DeleteOrganizationMutationResult {
  payload: DeleteOrganizationEventData;
  etag: string | null;
  headers: Record<string, string>;
}

interface CreateOrganizationVersionPayload {
  code: string;
  name: string;
  unitType: OrganizationUnit["unitType"];
  parentCode: string;
  description?: string;
  sortOrder?: number;
  effectiveDate: string;
  endDate?: string;
  operationReason?: string;
}

interface CreateOrganizationVersionResponse {
  recordId: string;
  code: string;
  name: string;
  effectiveDate: string;
  status: string;
}

// 新增组织单元
export const useCreateOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: OrganizationRequest,
    ): Promise<OrganizationUnit> => {
      logger.mutation("[Mutation] Creating organization:", data);
      const response = await unifiedRESTClient.request<
        APIResponse<OrganizationUnit>
      >("/organization-units", {
        method: "POST",
        body: JSON.stringify(data),
        headers: { "Content-Type": "application/json" },
      });
      logger.mutation("[Mutation] Create successful:", response);
      return ensureSuccess(response, "创建组织失败");
    },
    onSuccess: (organization) => {
      logger.mutation("[Mutation] Create settled, refreshing caches");
      invalidateOrganizationsCache(queryClient);
      if (organization?.code) {
        queryClient.setQueryData(
          organizationByCodeQueryKey(organization.code),
          organization,
        );
      }
    },
    onError: (error) => {
      logger.error("[Mutation] Create organization failed:", error);
    },
  });
};

// 更新组织单元
export const useUpdateOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      data: OrganizationRequest,
    ): Promise<OrganizationUnit> => {
      logger.mutation("[Mutation] Updating organization:", data);
      const response = await unifiedRESTClient.request<
        APIResponse<OrganizationUnit>
      >(`/organization-units/${data.code}`, {
        method: "PUT",
        body: JSON.stringify(data),
        headers: { "Content-Type": "application/json" },
      });
      logger.mutation("[Mutation] Update successful:", response);
      return ensureSuccess(response, "更新组织失败");
    },
    onSuccess: (organization, variables) => {
      logger.mutation("[Mutation] Update settled:", variables.code);
      invalidateOrganizationsCache(queryClient);
      invalidateOrganizationDetailCache(queryClient, variables.code ?? organization?.code ?? null);

      if (organization?.code) {
        queryClient.setQueryData(
          organizationByCodeQueryKey(organization.code),
          organization,
        );
      }
      // 触发审计历史刷新（按前缀失效，避免缺少 recordId 无法定向失效）
      queryClient.invalidateQueries({ queryKey: ["auditHistory"], exact: false });
    },
    onError: (error, variables) => {
      logger.error("[Mutation] Update organization failed:", {
        code: variables.code,
        error,
      });
    },
  });
};

// === 新增：操作驱动状态管理Hooks ===

// 停用组织
export const useSuspendOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation<
    OrganizationStateMutationResult,
    Error,
    OrganizationStateMutationVariables
  >({
    mutationFn: async ({
      code,
      operationReason,
      effectiveDate,
      currentETag,
      idempotencyKey,
    }: OrganizationStateMutationVariables) => {
      logger.mutation("[Mutation] Suspending organization:", code, effectiveDate);
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (currentETag) {
        const formatted = formatIfMatchHeader(currentETag);
        if (formatted) {
          headers["If-Match"] = formatted;
        }
      }
      if (idempotencyKey) {
        headers["Idempotency-Key"] = idempotencyKey;
      }

      const body: Record<string, string> = {
        effectiveDate,
      };
      if (operationReason) {
        body.operationReason = operationReason;
      }

      const { data, headers: responseHeaders } =
        await unifiedRESTClient.request<APIResponse<OrganizationUnit>>(
          `/organization-units/${code}/suspend`,
          {
            method: "POST",
            body: JSON.stringify(body),
            headers,
            includeRawResponse: true,
          },
        );

      const organization = ensureSuccess(data, "暂停组织失败");
      const etag = responseHeaders["etag"] ?? null;
      logger.mutation("[Mutation] Suspend successful:", { code, etag });

      return {
        organization,
        etag,
        headers: responseHeaders,
      };
    },
    onSuccess: (result, variables) => {
      logger.mutation("[Mutation] Suspend settled:", variables.code);
      invalidateOrganizationsCache(queryClient);
      invalidateOrganizationDetailCache(queryClient, variables.code);

      if (result?.organization?.code) {
        queryClient.setQueryData(
          organizationByCodeQueryKey(result.organization.code),
          result.organization,
        );
      }
    },
    onError: (error, variables) => {
      logger.error("[Mutation] Suspend organization failed:", {
        code: variables.code,
        error,
      });
    },
  });
};

// 重新启用组织
export const useActivateOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation<
    OrganizationStateMutationResult,
    Error,
    OrganizationStateMutationVariables
  >({
    mutationFn: async ({
      code,
      operationReason,
      effectiveDate,
      currentETag,
      idempotencyKey,
    }: OrganizationStateMutationVariables) => {
      logger.mutation("[Mutation] Activating organization:", code, effectiveDate);
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (currentETag) {
        const formatted = formatIfMatchHeader(currentETag);
        if (formatted) {
          headers["If-Match"] = formatted;
        }
      }
      if (idempotencyKey) {
        headers["Idempotency-Key"] = idempotencyKey;
      }

      const body: Record<string, string> = {
        effectiveDate,
      };
      if (operationReason) {
        body.operationReason = operationReason;
      }

      const { data, headers: responseHeaders } =
        await unifiedRESTClient.request<APIResponse<OrganizationUnit>>(
          `/organization-units/${code}/activate`,
          {
            method: "POST",
            body: JSON.stringify(body),
            headers,
            includeRawResponse: true,
          },
        );

      const organization = ensureSuccess(data, "重新启用组织失败");
      const etag = responseHeaders["etag"] ?? null;
      logger.mutation("[Mutation] Activate successful:", { code, etag });

      return {
        organization,
        etag,
        headers: responseHeaders,
      };
    },
    onSuccess: (result, variables) => {
      logger.mutation("[Mutation] Activate settled:", variables.code);
      invalidateOrganizationsCache(queryClient);
      invalidateOrganizationDetailCache(queryClient, variables.code);

      if (result?.organization?.code) {
        queryClient.setQueryData(
          organizationByCodeQueryKey(result.organization.code),
          result.organization,
        );
      }
    },
    onError: (error, variables) => {
      logger.error("[Mutation] Activate organization failed:", {
        code: variables.code,
        error,
      });
    },
  });
};

export const useDeleteOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation<
    DeleteOrganizationMutationResult,
    Error,
    DeleteOrganizationVariables
  >({
    mutationFn: async ({
      code,
      effectiveDate,
      operationReason,
      currentETag,
      idempotencyKey,
    }: DeleteOrganizationVariables) => {
      logger.mutation('[Mutation] Deleting organization:', { code, effectiveDate });

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (currentETag) {
        const formatted = formatIfMatchHeader(currentETag);
        if (formatted) {
          headers['If-Match'] = formatted;
        }
      }

      if (idempotencyKey) {
        headers['Idempotency-Key'] = idempotencyKey;
      }

      const body: {
        eventType: string;
        effectiveDate: string;
        changeReason?: string;
      } = {
        eventType: 'DELETE_ORGANIZATION',
        effectiveDate,
      };

      const finalReason = (operationReason ?? DEFAULT_DELETE_ORGANIZATION_REASON).trim();
      if (finalReason) {
        body.changeReason = finalReason;
      }

      const { data, headers: responseHeaders } =
        await unifiedRESTClient.request<APIResponse<DeleteOrganizationEventData>>(
          `/organization-units/${code}/events`,
          {
            method: 'POST',
            body: JSON.stringify(body),
            headers,
            includeRawResponse: true,
          },
        );

      const payload = ensureSuccess(
        data,
        '删除组织失败',
      );

      const etag = responseHeaders['etag'] ?? null;
      logger.mutation('[Mutation] Delete organization successful:', { code, etag });

      return {
        payload,
        etag,
        headers: responseHeaders,
      };
    },
    onSuccess: (_result, variables) => {
      logger.mutation('[Mutation] Delete organization settled:', variables.code);
      invalidateOrganizationsCache(queryClient);
      removeOrganizationDetailCache(queryClient, variables.code);
    },
    onError: (error, variables) => {
      logger.error('[Mutation] Delete organization failed:', {
        code: variables.code,
        error,
      });
    },
  });
};

export const useCreateOrganizationVersion = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      payload: CreateOrganizationVersionPayload,
    ): Promise<CreateOrganizationVersionResponse> => {
      const { code, ...requestBody } = payload;
      const response = await unifiedRESTClient.request<
        APIResponse<CreateOrganizationVersionResponse>
      >(`/organization-units/${code}/versions`, {
        method: "POST",
        body: JSON.stringify(requestBody),
        headers: { "Content-Type": "application/json" },
      });
      return ensureSuccess(response, "创建时态版本失败");
    },
    onSuccess: (_data, variables) => {
      logger.mutation(
        "[Mutation] Temporal version create settled:",
        variables.code,
      );

      invalidateOrganizationsCache(queryClient);
      invalidateOrganizationDetailCache(queryClient, variables.code);

      queryClient.invalidateQueries({
        queryKey: ["organization-history", variables.code],
        exact: false,
      });
    },
    onError: (error, variables) => {
      logger.error("[Mutation] Temporal version create failed:", {
        code: variables.code,
        error,
      });
    },
  });
};
