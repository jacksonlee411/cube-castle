# Plan 18 Phase 1.3 复测问题解决方案

**文档版本**: 1.0
**创建时间**: 2025-10-05 10:40
**关联报告**: `reports/iig-guardian/plan18-phase1.3-retest-20251005.md`

---

## 一、问题诊断结果

### 1.1 误判问题：服务健康端点

**初步判断**（❌ 错误）:
```bash
curl http://localhost:8080/health  # 命令服务
curl http://localhost:8081/health  # 查询服务
# 结果：无响应
```

**根因分析**（✅ 正确）:
- **命令服务**: 实际端口 **9090**，健康端点 `/health`
- **查询服务**: 实际端口 **8090**，健康端点 `/health`

**验证结果**:
```bash
$ curl http://localhost:9090/health
{"status": "healthy", "service": "organization-command-service", "timestamp": "2025-10-05T10:38:23+08:00"}

$ curl http://localhost:8090/health
{"status":"healthy","service":"postgresql-graphql","database":"postgresql","performance":"optimized","timestamp":"2025-10-05T10:38:23.482332787+08:00"}
```

**结论**: 服务健康状态正常，无需修复。

---

### 1.2 真实问题：E2E 测试创建组织失败

#### 失败现象
```
TimeoutError: page.waitForURL: Timeout 30000ms exceeded.
=========================== logs ===========================
waiting for navigation until "load"
============================================================

  59 |       await page.getByTestId('form-submit-button').click();
  60 |
> 61 |       await page.waitForURL(/\/organizations\/[0-9]{7}\/temporal$/);
     |                  ^
  62 |       await expect(page.getByTestId('organization-form')).toBeVisible();
```

#### 页面快照分析
```yaml
- heading "新建组织 - 编辑组织信息"
- textbox "请输入组织名称": 测试部门E2E-mgd30qjw
- textbox "搜索并选择上级组织...": 1000000 - 高谷集团
- combobox: 部门 [selected]
- button "创建组织"
```

**关键发现**:
1. ✅ 表单数据已正确填写
2. ❌ 点击"创建组织"按钮后，页面未发生导航
3. ❌ 30 秒超时后仍停留在新建页面

---

## 二、根因定位

### 2.1 前端代码链路分析

#### 链路 1: 创建组织API调用

**位置**: `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts:279-303`

```typescript
export const createOrganizationUnit = async (
  payload: OrganizationRequest,
): Promise<string | null> => {
  const result = await unifiedRESTClient.request<CreateOrganizationResponse>(
    "/organization-units",  // ✅ REST 端点正确
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );

  // ❓ 问题：多个兜底逻辑，返回值不确定
  if (result.data?.code) return result.data.code;
  if (result.data?.organization?.code) return result.data.organization.code;
  if (result.code) return result.code;
  return result.organization?.code ?? null;
};
```

**问题点**:
- API 响应结构不明确，存在 4 种可能的路径
- 如果所有路径都匹配失败，返回 `null`

#### 链路 2: 创建成功后的回调

**位置**: `frontend/src/features/temporal/components/hooks/useTemporalMasterDetail.ts:334-339`

```typescript
const newOrganizationCode = await createOrganizationUnit(requestBody);

if (newOrganizationCode && onCreateSuccess) {
  onCreateSuccess(newOrganizationCode);  // ⚠️ 仅在 code 非空时触发
  return;
}
```

**问题点**:
- 如果 `newOrganizationCode` 为 `null`，**不会调用** `onCreateSuccess`
- 导致页面不跳转，测试超时

#### 链路 3: 路由导航

**位置**: `frontend/src/features/organizations/OrganizationTemporalPage.tsx:48-51`

```typescript
const handleCreateSuccess = (newOrganizationCode: string) => {
  navigate(`/organizations/${newOrganizationCode}/temporal`, { replace: true });
};
```

