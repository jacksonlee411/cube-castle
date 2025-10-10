#!/usr/bin/env node
/**
 * openapi-to-json.js
 *
 * 将 docs/api/openapi.yaml 中与组织域相关的枚举和字段约束
 * 转换为统一的 JSON 中间契约（shared/contracts/organization.json）。
 *
 * 该文件作为后续 Go/TS 代码生成的唯一事实来源。
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const yaml = require('js-yaml');

const PROJECT_ROOT = path.resolve(__dirname, '../..');
const OPENAPI_PATH = path.join(PROJECT_ROOT, 'docs/api/openapi.yaml');
const OUTPUT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');

const ORGANIZATION_SCHEMA_NAME = 'OrganizationUnit';
const CREATE_REQUEST_SCHEMA_NAME = 'CreateOrganizationUnitRequest';

function readOpenAPIDoc() {
  const raw = fs.readFileSync(OPENAPI_PATH, 'utf8');
  const parsed = yaml.load(raw);
  return { raw, parsed };
}

function getSchema(schemas, name) {
  return (schemas && schemas[name]) || {};
}

function getEnumValues(schema) {
  if (!schema || !Array.isArray(schema.enum)) {
    return [];
  }
  return schema.enum.slice();
}

function extractStringConstraints(definition = {}) {
  const result = {};
  if (typeof definition.pattern === 'string') {
    result.pattern = definition.pattern;
  }
  if (typeof definition.maxLength === 'number') {
    result.maxLength = definition.maxLength;
  }
  if (typeof definition.minLength === 'number') {
    result.minLength = definition.minLength;
  }
  if (definition.format) {
    result.format = definition.format;
  }
  if (definition.default !== undefined) {
    result.default = definition.default;
  }
  if (definition.nullable === true) {
    result.nullable = true;
  }
  if (definition.example !== undefined) {
    result.example = definition.example;
  }
  if (definition.description) {
    result.description = definition.description;
  }
  return result;
}

function extractNumericConstraints(definition = {}) {
  const result = {};
  if (typeof definition.minimum === 'number') {
    result.min = definition.minimum;
  }
  if (typeof definition.maximum === 'number') {
    result.max = definition.maximum;
  }
  if (definition.default !== undefined) {
    result.default = definition.default;
  }
  if (definition.example !== undefined) {
    result.example = definition.example;
  }
  if (definition.description) {
    result.description = definition.description;
  }
  return result;
}

function removeEmpty(obj) {
  if (!obj || typeof obj !== 'object') {
    return obj;
  }
  const cleaned = {};
  Object.entries(obj).forEach(([key, value]) => {
    if (value === undefined) {
      return;
    }
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      const nested = removeEmpty(value);
      if (Object.keys(nested).length > 0) {
        cleaned[key] = nested;
      }
      return;
    }
    if (Array.isArray(value)) {
      if (value.length > 0) {
        cleaned[key] = value;
      }
      return;
    }
    cleaned[key] = value;
  });
  return cleaned;
}

function buildContract(openapi, rawContent) {
  const schemas = (openapi.components && openapi.components.schemas) || {};
  const organizationSchema = getSchema(schemas, ORGANIZATION_SCHEMA_NAME);
  const organizationProps = organizationSchema.properties || {};
  const createSchema = getSchema(schemas, CREATE_REQUEST_SCHEMA_NAME);
  const createProps = createSchema.properties || {};

  const metadata = {
    source: 'docs/api/openapi.yaml',
    openapiVersion: openapi.info && openapi.info.version ? openapi.info.version : null,
    generatedAt: new Date().toISOString(),
    schemaSha256: crypto.createHash('sha256').update(rawContent).digest('hex'),
  };

  const enums = {
    unitType: getEnumValues(getSchema(schemas, 'UnitType')),
    status: getEnumValues(getSchema(schemas, 'Status')),
    operationType: getEnumValues(getSchema(schemas, 'OperationType')),
  };

  const constraints = removeEmpty({
    code: extractStringConstraints(organizationProps.code),
    parentCode: extractStringConstraints(createProps.parentCode || organizationProps.parentCode),
    name: extractStringConstraints(organizationProps.name),
    description: extractStringConstraints(organizationProps.description),
    level: extractNumericConstraints(organizationProps.level),
    sortOrder: extractNumericConstraints(createProps.sortOrder || organizationProps.sortOrder),
    operationReason: extractStringConstraints(createProps.operationReason),
    effectiveDate: extractStringConstraints(createProps.effectiveDate || organizationProps.effectiveDate),
  });

  const contract = removeEmpty({
    metadata,
    enums,
    constraints,
  });

  return contract;
}

function writeContractFile(contract) {
  const outputDir = path.dirname(OUTPUT_PATH);
  fs.mkdirSync(outputDir, { recursive: true });
  fs.writeFileSync(OUTPUT_PATH, `${JSON.stringify(contract, null, 2)}\n`, 'utf8');
}

function main() {
  try {
    const { raw, parsed } = readOpenAPIDoc();
    const contract = buildContract(parsed, raw);
    writeContractFile(contract);

    console.log('[OpenAPI] ✓ 契约已提取');
    console.log(`  → ${OUTPUT_PATH}`);
    if (contract.enums && contract.enums.unitType) {
      console.log(`  → UnitType: ${contract.enums.unitType.join(', ')}`);
    }
    if (contract.enums && contract.enums.status) {
      console.log(`  → Status: ${contract.enums.status.join(', ')}`);
    }
  } catch (error) {
    console.error('[OpenAPI] ✗ 解析失败:', error.message);
    process.exit(1);
  }
}

main();
