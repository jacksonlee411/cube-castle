#!/usr/bin/env node
/**
 * Plan 252 – 权限契约校验器
 * - 校验 OpenAPI security scopes：路径引用→注册表一致性
 * - 生成 GraphQL Query→scope 映射（SSoT：docs/api/schema.graphql 注释）
 * - 校验 resolver 授权调用是否均有映射
 *
 * 依赖：Node.js（无三方库）
 */
const fs = require('fs');
const path = require('path');

function parseArgs(argv) {
  const args = {};
  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    if (a.startsWith('--')) {
      const [k, v] = a.split('=');
      const key = k.replace(/^--/, '');
      if (v !== undefined) args[key] = v;
      else args[key] = argv[i + 1], i++;
    }
  }
  return args;
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function readText(p) {
  return fs.readFileSync(p, 'utf8');
}

function writeJSON(p, obj) {
  ensureDir(path.dirname(p));
  fs.writeFileSync(p, JSON.stringify(obj, null, 2), 'utf8');
}

function writeText(p, txt) {
  ensureDir(path.dirname(p));
  fs.writeFileSync(p, txt, 'utf8');
}

// --- OpenAPI 解析 ---
function parseOpenAPIScopes(openapiText) {
  const lines = openapiText.split(/\r?\n/);
  // 1) 注册表：components.securitySchemes.OAuth2ClientCredentials.flows.clientCredentials.scopes
  const registry = new Set();
  let inScopes = false;
  let scopesIndent = 0;
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (!inScopes && line.match(/scopes:\s*$/)) {
      inScopes = true;
      scopesIndent = line.search(/\S/);
      continue;
    }
    if (inScopes) {
      if (line.trim() === '') continue; // 允许空行
      const indent = line.search(/\S/);
      if (indent <= scopesIndent) {
        inScopes = false;
        continue;
      }
      // 形如：'job-catalog:read': Read job catalog classifications
      const m = line.match(/['"]([^'"]+)['"]\s*:/);
      if (m) registry.add(m[1]);
    }
  }
  // 2) 使用：遍历所有 security: - OAuth2ClientCredentials: [...]
  const uses = new Set();
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const mInline = line.match(/-+\s*OAuth2ClientCredentials\s*:\s*\[([^\]]*)\]/);
    if (mInline) {
      const scopesStr = mInline[1];
      const re = /['"]([^'"]+)['"]/g;
      let ms;
      while ((ms = re.exec(scopesStr))) uses.add(ms[1]);
      continue;
    }
    const mStart = line.match(/-+\s*OAuth2ClientCredentials\s*:\s*$/);
    if (mStart) {
      const baseIndent = line.search(/\S/);
      let j = i + 1;
      for (; j < lines.length; j++) {
        const l2 = lines[j];
        const indent = l2.search(/\S/);
        if (indent <= baseIndent) break;
        const mItem = l2.match(/-\s*['"]?([^'"]+)['"]?/);
        if (mItem) {
          const val = mItem[1].trim();
          if (val) uses.add(val);
        }
      }
      i = j - 1;
      continue;
    }
  }
  return { registry, uses };
}

