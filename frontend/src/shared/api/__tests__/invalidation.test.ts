import { describe, it, expect, vi, beforeEach } from 'vitest';
import { QueryClient } from '@tanstack/react-query';
import {
  POSITIONS_QUERY_ROOT_KEY,
  VACANT_POSITIONS_QUERY_ROOT_KEY,
  POSITION_DETAIL_QUERY_ROOT_KEY,
  positionDetailQueryKey,
} from '@/shared/hooks/useEnterprisePositions';
import { temporalEntityDetailQueryKey } from '@/shared/hooks/useTemporalEntityDetail';
import { invalidateTemporalDetail } from '../invalidation';

describe('invalidateTemporalDetail (SSoT)', () => {
  let client: QueryClient;

  beforeEach(() => {
    client = new QueryClient();
  });

  it('invalidates position keys (list + detail variants)', async () => {
    const spy = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const code = 'P1234567';
    invalidateTemporalDetail(client, 'position', code);

    // 列表类键
    expect(spy).toHaveBeenCalledWith({ queryKey: POSITIONS_QUERY_ROOT_KEY, exact: false });
    expect(spy).toHaveBeenCalledWith({ queryKey: VACANT_POSITIONS_QUERY_ROOT_KEY, exact: false });
    expect(spy).toHaveBeenCalledWith({ queryKey: POSITION_DETAIL_QUERY_ROOT_KEY, exact: false });

    // 详情类键（职位自身）
    expect(spy).toHaveBeenCalledWith({ queryKey: positionDetailQueryKey(code, false), exact: false });
    expect(spy).toHaveBeenCalledWith({ queryKey: positionDetailQueryKey(code, true), exact: false });

    // 统一时态详情键（与 241 Hook/Loader 对齐）
    expect(spy).toHaveBeenCalledWith({
      queryKey: temporalEntityDetailQueryKey('position', code, {}),
      exact: false,
    });
    expect(spy).toHaveBeenCalledWith({
      queryKey: temporalEntityDetailQueryKey('position', code, { includeDeleted: true }),
      exact: false,
    });
  });

  it('no code: still invalidates list/root detail keys', async () => {
    const spy = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    invalidateTemporalDetail(client, 'position', undefined);
    expect(spy).toHaveBeenCalledWith({ queryKey: POSITIONS_QUERY_ROOT_KEY, exact: false });
    expect(spy).toHaveBeenCalledWith({ queryKey: VACANT_POSITIONS_QUERY_ROOT_KEY, exact: false });
    expect(spy).toHaveBeenCalledWith({ queryKey: POSITION_DETAIL_QUERY_ROOT_KEY, exact: false });
  });
});
