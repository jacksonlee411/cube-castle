import { test, expect } from '@playwright/test';

const POSITIONS_QUERY_NAME = 'EnterprisePositions';
const POSITION_DETAIL_QUERY_NAME = 'PositionDetail';

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
  });
});
