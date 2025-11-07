#!/usr/bin/env bash

set -euo pipefail

LOG_DIR="${LOG_DIR:-logs/219E}"
SUCCESS_THRESHOLD="${SUCCESS_THRESHOLD:-0.9}"
STRICT="${STRICT:-0}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "⚠️  当前环境缺少 python3，无法解析 REST Benchmark JSON Summary"
  exit 0
fi

latest_log="$(ls -t "${LOG_DIR}"/perf-rest-*.log 2>/dev/null | head -n 1 || true)"

if [[ -z "${latest_log}" ]]; then
  echo "⚠️  未找到 ${LOG_DIR}/perf-rest-*.log（尚未执行 scripts/perf/rest-benchmark.sh）"
  exit 0
fi

python3 - "${latest_log}" "${SUCCESS_THRESHOLD}" "${STRICT}" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
threshold = float(sys.argv[2])
strict = int(sys.argv[3])

lines = path.read_text(encoding='utf-8', errors='ignore').splitlines()
collect = False
buffer = []
depth = 0
for line in lines:
    if 'JSON Summary' in line:
        collect = True
        continue
    if not collect:
        continue
    if not buffer and not line.strip():
        continue
    buffer.append(line)
    depth += line.count('{') - line.count('}')
    if depth <= 0 and buffer:
        break

if not buffer:
    print(f"⚠️  {path} 未包含 JSON Summary 块")
    sys.exit(1 if strict else 0)

try:
    summary = json.loads('\n'.join(buffer))
except json.JSONDecodeError as exc:
    print(f"⚠️  无法解析 {path.name} 中的 JSON Summary: {exc}")
    sys.exit(1 if strict else 0)

success_rate = summary.get('successRate')
success_count = summary.get('successCount')
completed = summary.get('completed')
latency = summary.get('latency') or {}
p50 = latency.get('p50')
p95 = latency.get('p95')
p99 = latency.get('p99')
status_counts = summary.get('statusCounts') or {}

def fmt_percent(value):
    if value is None:
        return "n/a"
    return f"{value * 100:.2f}%"

def fmt_number(value):
    if value is None:
        return "n/a"
    return f"{value:.2f}"

status_overview = ', '.join(f"{code}={count}" for code, count in sorted(status_counts.items())) or "无"
issues = []

if success_rate is None:
    issues.append("缺少 successRate")
elif success_rate < threshold:
    issues.append(f"successRate={success_rate * 100:.1f}% < {threshold * 100:.0f}%")

if not success_count:
    issues.append("无成功请求")

if p95 is None or p99 is None:
    issues.append("P95/P99 不可用")

print(f"📈 最新 REST Benchmark: {path}")
print(f"   successRate: {fmt_percent(success_rate)}  (success={success_count or 0} / completed={completed or 0})")
print(f"   latency(ms): p50={fmt_number(p50)}  p95={fmt_number(p95)}  p99={fmt_number(p99)}")
print(f"   status 分布: {status_overview}")

if issues:
    print(f"   ⚠️  {', '.join(issues)}")
    if strict:
        sys.exit(1)
else:
    print("   ✅ 指标处于配置阈值之上")
PY
