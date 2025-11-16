#!/usr/bin/env node
/**
 * Plan 259 – FE Migration Scan (REST business GET → GraphQL)
 * - 扫描前端代码中对 REST 业务查询端点（/api/v1/positions/{code}/assignments）的使用
 * - 输出定位与替换建议（建议迁移到 GraphQL：positionAssignments/assignments）
 * - 仅扫描与输出报告，不修改代码
 *
 * 无第三方依赖
 */
const fs = require('fs');
const path = require('path');

const ROOT = process.cwd();
const TARGET_DIRS = ['frontend/src', 'frontend/tests'];
const OUT_PATH = path.join(ROOT, 'reports', 'plan259', 'fe-migration-suggestions.json');

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function listFiles(dir) {
  let results = [];
  if (!fs.existsSync(dir)) return results;
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const e of entries) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      results = results.concat(listFiles(p));
    } else if (e.isFile()) {
      if (/\.(ts|tsx|js|jsx|mjs|cjs|json|md)$/.test(e.name)) {
        results.push(p);
      }
    }
  }
  return results;
}

function scanFile(filePath) {
  const rel = path.relative(ROOT, filePath);
  const text = fs.readFileSync(filePath, 'utf8');
  const lines = text.split(/\r?\n/);
  const findings = [];

  const isE2E = rel.startsWith('frontend/tests/e2e/');

  const patterns = [
    {
      type: 'rest-business-get',
      re: /\/api\/v1\/positions\/[^"'`]+\/assignments/i,
      note: 'Direct REST business GET to assignments endpoint',
    },
    {
      type: 'rest-client-assignments',
      re: /UnifiedRESTClient.*assignments|assignments.*UnifiedRESTClient/i,
      note: 'UnifiedRESTClient used for assignments (should be GraphQL for reads)',
    },
    {
      type: 'raw-fetch-rest',
      re: /fetch\([^)]*\/api\/v1\/positions\/[^"'`]+\/assignments/i,
      note: 'fetch to REST assignments endpoint',
    },
  ];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    for (const p of patterns) {
      if (p.re.test(line)) {
        const suggestion = isE2E
          ? {
              action: 'Use GraphQL for read assertions',
              example: `fetch(\`\${base}/graphql\`, { method: 'POST', headers: { 'content-type': 'application/json', authorization: 'Bearer ...' }, body: JSON.stringify({ query: 'query($positionCode: PositionCode!, $page:Int!, $pageSize:Int!){ positionAssignments(positionCode:$positionCode,pagination:{page:$page,pageSize:$pageSize}){ data{ assignmentId status effectiveDate endDate } pagination{ total page pageSize } } }', variables: { positionCode: 'P1000001', page:1, pageSize:25 } }) })`,
              mapping: {
                assignmentTypes: 'filter.types',
                status: 'filter.status',
                asOfDate: 'filter.asOfDate',
                includeHistorical: 'filter.includeHistorical',
                includeActingOnly: 'filter.actingOnly',
              },
            }
          : {
              action: 'Replace with UnifiedGraphQLClient + domain facade',
              example: `unifiedGraphQLClient.request(query, { positionCode, page, pageSize })`,
              facadeHint: 'Add a domain API function (e.g., listPositionAssignments) under frontend/src/features/positions/api/ or shared facade, returning typed data for components.',
              mapping: {
                assignmentTypes: 'filter.types',
                status: 'filter.status',
                asOfDate: 'filter.asOfDate',
                includeHistorical: 'filter.includeHistorical',
                includeActingOnly: 'filter.actingOnly',
              },
            };

        findings.push({
          file: rel,
          line: i + 1,
          type: p.type,
          note: p.note,
          snippet: line.trim().slice(0, 300),
          suggestion,
          severity: isE2E ? 'info' : 'high',
          category: isE2E ? 'test-read' : 'app-read',
        });
      }
    }
  }

  return findings;
}

function main() {
  const allFiles = TARGET_DIRS.flatMap((d) => listFiles(path.join(ROOT, d)));
  const allFindings = [];
  for (const f of allFiles) {
    const fnd = scanFile(f);
    if (fnd.length) allFindings.push(...fnd);
  }
  const summary = {
    totalFilesScanned: allFiles.length,
    totalFindings: allFindings.length,
    byCategory: allFindings.reduce((acc, cur) => {
      acc[cur.category] = (acc[cur.category] || 0) + 1;
      return acc;
    }, {}),
  };
  const payload = {
    generatedAt: new Date().toISOString(),
    summary,
    findings: allFindings,
    whitelist: ['/.well-known/jwks.json', '/api/v1/operational/**', '/auth/**'],
    deprecation: {
      endpoint: 'GET /api/v1/positions/{code}/assignments',
      sunset: '2025-12-20T00:00:00Z',
      migration: 'Use GraphQL: positionAssignments/assignments',
    },
  };
  ensureDir(path.dirname(OUT_PATH));
  fs.writeFileSync(OUT_PATH, JSON.stringify(payload, null, 2), 'utf8');
  console.log(`Plan 259 – FE migration scan completed.\nFiles scanned: ${summary.totalFilesScanned}\nFindings: ${summary.totalFindings}\nOutput: ${path.relative(ROOT, OUT_PATH)}`);
  if (summary.totalFindings === 0) {
    console.log('No REST business GET usage detected in frontend codebase (app/tests).');
  }
}

main();

