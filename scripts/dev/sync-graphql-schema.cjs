#!/usr/bin/env node

/**
 * Fetches the running GraphQL schema via introspection and overwrites
 * docs/api/schema.graphql to keep the contract aligned with the live service.
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

async function main() {
  const data = await fetchSchema();
  const schemaSDL = `${printSchemaWithDirectives(buildClientSchema(data))}\n`;
  const targetPath = path.resolve(__dirname, '..', '..', 'docs', 'api', 'schema.graphql');
  fs.writeFileSync(targetPath, schemaSDL, 'utf-8');
  console.log(`✅ GraphQL schema updated at ${targetPath}`);
}

main().catch((error) => {
  console.error('❌ Failed to sync GraphQL schema:', error);
  process.exit(1);
});
