import { test, expect } from '@playwright/test';

const POSITIONS_QUERY_NAME = 'EnterprisePositions';
const POSITION_DETAIL_QUERY_NAME = 'PositionDetail';
const VACANT_POSITIONS_QUERY_NAME = 'VacantPositions';
const POSITION_HEADCOUNT_STATS_QUERY_NAME = 'PositionHeadcountStats';

const GRAPHQL_FIXTURES = {
  positions: {
    data: {
      positions: {
        data: [
          {
            code: 'P-LIFECYCLE-001',
            title: '生命周期演示岗位',
            jobFamilyGroupCode: 'OPER',
            jobFamilyGroupName: '运营管理',
            jobFamilyCode: 'OPER-OPS',
            jobFamilyName: '运营支持',
            jobRoleCode: 'OPER-OPS-SUPV',
            jobRoleName: '运营主管',
            jobLevelCode: 'M2',
            jobLevelName: '经理二级',
            organizationCode: 'ORG-A',
            organizationName: '生命周期演示组织',
            positionType: 'REGULAR',
            employmentType: 'FULL_TIME',
            headcountCapacity: 2,
            headcountInUse: 1,
            availableHeadcount: 1,
            gradeLevel: null,
            reportsToPositionCode: 'P-LEAD-001',
            status: 'ACTIVE',
            effectiveDate: '2024-01-01',
            endDate: null,
            isCurrent: true,
            isFuture: false,
            createdAt: '2024-01-01T00:00:00.000Z',
            updatedAt: '2024-03-15T10:00:00.000Z',
          },
        ],
        pagination: {
          total: 1,
          page: 1,
          pageSize: 25,
          hasNext: false,
          hasPrevious: false,
        },
        totalCount: 1,
      },
    },
  },
  positionDetail: {
    data: {
      position: {
        code: 'P-LIFECYCLE-001',
        title: '生命周期演示岗位',
        jobFamilyGroupCode: 'OPER',
        jobFamilyGroupName: '运营管理',
        jobFamilyCode: 'OPER-OPS',
        jobFamilyName: '运营支持',
        jobRoleCode: 'OPER-OPS-SUPV',
        jobRoleName: '运营主管',
        jobLevelCode: 'M2',
        jobLevelName: '经理二级',
        organizationCode: 'ORG-A',
        organizationName: '生命周期演示组织',
        positionType: 'REGULAR',
        employmentType: 'FULL_TIME',
        headcountCapacity: 2,
        headcountInUse: 1,
        availableHeadcount: 1,
        gradeLevel: null,
        reportsToPositionCode: 'P-LEAD-001',
        status: 'ACTIVE',
        effectiveDate: '2024-01-01',
        endDate: null,
        isCurrent: true,
        isFuture: false,
        createdAt: '2024-01-01T00:00:00.000Z',
        updatedAt: '2024-03-15T10:00:00.000Z',
        currentAssignment: {
          assignmentId: 'ASSIGN-001',
          positionCode: 'P-LIFECYCLE-001',
          positionRecordId: 'POS-REC-001',
          employeeId: 'EMP-001',
          employeeName: '张三',
          employeeNumber: 'E001',
          assignmentType: 'PRIMARY',
          assignmentStatus: 'ACTIVE',
          fte: 1,
          startDate: '2024-03-01',
          endDate: null,
          isCurrent: true,
          notes: '夜班负责人',
          createdAt: '2024-03-01T08:00:00.000Z',
          updatedAt: '2024-03-05T09:00:00.000Z',
        },
      },
      positionTimeline: [
        {
          recordId: 'TIMELINE-001',
          status: 'ACTIVE',
          title: '岗位创建',
          effectiveDate: '2024-01-01',
          endDate: null,
          changeReason: '年度人力预算批准',
          isCurrent: false,
        },
        {
          recordId: 'TIMELINE-002',
          status: 'ACTIVE',
          title: '岗位扩编',
          effectiveDate: '2024-03-01',
          endDate: null,
          changeReason: '夜班运营需求增加',
          isCurrent: true,
        },
      ],
      positionAssignments: {
        data: [
          {
            assignmentId: 'ASSIGN-001',
            positionCode: 'P-LIFECYCLE-001',
            positionRecordId: 'POS-REC-001',
            employeeId: 'EMP-001',
            employeeName: '张三',
            employeeNumber: 'E001',
            assignmentType: 'PRIMARY',
            assignmentStatus: 'ACTIVE',
            fte: 1,
            startDate: '2024-03-01',
            endDate: null,
            isCurrent: true,
            notes: '夜班负责人',
            createdAt: '2024-03-01T08:00:00.000Z',
            updatedAt: '2024-03-05T09:00:00.000Z',
          },
          {
            assignmentId: 'ASSIGN-000',
            positionCode: 'P-LIFECYCLE-001',
            positionRecordId: 'POS-REC-000',
            employeeId: 'EMP-000',
            employeeName: '李四',
            employeeNumber: 'E000',
            assignmentType: 'PRIMARY',
            assignmentStatus: 'ENDED',
            fte: 1,
            startDate: '2023-01-01',
            endDate: '2024-02-28',
            isCurrent: false,
            notes: '内部轮岗结束',
            createdAt: '2023-01-01T08:00:00.000Z',
            updatedAt: '2024-02-28T09:00:00.000Z',
          },
        ],
      },
      positionTransfers: {
        data: [
          {
            transferId: 'TRANSFER-001',
            positionCode: 'P-LIFECYCLE-001',
            fromOrganizationCode: 'ORG-A',
            toOrganizationCode: 'ORG-B',
            effectiveDate: '2024-05-01',
            initiatedBy: {
              id: 'USER-001',
              name: '王五',
            },
            operationReason: '业务线整合',
            createdAt: '2024-04-25T12:30:00.000Z',
          },
        ],
      },
    },
  },
  vacantPositions: {
    data: {
      vacantPositions: {
        data: [
          {
            positionCode: 'P-VAC-001',
            organizationCode: 'ORG-B',
            organizationName: '缺编演示组织',
            jobFamilyCode: 'OPER-OPS',
            jobRoleCode: 'OPER-OPS-SUPV',
            jobLevelCode: 'S1',
            vacantSince: '2024-06-01',
            headcountCapacity: 3,
            headcountAvailable: 2,
            totalAssignments: 5,
          },
        ],
        pagination: {
          total: 1,
          page: 1,
          pageSize: 25,
          hasNext: false,
          hasPrevious: false,
        },
        totalCount: 1,
      },
    },
  },
  headcountStats: {
    data: {
      positionHeadcountStats: {
        organizationCode: 'ORG-A',
        organizationName: '生命周期演示组织',
        totalCapacity: 5,
        totalFilled: 3,
        totalAvailable: 2,
        fillRate: 0.6,
        byLevel: [
          {
            jobLevelCode: 'S1',
            capacity: 2,
            utilized: 1,
            available: 1,
          },
          {
            jobLevelCode: 'M2',
            capacity: 3,
            utilized: 2,
            available: 1,
          },
        ],
        byType: [
          {
            positionType: 'REGULAR',
            capacity: 4,
            filled: 3,
            available: 1,
          },
          {
            positionType: 'CONTRACT',
            capacity: 1,
            filled: 0,
            available: 1,
          },
        ],
        byFamily: [
          {
            jobFamilyCode: 'OPER-OPS',
            jobFamilyName: '运营支持',
            capacity: 5,
            utilized: 3,
            available: 2,
          },
        ],
      },
    },
  },
};

