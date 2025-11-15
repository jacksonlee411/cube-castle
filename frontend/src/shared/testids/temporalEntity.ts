/**
 * Temporal Entity selectors - single source of truth for testids used by components and tests.
 * Phase 1: Map to existing testids to avoid breaking changes.
 * Phase 2: Gradually migrate component data-testid values to temporal-* prefixed IDs.
 *
 * NOTE: Do not hard-code data-testid strings in components/tests.
 * Import and use selectors defined here instead.
 *
 * // TODO-TEMPORARY: Keep organization/position mappings for 1 iteration, then switch to temporal-* IDs.
 * Deadline: next iteration end.
 */

export type TemporalSelectors = {
  page: {
    wrapper: string;
    timeline: string;
  };
  list: {
    table: string;
    rowPrefix: string;
    row: (code: string) => string;
  };
  action: {
    manageButton: (code: string) => string;
    deleteRecord: string;
  };
  // Compatibility namespaces (Phase 1)
  organization: {
    dashboardWrapper: string;
    dashboard: string;
    form: string;
    table: string;
  };
  position: {
    table: string;
    dashboard: string;
    temporalPageWrapper: string;
    temporalPage: string;
    overviewCard: string;
    versionToolbar: string;
    versionList: string;
    // Optional extensions used across components/tests
    mockBanner?: string;
    errorBox?: string;
    createButton?: string;
    searchInput?: string;
    statusFilter?: string;
    familyGroupFilter?: string;
    detailCard?: string;
    detailError?: string;
    editButton?: string;
    createVersionButton?: string;
    versionIncludeDeleted?: string;
    versionExportButton?: string;
    rowPrefix?: string;
    row?: (code: string) => string;
    versionRow?: (key: string) => string;
    versionRowPrefix?: string;
    tabVersions?: string;
    tabId?: (key: string) => string;
    vacancyBoard?: string;
    headcountDashboard?: string;
    // Transfer dialog
    transferOpen?: string;
    transferTarget?: string;
    transferDate?: string;
    transferReason?: string;
    transferReassign?: string;
    transferConfirm?: string;
  };
};

export const temporalEntitySelectors: TemporalSelectors = {
  page: {
    wrapper: 'temporal-master-detail-view',
    timeline: 'temporal-timeline',
  },
  list: {
    // Phase 1: map to existing table id if present in components; this can be migrated later
    table: 'temporal-entity-table',
    // Phase 1 compatibility: many tests use "table-row-{code}"
    rowPrefix: 'table-row-',
    row: (code: string) => `table-row-${code}`,
  },
  action: {
    manageButton: (code: string) => `temporal-manage-button-${code}`,
    deleteRecord: 'temporal-delete-record-button',
  },
  // Phase 1 compatibility mappings to existing component testids
  organization: {
    dashboardWrapper: 'temporal-organization-dashboard-wrapper',
    dashboard: 'temporal-organization-dashboard',
    form: 'organization-form',
    table: 'temporal-organization-table',
    createButton: 'temporal-organization-create-button',
    importButton: 'temporal-organization-import-button',
    exportButton: 'temporal-organization-export-button',
    treeRetryButton: 'temporal-organization-tree-retry-button',
  },
  position: {
    dashboard: 'temporal-position-dashboard',
    mockBanner: 'temporal-position-dashboard-mock-banner',
    errorBox: 'temporal-position-dashboard-error',
    createButton: 'temporal-position-create-button',
    searchInput: 'temporal-position-search-input',
    statusFilter: 'temporal-position-status-filter',
    familyGroupFilter: 'temporal-position-fg-filter',
    temporalPageWrapper: 'temporal-position-page-wrapper',
    temporalPage: 'temporal-position-page',
    overviewCard: 'temporal-position-overview-card',
    detailCard: 'temporal-position-detail-card',
    detailError: 'temporal-position-detail-error',
    editButton: 'temporal-position-edit-button',
    createVersionButton: 'temporal-position-create-version-button',
    versionToolbar: 'temporal-position-version-toolbar',
    versionList: 'temporal-position-version-list',
    tabVersions: 'position-tab-versions',
    versionIncludeDeleted: 'temporal-position-version-include-deleted',
    versionExportButton: 'temporal-position-version-export-button',
    tabId: (key: string) => `temporal-position-tab-${key}`,
    rowPrefix: 'temporal-position-row-',
    row: (code: string) => `temporal-position-row-${code}`,
    versionRow: (key: string) => `temporal-position-version-row-${key}`,
    versionRowPrefix: 'temporal-position-version-row-',
    vacancyBoard: 'temporal-position-vacancy-board',
    headcountDashboard: 'temporal-position-headcount-dashboard',
    transferOpen: 'temporal-position-transfer-open',
    transferTarget: 'temporal-position-transfer-target',
    transferDate: 'temporal-position-transfer-date',
    transferReason: 'temporal-position-transfer-reason',
    transferReassign: 'temporal-position-transfer-reassign-checkbox',
    transferConfirm: 'temporal-position-transfer-confirm',
  },
} as const;

export default temporalEntitySelectors;
