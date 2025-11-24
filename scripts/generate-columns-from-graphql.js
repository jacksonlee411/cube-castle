#!/usr/bin/env node

/**
 * Generate GraphQL column manifest information for UI tables.
 *
 * Usage: node scripts/generate-columns-from-graphql.js
 */

const fs = require('fs');
const path = require('path');
const { parse, buildASTSchema, isObjectType } = require('graphql');

const SCHEMA_PATH = path.resolve(__dirname, '..', 'docs', 'api', 'schema.graphql');
const OUTPUT_DIR = path.resolve(__dirname, '..', 'logs', 'plan402', 'ui');

const TYPE_WHITELIST = [
  'Organization',
  'OrganizationTimelineVersion',
  'OrganizationConnection',
  'Position',
  'PositionConnection',
  'StandardObject',
  'StandardObjectKernel',
  'StandardObjectVersion',
];

function timestamp() {
  const now = new Date();
  const pad = (n) => String(n).padStart(2, '0');
  return (
    now.getUTCFullYear().toString() +
    pad(now.getUTCMonth() + 1) +
    pad(now.getUTCDate()) +
    pad(now.getUTCHours()) +
    pad(now.getUTCMinutes()) +
    pad(now.getUTCSeconds())
  );
}

function loadSchema() {
  if (!fs.existsSync(SCHEMA_PATH)) {
    console.error(`[columns] GraphQL schema not found at ${SCHEMA_PATH}`);
    process.exit(1);
  }
  const sdl = fs.readFileSync(SCHEMA_PATH, 'utf8');
  const ast = parse(sdl);
  return buildASTSchema(ast);
}

function describeType(graphqlType) {
  if (!graphqlType) return 'unknown';
  if (graphqlType.toString) {
    return graphqlType.toString();
  }
  return String(graphqlType);
}

function collectColumns(schema) {
  return TYPE_WHITELIST.map((typeName) => {
    const type = schema.getType(typeName);
    if (!type || !isObjectType(type)) {
      return null;
    }
    const fields = Object.values(type.getFields()).map((field) => ({
      name: field.name,
      type: describeType(field.type),
      description: (field.description || '').replace(/\s+/g, ' ').trim(),
      deprecationReason: field.deprecationReason || undefined,
    }));
    return {
      typeName,
      fieldCount: fields.length,
      fields,
    };
  }).filter(Boolean);
}

function main() {
  const schema = loadSchema();
  const columns = collectColumns(schema);
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  const outFile = path.join(OUTPUT_DIR, `${timestamp()}-columns.log`);
  const payload = {
    generatedAt: new Date().toISOString(),
    source: path.relative(process.cwd(), SCHEMA_PATH),
    typesAnalyzed: columns.length,
    columns,
  };
  fs.writeFileSync(outFile, JSON.stringify(payload, null, 2));
  console.log(`[columns] GraphQL column manifest written to ${outFile}`);
}

main();

