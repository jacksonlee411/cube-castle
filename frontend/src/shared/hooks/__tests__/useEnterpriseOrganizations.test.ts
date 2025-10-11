import { describe, expect, it, vi, beforeEach } from 'vitest';
import type { Mock } from 'vitest';
import { OrganizationStatusEnum, OrganizationUnitTypeEnum } from '../../types/contract_gen';
import { graphqlEnterpriseAdapter } from '../../api/graphql-enterprise-adapter';
import { __internal } from '../useEnterpriseOrganizations';

vi.mock('../../api/graphql-enterprise-adapter', () => ({
  graphqlEnterpriseAdapter: {
    request: vi.fn(),
  },
}));

const {
  normalizeQueryParams,
  mergeQueryParams,
  buildGraphQLVariables,
  transformOrganizationsResponse,
  mapOrganizationStats,
  fetchOrganizationsWithParams,
  fetchOrganizationDetail,
  organizationsQueryKey,
  organizationByCodeQueryKey,
} = __internal;

const mockedRequest = graphqlEnterpriseAdapter.request as unknown as Mock;

beforeEach(() => {
  mockedRequest.mockReset();
});

describe('useEnterpriseOrganizations internals', () => {
  it('normalizes query params and clamps bounds', () => {
    const params = normalizeQueryParams({
      page: -3,
      pageSize: 5000,
      searchText: ' 研发部 ',
      unitType: OrganizationUnitTypeEnum.Department,
      status: OrganizationStatusEnum.Active,
      sortOrder: 'DESC',
    });

    expect(params.page).toBe(1);
    expect(params.pageSize).toBe(1000);
    expect(params.searchText).toBe('研发部');
    expect(params.unitType).toBe(OrganizationUnitTypeEnum.Department);
    expect(params.status).toBe(OrganizationStatusEnum.Active);
    expect(params.sortOrder).toBe('desc');
  });

  it('normalizes asOfDate from effectiveDate source', () => {
    const params = normalizeQueryParams({
      effectiveDate: '2025-02-01',
      sortBy: 'name',
    });

    expect(params.asOfDate).toBe('2025-02-01');
    expect(params.sortBy).toBe('name');
  });

  it('builds GraphQL variables with filter and pagination', () => {
    const params = normalizeQueryParams({
      page: 2,
      pageSize: 25,
      parentCode: 'ROOT001',
      level: 2,
      includeHistorical: true,
      asOfDate: '2025-01-01',
      sortBy: 'name',
      sortOrder: 'DESC',
    });

    const variables = buildGraphQLVariables(params);
    expect(variables.pagination).toEqual({
      page: 2,
      pageSize: 25,
      sortBy: 'name',
      sortOrder: 'desc',
    });
    expect(variables.filter).toEqual({
      parentCode: 'ROOT001',
      level: 2,
      includeFuture: true,
      asOfDate: '2025-01-01',
    });
    expect(variables.statsAsOfDate).toBe('2025-01-01');
    expect(variables.statsIncludeHistorical).toBe(true);
  });

  it('transforms GraphQL response to domain result', () => {
    const timestamp = '2025-10-10T00:00:00.000Z';
    const params = normalizeQueryParams();
    const result = transformOrganizationsResponse(
      {
        organizations: {
          data: [
            {
              code: 'UNIT001',
              parentCode: 'ROOT',
              tenantId: 'TENANT',
              name: '组织一部',
              unitType: OrganizationUnitTypeEnum.Department,
              status: OrganizationStatusEnum.Active,
              level: 1,
              sortOrder: 1,
              description: '测试部门',
              createdAt: timestamp,
              updatedAt: timestamp,
            },
          ],
          pagination: {
            total: 1,
            page: 1,
            pageSize: 50,
            hasNext: false,
            hasPrevious: false,
          },
          temporal: {
            asOfDate: timestamp,
            currentCount: 1,
            futureCount: 0,
            historicalCount: 0,
          },
        },
        organizationStats: {
          totalCount: 1,
          activeCount: 1,
          inactiveCount: 0,
          plannedCount: 0,
          deletedCount: 0,
          byType: [{ unitType: OrganizationUnitTypeEnum.Department, count: 1 }],
          byStatus: [{ status: OrganizationStatusEnum.Active, count: 1 }],
          byLevel: [{ level: 1, count: 1 }],
          temporalStats: {
            totalVersions: 1,
            averageVersionsPerOrg: 1,
            oldestEffectiveDate: timestamp,
            newestEffectiveDate: timestamp,
          },
        },
      },
      params,
      timestamp,
    );

    expect(result.organizations).toHaveLength(1);
    expect(result.organizations[0]).toMatchObject({
      code: 'UNIT001',
      name: '组织一部',
      unitType: OrganizationUnitTypeEnum.Department,
      status: OrganizationStatusEnum.Active,
    });
    expect(result.pagination.total).toBe(1);
    expect(result.stats?.totalCount).toBe(1);
    expect(result.temporal?.currentCount).toBe(1);
  });

  it('maps organization stats safely when enums mismatch', () => {
    const stats = mapOrganizationStats({
      totalCount: 2,
      activeCount: 1,
      inactiveCount: 1,
      plannedCount: 0,
      deletedCount: 0,
      byType: [
        { unitType: 'DEPARTMENT', count: 1 },
        { unitType: 'UNKNOWN', count: 1 },
      ],
      byStatus: [
        { status: 'ACTIVE', count: 1 },
        { status: 'INVALID', count: 1 },
      ],
      byLevel: [{ level: 1, count: 2 }],
      temporalStats: {
        totalVersions: 2,
        averageVersionsPerOrg: 1,
        oldestEffectiveDate: null,
        newestEffectiveDate: null,
      },
    });

    expect(stats?.byType).toHaveLength(1);
    expect(stats?.byType[0].unitType).toBe(OrganizationUnitTypeEnum.Department);
    expect(stats?.byStatus).toHaveLength(1);
    expect(stats?.byStatus[0].status).toBe(OrganizationStatusEnum.Active);
  });

  it('merges query params with patch updates', () => {
    const base = normalizeQueryParams({ page: 1, pageSize: 10, searchText: '初始' });
    const merged = mergeQueryParams(base, { page: 4, searchText: ' 更新 ' });

    expect(merged.page).toBe(4);
    expect(merged.pageSize).toBe(10);
    expect(merged.searchText).toBe('更新');
  });

  it('fetches organizations with parameters and transforms payload', async () => {
    const params = normalizeQueryParams({ page: 2, pageSize: 10, includeHistorical: true });
    mockedRequest.mockResolvedValueOnce({
      success: true,
      data: {
        organizations: {
          data: [
            {
              code: 'UNIT001',
              parentCode: 'ROOT',
              tenantId: 'TENANT',
              name: '组织一部',
              unitType: OrganizationUnitTypeEnum.Department,
              status: OrganizationStatusEnum.Active,
              level: 1,
              sortOrder: 1,
              description: '测试部门',
              createdAt: '2025-10-10T00:00:00.000Z',
              updatedAt: '2025-10-10T00:00:00.000Z',
            },
          ],
          pagination: {
            total: 1,
            page: 1,
            pageSize: 10,
            hasNext: false,
            hasPrevious: false,
          },
        },
        organizationStats: null,
      },
      message: 'ok',
      timestamp: '2025-10-10T00:00:00.000Z',
      requestId: 'req-1',
    });

    const result = await fetchOrganizationsWithParams(params);
    expect(result.organizations).toHaveLength(1);
    expect(result.pagination.pageSize).toBe(10);
    expect(mockedRequest).toHaveBeenCalledTimes(1);
  });

  it('throws query error when organization fetch fails', async () => {
    mockedRequest.mockResolvedValueOnce({
      success: false,
      error: { code: 'GRAPHQL_ERROR', message: 'boom' },
      timestamp: '2025-10-10T00:00:00.000Z',
      requestId: 'req-err',
    });

    await expect(fetchOrganizationsWithParams(normalizeQueryParams())).rejects.toThrow('boom');
  });

  it('fetches organization detail and maps response', async () => {
    mockedRequest.mockResolvedValueOnce({
      success: true,
      data: {
        organization: {
          code: 'UNIT002',
          parentCode: 'ROOT',
          name: '组织二部',
          unitType: OrganizationUnitTypeEnum.Department,
          status: OrganizationStatusEnum.Active,
          level: 1,
          sortOrder: 1,
          description: null,
          createdAt: '2025-10-10T00:00:00.000Z',
          updatedAt: '2025-10-10T00:00:00.000Z',
        },
      },
      timestamp: '2025-10-10T00:00:00.000Z',
      requestId: 'req-detail',
    });

    const detail = await fetchOrganizationDetail('UNIT002', undefined, undefined);
    expect(detail?.code).toBe('UNIT002');
    expect(mockedRequest).toHaveBeenCalledTimes(1);
  });

  it('throws when fetching organization detail fails', async () => {
    mockedRequest.mockResolvedValueOnce({
      success: false,
      error: { code: 'NOT_FOUND', message: 'not found' },
      timestamp: '2025-10-10T00:00:00.000Z',
      requestId: 'req-detail-err',
    });

    await expect(fetchOrganizationDetail('UNKNOWN')).rejects.toThrow('not found');
  });

  it('returns null when organization detail payload is missing', async () => {
    mockedRequest.mockResolvedValueOnce({
      success: true,
      data: { organization: null },
      timestamp: '2025-10-10T00:00:00.000Z',
      requestId: 'req-null',
    });

    const detail = await fetchOrganizationDetail('MISSING');
    expect(detail).toBeNull();
  });

  it('produces stable query keys', () => {
    const params = normalizeQueryParams({ page: 5, includeHistorical: true });
    const queryKey = organizationsQueryKey(params);
    expect(queryKey[0]).toBe('organizations');
    expect(queryKey[1].page).toBe(5);

    const detailKey = organizationByCodeQueryKey('UNIT123', '2025-01-01');
    expect(detailKey).toEqual(['organization-detail', 'UNIT123', '2025-01-01']);
  });
});
