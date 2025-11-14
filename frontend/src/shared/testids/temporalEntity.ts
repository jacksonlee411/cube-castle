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
  };
  position: {
    dashboard: string;
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
    dashboardWrapper: 'organization-dashboard-wrapper',
  },
  position: {
    dashboard: 'position-dashboard',
  },
} as const;

export default temporalEntitySelectors;
