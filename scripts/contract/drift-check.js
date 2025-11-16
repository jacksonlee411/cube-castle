#!/usr/bin/env node
/**
 * drift-check.js
 *
 * 对比 OpenAPI ↔ GraphQL 的契约片段（以 shared/contracts/organization.json 为输入），
 * 检测主要枚举（UnitType/Status/OperationType）的差异，输出报告并可在发现差异时失败。
 *
 * 用法:
 *   node scripts/contract/drift-check.js [--fail-on-diff]
 *
 * 产出:
 *   reports/contracts/drift-report.json
 */

const fs = require('fs');
const path = require('path');

const PROJECT_ROOT = path.resolve(__dirname, '../..');
const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');
const REPORT_DIR = path.join(PROJECT_ROOT, 'reports', 'contracts');
const REPORT_PATH = path.join(REPORT_DIR, 'drift-report.json');

const ENUM_KEYS = ['unitType', 'status', 'operationType'];

function readJSON(p) {
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

function computeEnumDiff(restArr = [], gqlArr = []) {
  const rest = new Set(restArr);
  const gql = new Set(gqlArr);
  const missingInRest = gqlArr.filter(v => !rest.has(v));
  const missingInGql = restArr.filter(v => !gql.has(v));
  return { missingInRest, missingInGql };
}

function main() {
  const failOnDiff = process.argv.includes('--fail-on-diff');
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
    results[key] = diff;
    if ((diff.missingInRest && diff.missingInRest.length) ||
        (diff.missingInGql && diff.missingInGql.length)) {
      hasDiff = true;
    }
  }

  const report = {
    generatedAt: new Date().toISOString(),
    source: path.relative(PROJECT_ROOT, CONTRACT_PATH),
    results,
  };
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

