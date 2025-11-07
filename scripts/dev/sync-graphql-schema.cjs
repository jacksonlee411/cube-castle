#!/usr/bin/env node

/**
 * Fetches the running GraphQL schema via introspection and stores snapshots
 * under logs/graphql-snapshots for diff/diagnostics. Never overwrites
 * docs/api/schema.graphql, which remains the single source of truth.
 */

const fs = require('fs');
const path = require('path');
const { getIntrospectionQuery, buildClientSchema } = require('graphql');
const { printSchemaWithDirectives } = require('@graphql-tools/utils');

const GRAPHQL_ENDPOINT =
  process.env.GRAPHQL_ENDPOINT || process.env.E2E_GRAPHQL_API_URL || 'http://localhost:8090/graphql';
const DEFAULT_TENANT = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

const tenantId = process.env.PW_TENANT_ID || DEFAULT_TENANT;
const bearer = (process.env.PW_JWT || '').trim();

async function fetchSchema() {
  const headers = {
    'content-type': 'application/json',
    'x-tenant-id': tenantId,
  };
  if (bearer) {
    headers.Authorization = bearer.startsWith('Bearer ') ? bearer : `Bearer ${bearer}`;
  }

  const response = await fetch(GRAPHQL_ENDPOINT, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query: getIntrospectionQuery({ descriptions: true }) }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`GraphQL introspection failed (${response.status}): ${text}`);
  }

  const payload = await response.json();
  if (!payload || !payload.data) {
    throw new Error(`GraphQL introspection returned no data: ${JSON.stringify(payload)}`);
  }

  return payload.data;
}

function writeSnapshot(contents) {
  const outputDir = path.resolve(__dirname, '..', '..', 'logs', 'graphql-snapshots');
  fs.mkdirSync(outputDir, { recursive: true });

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const latestPath = path.join(outputDir, 'schema.latest.graphql');
  const timestampedPath = path.join(outputDir, `schema-${timestamp}.graphql`);

  fs.writeFileSync(latestPath, contents, 'utf-8');
  fs.writeFileSync(timestampedPath, contents, 'utf-8');

  console.log(`✅ GraphQL schema snapshot saved to ${latestPath}`);
  console.log(`🗃  Historical snapshot saved to ${timestampedPath}`);
}

async function main() {
  const data = await fetchSchema();
  const schemaSDL = `${printSchemaWithDirectives(buildClientSchema(data))}\n`;
  writeSnapshot(schemaSDL);
}

main().catch((error) => {
  console.error('❌ Failed to sync GraphQL schema:', error);
  process.exit(1);
});
