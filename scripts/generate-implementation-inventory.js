#!/usr/bin/env node
/**
 * Generate draft entries for IMPLEMENTATION-INVENTORY.md
 * Scans OpenAPI, GraphQL schema, Go handlers/services, and TS exports.
 * Output: Markdown snippet to stdout (non-destructive).
 */

const fs = require('fs');
const path = require('path');

const repoRoot = path.resolve(__dirname, '..');

function readSafe(file) {
  try { return fs.readFileSync(file, 'utf8'); } catch { return ''; }
}

function extractOpenApiPaths() {
  const p = path.join(repoRoot, 'docs/api/openapi.yaml');
  const text = readSafe(p);
  if (!text) return [];
  const lines = text.split(/\r?\n/);
  const paths = [];
  let inPaths = false;
  for (const line of lines) {
    if (/^paths:\s*$/.test(line)) { inPaths = true; continue; }
    if (inPaths) {
      // path entries look like: "  /api/v1/organization-units:"
      const m = line.match(/^\s{2,}(\/[^\s:]+):\s*$/);
      if (m) {
        paths.push(m[1]);
      } else if (/^\S/.test(line)) {
        // left section
        break;
      }
    }
  }
  return paths;
}

function extractGraphQLQueries() {
  const p = path.join(repoRoot, 'docs/api/schema.graphql');
  const text = readSafe(p);
  if (!text) return [];
  const queries = [];
  const queryBlock = /type\s+Query\s*\{([\s\S]*?)\}/m.exec(text);
  if (!queryBlock) return queries;
  let body = queryBlock[1];
  // remove triple-quoted description blocks to avoid capturing words inside docs
  body = body.replace(/"""[\s\S]*?"""/g, '\n');
  body.split(/\n/).forEach((line) => {
    const clean = line.trim();
    if (!clean || clean.startsWith('#')) return;
    // e.g. organizations( ... ): OrganizationConnection!
    const m = clean.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*\(/) || clean.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*:/);
    if (m) queries.push(m[1]);
  });
  return queries;
}

function rgExportedFunctions(root, exts, patterns) {
  const results = [];
  function walk(dir) {
    const ents = fs.existsSync(dir) ? fs.readdirSync(dir, { withFileTypes: true }) : [];
    for (const ent of ents) {
      const p = path.join(dir, ent.name);
      if (ent.isDirectory()) walk(p);
      else if (exts.some((e) => ent.name.endsWith(e))) {
        const text = readSafe(p);
        if (!text) continue;
        for (const { label, regex } of patterns) {
          const re = new RegExp(regex, 'g');
          let m;
          while ((m = re.exec(text))) {
            results.push({ file: p, name: m[1], kind: label });
          }
        }
      }
    }
  }
  walk(root);
  return results;
}

function dedupeBy(arr, key) {
  const seen = new Set();
  return arr.filter((x) => { const k = key(x); if (seen.has(k)) return false; seen.add(k); return true; });
}

function main() {
  const openapiPaths = extractOpenApiPaths();
  const gqlQueries = extractGraphQLQueries();

  const goHandlers = rgExportedFunctions(
    path.join(repoRoot, 'cmd/organization-command-service/internal/handlers'),
    ['.go'],
    [
      { label: 'method', regex: String.raw`func\s*\([^)]*\)\s+([A-Z][A-Za-z0-9_]*)\s*\(` },
    ]
  );

  const goServices = rgExportedFunctions(
    path.join(repoRoot, 'cmd/organization-command-service/internal/services'),
    ['.go'],
    [
      { label: 'type', regex: String.raw`type\s+([A-Z][A-Za-z0-9_]*)\s+struct\b` },
    ]
  );

  const tsExports = rgExportedFunctions(
    path.join(repoRoot, 'frontend/src'),
    ['.ts', '.tsx'],
    [
      { label: 'class', regex: String.raw`export\s+class\s+([A-Z][A-Za-z0-9_]*)\b` },
      { label: 'func', regex: String.raw`export\s+function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(` },
      { label: 'const', regex: String.raw`export\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=` },
    ]
  );

  // Prepare JSON report (deduped and relative paths)
  const rel = (p) => path.relative(repoRoot, p);
  const jsonReport = {
    timestamp: new Date().toISOString(),
    summary: {
      openapiPaths: openapiPaths.length,
      graphqlQueries: gqlQueries.length,
      goHandlers: 0,
      goServices: 0,
      tsExports: 0
    },
    openapiPaths,
    graphqlQueries: gqlQueries,
    goHandlers: [],
    goServices: [],
    tsExports: []
  };

  const out = [];
  out.push('## Draft – Command API (from OpenAPI)');
  if (openapiPaths.length) {
    openapiPaths.forEach((p) => out.push(`- \`${p}\``));
  } else {
    out.push('- (no paths found)');
  }

  out.push('\n## Draft – GraphQL Queries (from schema.graphql)');
  if (gqlQueries.length) {
    gqlQueries.forEach((q) => out.push(`- \`${q}\``));
  } else {
    out.push('- (no queries found)');
  }

  out.push('\n## Draft – Go Handlers (exported methods)');
  const dedupGoHandlers = dedupeBy(goHandlers, (x) => `${x.name}@${x.file}`);
  if (dedupGoHandlers.length) {
    dedupGoHandlers.forEach((h) => out.push(`- ${h.name} — ${rel(h.file)}`));
  } else {
    out.push('- (no handlers found)');
  }
  jsonReport.goHandlers = dedupGoHandlers.map((h) => ({ name: h.name, file: rel(h.file), kind: h.kind }));
  jsonReport.summary.goHandlers = jsonReport.goHandlers.length;

  out.push('\n## Draft – Go Services (exported types)');
  const dedupGoServices = dedupeBy(goServices, (x) => `${x.name}@${x.file}`);
  if (dedupGoServices.length) {
    dedupGoServices.forEach((s) => out.push(`- ${s.name} — ${rel(s.file)}`));
  } else {
    out.push('- (no services found)');
  }
  jsonReport.goServices = dedupGoServices.map((s) => ({ name: s.name, file: rel(s.file), kind: s.kind }));
  jsonReport.summary.goServices = jsonReport.goServices.length;

  out.push('\n## Draft – Frontend Exports (classes/functions/const)');
  const dedupTs = dedupeBy(tsExports, (x) => `${x.name}@${x.file}`);
  if (dedupTs.length) {
    dedupTs.forEach((e) => out.push(`- [${e.kind}] ${e.name} — ${rel(e.file)}`));
  } else {
    out.push('- (no TS exports found)');
  }
  jsonReport.tsExports = dedupTs.map((e) => ({ name: e.name, file: rel(e.file), kind: e.kind }));
  jsonReport.summary.tsExports = jsonReport.tsExports.length;

  // write JSON report
  try {
    const reportsDir = path.join(repoRoot, 'reports');
    if (!fs.existsSync(reportsDir)) fs.mkdirSync(reportsDir, { recursive: true });
    fs.writeFileSync(path.join(reportsDir, 'implementation-inventory.json'), JSON.stringify(jsonReport, null, 2));
  } catch (err) {
    // non-fatal; continue output markdown
  }

  console.log(out.join('\n'));
}

main();
