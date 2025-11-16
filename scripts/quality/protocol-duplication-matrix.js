#!/usr/bin/env node
/**
 * Plan 259A – 协议重复矩阵生成与白名单固化
 * - 解析 OpenAPI 获取所有 GET 路径
 * - 依据白名单过滤得到“业务 GET（REST）”清单与计数（目标=0）
 * - 解析 GraphQL schema 获取 Query 列表
 * - 输出 JSON 矩阵与文本摘要；可配置失败阈值（软→硬门禁）
 *
 * 无第三方依赖；基于行级解析的保守实现。
 */
const fs = require('fs');
const path = require('path');

function parseArgs(argv) {
  const args = {};
  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    if (!a.startsWith('--')) continue;
    const [k, vRaw] = a.split('=');
    const key = k.replace(/^--/, '');
    let val = vRaw;
    if (val === undefined) {
      val = argv[i + 1];
      i++;
    }
    args[key] = val;
  }
  return args;
}

function readText(p) {
  return fs.readFileSync(p, 'utf8');
}

function ensureDir(d) {
  fs.mkdirSync(d, { recursive: true });
}

function nowTS() {
  const d = new Date();
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getUTCFullYear()}${pad(d.getUTCMonth()+1)}${pad(d.getUTCDate())}_${pad(d.getUTCHours())}${pad(d.getUTCMinutes())}${pad(d.getUTCSeconds())}`;
}

function toJSON(p, obj) {
  ensureDir(path.dirname(p));
  fs.writeFileSync(p, JSON.stringify(obj, null, 2), 'utf8');
}

function toText(p, txt) {
  ensureDir(path.dirname(p));
  fs.writeFileSync(p, txt, 'utf8');
}

function parseOpenApiGetPaths(openapiText) {
  const lines = openapiText.split(/\r?\n/);
  const all = [];
  let currentPath = null;
  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const mPath = raw.match(/^\s*(\/[A-Za-z0-9._\-{}\/]+)\s*:\s*$/);
    if (mPath) {
      currentPath = mPath[1];
      continue;
    }
    // method lines
    if (currentPath && /^\s*get\s*:\s*$/i.test(raw)) {
      all.push(currentPath);
      continue;
    }
    // reset when next top-level key appears (optional; conservative)
    if (/^\s*[A-Za-z0-9._\-]+\s*:\s*$/.test(raw) && !/^\s*(get|post|put|patch|delete)\s*:\s*$/i.test(raw)) {
      // do nothing; keeping currentPath until next path line
    }
  }
  // unique
  return Array.from(new Set(all));
}

function parseGraphQLQueries(schemaText) {
  const lines = schemaText.split(/\r?\n/);
  let inQuery = false;
  let depth = 0;
  let parenDepth = 0;
  const names = new Set();
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!inQuery) {
      if (/^type\s+Query\b/.test(line)) {
        inQuery = true;
      }
      continue;
    }
    // skip docstring delimiters
    if (line.startsWith('"""')) {
      // consume until next """
      while (i + 1 < lines.length) {
        i++;
        const dl = lines[i].trim();
        if (dl.startsWith('"""')) break;
      }
      continue;
    }
    // track braces
    if (line.includes('{')) depth++;
    if (line.includes('}')) {
      depth--;
      if (depth <= 0) break;
      continue;
    }
    // capture field name only when not inside arg list (parenDepth==0)
    if (parenDepth === 0) {
      const m = line.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*[\(:]/);
      if (m) names.add(m[1]);
    }
    // update paren depth after potential capture
    for (const ch of line) {
      if (ch === '(') parenDepth++;
      else if (ch === ')') parenDepth = Math.max(0, parenDepth - 1);
    }
  }
  return Array.from(names);
}

function compileWhitelist(list) {
  return list.map((pat) => {
    if (pat.endsWith('/**')) {
      const prefix = pat.slice(0, -3);
      return { type: 'prefix', value: prefix };
    }
    return { type: 'exact', value: pat };
  });
}

function isWhitelisted(pathname, compiled) {
  for (const w of compiled) {
    if (w.type === 'exact' && pathname === w.value) return true;
    if (w.type === 'prefix' && pathname.startsWith(w.value)) return true;
  }
  return false;
}

function main() {
  const args = parseArgs(process.argv);
  const openapi = args.openapi || 'docs/api/openapi.yaml';
  const graphql = args.graphql || 'docs/api/schema.graphql';
  const whitelistArg = args.whitelist || '/.well-known/jwks.json,/api/v1/operational/**,/auth/**';
  const out = args.out || 'reports/plan259/protocol-duplication-matrix.json';
  const summaryOutDir = args.summaryDir || 'logs/plan259';
  const envThreshold = process.env.PLAN259_BUSINESS_GET_THRESHOLD;
  const failThreshold = Number(args['fail-threshold'] || (envThreshold || 1));

  const ts = nowTS();
  const openText = readText(openapi);
  const gqlText = readText(graphql);
  const restGetPaths = parseOpenApiGetPaths(openText).sort();
  const wl = whitelistArg.split(',').map(s => s.trim()).filter(Boolean);
  const compiled = compileWhitelist(wl);
  const restBusiness = restGetPaths
    .filter(p => p.startsWith('/api/v1/'))
    .filter(p => !isWhitelisted(p, compiled))
    .sort();
  const gqlQueries = parseGraphQLQueries(gqlText).sort();

  // heuristic duplicates
  const heuristics = {};
  for (const pth of restBusiness) {
    if (/^\/api\/v1\/positions\/\{?code\}?\/assignments\/?$/.test(pth)) {
      heuristics[pth] = ['positionAssignments', 'assignments'];
    }
  }

  const payload = {
    timestamp: ts,
    whitelist: wl,
    failThreshold,
    restGetPathsAll: restGetPaths,
    restBusinessGetPaths: restBusiness,
    restBusinessGetCount: restBusiness.length,
    graphqlQueries: gqlQueries,
    heuristicDuplicateMapping: heuristics,
    failed: restBusiness.length > failThreshold
  };
  toJSON(out, payload);
  const summary = [
    `Plan 259A – Protocol Duplication Matrix @ ${ts}`,
    `REST GET (all): ${restGetPaths.length}`,
    `REST business GET (filtered): ${restBusiness.length} (threshold=${failThreshold})`,
    `Whitelist: ${wl.join(', ')}`,
    `Output JSON: ${out}`
  ].join('\n') + '\n';
  toText(path.join(summaryOutDir, `protocol-duplication-summary-${ts}.txt`), summary);
  process.stdout.write(summary);
  if (payload.failed) process.exit(1);
}

main();