**依赖关系**:
```
API 返回正确的 code
  ↓
createOrganizationUnit 解析成功
  ↓
onCreateSuccess 被调用
  ↓
navigate 触发路由跳转
  ↓
测试通过
```

**关键缺失**: 如果 API 响应结构与预期不符，整个链路中断。

---

### 2.2 后端 API 响应结构验证

#### OpenAPI 契约定义

**位置**: `docs/api/openapi.yaml`

```yaml
/organization-units:
  post:
    summary: 创建新组织单元
    responses:
      '201':
        description: 组织单元创建成功
        content:
          application/json:
            schema:
              type: object
              required:
                - success
                - data
              properties:
                success:
                  type: boolean
                  example: true
                data:
                  type: object
                  required:
                    - code
                    - name
                    - unitType
                  properties:
                    code:
                      type: string
                      pattern: '^[0-9]{7}$'
                      example: "1000123"
```

**契约期望**: `response.data.code`（七位数字字符串）

#### 实际后端实现

**需要验证的文件**:
- `cmd/organization-command-service/internal/handlers/organization_create.go`

**验证步骤**:
```bash
# 手动测试创建组织 API
curl -X POST http://localhost:9090/api/organization-units \
  -H "Authorization: Bearer $(cat /tmp/dev-token.txt)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "手动测试部门",
    "parentCode": "1000000",
    "unitType": "DEPARTMENT",
    "effectiveDate": "2025-10-05",
    "description": "验证响应结构"
  }' | jq .
```

**预期响应**:
```json
{
  "success": true,
  "data": {
    "code": "1000456",
    "name": "手动测试部门",
    "unitType": "DEPARTMENT",
    "parentCode": "1000000",
    "effectiveDate": "2025-10-05",
    "status": "PLANNED",
    ...
  }
}
```

---

## 三、解决方案（三选一）

### 方案 A：修复后端响应结构（推荐 ⭐⭐⭐⭐⭐）

**优先级**: P0
**工作量**: 低
**风险**: 低
**契约合规性**: 高

#### 实施步骤

1. **检查后端实现**

```bash
grep -A 20 "func.*CreateOrganizationHandler\|HandleCreate" \
  cmd/organization-command-service/internal/handlers/organization_create.go
```

2. **确认响应格式**

预期代码结构：
```go
// ✅ 正确示例
response := map[string]interface{}{
    "success": true,
    "data": map[string]interface{}{
        "code":          newOrgCode,  // 必须字段
        "name":          req.Name,
        "unitType":      req.UnitType,
        "parentCode":    req.ParentCode,
        "effectiveDate": req.EffectiveDate,
        "status":        "PLANNED",
        // ...
    },
}

// ❌ 错误示例（前端无法解析）
response := map[string]interface{}{
    "success": true,
    "organization": map[string]interface{}{  // ❌ 应为 "data"
        "code": newOrgCode,
    },
}
```

3. **修复响应结构（如需要）**

编辑文件：`cmd/organization-command-service/internal/handlers/organization_create.go`

```go
// 确保响应符合 OpenAPI 契约
response := models.SuccessResponse{
    Success: true,
    Data: map[string]interface{}{
        "code":          newOrgCode,      // 前端解析：result.data.code
        "name":          createReq.Name,
        "unitType":      createReq.UnitType,
        "parentCode":    createReq.ParentCode,
        "effectiveDate": createReq.EffectiveDate,
        "status":        "PLANNED",
        "description":   createReq.Description,
        "createdAt":     time.Now().Format(time.RFC3339),
    },
}
```

4. **验证修复**

```bash
# 重启命令服务
pkill -f organization-command-service
go run ./cmd/organization-command-service/main.go &

# 手动测试API
curl -X POST http://localhost:9090/api/organization-units \
  -H "Authorization: Bearer $(scripts/plan18/get-dev-token.sh)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "验证修复-测试部门",
    "parentCode": "1000000",
    "unitType": "DEPARTMENT",
    "effectiveDate": "2025-10-05",
    "description": "验证响应格式"
  }' | jq '.data.code'

# 预期输出：
# "1000789"  （七位数字字符串）
```

