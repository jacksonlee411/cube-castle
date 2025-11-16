#!/usr/bin/env node
/**
 * drift-matrix-gate.js (Plan 258 - 阻断门禁)
 *
 * 对比 OpenAPI(REST) 与 GraphQL 的字段矩阵（类型/必填/可空/描述），
 * 按映射关系输出差异，并在未被允许（allowlist）覆盖的差异存在时阻断。
 *
 * 用法:
 *   node scripts/contract/drift-matrix-gate.js \
 *     --openapi docs/api/openapi.yaml \
 *     --graphql docs/api/schema.graphql \
 *     --allow scripts/contract/drift-allowlist.json \
 *     --out reports/contracts/drift-matrix-report.json \
 *     --fail-on-diff
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');
const { parse } = require('graphql');

const PROJECT_ROOT = path.resolve(__dirname, '../..');

function readArgs() {
  const argv = process.argv.slice(2);
  function get(name, defVal) {
    const i = argv.indexOf(name);
    return i !== -1 ? argv[i + 1] : defVal;
  }
  const args = {
    openapiPath: get('--openapi', path.join(PROJECT_ROOT, 'docs/api/openapi.yaml')),
    graphqlPath: get('--graphql', path.join(PROJECT_ROOT, 'docs/api/schema.graphql')),
    allowlistPath: get('--allow', path.join(PROJECT_ROOT, 'scripts/contract/drift-allowlist.json')),
    outPath: get('--out', path.join(PROJECT_ROOT, 'reports/contracts/drift-matrix-report.json')),
    failOnDiff: argv.includes('--fail-on-diff'),
  };
  const typesPath = get('--types', null);
  if (typesPath) {
    args.typesPath = path.isAbsolute(typesPath) ? typesPath : path.join(PROJECT_ROOT, typesPath);
  }
  return args;
}

function mapScalarToGraphQL(type) {
  return {
    string: 'String',
    integer: 'Int',
    number: 'Float',
    boolean: 'Boolean',
    object: 'JSON',
  }[String(type).toLowerCase()] || type || 'Unknown';
}

function extractRefName(ref) {
  if (!ref) return null;
  return String(ref).split('/').pop();
}

function readOpenAPI(openapiPath) {
  const text = fs.readFileSync(openapiPath, 'utf8');
  const obj = yaml.load(text);
  const schemas = (obj.components && obj.components.schemas) || {};
  const out = {};
  for (const [name, s] of Object.entries(schemas)) {
    if (!s || s.type !== 'object' || !s.properties) continue;
    const required = new Set(Array.isArray(s.required) ? s.required : []);
    const fields = {};
    for (const [pname, prop] of Object.entries(s.properties)) {
      const info = { isList: false };
      if (prop.$ref) {
        info.type = extractRefName(prop.$ref);
        info.baseType = info.type;
      } else if (prop.type === 'array') {
        info.isList = true;
        let itemType = 'Any';
        if (prop.items && prop.items.$ref) {
          itemType = extractRefName(prop.items.$ref) || 'Any';
        } else if (prop.items && prop.items.type) {
          itemType = mapScalarToGraphQL(prop.items.type);
        }
        info.baseType = itemType;
        info.type = `List<${itemType}>`;
      } else if (prop.type) {
        info.type = mapScalarToGraphQL(prop.type);
        info.baseType = info.type;
      } else {
        info.type = 'Unknown';
        info.baseType = 'Unknown';
      }
      info.required = required.has(pname); // required implies non-null (!)
      info.nullable = prop.nullable === true ? true : false;
      info.description = prop.description || null;
      fields[pname] = info;
    }
    out[name] = { kind: 'object', fields };
  }
  return out;
}

function unwrapType(t) {
  // Returns { baseType, isList, isNonNull }
  let isNonNull = false;
  let isList = false;
  let node = t;
  // NonNull
  if (node.kind === 'NonNullType') {
    isNonNull = true;
    node = node.type;
  }
  // List
  if (node.kind === 'ListType') {
    isList = true;
    // Unwrap inner
    node = node.type;
    if (node.kind === 'NonNullType') {
      // NonNull of inner list element doesn't change top-level required
      node = node.type;
    }
  }
  const base = node.kind === 'NamedType' ? node.name.value : 'Unknown';
  return { baseType: base, isList, isNonNull };
}

function readGraphQL(graphqlPath) {
  const text = fs.readFileSync(graphqlPath, 'utf8');
  const ast = parse(text, { noLocation: false });
  const out = {};
  for (const def of ast.definitions) {
    if (def.kind !== 'ObjectTypeDefinition' && def.kind !== 'InputObjectTypeDefinition') continue;
    const name = def.name.value;
    const fields = {};
    const fieldNodes = def.fields || [];
    for (const f of fieldNodes) {
      const fname = f.name.value;
      const { baseType, isList, isNonNull } = unwrapType(f.type);
      const desc = f.description ? f.description.value : null;
      fields[fname] = {
        type: baseType,
        baseType,
        isList,
        required: isNonNull,
        nullable: !isNonNull,
        description: desc,
      };
    }
    out[name] = { kind: def.kind === 'InputObjectTypeDefinition' ? 'input' : 'object', fields };
  }
  return out;
}

function loadAllowlist(allowlistPath) {
  if (!fs.existsSync(allowlistPath)) return { items: [] };
  try {
    const obj = JSON.parse(fs.readFileSync(allowlistPath, 'utf8'));
    return obj && obj.items && Array.isArray(obj.items) ? obj : { items: [] };
  } catch (e) {
    console.warn('[allowlist] 无法解析，忽略:', e.message);
    return { items: [] };
  }
}

function isAllowed(diff, allow) {
  const today = new Date().toISOString().slice(0, 10);
  return allow.items.some((it) => {
    if (it.typeMap !== diff.typeMap) return false;
    if (it.field !== diff.field) return false;
    if (it.kind !== diff.kind) return false;
    if (!it.expires || it.expires < today) return false;
    if (!it.reason || String(it.reason).trim().length < 10) return false;
    return true;
  });
}

function compareTypePair(oas, gql, oasName, gqlName, allow) {
  const result = {
    typeMap: `${oasName}:${gqlName}`,
    missingInOpenAPI: [],
    missingInGraphQL: [],
    mismatches: [], // array of { field, kind, detail, severity }
  };
  const oFields = (oas[oasName] && oas[oasName].fields) || {};
  const gFields = (gql[gqlName] && gql[gqlName].fields) || {};
  const oKeys = new Set(Object.keys(oFields));
  const gKeys = new Set(Object.keys(gFields));
  for (const k of gKeys) {
    if (!oKeys.has(k)) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'missingInOpenAPI' };
      if (!isAllowed(diff, allow)) result.missingInOpenAPI.push(k);
    }
  }
  for (const k of oKeys) {
    if (!gKeys.has(k)) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'missingInGraphQL' };
      if (!isAllowed(diff, allow)) result.missingInGraphQL.push(k);
    }
  }
  const common = [...oKeys].filter((k) => gKeys.has(k));
  for (const k of common) {
    const o = oFields[k], g = gFields[k];
    // list semantics
    const listMismatch = Boolean(o.isList) !== Boolean(g.isList);
    if (listMismatch) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'list-mismatch', severity: 'error', detail: { openapiList: o.isList, graphqlList: g.isList } };
      if (!isAllowed(diff, allow)) result.mismatches.push(diff);
      continue;
    }
    // base type semantics
    const oBase = (o.baseType || o.type || '').toLowerCase();
    const gBase = (g.baseType || g.type || '').toLowerCase();
    const typeMismatch = oBase !== gBase;
    // nullability mapping：OpenAPI 非空 = required=true 且 nullable=false
    const oasNonNull = Boolean(o.required) && !Boolean(o.nullable);
    const nullMismatch = oasNonNull !== Boolean(g.required);
    if (typeMismatch) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'type-mismatch', severity: 'error', detail: { openapi: o.baseType || o.type, graphql: g.baseType || g.type } };
      if (!isAllowed(diff, allow)) result.mismatches.push(diff);
    }
    if (nullMismatch) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'nullability-mismatch', severity: 'error', detail: { openapiRequired: o.required, openapiNullable: o.nullable, graphqlRequired: g.required } };
      if (!isAllowed(diff, allow)) result.mismatches.push(diff);
    }
    // description difference（信息级）
    const oDesc = (o.description || '').trim();
    const gDesc = (g.description || '').trim();
    if (oDesc && gDesc && oDesc !== gDesc) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'description-mismatch', severity: 'info' };
      if (!isAllowed(diff, allow)) result.mismatches.push(diff);
    }
  }
  return result;
}

function main() {
  const args = readArgs();
  const oas = readOpenAPI(args.openapiPath);
  const gql = readGraphQL(args.graphqlPath);
  const allow = loadAllowlist(args.allowlistPath);

  // 类型映射（按项目约定）
  let TYPE_MAP = [
    { openapi: 'OrganizationUnit', graphql: 'Organization' },
    // 后续如需扩展，追加映射项
  ];
  if (args.typesPath && fs.existsSync(args.typesPath)) {
    try {
      const cfg = JSON.parse(fs.readFileSync(args.typesPath, 'utf8'));
      if (Array.isArray(cfg) && cfg.length > 0) {
        TYPE_MAP = cfg;
      }
    } catch (e) {
      console.warn('[types] 配置解析失败，使用默认映射:', e.message);
    }
  }

  const report = {
    generatedAt: new Date().toISOString(),
    pairs: [],
    summary: { totalPairs: 0, withDiffs: 0, totalIssues: 0 },
  };

  let hasBlocking = false;
  for (const m of TYPE_MAP) {
    const r = compareTypePair(oas, gql, m.openapi, m.graphql, allow);
    report.pairs.push(r);
    const blockingIssues =
      r.missingInOpenAPI.length +
      r.missingInGraphQL.length +
      r.mismatches.filter((x) => x.severity !== 'info').length;
    const totalIssues =
      r.missingInOpenAPI.length + r.missingInGraphQL.length + r.mismatches.length;
    if (totalIssues > 0) {
      report.summary.withDiffs += 1;
      report.summary.totalIssues += totalIssues;
      hasBlocking = hasBlocking || blockingIssues > 0;
    }
  }
  report.summary.totalPairs = TYPE_MAP.length;

  // 输出报告
  const outDir = path.dirname(args.outPath);
  fs.mkdirSync(outDir, { recursive: true });
  fs.writeFileSync(args.outPath, JSON.stringify(report, null, 2));

  if (hasBlocking && args.failOnDiff) {
    console.error('❌ Plan 258 Gate: 发现未允许的契约漂移差异，阻断合并。');
    console.error(`报告: ${path.relative(PROJECT_ROOT, args.outPath)}`);
    process.exit(5);
  } else if (hasBlocking) {
    console.warn('⚠ Plan 258 Gate: 发现差异，但 failOnDiff 未开启。');
  } else {
    console.log('✅ Plan 258 Gate: 未发现差异。');
  }
}

try {
  main();
} catch (e) {
  console.error('✗ Plan 258 Gate 执行失败:', e.message);
  process.exit(1);
}
