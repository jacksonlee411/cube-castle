#!/usr/bin/env node
/**
 * drift-check.js
 *
 * 对比 OpenAPI ↔ GraphQL 的契约片段（以 shared/contracts/organization.json 为输入），
 * 检测主要枚举（UnitType/Status/OperationType）的差异，输出报告并可在发现差异时失败。
 * 可选：启用字段矩阵（OrganizationUnit ↔ Organization）对比（报告模式或阻断）。
 *
 * 用法:
 *   node scripts/contract/drift-check.js [--fail-on-diff] [--include-fields] [--fail-on-fields]
 *
 * 产出:
 *   reports/contracts/drift-report.json
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const PROJECT_ROOT = path.resolve(__dirname, '../..');
const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');
const REPORT_DIR = path.join(PROJECT_ROOT, 'reports', 'contracts');
const REPORT_PATH = path.join(REPORT_DIR, 'drift-report.json');
const ALLOWLIST_PATH = path.join(PROJECT_ROOT, 'scripts', 'contract', 'drift-allowlist.json');
const OPENAPI_PATH = path.join(PROJECT_ROOT, 'docs', 'api', 'openapi.yaml');
const GRAPHQL_SCHEMA_PATH = path.join(PROJECT_ROOT, 'docs', 'api', 'schema.graphql');

const ENUM_KEYS = ['unitType', 'status', 'operationType'];

function readJSON(p) {
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

function readAllowlist() {
  try {
    if (!fs.existsSync(ALLOWLIST_PATH)) {
      return { enum: {}, fields: {} };
    }
    const obj = readJSON(ALLOWLIST_PATH);
    // 结构容错
    return {
      enum: (obj && obj.enum) || {},
      fields: (obj && obj.fields) || {},
    };
  } catch (e) {
    console.warn('⚠ 漂移白名单解析失败（忽略处理）:', e.message);
    return { enum: {}, fields: {} };
  }
}

function computeEnumDiff(restArr = [], gqlArr = []) {
  const rest = new Set(restArr);
  const gql = new Set(gqlArr);
  const missingInRest = gqlArr.filter(v => !rest.has(v));
  const missingInGql = restArr.filter(v => !gql.has(v));
  return { missingInRest, missingInGql };
}

// ----------------------------
// 字段矩阵对比（OrganizationUnit ↔ Organization）
// ----------------------------
function loadOpenAPI() {
  const raw = fs.readFileSync(OPENAPI_PATH, 'utf8');
  return yaml.load(raw);
}

function getOpenAPISchema(openapi, name) {
  const schemas = (openapi.components && openapi.components.schemas) || {};
  return schemas[name] || {};
}

function normalizeRestField(name, def, requiredSet) {
  const out = { name, list: false, base: null, nullable: false, required: false, description: def && def.description || null };
  let type = def && def.type;
  if (def && def.$ref) {
    const ref = def.$ref || '';
    const enumName = ref.split('/').pop();
    out.base = `enum:${enumName}`;
  } else if (type === 'array') {
    out.list = true;
    const items = def.items || {};
    if (items.$ref) {
      out.base = `enum:${String(items.$ref).split('/').pop()}`;
    } else {
      out.base = (items.type || 'object');
    }
  } else {
    out.base = type || 'object';
  }
  out.nullable = def && def.nullable === true;
  out.required = requiredSet.has(name) && !out.nullable;
  return out;
}

function extractRestFields(openapi, schemaName) {
  const schema = getOpenAPISchema(openapi, schemaName);
  const props = (schema && schema.properties) || {};
  const required = new Set((schema && schema.required) || []);
  const fields = {};
  for (const [name, def] of Object.entries(props)) {
    fields[name] = normalizeRestField(name, def, required);
  }
  return fields;
}

function extractGraphQLTypeBlock(schemaText, typeName) {
  const pattern = new RegExp(`type\\s+${typeName}\\s*\\{([\\s\\S]*?)\\}`, 'm');
  const match = pattern.exec(schemaText);
  return match ? match[1] : '';
}

function parseGraphQLFieldLine(line) {
  // remove comments after '#'
  const code = line.split('#')[0].trim();
  if (!code) return null;
  // match: name: Type
  const m = /^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*(.+)$/.exec(code);
  if (!m) return null;
  const name = m[1];
  const typeExpr = m[2].trim();
  // detect list and non-null
  let list = false;
  let nonNull = false;
  let base = typeExpr;
  // strip exclamation at end
  if (base.endsWith('!')) {
    nonNull = true;
    base = base.slice(0, -1).trim();
  }
  // list form: [Type] or [Type!]!
  const listMatch = /^\[(.+)\]!?$/.exec(base);
  if (listMatch) {
    list = true;
    base = listMatch[1].trim();
    if (base.endsWith('!')) {
      base = base.slice(0, -1).trim();
    }
  }
  return { name, base, list, nonNull };
}

function extractGraphQLFields(schemaText, typeName) {
  const block = extractGraphQLTypeBlock(schemaText, typeName);
  const lines = block.split('\n');
  const fields = {};
  for (const line of lines) {
    const f = parseGraphQLFieldLine(line);
    if (!f) continue;
    fields[f.name] = f;
  }
  return fields;
}

function mapGraphQLBaseToRest(base) {
  // normalize GraphQL → REST primitive names
  switch (base) {
    case 'String': return 'string';
    case 'Int': return 'integer';
    case 'Float': return 'number';
    case 'Boolean': return 'boolean';
    case 'ID': return 'string';
    default:
      // enums and object types keep name; enums will compare as enum:Name
      return `enum:${base}`;
  }
}

function computeFieldMatrixDiff(restFields = {}, gqlFields = {}, allow = {}) {
  const restNames = new Set(Object.keys(restFields));
  const gqlNames = new Set(Object.keys(gqlFields));
  const missingInRestRaw = [...gqlNames].filter(n => !restNames.has(n));
  const missingInGqlRaw = [...restNames].filter(n => !gqlNames.has(n));

  // apply allowlist for presence
  const allowMissingInRest = new Set(((allow && allow.missingInRest) || []).map(String));
  const allowMissingInGql = new Set(((allow && allow.missingInGql) || []).map(String));
  const missingInRest = missingInRestRaw.filter(n => !allowMissingInRest.has(n));
  const missingInGql = missingInGqlRaw.filter(n => !allowMissingInGql.has(n));

  // type/nullability/list mismatches for intersection
  const common = [...restNames].filter(n => gqlNames.has(n));
  const typeMismatches = [];
  const nullabilityMismatches = [];
  const listMismatches = [];

  const allowType = new Set(((allow && allow.typeMismatch) || []).map(x => typeof x === 'string' ? x : (x && x.field)).filter(Boolean));
  const allowNull = new Set(((allow && allow.nullabilityMismatch) || []).map(x => typeof x === 'string' ? x : (x && x.field)).filter(Boolean));
  const allowList = new Set(((allow && allow.listMismatch) || []).map(x => typeof x === 'string' ? x : (x && x.field)).filter(Boolean));

  for (const name of common) {
    const r = restFields[name];
    const g = gqlFields[name];
    const gqlBaseAsRest = mapGraphQLBaseToRest(g.base);

    if (!allowList.has(name)) {
      if (Boolean(r.list) !== Boolean(g.list)) {
        listMismatches.push({ field: name, rest: r.list ? 'list' : 'single', gql: g.list ? 'list' : 'single' });
      }
    }

    if (!allowType.has(name)) {
      // For REST enums we got "enum:Name"; for GraphQL non-enum object types we also map to enum:Name by default.
      const restType = r.base;
      const gqlType = gqlBaseAsRest;
      if (restType !== gqlType) {
        typeMismatches.push({ field: name, rest: restType, gql: gqlType });
      }
    }

    if (!allowNull.has(name)) {
      const restNonNull = r.required === true;
      const gqlNonNull = g.nonNull === true;
      if (restNonNull !== gqlNonNull) {
        nullabilityMismatches.push({ field: name, rest: restNonNull ? 'nonNull' : 'nullable', gql: gqlNonNull ? 'nonNull' : 'nullable' });
      }
    }
  }

  return { missingInRest, missingInGql, typeMismatches, nullabilityMismatches, listMismatches };
}

function main() {
  const failOnDiff = process.argv.includes('--fail-on-diff');
  const includeFields = process.argv.includes('--include-fields') || process.env.DRIFT_INCLUDE_FIELDS === '1';
  const failOnFields = process.argv.includes('--fail-on-fields') || process.env.DRIFT_FAIL_ON_FIELDS === '1';
  const allow = readAllowlist();
  if (!fs.existsSync(CONTRACT_PATH)) {
    console.error('✗ 未找到契约中间层文件，请先运行 scripts/contract/sync.sh');
    process.exit(2);
  }

  const contract = readJSON(CONTRACT_PATH);
  const restEnums = contract.enums || {};
  const gqlEnums = (contract.graphql && contract.graphql.enums) || {};

  const results = {};
  let hasDiff = false;
  for (const key of ENUM_KEYS) {
    const diff = computeEnumDiff(restEnums[key] || [], gqlEnums[key] || []);
    // 应用白名单过滤（仅对枚举）
    const allowEnum = allow.enum && allow.enum[key] ? allow.enum[key] : {};
    const allowMissingInRest = new Set((allowEnum.missingInRest || []).map(String));
    const allowMissingInGql = new Set((allowEnum.missingInGql || []).map(String));
    const filtered = {
      missingInRest: (diff.missingInRest || []).filter(v => !allowMissingInRest.has(String(v))),
      missingInGql: (diff.missingInGql || []).filter(v => !allowMissingInGql.has(String(v))),
    };
    results[key] = filtered;
    if ((filtered.missingInRest && filtered.missingInRest.length) ||
        (filtered.missingInGql && filtered.missingInGql.length)) {
      hasDiff = true;
    }
  }

  // 枚举差异
  const report = {
    generatedAt: new Date().toISOString(),
    source: path.relative(PROJECT_ROOT, CONTRACT_PATH),
    results,
  };

  // 字段矩阵（报告或阻断）
  if (includeFields) {
    try {
      const openapi = loadOpenAPI();
      const schemaText = fs.readFileSync(GRAPHQL_SCHEMA_PATH, 'utf8');
      const restFields = extractRestFields(openapi, 'OrganizationUnit');
      const gqlFields = extractGraphQLFields(schemaText, 'Organization');
      const allowFields = (allow.fields && allow.fields['OrganizationUnit|Organization']) || {};
      const matrix = computeFieldMatrixDiff(restFields, gqlFields, allowFields);
      report.fieldMatrix = {
        entityPair: 'OrganizationUnit|Organization',
        ...matrix,
      };
      if (failOnFields) {
        const hasFieldDiff =
          (matrix.missingInRest && matrix.missingInRest.length > 0) ||
          (matrix.missingInGql && matrix.missingInGql.length > 0) ||
          (matrix.typeMismatches && matrix.typeMismatches.length > 0) ||
          (matrix.nullabilityMismatches && matrix.nullabilityMismatches.length > 0) ||
          (matrix.listMismatches && matrix.listMismatches.length > 0);
        if (hasFieldDiff) {
          hasDiff = true;
        }
      }
    } catch (e) {
      console.warn('⚠ 字段矩阵分析失败（忽略）：', e.message);
    }
  }

  fs.mkdirSync(REPORT_DIR, { recursive: true });
  fs.writeFileSync(REPORT_PATH, JSON.stringify(report, null, 2));

  if (hasDiff) {
    console.warn('⚠ 契约漂移检测发现差异，详情见:', path.relative(PROJECT_ROOT, REPORT_PATH));
    console.warn(JSON.stringify(results, null, 2));
    if (failOnDiff) {
      process.exit(3);
    }
  } else {
    console.log('✅ 契约漂移检测通过，OpenAPI 与 GraphQL 枚举一致');
  }
}

try {
  main();
} catch (err) {
  console.error('✗ 漂移检测失败:', err.message);
  process.exit(1);
}
