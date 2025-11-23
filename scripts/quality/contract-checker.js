#!/usr/bin/env node
/**
 * contract-checker.js
 *
 * Phase 402A 要求输出 REST ↔ GraphQL 契约一致性日志（logs/plan402/mapping/api-contract.log）。
 * 本脚本包装既有 drift-check，方便在 CI/本地记录可追溯证据。
 */
const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const PROJECT_ROOT = path.resolve(__dirname, "../..");
const LOG_DIR = path.join(PROJECT_ROOT, "logs/plan402/mapping");
const LOG_FILE = path.join(LOG_DIR, "api-contract.log");
const COMMANDS = [
  {
    name: "contract-drift-check",
    cmd: "node",
    args: ["scripts/contract/drift-check.js", "--include-fields"],
  },
  {
    name: "frontend-contract-generate",
    cmd: "npm",
    args: ["--prefix", "frontend", "run", "contract:generate"],
  },
];

fs.mkdirSync(LOG_DIR, { recursive: true });

function run(step) {
  const startedAt = new Date().toISOString();
  const result = spawnSync(step.cmd, step.args, {
    cwd: PROJECT_ROOT,
    encoding: "utf8",
    stdio: "pipe",
  });
  const finishedAt = new Date().toISOString();
  return {
    name: step.name,
    command: step.cmd + " " + step.args.join(" "),
    startedAt,
    finishedAt,
    status: result.status || 0,
    stdout: result.stdout || "",
    stderr: result.stderr || "",
  };
}

const steps = COMMANDS.map(run);
const payload = {
  timestamp: new Date().toISOString(),
  steps,
};

fs.writeFileSync(LOG_FILE, JSON.stringify(payload, null, 2));

steps.forEach((step) => {
  if (step.stdout) process.stdout.write(step.stdout);
  if (step.stderr) process.stderr.write(step.stderr);
});

const failed = steps.find((step) => step.status !== 0);
process.exit(failed ? failed.status : 0);
