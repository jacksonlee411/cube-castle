#!/usr/bin/env bash
set -euo pipefail

# 审计历史 GraphQL 查询性能基线脚本
# 单一事实来源：`docs/api/schema.graphql`
# 使用方式：
#   JWT_FILE=.cache/dev.jwt RECORD_ID=<record-id> ./tests/perf/graphql-audit-history-benchmark.sh
# 可选变量：
#   GRAPHQL_URL (默认 http://localhost:8090/graphql)
#   ITERATIONS   (默认 10 次)
#   PAYLOAD_FILE (自定义 GraphQL Query 文件)
#   CONCURRENCY  (并发数，默认 1)

GRAPHQL_URL=${GRAPHQL_URL:-"http://localhost:8090/graphql"}
ITERATIONS=${ITERATIONS:-10}
CONCURRENCY=${CONCURRENCY:-1}
JWT_FILE=${JWT_FILE:-".cache/dev.jwt"}

if [[ ! -f "$JWT_FILE" ]]; then
  echo "❌ 找不到 JWT 文件: $JWT_FILE" >&2
  exit 1
fi

JWT_TOKEN=$(cat "$JWT_FILE")

if [[ -n ${PAYLOAD_FILE:-} ]]; then
  GRAPHQL_QUERY=$(cat "$PAYLOAD_FILE")
else
  if [[ -z ${RECORD_ID:-} ]]; then
    echo "❌ 请设置 RECORD_ID 或提供 PAYLOAD_FILE" >&2
    exit 1
  fi
  read -r -d '' GRAPHQL_QUERY <<EOF_QUERY || true
{
  "query": "query AuditHistoryBaseline(\$recordId: String!, \$limit: Int) { auditHistory(recordId: \$recordId, limit: \$limit) { auditId recordId operation timestamp operationReason modifiedFields changes { field dataType oldValue newValue } } }",
  "variables": {
    "recordId": "${RECORD_ID}",
    "limit": ${LIMIT:-50}
  }
}
EOF_QUERY
fi

print_header() {
  printf '\n%-8s %-12s %-12s\n' "迭代" "耗时(ms)" "状态码"
  printf '%-8s %-12s %-12s\n' "--------" "------------" "------------"
}

measure_once() {
  local iteration=$1
  local start end duration status
  start=$(date +%s%3N)
  status=$(curl -s -o /tmp/audit-history-benchmark-${iteration}.json -w "%{http_code}" \
    -X POST "$GRAPHQL_URL" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    --data "$GRAPHQL_QUERY")
  end=$(date +%s%3N)
  duration=$((end - start))
  printf '%-8s %-12s %-12s\n' "$iteration" "$duration" "$status"
}

run_serial() {
  print_header
  for ((i=1; i<=ITERATIONS; i++)); do
    measure_once "$i"
  done
}

run_parallel() {
  print_header
  seq 1 "$ITERATIONS" | xargs -I{} -P "$CONCURRENCY" bash -c 'measure_once "$@"' _ {}
}

if [[ $CONCURRENCY -le 1 ]]; then
  run_serial
else
  run_parallel
fi

python3 <<'PY_END'
import json
import os
from statistics import mean

results = []
for name in os.listdir('/tmp'):
    if name.startswith('audit-history-benchmark-') and name.endswith('.json'):
        path = os.path.join('/tmp', name)
        try:
            with open(path, 'r', encoding='utf-8') as f:
                data = json.load(f)
            results.append({
                'file': path,
                'errors': data.get('errors'),
                'count': len(data.get('data', {}).get('auditHistory', []) if isinstance(data.get('data'), dict) else []),
            })
        except Exception as exc:  # noqa: BLE001
            print(f"⚠️  无法解析 {path}: {exc}")

if results:
    failures = [r for r in results if r['errors']]
    print('\n--- 汇总 ---')
    print(f"总样本: {len(results)}")
    print(f"含错误的响应: {len(failures)}")
    if failures:
        for item in failures[:3]:
            print(f"  ✖ 错误样本: {item['file']}")
    counts = [r['count'] for r in results]
    print(f"平均返回记录数: {mean(counts):.2f}")
else:
    print('\n⚠️  未产生任何响应文件 (检查 curl 调用)')
PY_END

echo '\n✅ 性能基线采集完成。请将结果记录至 reports/temporal/audit-history-nullability.md 的性能基线表格。'