// --- GraphQL 解析 ---
function parseGraphQLQueryPermissions(schemaText) {
  const lines = schemaText.split(/\r?\n/);
  let inQuery = false;
  let braceDepth = 0;
  let lastPermissions = null;
  const mapping = {};
  const allFields = new Set();
  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const line = raw.trim();
    // 进入 Query
    if (!inQuery && line.startsWith('type Query')) {
      inQuery = true;
      // next lines until closing }
      continue;
    }
    if (inQuery) {
      if (line.startsWith('{')) {
        braceDepth++;
        continue;
      }
      if (line.startsWith('}')) {
        braceDepth--;
        if (braceDepth <= 0) {
          inQuery = false;
          braceDepth = 0;
        }
        continue;
      }
      // 捕获注释中的权限说明
      if (line.startsWith('"""')) {
        // consume docstring
        let j = i + 1;
        let perms = null;
        for (; j < lines.length; j++) {
          const dl = lines[j].trim();
          if (dl.startsWith('"""')) { // end
            break;
          }
          const m1 = dl.match(/Permissions Required:\s*([A-Za-z0-9:_-]+)/i);
          const m2 = dl.match(/Requires scope:\s*([A-Za-z0-9:_-]+)/i);
          if (m1) perms = m1[1];
          else if (m2) perms = m2[1];
        }
        lastPermissions = perms;
        i = j; // move to end of docstring
        continue;
      }
      // 捕获字段定义
      if (/^[A-Za-z_][A-Za-z0-9_]*\s*\(/.test(line) || /^[A-Za-z_][A-Za-z0-9_]*\s*:/.test(line)) {
        const name = line.split(/[\s(:]/)[0];
        if (name) {
          allFields.add(name);
        }
        if (name && lastPermissions) {
          mapping[name] = lastPermissions;
        }
        lastPermissions = null;
      }
    }
  }
  return { mapping, allFields };
}

// --- Resolver 授权扫描 ---
function parseResolverAuthorizations(resolverText) {
  const re = /authorize\s*\(\s*ctx\s*,\s*"(.*?)"\s*/g;
  const names = new Set();
  let m;
  while ((m = re.exec(resolverText))) {
    names.add(m[1]);
  }
  return names;
}

function main() {
  const args = parseArgs(process.argv);
  const openapiPath = args.openapi || 'docs/api/openapi.yaml';
  const graphqlPath = args.graphql || 'docs/api/schema.graphql';
  const resolverDirs = (args['resolver-dirs'] || 'internal/organization/resolver,cmd/hrms-server/query/internal/auth')
    .split(',').map(s => s.trim()).filter(Boolean);
  const outDir = args.out || 'reports/permissions';
  const failOn = new Set((args['fail-on'] || 'missing-scope,unregistered-scope,mapping-missing,resolver-bypass')
    .split(',').map(s => s.trim()).filter(Boolean));

  const summary = [];
  let fail = false;

  // OpenAPI
  const openapiText = readText(openapiPath);
  const { registry, uses } = parseOpenAPIScopes(openapiText);
  const unregistered = [...uses].filter(s => !registry.has(s)).sort();
  const unused = [...registry].filter(s => !uses.has(s)).sort();
  writeJSON(path.join(outDir, 'openapi-scope-usage.json'), {
    used: [...uses].sort()
  });
  writeJSON(path.join(outDir, 'openapi-scope-registry.json'), {
    registry: [...registry].sort(),
    unused
  });
  if (unregistered.length > 0 && failOn.has('unregistered-scope')) {
    fail = true;
    summary.push(`✖ OpenAPI 未注册即使用 scopes: ${unregistered.join(', ')}`);
  } else {
    summary.push(`✔ OpenAPI 引用→注册一致性检查（未注册引用=${unregistered.length}）`);
  }
  if (unused.length > 0) {
    summary.push(`ℹ️ OpenAPI 已注册但未引用 scopes: ${unused.join(', ')}`);
  }

  // GraphQL
  const gqlText = readText(graphqlPath);
  const gqlParsed = parseGraphQLQueryPermissions(gqlText);
  const gqlMapping = gqlParsed.mapping;
  const gqlFields = gqlParsed.allFields;
  writeJSON(path.join(outDir, 'graphql-query-permissions.json'), gqlMapping);
  // 将映射写入运行时消费路径（供 go:embed 使用）
  const runtimeMappingPath = path.join('cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json');
  writeJSON(runtimeMappingPath, gqlMapping);
  // 映射覆盖率（近似）：存在权限注释的字段数量
  if (Object.keys(gqlMapping).length === 0 && failOn.has('mapping-missing')) {
    fail = true;
    summary.push('✖ GraphQL 权限映射缺失（未解析到任何 "Permissions Required"/"Requires scope" 注释）');
  } else {
    summary.push(`✔ GraphQL 权限映射生成（entries=${Object.keys(gqlMapping).length}）`);
  }

  // Resolver 授权覆盖
  const resolverAuthCalls = new Set();
  for (const dir of resolverDirs) {
    if (!fs.existsSync(dir)) continue;
    const files = fs.readdirSync(dir)
      .filter(f => f.endsWith('.go'))
      .map(f => path.join(dir, f));
    for (const f of files) {
      const txt = readText(f);
      for (const name of parseResolverAuthorizations(txt)) resolverAuthCalls.add(name);
    }
  }
  const missingInMapping = [...resolverAuthCalls].filter(n => !(n in gqlMapping)).sort();
  const resolverOnSchemaMissingPerms = missingInMapping.filter(n => gqlFields.has(n));
  writeJSON(path.join(outDir, 'resolver-permission-calls.json'), {
    authorizeCalls: [...resolverAuthCalls].sort(),
    missingInMapping,
    missingOnSchemaWithNoPerms: resolverOnSchemaMissingPerms
  });
  if (resolverOnSchemaMissingPerms.length > 0 && failOn.has('resolver-bypass')) {
    fail = true;
    summary.push(`✖ Resolver 授权调用缺少映射（或 Query 注释缺失）：${resolverOnSchemaMissingPerms.join(', ')}`);
  } else {
    summary.push(`✔ Resolver 授权调用覆盖（未匹配映射=${missingInMapping.length}；在 Schema 中但缺权限注释=${resolverOnSchemaMissingPerms.length})`);
  }

  // 汇总
  const summaryText = summary.join('\n') + '\n';
  writeText(path.join(outDir, 'summary.txt'), summaryText);
  process.stdout.write(summaryText);
  if (fail) process.exit(1);
}

main();
