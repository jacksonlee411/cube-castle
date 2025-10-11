import { describe, expect, it } from 'vitest';
import { OrganizationStatusEnum, OrganizationUnitTypeEnum } from '../../types/contract_gen';
import { __internal } from '../useEnterpriseOrganizations';

const {
  normalizeQueryParams,
  buildGraphQLVariables,
  transformOrganizationsResponse,
  mapOrganizationStats,
} = __internal;

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

  it('builds GraphQL variables with filter and pagination', () => {
    const params = normalizeQueryParams({
      page: 2,
      pageSize: 25,
      parentCode: 'ROOT001',
      level: 2,
      includeHistorical: true,
      asOfDate: '2025-01-01',
    });

    const variables = buildGraphQLVariables(params);
    expect(variables.pagination).toEqual({
      page: 2,
      pageSize: 25,
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
});