5. **重新执行 E2E 测试**

```bash
scripts/plan18/run-business-flow-e2e.sh
```

**优势**:
- ✅ 符合 API 契约（单一事实来源）
- ✅ 根治问题，一次修复，所有依赖方受益
- ✅ 不影响前端代码

**劣势**:
- 需要后端开发权限

---

### 方案 B：增强前端解析逻辑（临时方案 ⭐⭐⭐）

**优先级**: P1
**工作量**: 低
**风险**: 中
**契约合规性**: 低（掩盖后端问题）

#### 实施步骤

1. **编辑前端 API 文件**

文件：`frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`

```typescript
export const createOrganizationUnit = async (
  payload: OrganizationRequest,
): Promise<string | null> => {
  const result = await unifiedRESTClient.request<CreateOrganizationResponse>(
    "/organization-units",
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );

  // ✅ 增强：添加调试日志
  console.log('[DEBUG] Create Organization API Response:', JSON.stringify(result, null, 2));

  // ✅ 增强：按优先级解析 code
  const code =
    result.data?.code ||                    // 优先级 1: 契约期望路径
    result.data?.organization?.code ||      // 优先级 2: 嵌套路径
    result.code ||                          // 优先级 3: 顶层路径
    result.organization?.code ||            // 优先级 4: 兜底路径
    null;

  // ✅ 增强：解析失败时记录错误
  if (!code) {
    console.error('[ERROR] Failed to extract organization code from response:', result);
  }

  return code;
};
```

2. **验证日志输出**

```bash
# 运行 Playwright 测试并查看控制台
cd frontend
npx playwright test tests/e2e/business-flow-e2e.spec.ts --headed

# 观察浏览器控制台输出：
# [DEBUG] Create Organization API Response: { ... }
```

3. **根据日志调整解析路径**

**优势**:
- ✅ 快速实施，无需后端改动
- ✅ 保留调试信息，便于后续排查

**劣势**:
- ❌ 违反"API契约优先"原则
- ❌ 掩盖后端问题，增加维护成本
- ❌ 临时方案，需后续清理

---

### 方案 C：完整诊断流程（彻底解决 ⭐⭐⭐⭐⭐）

**优先级**: P0
**工作量**: 中
**风险**: 低
**契约合规性**: 高

#### 分阶段实施

##### Phase 1: 诊断阶段

1. **验证后端 API 实际响应**

```bash
# 脚本：scripts/plan18/diagnose-create-api.sh
#!/bin/bash
set -e

echo "=== Phase 1: 诊断创建组织 API 响应结构 ==="

# 1. 获取 JWT Token
TOKEN=$(curl -s http://localhost:9090/auth/dev-token | jq -r '.token')
echo "✅ JWT Token 已获取"

# 2. 调用创建组织 API
RESPONSE=$(curl -s -X POST http://localhost:9090/api/organization-units \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "诊断测试部门",
    "parentCode": "1000000",
    "unitType": "DEPARTMENT",
    "effectiveDate": "'$(date +%Y-%m-%d)'",
    "description": "API 响应结构诊断"
  }')

echo "📋 完整响应："
echo "$RESPONSE" | jq .

# 3. 验证契约合规性
CODE=$(echo "$RESPONSE" | jq -r '.data.code // empty')
if [[ "$CODE" =~ ^[0-9]{7}$ ]]; then
  echo "✅ 契约合规：response.data.code = $CODE"
else
  echo "❌ 契约违规：无法从 response.data.code 获取七位编码"
  echo "   实际值：$CODE"

  # 尝试其他路径
  ALT_CODE=$(echo "$RESPONSE" | jq -r '.code // .organization.code // .data.organization.code // empty')
  if [[ -n "$ALT_CODE" ]]; then
    echo "⚠️  发现备用路径：$ALT_CODE"
    echo "   需要修复后端以符合契约"
  fi
fi
```

