#!/usr/bin/env node

/**
 * Generate UI manifest metadata from OpenAPI schemas.
 * Outputs a JSON log summarizing Organization/Position form fields.
 *
 * Usage: node scripts/generate-forms-from-openapi.js
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const OPENAPI_PATH = path.resolve(__dirname, '..', 'docs', 'api', 'openapi.yaml');
const OUTPUT_DIR = path.resolve(__dirname, '..', 'logs', 'plan400', 'manifest');

function timestamp() {
  const now = new Date();
  const pad = (n) => String(n).padStart(2, '0');
  return (
    now.getUTCFullYear().toString() +
    pad(now.getUTCMonth() + 1) +
    pad(now.getUTCDate()) +
    pad(now.getUTCHours()) +
    pad(now.getUTCMinutes()) +
    pad(now.getUTCSeconds())
  );
}

function loadOpenAPI() {
  const file = fs.readFileSync(OPENAPI_PATH, 'utf8');
  return yaml.load(file);
}

function resolveRef(ref, schemas, seen = new Set()) {
  if (!ref) return null;
  const refName = ref.split('/').pop();
  if (!refName || seen.has(refName)) {
    return null;
  }
  seen.add(refName);
  return schemas[refName] || null;
}

function describeType(definition, schemas, seenRefs = new Set()) {
  if (!definition) {
    return 'unknown';
  }
  if (definition.$ref) {
    const refName = definition.$ref.split('/').pop();
    return `ref:${refName || 'unknown'}`;
  }
  if (definition.type === 'array') {
    return `array<${describeType(definition.items, schemas, seenRefs)}>`;
  }
  if (definition.type === 'object' && definition.properties) {
    const keys = Object.keys(definition.properties);
    return `object{${keys.slice(0, 3).join(', ')}${keys.length > 3 ? ', …' : ''}}`;
  }
  if (definition.enum) {
    return `enum(${definition.enum.join('|')})`;
  }
  const base = definition.type || 'unknown';
  return definition.format ? `${base}<${definition.format}>` : base;
}

function collectShape(schema, schemas, seen = new Set()) {
  if (!schema) {
    return { properties: {}, required: [] };
  }
  if (schema.$ref) {
    const resolved = resolveRef(schema.$ref, schemas, seen);
    return collectShape(resolved, schemas, seen);
  }
  if (schema.allOf) {
    return schema.allOf.reduce(
      (acc, fragment) => {
        const next = collectShape(fragment, schemas, seen);
        return {
          properties: { ...acc.properties, ...next.properties },
          required: acc.required.concat(next.required),
        };
      },
      { properties: {}, required: [] },
    );
  }
  return {
    properties: schema.properties || {},
    required: Array.isArray(schema.required) ? schema.required : [],
  };
}

function normalizeDescription(text) {
  if (!text) return '';
  return String(text).replace(/\s+/g, ' ').trim();
}

function buildForms(doc) {
  const schemas = (doc.components && doc.components.schemas) || {};
  const includeKeywords = ['OrganizationUnit', 'Position'];

  return Object.entries(schemas)
    .filter(([name]) => includeKeywords.some((keyword) => name.includes(keyword)) && /Request$/.test(name))
    .map(([name, schemaDef]) => {
      const shape = collectShape(schemaDef, schemas);
      const requiredSet = new Set(shape.required);
      const fields = Object.entries(shape.properties).map(([fieldName, definition]) => ({
        name: fieldName,
        type: describeType(definition, schemas),
        required: requiredSet.has(fieldName),
        description: normalizeDescription(definition.description),
        example:
          definition.example !== undefined
            ? definition.example
            : definition.default !== undefined
              ? definition.default
              : undefined,
      }));
      return {
        schema: name,
        totalFields: fields.length,
        requiredCount: requiredSet.size,
        fields,
      };
    })
    .sort((a, b) => a.schema.localeCompare(b.schema));
}

function main() {
  if (!fs.existsSync(OPENAPI_PATH)) {
    console.error(`[forms] OpenAPI file not found at ${OPENAPI_PATH}`);
    process.exit(1);
  }
  const spec = loadOpenAPI();
  const forms = buildForms(spec);
  if (forms.length === 0) {
    console.warn('[forms] No Organization/Position request schemas found.');
  }
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  const outFile = path.join(OUTPUT_DIR, `${timestamp()}-forms.log`);
  const payload = {
    generatedAt: new Date().toISOString(),
    source: path.relative(process.cwd(), OPENAPI_PATH),
    totalForms: forms.length,
    forms,
  };
  fs.writeFileSync(outFile, JSON.stringify(payload, null, 2));
  console.log(`[forms] Manifest metadata written to ${outFile}`);
}

main();

