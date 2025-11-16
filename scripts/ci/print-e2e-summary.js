#!/usr/bin/env node
/**
 * 输出 Playwright JSON 报告的机器可读 SUMMARY
 * 用法：node scripts/ci/print-e2e-summary.js <planId> [logsDir]
 * 默认 logsDir=logs/plan<planId>
 */
const fs = require('fs');
const path = require('path');

function findResultFiles(dir) {
  try {
    const files = fs.readdirSync(dir).filter(f => /^results-.*\.json$/.test(f));
    return files.map(f => path.join(dir, f)).sort();
  } catch {
    return [];
  }
}

function walkForStatuses(obj, acc) {
  if (!obj || typeof obj !== 'object') return;
  // 统计 tests/ results 中的 status
  if (Array.isArray(obj.tests)) {
    obj.tests.forEach(t => {
      if (t && typeof t === 'object') {
        if (typeof t.status === 'string') {
          acc.all += 1;
          if (t.status === 'passed') acc.passed += 1;
          if (t.status === 'failed') acc.failed += 1;
          if (t.status === 'skipped') acc.skipped += 1;
          if (t.status === 'flaky') acc.flaky += 1;
        }
        if (Array.isArray(t.results)) {
          t.results.forEach(r => {
            if (r && typeof r.status === 'string') {
              acc.runs += 1;
              if (r.status === 'passed') acc.passedRuns += 1;
              if (r.status === 'failed') acc.failedRuns += 1;
              if (r.status === 'skipped') acc.skippedRuns += 1;
            }
          });
        }
      }
    });
  }
  if (Array.isArray(obj.suites)) {
    obj.suites.forEach(s => walkForStatuses(s, acc));
  }
  // 递归其他子对象，尽量容错
  Object.keys(obj).forEach(k => {
    const v = obj[k];
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      walkForStatuses(v, acc);
    }
  });
}

function summarize(file) {
  try {
    const raw = fs.readFileSync(file, 'utf8');
    const data = JSON.parse(raw);
    const acc = {
      all: 0, passed: 0, failed: 0, skipped: 0, flaky: 0,
      runs: 0, passedRuns: 0, failedRuns: 0, skippedRuns: 0
    };
    walkForStatuses(data, acc);
    const total = acc.all || acc.runs;
    const passed = acc.passed || acc.passedRuns;
    const failed = acc.failed || acc.failedRuns;
    const skipped = acc.skipped || acc.skippedRuns;
    // 机器可读一行输出
    console.log(`SUMMARY file=${path.basename(file)} total=${total} passed=${passed} failed=${failed} skipped=${skipped}`);
    return { total, passed, failed, skipped };
  } catch (e) {
    console.log(`SUMMARY file=${path.basename(file)} parse=error`);
    return null;
  }
}

function main() {
  const planId = process.argv[2] || '254';
  const root = process.argv[3] || path.join('logs', `plan${planId}`);
  const files = findResultFiles(root);
  if (files.length === 0) {
    console.log(`SUMMARY dir=${root} files=0`);
    process.exit(0);
  }
  let agg = { total: 0, passed: 0, failed: 0, skipped: 0 };
  files.forEach(f => {
    const s = summarize(f);
    if (s) {
      agg.total += s.total;
      agg.passed += s.passed;
      agg.failed += s.failed;
      agg.skipped += s.skipped;
    }
  });
  console.log(`SUMMARY_ALL plan=${planId} total=${agg.total} passed=${agg.passed} failed=${agg.failed} skipped=${agg.skipped}`);
}

if (require.main === module) {
  main();
}