运行诊断：
```bash
chmod +x scripts/plan18/diagnose-create-api.sh
scripts/plan18/diagnose-create-api.sh
```

##### Phase 2: 修复阶段

**根据诊断结果选择行动**:

| 诊断结果 | 行动方案 |
|---------|---------|
| ✅ `response.data.code` 正确返回 | → 跳转到 Phase 3（前端问题） |
| ❌ 响应结构不符合契约 | → 执行方案 A（修复后端） |
| ❌ API 返回错误状态码 | → 检查权限/数据库/业务规则 |

##### Phase 3: 前端调试

1. **启用前端请求拦截日志**

编辑 `frontend/src/shared/api/restClient.ts`:

```typescript
// 在 request 方法中添加
console.log(`[REST] ${method} ${url}`, { body, response });
```

2. **检查浏览器 DevTools Network 面板**

```bash
# Playwright 测试模式（保留浏览器窗口）
cd frontend
npx playwright test tests/e2e/business-flow-e2e.spec.ts \
  --headed \
  --debug
```

在测试暂停时：
- 打开 DevTools → Network 标签
- 筛选 `organization-units`
- 查看 Response 标签

3. **验证前端解析逻辑**

在 `temporalMasterDetailApi.ts` 中添加断点：
```typescript
const code = result.data?.code;  // 在此行设置断点
console.log('Parsed code:', code, 'from result:', result);
```

##### Phase 4: 回归测试

```bash
# 完整流程测试
scripts/plan18/run-business-flow-e2e.sh

# 查看测试报告
ls -lh reports/iig-guardian/plan18-business-flow-*.log | tail -1
```

---

## 四、快速修复脚本（推荐使用）

### 4.1 一键诊断与修复脚本

**文件**: `scripts/plan18/fix-create-organization.sh`

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🔧 Plan 18 Phase 1.3 - 创建组织问题修复脚本"
echo "=============================================="

