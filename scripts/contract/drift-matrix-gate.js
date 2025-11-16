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

const PROJECT_ROOT = path.resolve(__dirname, '../..');

function readArgs() {
  const argv = process.argv.slice(2);
  function get(name, defVal) {
    const i = argv.indexOf(name);
    return i !== -1 ? argv[i + 1] : defVal;
  }
  return {
    openapiPath: get('--openapi', path.join(PROJECT_ROOT, 'docs/api/openapi.yaml')),
    graphqlPath: get('--graphql', path.join(PROJECT_ROOT, 'docs/api/schema.graphql')),
    allowlistPath: get('--allow', path.join(PROJECT_ROOT, 'scripts/contract/drift-allowlist.json')),
    outPath: get('--out', path.join(PROJECT_ROOT, 'reports/contracts/drift-matrix-report.json')),
    failOnDiff: argv.includes('--fail-on-diff'),
  };
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
      const info = {};
      if (prop.$ref) {
        info.type = String(prop.$ref).split('/').pop();
      } else if (prop.type) {
        // map to GraphQL-ish types (best-effort)
        info.type = {
          string: 'String',
          integer: 'Int',
          number: 'Float',
          boolean: 'Boolean',
          object: 'JSON',
          array: 'List',
        }[prop.type] || prop.type;
      } else {
        info.type = 'Unknown';
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

function readGraphQL(graphqlPath) {
  const text = fs.readFileSync(graphqlPath, 'utf8');
  const out = {};
  const typeRe = /^(type|input)\s+([A-Za-z0-9_]+)\s*\{([\s\S]*?)^\}/gm;
  let m;
  while ((m = typeRe.exec(text)) !== null) {
    const kind = m[1];
    const name = m[2];
    const body = m[3];
    const fields = {};
    const lines = body.split('\n');
    for (let raw of lines) {
      const line = raw.split('#')[0].trim();
      if (!line || line.startsWith('"""') || line.startsWith('"')) continue;
      // match: fieldName: Type! / [Type!]!
      const f = /^([A-Za-z0-9_]+)\s*:\s*([!\[\]A-Za-z0-9_]+)(?:\s|$)/.exec(line);
      if (!f) continue;
      const fname = f[1];
      const ftypeRaw = f[2];
      const nonNull = /!$/.test(ftypeRaw);
      let baseType = ftypeRaw.replace(/[!\[\]]/g, '');
      // map to canonical names
      baseType = {
        ID: 'ID',
        String: 'String',
        Int: 'Int',
        Float: 'Float',
        Boolean: 'Boolean',
      }[baseType] || baseType;
      fields[fname] = {
        type: baseType,
        required: nonNull,
        nullable: !nonNull,
        description: null,
      };
    }
    out[name] = { kind, fields };
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
    mismatches: [],
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
    const typeMismatch = (o.type || '').toLowerCase() !== (g.type || '').toLowerCase();
    const nullMismatch = Boolean(o.required) !== Boolean(g.required);
    if (typeMismatch) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'type-mismatch', detail: { openapi: o.type, graphql: g.type } };
      if (!isAllowed(diff, allow)) result.mismatches.push(diff);
    }
    if (nullMismatch) {
      const diff = { typeMap: result.typeMap, field: k, kind: 'nullability-mismatch', detail: { openapiRequired: o.required, graphqlRequired: g.required } };
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
  const TYPE_MAP = [
    { openapi: 'OrganizationUnit', graphql: 'Organization' },
    // 后续如需扩展，追加映射项
  ];

  const report = {
    generatedAt: new Date().toISOString(),
    pairs: [],
    summary: { totalPairs: 0, withDiffs: 0, totalIssues: 0 },
  };

  let hasBlocking = false;
  for (const m of TYPE_MAP) {
    const r = compareTypePair(oas, gql, m.openapi, m.graphql, allow);
    report.pairs.push(r);
    const issues = r.missingInOpenAPI.length + r.missingInGraphQL.length + r.mismatches.length;
    if (issues > 0) {
      report.summary.withDiffs += 1;
      report.summary.totalIssues += issues;
      hasBlocking = true;
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

