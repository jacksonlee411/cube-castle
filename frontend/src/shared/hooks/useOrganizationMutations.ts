import { logger } from '@/shared/utils/logger';
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { unifiedRESTClient } from "../api";
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
    const error = new Error(
      response.error?.message ?? fallbackMessage,
    ) as Error & {
      code?: string;
      details?: JsonValue;
    };
    if (response.error?.code) {
      error.code = response.error.code;
    }
    if (response.error?.details) {
      error.details = response.error.details;
    }
    throw error;
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
    onSettled: () => {
      logger.mutation("[Mutation] Create settled, invalidating queries");

      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({
        queryKey: ["organizations"],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-stats"],
        exact: false,
      });

      // 强制重新获取数据以确保立即显示新创建的组织
      queryClient.refetchQueries({
        queryKey: ["organizations"],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization-stats"],
        type: "active",
      });

      logger.mutation("[Mutation] Create cache invalidation and refetch completed");
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
    onSettled: (data, error, variables) => {
      logger.mutation("[Mutation] Update settled:", variables.code);

      // 立即失效所有相关查询缓存
      queryClient.invalidateQueries({
        queryKey: ["organizations"],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization", variables.code!],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-stats"],
        exact: false,
      });

      // 强制重新获取数据以确保立即显示更新的组织
      queryClient.refetchQueries({
        queryKey: ["organizations"],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization-stats"],
        type: "active",
      });

      // 新增：直接设置缓存数据以提供即时反馈
      if (data) {
        queryClient.setQueryData(["organization", variables.code!], data);
      }

      // 新增：移除过时的缓存数据
      queryClient.removeQueries({
        queryKey: ["organizations"],
        exact: false,
        type: "inactive",
      });

      logger.mutation("[Mutation] Update cache invalidation and refetch completed");
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
    onSettled: (result, error, variables) => {
      logger.mutation("[Mutation] Suspend settled:", variables.code);

      queryClient.invalidateQueries({
        queryKey: ["organizations"],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization", variables.code!],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-stats"],
        exact: false,
      });

      queryClient.refetchQueries({
        queryKey: ["organizations"],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization-stats"],
        type: "active",
      });

      if (result?.organization) {
        queryClient.setQueryData(
          ["organization", variables.code!],
          result.organization,
        );
      }

      logger.mutation(
        "[Mutation] Suspend cache invalidation and refetch completed",
      );
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
    onSettled: (result, error, variables) => {
      logger.mutation("[Mutation] Activate settled:", variables.code);

      queryClient.invalidateQueries({
        queryKey: ["organizations"],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization", variables.code!],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-stats"],
        exact: false,
      });

      queryClient.refetchQueries({
        queryKey: ["organizations"],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization-stats"],
        type: "active",
      });

      if (result?.organization) {
        queryClient.setQueryData(
          ["organization", variables.code!],
          result.organization,
        );
      }

      logger.mutation(
        "[Mutation] Activate cache invalidation and refetch completed",
      );
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
    onSettled: (_data, _error, variables) => {
      logger.mutation(
        "[Mutation] Temporal version create settled:",
        variables.code,
      );

      queryClient.invalidateQueries({
        queryKey: ["organizations"],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization", variables.code],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-history", variables.code],
        exact: false,
      });

      queryClient.invalidateQueries({
        queryKey: ["organization-stats"],
        exact: false,
      });

      queryClient.refetchQueries({
        queryKey: ["organizations"],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization", variables.code],
        type: "active",
      });

      queryClient.refetchQueries({
        queryKey: ["organization-stats"],
        type: "active",
      });

      logger.mutation("[Mutation] Temporal version cache refresh completed");
    },
  });
};
