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
  const schemaPath = path.join(repoRoot, 'docs/api/schema.graphql');
  const text = readSafe(schemaPath);
  if (!text) return [];

  const typeIndex = text.indexOf('type Query');
  if (typeIndex === -1) return [];
  const braceStart = text.indexOf('{', typeIndex);
  if (braceStart === -1) return [];

  let cursor = braceStart + 1;
  let depth = 1;
  const end = text.length;
  let body = '';

  while (cursor < end && depth > 0) {
    // Block strings (""" ... """), ignore internal braces
    if (text.startsWith('"""', cursor)) {
      cursor += 3;
      while (cursor < end && !text.startsWith('"""', cursor)) cursor += 1;
      cursor = Math.min(end, cursor + 3);
      continue;
    }

    // Regular quoted strings "...", skip escaped quotes
    if (text[cursor] === '"') {
      cursor += 1;
      while (cursor < end) {
        if (text[cursor] === '\\') { cursor += 2; continue; }
        if (text[cursor] === '"') { cursor += 1; break; }
        cursor += 1;
      }
      continue;
    }

    const ch = text[cursor];
    if (ch === '{') depth += 1;
    else if (ch === '}') {
      depth -= 1;
      if (depth === 0) break;
    }

    if (depth > 0) body += ch;
    cursor += 1;
  }

  const queries = [];
  let i = 0;
  while (i < body.length) {
    while (i < body.length && /\s/.test(body[i])) i += 1;
    if (i >= body.length) break;

    if (body.startsWith('#', i)) {
      while (i < body.length && body[i] !== '\n') i += 1;
      continue;
    }

    if (body.startsWith('"""', i)) {
      i += 3;
      while (i < body.length && !body.startsWith('"""', i)) i += 1;
      if (i < body.length) i += 3;
      continue;
    }

    const identifierMatch = body.slice(i).match(/^([A-Za-z_][A-Za-z0-9_]*)/);
    if (!identifierMatch) {
      i += 1;
      continue;
    }

    const name = identifierMatch[1];
    i += name.length;

    while (i < body.length && /\s/.test(body[i])) i += 1;

    if (body[i] === '(') {
      let parenDepth = 0;
      while (i < body.length) {
        const ch = body[i];
        if (ch === '(') parenDepth += 1;
        else if (ch === ')') {
          parenDepth -= 1;
          if (parenDepth === 0) { i += 1; break; }
        }
        i += 1;
      }
      while (i < body.length && /\s/.test(body[i])) i += 1;
    }

    if (body[i] === ':') {
      queries.push(name);
      while (i < body.length && body[i] !== '\n') i += 1;
    }
  }

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
