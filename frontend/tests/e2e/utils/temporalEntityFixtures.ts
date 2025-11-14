/**
 * Temporal Entity Fixtures (Phase 1)
 * - Single source for E2E fixtures with entity-type aware accessors
 * - Phase 1 bridges to existing position fixtures without breaking tests
 *
 * NOTE:
 * - Do not duplicate field names or enums here; all shapes must follow schema.graphql
 * - In Phase 2, organization/job-catalog fixtures can be added alongside position
 */

import {
  POSITION_FIXTURE_CODE,
  POSITION_GRAPHQL_FIXTURES,
  POSITIONS_QUERY_NAME,
  POSITION_DETAIL_QUERY_NAME,
  VACANT_POSITIONS_QUERY_NAME,
  POSITION_HEADCOUNT_STATS_QUERY_NAME,
} from './positionFixtures';

type EntityType = 'position' | 'organization';

export type TemporalEntityFixtures = {
  graphql: unknown;
  helpers?: Record<string, unknown>;
};

export function createFixtures(entity: EntityType): TemporalEntityFixtures {
  if (entity === 'position') {
    return {
      graphql: POSITION_GRAPHQL_FIXTURES,
      helpers: {
        code: POSITION_FIXTURE_CODE,
        queries: {
          positions: POSITIONS_QUERY_NAME,
          positionDetail: POSITION_DETAIL_QUERY_NAME,
          vacantPositions: VACANT_POSITIONS_QUERY_NAME,
          headcountStats: POSITION_HEADCOUNT_STATS_QUERY_NAME,
        },
      },
    };
  }

  // Phase 1: organization fixtures are covered in tests inline (e.g., temporal-management-integration)
  // Returning an empty object keeps the API stable for gradual adoption
  return {
    graphql: {},
  };
}

// Re-export position constants for gradual adoption in existing tests if needed.
export {
  POSITION_FIXTURE_CODE,
  POSITION_GRAPHQL_FIXTURES,
  POSITIONS_QUERY_NAME,
  POSITION_DETAIL_QUERY_NAME,
  VACANT_POSITIONS_QUERY_NAME,
  POSITION_HEADCOUNT_STATS_QUERY_NAME,
};