test.describe('职位生命周期视图', () => {
  test('展示任职与调动历史', async ({ page }) => {
    await page.route('**/graphql', async route => {
      const request = route.request();
      let body: { query?: string } | undefined;
      try {
        body = request.postDataJSON();
      } catch (_error) {
        return route.continue();
      }

      const query = body?.query;
      if (!query) {
        return route.continue();
      }

      if (query.includes(POSITIONS_QUERY_NAME)) {
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(GRAPHQL_FIXTURES.positions),
        });
      }

      if (query.includes(POSITION_DETAIL_QUERY_NAME)) {
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(GRAPHQL_FIXTURES.positionDetail),
        });
      }

      if (query.includes(VACANT_POSITIONS_QUERY_NAME)) {
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(GRAPHQL_FIXTURES.vacantPositions),
        });
      }

      if (query.includes(POSITION_HEADCOUNT_STATS_QUERY_NAME)) {
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(GRAPHQL_FIXTURES.headcountStats),
        });
      }

      return route.continue();
    });

    await page.goto('/positions');

    await expect(page.getByRole('heading', { name: '职位管理（Stage 1 数据接入）' })).toBeVisible();
    await expect(page.getByTestId('position-row-P-LIFECYCLE-001')).toBeVisible();

    const detailCard = page.getByTestId('position-detail-card');
    await expect(detailCard).toContainText('生命周期演示岗位');
    await expect(detailCard).toContainText('当前任职');
    await expect(detailCard).toContainText('张三');
    await expect(detailCard).toContainText('夜班负责人');
    await expect(detailCard).toContainText('任职历史');
    await expect(detailCard).toContainText('李四');
    await expect(detailCard).toContainText('调动记录');
    await expect(detailCard).toContainText('ORG-A → ORG-B');
    await expect(detailCard).toContainText('业务线整合');
    await expect(detailCard.getByRole('button', { name: '发起职位转移' })).toBeVisible();

    const vacancyBoard = page.getByTestId('position-vacancy-board');
    await expect(vacancyBoard).toContainText('空缺职位看板');
    await expect(vacancyBoard).toContainText('P-VAC-001');
    await expect(vacancyBoard).toContainText('缺编演示组织');
    await expect(vacancyBoard).toContainText('空缺职位数');

    const headcountDashboard = page.getByTestId('position-headcount-dashboard');
    await expect(headcountDashboard).toContainText('职位编制统计');
    await expect(headcountDashboard).toContainText('总编制');
    await expect(headcountDashboard).toContainText('5');
    await expect(headcountDashboard).toContainText('已占用');
    await expect(headcountDashboard).toContainText('3');
    await expect(page.getByTestId('headcount-level-table')).toContainText('S1');
    await expect(page.getByTestId('headcount-type-table')).toContainText('REGULAR');
    await expect(page.getByTestId('headcount-family-table')).toContainText('OPER-OPS');
  });
});
