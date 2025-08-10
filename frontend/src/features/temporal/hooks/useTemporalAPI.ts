import { useMutation, useQueryClient } from '@tanstack/react-query';
// import { organizationsApi } from '../../shared/api/organizations-simplified';
import type { PlannedOrganizationData } from '../components/PlannedOrganizationForm';

/**
 * 创建计划组织的钩子
 */
export const useCreatePlannedOrganization = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: PlannedOrganizationData) => {
      // 调用计划组织专用API端点
      const response = await fetch('http://localhost:9090/api/v1/organization-units/planned', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || errorData.message || '创建计划组织失败');
      }

      return response.json();
    },

    onSuccess: () => {
      // 成功后刷新组织列表
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organizationStats'] });
    },
  });
};

/**
 * 时态查询钩子
 */
export const useTemporalQuery = () => {
  const queryClient = useQueryClient();

  return {
    // 按时间点查询组织
    queryByPointInTime: async (pointInTime: string) => {
      const response = await fetch(`http://localhost:8090/graphql`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query($pointInTime: String!) {
              organizations(pointInTime: $pointInTime) {
                code name status effective_date end_date
              }
            }
          `,
          variables: { pointInTime }
        })
      });
      return response.json();
    },

    // 按时间范围查询
    queryByDateRange: async (effectiveDateFrom?: string, effectiveDateTo?: string) => {
      const response = await fetch(`http://localhost:8090/graphql`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query($effectiveDateFrom: String, $effectiveDateTo: String) {
              organizations(effectiveDateFrom: $effectiveDateFrom, effectiveDateTo: $effectiveDateTo) {
                code name status effective_date end_date
              }
            }
          `,
          variables: { effectiveDateFrom, effectiveDateTo }
        })
      });
      return response.json();
    },

    // 按时态状态查询
    queryByTemporalStatus: async (temporalStatus: string) => {
      const response = await fetch(`http://localhost:8090/graphql`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query($status: String!) {
              organizations(status: $status) {
                code name status effective_date end_date
              }
            }
          `,
          variables: { status: temporalStatus }
        })
      });
      return response.json();
    },

    // 仅查询时态组织
    queryTemporalOnly: async () => {
      const response = await fetch(`http://localhost:8090/graphql`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query {
              organizations(temporalOnly: true) {
                code name status effective_date end_date is_temporal
              }
            }
          `
        })
      });
      return response.json();
    },
  };
};

/**
 * 更新组织时态信息的钩子
 */
export const useUpdateTemporalInfo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ code, ...data }: { code: string } & Partial<PlannedOrganizationData>) => {
      const response = await fetch(`http://localhost:9090/api/v1/organization-units/${code}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || errorData.message || '更新时态信息失败');
      }

      return response.json();
    },

    onSuccess: () => {
      // 成功后刷新相关查询
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['organizationStats'] });
    },
  });
};