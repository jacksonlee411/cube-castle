import type { QueryClient } from '@tanstack/react-query';
import {
  POSITIONS_QUERY_ROOT_KEY,
  VACANT_POSITIONS_QUERY_ROOT_KEY,
  POSITION_DETAIL_QUERY_ROOT_KEY,
  positionDetailQueryKey,
} from '@/shared/hooks/useEnterprisePositions';
import { temporalEntityDetailQueryKey } from '@/shared/hooks/useTemporalEntityDetail';

export type TemporalEntity = 'position' | 'organization';

/**
 * SSoT: 统一的缓存失效工具
 * - 仅在此处集中罗列命令→查询键映射，禁止在业务处手写键名
 * - 目前实现 position；organization 预留扩展
 */
export const invalidateTemporalDetail = (
  client: QueryClient,
  entity: TemporalEntity,
  code?: string,
): void => {
  if (entity === 'position') {
    // 列表与详情根键
    client.invalidateQueries({ queryKey: POSITIONS_QUERY_ROOT_KEY, exact: false });
    client.invalidateQueries({ queryKey: VACANT_POSITIONS_QUERY_ROOT_KEY, exact: false });
    client.invalidateQueries({ queryKey: POSITION_DETAIL_QUERY_ROOT_KEY, exact: false });

    if (code) {
      // 详情（含 includeDeleted 变体）
      client.invalidateQueries({ queryKey: positionDetailQueryKey(code, false), exact: false });
      client.invalidateQueries({ queryKey: positionDetailQueryKey(code, true), exact: false });
      // 统一时态详情键（供 241 Hook/Loader 使用）
      client.invalidateQueries({ queryKey: temporalEntityDetailQueryKey('position', code, {}), exact: false });
      client.invalidateQueries({ queryKey: temporalEntityDetailQueryKey('position', code, { includeDeleted: true }), exact: false });
    }
    return;
  }

  if (entity === 'organization') {
    // 组织：预留（落地 241/244 后补全）；此处仅示意 entity 维度键失效
    if (code) {
      client.invalidateQueries({ queryKey: temporalEntityDetailQueryKey('organization', code, {}), exact: false });
      client.invalidateQueries({ queryKey: temporalEntityDetailQueryKey('organization', code, { asOfDate: null }), exact: false });
    }
    return;
  }
};

export default {
  invalidateTemporalDetail,
};

