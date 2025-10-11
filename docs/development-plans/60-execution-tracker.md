# 60号计划执行跟踪

**启动日期**: 2025-10-10
**执行模式**: 单人全栈工程师
**当前阶段**: 第二阶段（后端观测与运维巩固）
**预计完成**: 2025-12-20（10周）

## 进度看板

### 阶段零：启动准备（3-5天）
- [x] Step 0.1: 计划文档正式化
- [x] Step 0.2: 组建跨团队小组（单人执行，已跳过）
- [x] Step 0.3: 评估前置条件
- [x] Step 0.4: 建立迭代跟踪

### 第一阶段：契约与类型统一（2周）
- [x] Week 1: 契约同步脚本开发
- [x] Week 2: 枚举对齐与代码生成集成（进行中）

### 第二阶段：后端服务与中间件收敛（3周）
- [x] Prometheus 指标补充与暴露 `/metrics` ✅ 2025-10-10
  - 实现 `temporal_operations_total`、`audit_writes_total`、`http_requests_total` 三类 Counter
  - 在 main.go 中暴露 `/metrics` 端点
  - 创建 `scripts/quality/validate-metrics.sh` 自动化验证脚本
- [ ] ~~运维开关与熔断策略调研~~（已取消：判定为过度设计）
- [x] Phase 2 验收草稿与监控文档完善 ✅ 2025-10-10
  - 完成 `62-phase-2-acceptance-draft.md` v0.2（含运行时验证结果）
  - 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 添加"📊 运行监控（Prometheus）"章节
  - 更新 `62-backend-middleware-refactor-plan.md` v2.1 并标记已完成项

### 第三阶段：前端 API/Hooks/配置整治（2-3周）
- [x] Week 6：统一 React Query 客户端与错误包装（`shared/api/queryClient.ts` 已落地）
- [x] Week 7：组织相关 Hooks 迁移（查询与写操作整合，现有调用方无需额外桥接层）
- [x] Week 8：端口/环境助手重写（`shared/config/environment.ts` 等已更新）
- [x] QA 冒烟巡检：`npm run test:e2e:smoke` 通过，报告归档于 `frontend/playwright-report/`
- [x] Vitest 覆盖率 ≥ 75%（2025-10-11：Phase 3 模块语句覆盖率 84.1%，记录见 06 号文档）
- [ ] 代码包体积下降 ≥ 5%（`npm run build:analyze` 已恢复可用，待比对最新输出与基线）

### 第四阶段：工具与验证体系巩固（1-2周）
- [ ] 待启动

## 本周进展（Week 41, 2025-10-10）

### 已完成
**第一阶段（契约与类型统一）**:
- ✅ 创建 60 号计划文档 v1.1
- ✅ 创建 61 号执行计划
- ✅ 创建执行分支 `feature/plan-61-system-quality-refactor`
- ✅ 更新开发计划索引（00-README.md）
- ✅ 验证 API 契约状态（干净无变更）
- ✅ 检查 53/56 号计划（无阻塞项）
- ✅ 生成实现清单基线（269行）
- ✅ 建立执行跟踪机制（60-execution-tracker.md）
- ✅ 搭建契约同步脚本框架（sync.sh + 占位子脚本）
- ✅ 实现 OpenAPI 契约解析器（openapi-to-json.js）
- ✅ 实现 GraphQL 契约解析器（graphql-to-json.js，发现 REST/GraphQL 枚举差异）
- ✅ 实现 Go 类型生成器（generate-go-types.js）
- ✅ 实现 TypeScript 类型生成器（generate-ts-types.js）
- ✅ 建立契约快照基线与校验脚本（tests/contract/）
- ✅ 将契约快照校验纳入 CI（contract-snapshot job）

**第二阶段（后端观测与运维巩固）**:
- ✅ 命令服务 Prometheus 指标实现（2025-10-10）
  - `temporal_operations_total{operation, status}` —— 时态操作计数
  - `audit_writes_total{status}` —— 审计写入计数
  - `http_requests_total{method, route, status}` —— HTTP 请求计数
  - 代码位置：`cmd/organization-command-service/internal/utils/metrics.go`
- ✅ `/metrics` 端点暴露（main.go:202-207）
- ✅ 指标验证脚本开发（`scripts/quality/validate-metrics.sh`）
  - 支持关键指标与业务触发指标分类验证
  - CI 集成友好（适当退出码）
- ✅ 运行时指标验证与采样（2025-10-10）
  - 验证 `http_requests_total` 立即可见性
  - 记录业务触发指标的 Prometheus Counter 行为
- ✅ 监控文档完善
  - 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 添加"📊 运行监控（Prometheus）"完整章节
  - 记录指标定义、触发条件、验证方法、技术说明
- ✅ Phase 2 验收报告草稿（`62-phase-2-acceptance-draft.md` v0.2）
  - 完整验收清单与运行时验证结果
  - 风险识别与缓解措施
  - 后续改进建议
- ✅ 62 号计划文档更新（v2.1）
  - 标记已完成任务与交付物
  - 更新计划状态为"第一批交付完成"

### 进行中
- 🔄 第三阶段跟进事项：
  - 对比 `npm run build:analyze` 输出与历史基线，形成 5% 体积结论
  - 在 QA/运行手册中补充 HTTPS 场景示例并复核 06 号文档
  - 将 64 号验收草案推进至 v0.2（完成剩余待办后申请评审）

### 下周计划
- 评估第二阶段剩余工作优先级（运维开关）
- 考虑是否启动第三阶段（前端 API/Hooks/配置整治）
- 根据业务需求决定是否进一步完善 62 号计划剩余项

## 风险与问题日志

| ID | 风险/问题 | 影响 | 状态 | 负责人 | 应对措施 |
|----|----------|------|------|--------|---------|
| R01 | 契约脚本开发延期 | 中 | 监控中 | 全栈工程师 | 保留人工校对备选 |
| R02 | 单人执行时间压力 | 高 | 监控中 | 全栈工程师 | 按阶段门禁严格验收，必要时调整范围 |

## 执行环境

- **分支**: `feature/plan-61-system-quality-refactor`
- **基线提交**: d6714146
- **当前提交**: 5ca98ad5
- **基线文件**: `.baseline-before-refactor.md` (269行)

## 变更记录

- 2025-10-10: 初始化跟踪文档
- 2025-10-10: 完成阶段零 Step 0.1-0.4
- 2025-10-10: 启动第一阶段 Week 1（契约脚本框架与 OpenAPI 解析器）
- 2025-10-10: 完成第二阶段首批交付（Prometheus 指标、验证脚本、监控文档、Phase 2 验收报告 v0.2）
  - 实现三类 Prometheus Counter 指标
  - 创建自动化验证脚本 `scripts/quality/validate-metrics.sh`
  - 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 添加运行监控章节
  - 完成 `62-phase-2-acceptance-draft.md` v0.2 含运行时验证结果
  - 更新 `62-backend-middleware-refactor-plan.md` v2.1 标记已完成项
