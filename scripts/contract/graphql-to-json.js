#!/usr/bin/env node
/**
 * graphql-to-json.js
 *
 * 将 docs/api/schema.graphql 中的组织域枚举写入共享契约文件，
 * 并对比 REST 契约中的枚举差异，帮助统一跨层定义。
 */

const fs = require('fs');
const path = require('path');

const PROJECT_ROOT = path.resolve(__dirname, '../..');
const GRAPHQL_SCHEMA_PATH = path.join(PROJECT_ROOT, 'docs/api/schema.graphql');
const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');

const ENUM_NAMES = [
  { key: 'unitType', name: 'UnitType' },
  { key: 'status', name: 'Status' },
  { key: 'operationType', name: 'OperationType' },
];

function readSchema() {
  return fs.readFileSync(GRAPHQL_SCHEMA_PATH, 'utf8');
}

function loadContract() {
  if (!fs.existsSync(CONTRACT_PATH)) {
    return {};
  }
  try {
    return JSON.parse(fs.readFileSync(CONTRACT_PATH, 'utf8'));
  } catch (error) {
    console.warn('[GraphQL] ⚠ 现有契约文件解析失败，将重新生成:', error.message);
    return {};
  }
}

function extractSchemaVersion(schemaText) {
  const match = schemaText.match(/Version:\s*([0-9.]+)/);
  return match ? match[1] : null;
}

function extractEnumValues(schemaText, enumName) {
  const pattern = new RegExp(`enum\\s+${enumName}\\s*\\{([\\s\\S]*?)\\}`, 'm');
  const match = pattern.exec(schemaText);
  if (!match) {
    return [];
  }
  const body = match[1];
  return body
    .split('\n')
    .map((line) => {
      const trimmed = line.split('#')[0].trim();
      if (!trimmed) return null;
      const token = trimmed.split(/\s+/)[0];
      return token || null;
    })
    .filter((value, index, arr) => value && arr.indexOf(value) === index);
}

function computeDiff(restValues = [], gqlValues = []) {
  const restSet = new Set(restValues);
  const gqlSet = new Set(gqlValues);

  const missingInRest = gqlValues.filter((value) => !restSet.has(value));
  const missingInGql = restValues.filter((value) => !gqlSet.has(value));

  return { missingInRest, missingInGql };
}

function main() {
  try {
    const schema = readSchema();
    const contract = loadContract();

    const graphqlEnums = {};
    ENUM_NAMES.forEach(({ key, name }) => {
      graphqlEnums[key] = extractEnumValues(schema, name);
    });

    const graphqlSection = {
      source: 'docs/api/schema.graphql',
      schemaVersion: extractSchemaVersion(schema),
      generatedAt: new Date().toISOString(),
      enums: graphqlEnums,
    };

    contract.graphql = graphqlSection;

    const output = JSON.stringify(contract, null, 2);
    fs.writeFileSync(CONTRACT_PATH, `${output}\n`, 'utf8');

    console.log('[GraphQL] ✓ 枚举已提取');
    Object.entries(graphqlEnums).forEach(([key, values]) => {
      console.log(`  → ${key}: ${values.join(', ') || '(空)'}`);
    });

    // 差异提示
    if (contract.enums) {
      ENUM_NAMES.forEach(({ key }) => {
        const restValues = contract.enums[key] || [];
        const gqlValues = graphqlEnums[key] || [];
        const { missingInRest, missingInGql } = computeDiff(restValues, gqlValues);

        if (missingInRest.length > 0) {
          console.warn(`[GraphQL] ⚠ ${key} 枚举在 REST 契约中缺失: ${missingInRest.join(', ')}`);
        }
        if (missingInGql.length > 0) {
          console.warn(`[GraphQL] ⚠ ${key} 枚举在 GraphQL 契约中缺失: ${missingInGql.join(', ')}`);
        }
      });
    }
  } catch (error) {
    console.error('[GraphQL] ✗ 解析失败:', error.message);
    process.exit(1);
  }
}

main();
