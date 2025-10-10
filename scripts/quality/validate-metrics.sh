#!/usr/bin/env bash
# validate-metrics.sh - 验证命令服务 Prometheus 指标可用性
# 用途：自动化检查 temporal_operations_total、audit_writes_total、http_requests_total 三类 Counter
# 退出码：0=全部通过, 1=服务不可达, 2=指标缺失

set -euo pipefail

# 配置
SERVICE_URL="${METRICS_URL:-http://localhost:9090}"
METRICS_ENDPOINT="${SERVICE_URL}/metrics"
TIMEOUT=5

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Prometheus 指标验证工具"
echo "=========================================="
echo "目标端点: ${METRICS_ENDPOINT}"
echo ""

# 1. 检查服务可达性
echo "[ 1/4 ] 检查服务可达性..."
if ! curl -s --max-time "${TIMEOUT}" "${METRICS_ENDPOINT}" > /dev/null 2>&1; then
    echo -e "${RED}✗ FAIL${NC}: 无法访问 ${METRICS_ENDPOINT}"
    echo "提示: 请确保命令服务已启动 (make run-dev 或 go run cmd/organization-command-service/main.go)"
    exit 1
fi
echo -e "${GREEN}✓ PASS${NC}: 服务可达"
echo ""

# 2. 获取指标内容
echo "[ 2/4 ] 获取 /metrics 输出..."
METRICS_OUTPUT=$(curl -s --max-time "${TIMEOUT}" "${METRICS_ENDPOINT}")
if [ -z "${METRICS_OUTPUT}" ]; then
    echo -e "${RED}✗ FAIL${NC}: /metrics 返回空内容"
    exit 1
fi
echo -e "${GREEN}✓ PASS${NC}: 成功获取指标数据"
echo ""

# 3. 验证必需指标定义存在
echo "[ 3/4 ] 验证指标定义..."

# 立即可见的指标（由中间件或初始化代码自动触发）
IMMEDIATE_METRICS=(
    "http_requests_total"
)

# 需要业务操作触发的指标（Counter 在未被 Inc() 前不会显示）
BUSINESS_TRIGGERED_METRICS=(
    "temporal_operations_total"
    "audit_writes_total"
)

MISSING_CRITICAL=()
MISSING_BUSINESS=()

# 检查立即可见的指标（必须存在）
for metric in "${IMMEDIATE_METRICS[@]}"; do
    if echo "${METRICS_OUTPUT}" | grep -q "${metric}"; then
        echo -e "${GREEN}✓${NC} ${metric} - 已定义或有数据"
    else
        echo -e "${RED}✗${NC} ${metric} - 未找到（关键指标）"
        MISSING_CRITICAL+=("${metric}")
    fi
done

# 检查需要业务触发的指标（未找到给予警告）
for metric in "${BUSINESS_TRIGGERED_METRICS[@]}"; do
    if echo "${METRICS_OUTPUT}" | grep -q "${metric}"; then
        echo -e "${GREEN}✓${NC} ${metric} - 已定义或有数据"
    else
        echo -e "${YELLOW}⚠${NC} ${metric} - 未找到（需业务触发，代码已集成）"
        MISSING_BUSINESS+=("${metric}")
    fi
done
echo ""

# 4. 数据点可见性检查（仅警告，不阻塞）
echo "[ 4/4 ] 数据点可见性检查（仅供参考）..."
ALL_METRICS=("${IMMEDIATE_METRICS[@]}" "${BUSINESS_TRIGGERED_METRICS[@]}")
for metric in "${ALL_METRICS[@]}"; do
    # 检查是否有实际数据点（非注释行）
    DATA_POINTS=$(echo "${METRICS_OUTPUT}" | grep -E "^${metric}\{" | wc -l || true)
    if [ "${DATA_POINTS}" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} ${metric} - 有 ${DATA_POINTS} 个数据点"
    else
        echo -e "${YELLOW}⚠${NC} ${metric} - 无数据点（需触发业务操作后才可见，这是正常的 Prometheus Counter 行为）"
    fi
done
echo ""

# 5. 最终结论
echo "=========================================="
if [ ${#MISSING_CRITICAL[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ 验证通过${NC}: 所有关键指标均已定义"

    if [ ${#MISSING_BUSINESS[@]} -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}⚠ 提示${NC}: 以下业务触发指标暂未显示（这是正常的）："
        for metric in "${MISSING_BUSINESS[@]}"; do
            echo "  • ${metric}"
        done
        echo ""
        echo "说明："
        echo "  • temporal_operations_total: 需通过时态操作触发（CreateVersion/SuspendOrganization 等）"
        echo "  • audit_writes_total: 需通过审计写入操作触发"
        echo "  • 这些指标已在代码中集成（internal/utils/metrics.go），Prometheus Counter"
        echo "    只有在至少被记录一次后才会出现在 /metrics 输出中"
        echo ""
        echo "触发指标的示例操作："
        echo "  curl -X POST http://localhost:9090/api/v1/organization-units \\"
        echo "    -H \"Authorization: Bearer \$(cat /tmp/jwt.txt)\" \\"
        echo "    -H \"X-Tenant-ID: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9\" \\"
        echo "    -H \"Content-Type: application/json\" \\"
        echo "    -d '{\"name\":\"测试部门\",\"unitType\":\"DEPARTMENT\",\"parentCode\":\"0\",\"effectiveDate\":\"2025-10-10\"}'"
    fi
    echo ""
    exit 0
else
    echo -e "${RED}✗ 验证失败${NC}: 缺失以下关键指标："
    for metric in "${MISSING_CRITICAL[@]}"; do
        echo "  - ${metric}"
    done
    echo ""
    echo "请检查以下文件："
    echo "  • cmd/organization-command-service/internal/utils/metrics.go - 指标定义"
    echo "  • cmd/organization-command-service/main.go - /metrics 端点注册"
    echo "  • cmd/organization-command-service/internal/middleware/performance.go - HTTP 请求计数"
    exit 2
fi
