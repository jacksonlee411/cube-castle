/**
 * Temporal Entity selectors - single source of truth for testids used by components and tests.
 * Phase 1: Map to existing testids to avoid breaking changes.
 * Phase 2: Gradually migrate component data-testid values to temporal-* prefixed IDs.
 *
 * NOTE: Do not hard-code data-testid strings in components/tests.
 * Import and use selectors defined here instead.
 *
 * // TODO-TEMPORARY(2025-12-15): Keep organization/position mappings for one iteration as compatibility.
 * Reason: allow gradual migration from legacy testids to temporal-* without breaking tests.
 * Plan: replace remaining position-* usages in tests, then remove compatibility mappings.
 */

export type TemporalSelectors = {
  page: {
    wrapper: string;
    timeline: string;
    timelineNode?: string;
    lifecycleBadge?: string;
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
  form?: {
    field?: {
      name?: string;
      description?: string;
      effectiveDate?: string;
      unitType?: string;
    };
    actions?: {
      startInsertVersion?: string;
      editHistoryToggle?: string;
      formClose?: string;
      cancelEditHistory?: string;
      submitEditHistory?: string;
      formCancel?: string;
      formSubmit?: string;
      deleteRecordWrapper?: string;
      deleteOrganization?: string;
      deleteRecord?: string;
    };
    messages?: {
      error?: string;
      success?: string;
    };
    contentWrapper?: string;
  };
  // Compatibility namespaces (Phase 1)
  organization: {
    dashboardWrapper: string;
    dashboard: string;
    form: string;
    stateChange?: {
      activateButton?: string;
      suspendButton?: string;
      dateInput?: string;
      reasonInput?: string;
      cancel?: string;
      confirm?: string;
    };
    table: string;
    createButton?: string;
    importButton?: string;
    exportButton?: string;
    treeRetryButton?: string;
  };
  position: {
    table: string;
    dashboard: string;
    temporalPageWrapper: string;
    temporalPage: string;
    overviewCard: string;
    versionToolbar: string;
    versionList: string;
    form?: (mode: string) => string;
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
    // headcount dashboard internals (Phase 1)
    headcountOrgInput?: string;
    headcountIncludeSubordinates?: string;
    headcountExportButton?: string;
    headcountLevelTable?: string;
    headcountTypeTable?: string;
    headcountFamilyTable?: string;
    vacantTable?: string;
    // Transfer dialog
    transferOpen?: string;
    transferTarget?: string;
    transferDate?: string;
    transferReason?: string;
    transferReassign?: string;
    transferConfirm?: string;
  };
  widgets?: {
    combobox?: {
      root?: string;
      input?: string;
      menu?: string;
      items?: string;
      item?: (code: string) => string;
      empty?: string;
    };
  };
};

export const temporalEntitySelectors: TemporalSelectors = {
  page: {
    wrapper: 'temporal-master-detail-view',
    timeline: 'temporal-timeline',
    timelineNode: 'temporal-timeline-node',
    lifecycleBadge: 'temporal-lifecycle-badge',
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
  form: {
    field: {
      name: 'form-field-name',
      description: 'form-field-description',
      effectiveDate: 'form-field-effective-date',
      unitType: 'form-field-unit-type',
    },
    actions: {
      startInsertVersion: 'start-insert-version-button',
      editHistoryToggle: 'edit-history-toggle-button',
      formClose: 'form-close-button',
      cancelEditHistory: 'cancel-edit-history-button',
      submitEditHistory: 'submit-edit-history-button',
      formCancel: 'form-cancel-button',
      formSubmit: 'form-submit-button',
      deleteRecordWrapper: 'temporal-delete-record-button-wrapper',
      deleteOrganization: 'temporal-delete-organization-button',
      deleteRecord: 'temporal-delete-record-button',
    },
    messages: {
      error: 'temporal-form-error',
      success: 'temporal-form-success',
    },
    contentWrapper: 'organization-form-content',
  },
  // Phase 1 compatibility mappings to existing component testids
  organization: {
    dashboardWrapper: 'temporal-organization-dashboard-wrapper',
    dashboard: 'temporal-organization-dashboard',
    form: 'organization-form',
    stateChange: {
      activateButton: 'activate-organization-button',
      suspendButton: 'suspend-organization-button',
      dateInput: 'status-change-date-input',
      reasonInput: 'status-change-reason-input',
      cancel: 'status-change-cancel',
      confirm: 'status-change-confirm',
    },
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
    form: (mode: string) => `temporal-position-form-${mode}`,
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
    headcountOrgInput: 'headcount-org-input',
    headcountIncludeSubordinates: 'headcount-include-subordinates',
    headcountExportButton: 'headcount-export',
    headcountLevelTable: 'headcount-level-table',
    headcountTypeTable: 'headcount-type-table',
    headcountFamilyTable: 'headcount-family-table',
    vacantTable: 'vacant-position-table',
    transferOpen: 'temporal-position-transfer-open',
    transferTarget: 'temporal-position-transfer-target',
    transferDate: 'temporal-position-transfer-date',
    transferReason: 'temporal-position-transfer-reason',
    transferReassign: 'temporal-position-transfer-reassign-checkbox',
    transferConfirm: 'temporal-position-transfer-confirm',
  },
  widgets: {
    combobox: {
      root: 'combobox',
      input: 'combobox-input',
      menu: 'combobox-menu',
      items: 'combobox-items',
      item: (code: string) => `combobox-item-${code}`,
      empty: 'combobox-empty',
    },
  },
} as const;

export default temporalEntitySelectors;