# Step 1: 诊断 API 响应
echo ""
echo "📋 Step 1: 诊断后端 API 响应结构..."
TOKEN=$(curl -s http://localhost:9090/auth/dev-token | jq -r '.token')
RESPONSE=$(curl -s -X POST http://localhost:9090/api/organization-units \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "诊断测试-'$(date +%s)'",
    "parentCode": "1000000",
    "unitType": "DEPARTMENT",
    "effectiveDate": "'$(date +%Y-%m-%d)'",
    "description": "自动诊断"
  }')

echo "$RESPONSE" | jq .

# Step 2: 验证契约合规性
CODE=$(echo "$RESPONSE" | jq -r '.data.code // empty')
if [[ "$CODE" =~ ^[0-9]{7}$ ]]; then
  echo "✅ API 响应符合契约：response.data.code = $CODE"
  echo ""
  echo "🎯 问题可能在前端，建议："
  echo "   1. 检查浏览器 DevTools Network 面板"
  echo "   2. 启用前端日志（见方案 B）"
  echo "   3. 使用 Playwright --debug 模式"
else
  echo "❌ API 响应不符合契约"
  echo ""
  echo "🔧 需要修复后端 API 响应结构："
  echo "   文件: cmd/organization-command-service/internal/handlers/organization_create.go"
  echo "   确保响应格式："
  echo "   {"
  echo "     \"success\": true,"
  echo "     \"data\": {"
  echo "       \"code\": \"1000XXX\","
  echo "       ..."
  echo "     }"
  echo "   }"
fi

# Step 3: 提供后续行动
echo ""
echo "📚 详细解决方案请参考："
echo "   reports/iig-guardian/plan18-phase1.3-solution-20251005.md"
```

### 4.2 使用方法

```bash
chmod +x scripts/plan18/fix-create-organization.sh
scripts/plan18/fix-create-organization.sh
```

---

## 五、验收标准

### 5.1 功能验收

| 验收项 | 标准 | 验证方法 |
|-------|------|---------|
| API 响应结构 | 符合 OpenAPI 契约 `response.data.code` | 手动 curl 测试 |
| 前端解析成功 | `createOrganizationUnit` 返回七位编码 | 浏览器控制台日志 |
| 路由跳转成功 | 创建后自动跳转到 `/organizations/{code}/temporal` | Playwright 测试通过 |
| E2E 测试通过 | 10/10 测试通过，0 失败 | `run-business-flow-e2e.sh` 输出 |

### 5.2 性能验收

| 指标 | 标准 | 当前值 | 目标值 |
|------|------|--------|--------|
| 创建组织 API 响应时间 | < 500ms | 待测 | < 300ms |
| 路由跳转延迟 | < 200ms | 待测 | < 100ms |
| E2E 测试总耗时 | < 60s | 43.0s | < 45s |

### 5.3 回归测试清单

```bash
# ✅ 1. 服务健康检查
curl http://localhost:9090/health
curl http://localhost:8090/health

# ✅ 2. 创建组织 API 手动测试
curl -X POST http://localhost:9090/api/organization-units \
  -H "Authorization: Bearer $(curl -s http://localhost:9090/auth/dev-token | jq -r '.token')" \
  -H "Content-Type: application/json" \
  -d '{"name":"回归测试","parentCode":"1000000","unitType":"DEPARTMENT","effectiveDate":"2025-10-05","description":"test"}' \
  | jq '.data.code'

# ✅ 3. E2E 完整流程
scripts/plan18/run-business-flow-e2e.sh

# ✅ 4. 检查日志
tail -50 reports/iig-guardian/plan18-business-flow-$(date +%Y%m%d)*.log
```

---

## 六、后续优化建议

### 6.1 短期优化（1 周内）

1. **增强脚本健康检查**
   - 在 `run-business-flow-e2e.sh` 中添加服务就绪轮询
   - 超时后自动输出服务日志

2. **完善错误提示**
   - 前端 API 调用失败时显示具体错误
   - 区分网络错误、权限错误、业务错误

3. **补充单元测试**
   - 后端：`organization_create_test.go` 验证响应格式
   - 前端：`temporalMasterDetailApi.test.ts` 验证 code 解析

### 6.2 中期优化（2-4 周）

1. **API 响应标准化**
   - 所有 REST API 统一响应格式
   - 自动化契约测试（Spectral + Dredd）

2. **前端类型安全**
   - 为 `CreateOrganizationResponse` 添加严格类型定义
   - 使用 TypeScript 编译时检查响应结构

3. **监控与告警**
   - 生产环境 API 响应结构监控
   - 契约违规自动告警

### 6.3 长期优化（1-3 个月）

1. **契约测试自动化**
   - CI/CD 集成契约测试
   - 前后端契约变更自动阻塞

2. **端到端可观测性**
   - 分布式追踪（Jaeger/OpenTelemetry）
   - 用户会话回放（Sentry/LogRocket）

3. **自愈能力**
   - API 响应格式自动适配（有限兜底）
   - 降级方案（本地状态管理）

---

## 七、参考资料

### 7.1 相关文档

- 复测报告: `reports/iig-guardian/plan18-phase1.3-retest-20251005.md`
- API 契约: `docs/api/openapi.yaml` (Line 145-180: `/organization-units` POST)
- 前端 API 层: `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts:279`
- 后端处理器: `cmd/organization-command-service/internal/handlers/organization_create.go`

### 7.2 工具链

- Playwright 文档: https://playwright.dev/docs/api/class-page#page-wait-for-url
- OpenAPI Spec: https://spec.openapis.org/oas/v3.1.0
- jq 手册: https://stedolan.github.io/jq/manual/

### 7.3 团队联系

- 后端 API 问题: Organization Command Service Team
- 前端路由问题: Frontend Architecture Team
- E2E 测试问题: QA Automation Team

---

**文档维护**:
- 初版: 2025-10-05 (Implementation Inventory Guardian)
- 更新: 根据实际修复结果更新验收状态
