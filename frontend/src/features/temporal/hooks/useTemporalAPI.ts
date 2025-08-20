import { useMutation, useQueryClient } from '@tanstack/react-query';
// import { organizationsApi } from '../../shared/api/organizations';
import type { PlannedOrganizationData } from '../components/PlannedOrganizationForm';

// ❌ 已移除 useCreatePlannedOrganization - 简化时态管理API
// 使用基础创建API统一处理，通过status字段区分
// 
// 原功能替代方案：
// const response = await fetch('http://localhost:9090/api/v1/organization-units', {
//   method: 'POST',
//   body: JSON.stringify({ ...data, status: 'PLANNED' }),
// });

/**
 * 时态查询钩子
 */
export const useTemporalQuery = () => {
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

// ❌ 已移除 useUpdateTemporalInfo - 功能重复
// 使用基础更新API统一处理时态信息更新
// 
// 原功能保留：直接使用 PUT /api/v1/organization-units/{code} 即可