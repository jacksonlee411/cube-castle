# 16号计划：代码异味分析与改进计划（Go工程实践优化版）

## 计划概述

**计划名称**: 代码异味分析与改进计划（Go工程实践优化版）
**计划编号**: 16
**创建日期**: 2025-09-29
**更新日期**: 2025-09-30（修复评审意见，对齐唯一事实来源）
**优先级**: P2（中高优先级 - 平衡质量与实用性）
**预计完成时间**: 5-6周（增加20%缓冲）
**负责团队**: 架构组（Owner） / 前端团队 / 后端团队 / QA 联合小组
**进展同步频率**: 每周五更新至 `docs/development-plans/06-integrated-teams-progress-log.md`

## 执行摘要

通过分析项目代码库，发现当前平均文件行数已达到312.7行（Go）和163.0行（TypeScript）。基于Go语言工程实践的最佳标准，本计划制定了平衡的重构策略，重点解决超大文件问题，将代码质量提升到行业标准水平。

**基线数据来源**: `reports/iig-guardian/code-smell-baseline-20250929.md`（唯一事实来源）
- Go后端：54个文件，16,888行，红灯3个（27.5%），橙灯5个（22.1%）
- 前端TS：112个文件，18,254行，红灯2个（12.2%），橙灯9个（26.2%）

## 问题识别与分析

### 🔍 代码异味识别结果

#### 1. 文件复杂度异味（基于Go工程实践标准）
**问题描述**: 部分文件超出Go社区最佳实践标准，影响代码可维护性

**Go后端文件分析**（基于Go工程实践标准）:
- **红灯区域**（>800行，强制重构）:
  - `cmd/organization-query-service/main.go`: 2,264行 ⚠️ 严重超标
  - `cmd/organization-command-service/internal/handlers/organization.go`: 1,399行 ⚠️ 严重超标
  - `cmd/organization-command-service/internal/repository/organization.go`: 817行 ⚠️ 需重构
- **橙灯区域**（500-800行，需要架构师评估）:
  - `cmd/organization-command-service/internal/services/temporal.go`: 773行
  - `cmd/organization-command-service/internal/repository/temporal_timeline.go`: 685行
  - `cmd/organization-command-service/internal/validators/business.go`: 596行
  - `cmd/organization-command-service/internal/audit/logger.go`: 595行
  - `cmd/organization-command-service/internal/authbff/handler.go`: 589行
- **黄灯区域**（300-500行，关注结构优化）: 其他17个文件
- **统计数据**: 54个Go文件，平均312.7行（基线参见 `reports/iig-guardian/code-smell-baseline-20250929.md`）

**前端TypeScript文件分析**:
- **超大文件**（>800行，需要重构）:
  - `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`: 1,157行 ⚠️
  - `frontend/src/features/temporal/components/InlineNewVersionForm.tsx`: 1,067行 ⚠️
- **大文件**（400-800行，评估拆分价值）:
  - `frontend/src/features/organizations/components/OrganizationTree.tsx`: 586行
  - `frontend/src/shared/hooks/useEnterpriseOrganizations.ts`: 491行
  - `frontend/src/shared/api/unified-client.ts`: 486行
- **中等文件**（300-400行）: 其他多个文件
- **统计数据**: 112个TypeScript文件，平均163.0行（基线参见同一报告）

**影响评估（合理级别）**:
- 超大文件影响代码可读性和团队协作效率
- 单元测试复杂度较高，需要优化
- 代码审查难度增加，特别是超800行的文件
- 部分违反Go社区推荐的文件组织方式

#### 2. 类型安全异味
**问题描述**: TypeScript代码中存在大量弱类型使用
- 检测到171处`any`或`unknown`类型使用（详见 `reports/iig-guardian/code-smell-baseline-20250929.md`）
- 类型守卫使用不够充分
- 缺乏严格的类型检查配置

#### 3. 架构一致性异味
**问题描述**: 部分代码违反项目架构原则
- CQRS分离不够彻底（部分混用场景）
- API命名不一致问题（camelCase vs snake_case）
- 跨层依赖关系复杂

#### 4. 技术债务异味
**问题描述**:
- 控制台日志使用过多（47处console.log相关调用）
- 部分脚本文件包含TODO标记
- 依赖包存在弃用警告

## 改进计划（基于Go工程实践）

### 📊 Go工程实践标准定义

#### 文件大小分级管理
```yaml
绿灯区域 (0-400行): 符合Go社区标准，无需干预
黄灯区域 (400-600行): 关注代码结构，优化组织方式
橙灯区域 (600-800行): 需要架构师评估，建议拆分
红灯区域 (800行以上): 强制重构，违反工程实践标准
```

#### 函数大小标准
```yaml
推荐函数大小: ≤50行（一屏显示）
可接受范围: 50-100行
需要重构: >100行
```

### Phase 0：基线确认与前置检查（1-2天）

| 任务 | 必须执行的命令/输出 | Owner | 验收标准 |
| --- | --- | --- | --- |
| ✅ 基线报告已生成 | `reports/iig-guardian/code-smell-baseline-20250929.md` | 架构组 | Go 54文件/TS 112文件统计完成 |
| 运行实现清单 | `node scripts/generate-implementation-inventory.js` → `reports/implementation-inventory.json` | 架构组 | JSON文件更新，无重复资源 |
| 临时治理巡检 | `bash scripts/check-temporary-tags.sh`（确认无新增 TODO） | 架构组 + QA | 输出无新增临时标记 |
| 工作量复核 | 评估实际可用人力与时间缓冲（建议+20%） | 架构组 + PM | 确认团队30%工作量可投入 |
| Git标签基线 | `git tag plan16-phase0-baseline` | 架构组 | 标签创建成功，用于回滚 |
| 弱类型统计 | `rg "\bany\b|\bunknown\b" frontend/src --stats` → 更新至基线报告 | 前端团队 | 确认171处统计无偏差 |
| 更新进展日志 | 在 `docs/development-plans/06-integrated-teams-progress-log.md` 添加Plan 16进展模板（见下方） | 计划 Owner | 模板填写完整 |

**进展日志模板**（添加至06号文档）:
```markdown
## Plan 16 代码异味治理进展
- **Phase 0 完成时间**: 2025-10-01
- **责任人**: [架构组负责人姓名]
- **基线报告**: `reports/iig-guardian/code-smell-baseline-20250929.md`
- **红灯文件**: Go 3个, TS 2个（需Phase 1清零）
- **下一检查点**: Phase 1.1 main.go拆分完成（预计2025-10-08）
- **风险提示**: 测试工作量需额外30%时间
- **每周五同步**: 更新进展、阻塞项、风险变化
```

> 未完成以上步骤（含报告落盘）不得进入 Phase 1；所有报告需在 PR 说明中引用。

### 🎯 Phase 1：重点文件重构（3周，含测试缓冲）

#### 目标
优先解决红灯区域文件，清零红灯并将前端平均文件行数降至150行以下

#### 预防措施（每次重构前强制执行）
```bash
# 1. 运行完整测试套件
make test && npm --prefix frontend run test

# 2. 运行契约测试
npm --prefix frontend run test:contract

# 3. 创建回滚点
git tag plan16-phase1-task[N]-before

# 4. 重构后立即验证
make test-integration && npm --prefix frontend run test:e2e
```

#### 详细执行计划

**第一优先级 - 红灯区域强制重构**（2周，含测试时间）

1. **Go后端红灯文件重构**

   **A. main.go重构详细计划** (2,264行 → 6-8个文件，目标<400行/文件)
   ```
   第1-2天: 分析现有main.go结构，识别功能模块
   第3-4天: 拆分服务器核心逻辑和配置管理
   第5-6天: 拆分路由定义和中间件
   第7-8天: 拆分数据库管理和健康检查
   第9-11天: 单元测试编写（覆盖率≥80%）
   第12-13天: 集成测试验证和代码优化
   ```
   目标文件结构:
   - `internal/server/server.go` - 服务器核心逻辑 (~350行)
   - `internal/server/config.go` - 配置管理 (~300行)
   - `internal/routes/routes.go` - 路由定义 (~400行)
   - `internal/middleware/middleware.go` - 中间件集成 (~350行)
   - `internal/database/database.go` - 数据库管理 (~300行)
   - `internal/health/health.go` - 健康检查 (~200行)
   - `cmd/organization-query-service/main.go` - 简化主函数 (~264行)

   **B. organization.go处理器重构** (1,399行 → 4个文件)
   ```
   第11-12天: 按业务功能分析和分组
   第13-14天: CRUD操作和时态操作拆分
   第15天: 事件处理和验证逻辑拆分
   ```
   目标文件结构:
   - `handlers/organization/crud.go` - CRUD操作 (~400行)
   - `handlers/organization/temporal.go` - 时态操作 (~400行)
   - `handlers/organization/events.go` - 事件处理 (~300行)
   - `handlers/organization/validation.go` - 验证逻辑 (~299行)

   **C. repository重构** (817行 → 3个文件)
   ```
   第16-17天: 按数据访问模式拆分
   第18天: 测试和优化
   ```
   目标文件结构:
   - `repository/organization/queries.go` - 查询操作 (~300行)
   - `repository/organization/commands.go` - 命令操作 (~300行)
   - `repository/organization/temporal.go` - 时态操作 (~217行)

   **架构边界与依赖控制**
   - 新增 `internal/server/*` 与 `internal/routes/*` 仅暴露于 `cmd/organization-query-service`，禁止被命令服务引用（通过 `go list ./cmd/...` 依赖检查验证）。
   - handler 拆分后模块统一归档在 `cmd/organization-command-service/internal/handlers/organization/`，只依赖所在服务的 `services`/`repository` 层，不直接访问 GraphQL 查询层。
   - repository 拆分文件保持 `repository/organization` 包级别，禁止跨层调用 `handlers`，通过 `golangci-lint` import rules 校验。
   - Phase 1 每个 PR 添加架构依赖图（使用 `go mod graph` 与 `golangci-lint run --config .golangci.yml`）并在评审清单中勾选“CQRS 边界无交叉”项。

2. **前端超大组件重构**
   - **TemporalMasterDetailView.tsx拆分** (1,157行 → 3-4个文件)
     - `components/temporal/TemporalMasterView.tsx` - 主视图 (~400行)
     - `components/temporal/TemporalDetailView.tsx` - 详情视图 (~400行)
     - `hooks/useTemporalMasterDetail.ts` - 业务逻辑 (~357行)

   - **InlineNewVersionForm.tsx重构** (1,067行 → 3个文件)
     - `components/forms/VersionFormCore.tsx` - 核心表单 (~400行)
     - `components/forms/VersionFormLogic.tsx` - 表单逻辑 (~367行)
     - `hooks/useVersionForm.ts` - 表单状态管理 (~300行)

**第二优先级 - 橙灯区域评估优化**（1周）

3. **Go后端橙灯文件优化**
   - `internal/services/temporal.go` (773行 → 2个文件)
     - `services/temporal/service.go` (~400行)
     - `services/temporal/operations.go` (~373行)

   - `internal/repository/temporal_timeline.go` (685行 → 2个文件)
   - `internal/validators/business.go` (596行 → 保持单文件，优化结构)
   - `internal/audit/logger.go` (595行 → 保持单文件，优化函数)
   - `internal/authbff/handler.go` (589行 → 保持单文件，优化组织)

4. **前端大文件合理拆分**
   - `OrganizationTree.tsx` (586行 → 保持单文件，提取子组件)
   - `useEnterpriseOrganizations.ts` (491行 → 保持单文件，优化逻辑)
   - `unified-client.ts` (486行 → 保持单文件，按协议分组)

#### 验收标准（可验证版本）
- **红灯消除**: 所有文件≤800行（通过 `find cmd -name '*.go' -exec wc -l {} +` 验证）
- **目标平均**: Go文件平均≤350行，前端TypeScript平均≤150行
- **函数优化**: 超过100行的函数减少80%（通过 `scripts/code-smell-monitor.sh --functions` 验证）
- **质量保证**: 保持现有功能100%完整性（通过 `make test && make test-integration` 验证）
- **测试覆盖**: 单元测试覆盖率≥80%（通过 `make coverage` 验证）
- **回归测试**: 契约测试通过（通过 `npm run test:contract` 验证）
- **代码审查**: PR review时间降低30%（基于GitHub PR metrics统计）

### 🛡️ Phase 2：类型安全提升（1.5周）

#### 目标
消除弱类型使用，建立严格的类型检查机制

#### 具体行动
1. **any/unknown类型清理**
   - 为171处弱类型使用建立具体类型定义
   - 加强类型守卫使用
   - 配置更严格的TypeScript编译选项

2. **类型系统完善**
   - 建立统一的类型导出索引
   - 完善API响应类型定义
   - 加强运行时类型验证

3. **工具配置优化**
   - 启用strict、exactOptionalPropertyTypes等严格配置
   - 集成类型检查到CI/CD流程
   - 建立类型覆盖率监控

#### 验收标准（平衡版本）
- any/unknown使用减少至30个以下（合理目标）
- TypeScript编译零错误（警告可接受）
- 逐步启用更严格的类型检查配置
- 重要API端点的运行时类型验证覆盖率≥80%

### 🏗️ Phase 3：架构一致性修复（1周）

#### 目标
修复架构违规问题，建立持续质量监控机制

#### 具体行动
1. **CQRS分离强化**
   - 审查并修复混用场景
   - 完善API契约一致性
   - 加强命名规范执行

2. **质量监控系统建立**
   - 集成文件大小自动监控
   - 建立违规文件自动报警
   - 配置代码审查强制规则

## 5. 责任矩阵与里程碑

### 5.1 责任矩阵
| 阶段 | Owner | 支持团队 | 主要交付物 |
| --- | --- | --- | --- |
| Phase 0 | 架构组（A. Chen） | QA（L. Wu） | `code-smell-baseline-<date>.md`、进展日志更新 |
| Phase 1 | 架构组（A. Chen） | 后端团队（B. Yang）、QA（L. Wu） | 重构 PR、Go 测试报告、`code-smell-progress-<date>.md` |
| Phase 2 | 前端团队（C. Zhang） | 架构组、QA | 类型治理报告、`npm run test`/`npm run lint` 结果 |
| Phase 3 | 架构组（A. Chen） | 平台工程（D. Li） | 规模监控脚本、`code-smell-ci-<date>.md`、巡检模板 |

### 5.2 里程碑
| 日期 | 里程碑 | 验证方式 |
| --- | --- | --- |
| 2025-10-07 | Phase 1 完成 | `reports/iig-guardian/code-smell-progress-20251007.md` + Go 测试记录 |
| 2025-10-18 | Phase 2 完成 | TypeScript 类型治理日志 + 前端测试报告 |
| 2025-10-25 | Phase 3 完成 | 监控脚本 PR + `code-smell-ci-20251025.md` |

## 6. 风险评估与缓解

### 高风险项
1. **激进重构引发关键 API 异常**
   - **触发条件**: 重构后出现 5xx 或 E2E 关键用例失败
   - **预防措施**:
     - 每次重构前运行 `make test` + `npm run test:contract`
     - 重构后立即运行集成测试 `make test-integration`
     - 保持 git 标签 `plan16-phaseX-taskY-before` 用于快速回滚
   - **响应对策**: 在 4 小时内回滚至最近标签，补充测试并复盘；回滚记录在 `reports/iig-guardian/code-smell-rollback-<date>.md`

2. **开发效率显著下降**
   - **触发条件**: 周会中报告的阻塞 >2 次/周
   - **对策**: 缩减当期目标，仅处理红灯文件，其余调度至后续迭代；同步 `docs/development-plans/06`。

3. **前端类型改动引发编译失败**
   - **触发条件**: Phase 2 类型治理导致前端大面积编译错误
   - **预防措施**: 分批次改动（每批≤20个文件），每批后运行 `npm run build`
   - **响应对策**: 回滚当批改动，重新设计渐进式类型迁移路径

### 中等风险项
1. **团队学习成本**
   - **对策**: 组织 Workshop，提供拆分模板示例，更新 Review checklist。

2. **监控脚本误报**
   - **对策**: 在 `code-smell-ci-<date>.md` 中记录沙盒结果，维护白名单。

## 7. 成功指标（可验证版本）

### 量化指标（基于基线 `reports/iig-guardian/code-smell-baseline-20250929.md`）
- **平均文件行数**: Go ≤ 350行（当前312.7），前端平均≤150行（当前163.0）
- **红灯区域文件**: 0个（当前Go 3个, TS 2个）
- **橙灯区域文件**: ≤5个（当前Go 5个, TS 9个，目标Go≤3, TS≤5）
- any/unknown类型使用 ≤ 30处（当前171处，Phase 2目标）
- 单元测试覆盖率 ≥ 80%（通过 `make coverage` 验证）
- 契约测试通过率 100%（通过 `npm run test:contract` 验证）
- 函数超100行数量减少80%（通过监控脚本验证）

### 过程指标（可度量）
- **PR review时间**: 降低30%（基于GitHub PR平均review时间统计，Phase完成前后对比）
- **单元测试编写速度**: 提升（基于function:test行数比，目标1:0.8以上）
- **Bug修复周期**: 降低（基于Jira/GitHub Issue关闭时间统计）
- **团队反馈**: 通过问卷调查（Phase 3结束后执行）

## 8. 资源需求（平衡版本）

### 人力资源
- **技术架构师**: 1人（全程指导）
- **后端开发工程师**: 2人（Phase 1 & Phase 3）
- **前端开发工程师**: 2人（Phase 1 & Phase 2）
- **质量保证工程师**: 1人（全程）

### 时间分配（含20%缓冲）
- **Phase 0**: 2天（基线确认+工具准备）
- **Phase 1**: 3周（红灯文件重构+测试，50%工作量）
- **Phase 2**: 1.5周（类型治理，30%工作量）
- **Phase 3**: 1周（监控系统，20%工作量）
- **总计**: 5.5-6周（原计划4-5周+20%缓冲）

## 后续维护计划

### 监控机制（实用版）
1. **文件大小监控**（每周执行 `find ... wc -l`，记录在 `code-smell-weekly-<date>.md`）
2. **质量报告**（月度汇总至 `reports/iig-guardian/`）
3. **架构合规性检查**（PR 时触发 `node scripts/generate-implementation-inventory.js` + `scripts/check-temporary-tags.sh`）
4. **类型安全度量**（季度评估 `rg any|unknown --stats` 报告）

### 预防措施
1. **合理的质量门禁**（平衡开发效率）
2. **团队培训与最佳实践分享**
3. **代码规范文档化**
4. **定期架构评审**

## 9. 质量标准与验收清单

### 建立Go工程实践标准
```markdown
项目长期文件大小标准：
- 推荐范围: 200-400行
- 可接受范围: 400-600行
- 需要评估: 600-800行
- 强制重构: >800行

函数复杂度标准：
- 推荐: ≤30行
- 可接受: 30-50行
- 需要重构: >100行
```

### 验收清单
- [ ] ✅ `reports/iig-guardian/code-smell-baseline-20250929.md` 已生成（Phase 0完成）
- [ ] `reports/iig-guardian/code-smell-progress-<date>.md` 记录红/橙灯文件状态并附测试结果
- [ ] `reports/iig-guardian/code-smell-types-<date>.md` 显示 any/unknown 统计（≤30）
- [ ] `reports/iig-guardian/code-smell-progress-<date>.md` 中记录 TypeScript 平均行数（≤150）
- [ ] `reports/iig-guardian/code-smell-ci-<date>.md` 证明文件规模脚本已纳入 CI
- [ ] `docs/development-plans/06-integrated-teams-progress-log.md` 每周五更新进展、风险、阻塞项
- [ ] `node scripts/generate-implementation-inventory.js` 复跑并更新实现清单，确保导出一致
- [ ] `scripts/code-smell-monitor.sh` 脚本创建并验证（Phase 3交付）
- [ ] git标签体系建立：`plan16-phase[0-3]-baseline` 和 `plan16-phase[1-3]-task[N]-before`

## 10. 相关文档

- [项目架构指导原则](../../CLAUDE.md)
- [实现清单护卫系统](../reference/02-IMPLEMENTATION-INVENTORY.md)
- [API契约文档](../api/openapi.yaml)
- [开发者快速参考](../reference/01-DEVELOPER-QUICK-REFERENCE.md)

## 批准与执行

**计划状态**: 待批准
**预期开始日期**: 2025-10-01
**预期完成日期**: 2025-11-08（原2025-10-29 + 1.5周缓冲）
**基线数据**: `reports/iig-guardian/code-smell-baseline-20250929.md`（唯一事实来源）

### 批准前置条件
在技术团队评审前，请确认以下内容已准备就绪：
- [x] 基线报告已生成并验证（`reports/iig-guardian/code-smell-baseline-20250929.md`）
- [x] 监控脚本已创建并测试（`scripts/code-smell-monitor.sh` + `scripts/code-smell-check-quick.sh`）
- [x] 进展跟踪模板已添加至06号日志（`docs/development-plans/06-integrated-teams-progress-log.md` 第79-122行）
- [ ] Phase 0剩余任务执行计划已确认（见下方"批准后立即行动"）

**批准人签名区域**:
- [ ] 技术架构负责人（评审重点：技术方案可行性、风险评估充分性）
- [ ] 项目经理（评审重点：时间表合理性、资源分配可行性）
- [ ] 质量保证负责人（评审重点：验收标准完整性、测试策略充分性）

### 批准后立即行动（Phase 0剩余任务，1-2天内完成）

#### Phase 0任务清单（更新版）
1. ✅ 基线报告已生成（`reports/iig-guardian/code-smell-baseline-20250929.md`，已完成）
2. ☐ 运行实现清单（30分钟）
3. ☐ 临时治理巡检（15分钟）
4. ☐ 创建 `golangci-lint` 配置（15分钟，新增）
5. ☐ 工作量复核会议（1小时）
6. ☐ Git 标签基线（5分钟）
7. ✅ 弱类型统计复核（171处，已完成）
8. ☐ 更新进展日志（10分钟）

#### 执行建议
```bash
# 批准当天（快速通道 ~45分钟）
node scripts/generate-implementation-inventory.js      # 30分钟
bash scripts/check-temporary-tags.sh                   # 5分钟
cat <<'EOF' > .golangci.yml                            # 见下方配置说明
run:
  timeout: 5m

linters:
  enable:
    - depguard
    - importas
    - revive

linters-settings:
  depguard:
    list-type: blacklist
    packages:
      - cube-castle-deployment-test/cmd/organization-query-service/internal
    packages-with-error-message:
      cube-castle-deployment-test/cmd/organization-query-service/internal: "命令服务不得依赖查询服务实现"
    ignore-tests: true
# TODO: 根据实际包路径补充/调整限制条目
EOF

# 批准后1天内
- 工作量复核会议（1小时）
- git tag plan16-phase0-baseline && git push origin plan16-phase0-baseline
- 更新 docs/development-plans/06-integrated-teams-progress-log.md
```

> `.golangci.yml` 示例仅提供最小依赖守卫，请在提交前根据实际包路径增补其他 CQRS 规则（例如禁止查询服务引用命令层实现）。

#### 任务清单（按优先级排序）
| 序号 | 任务 | 负责人 | 执行命令 | 验收标准 | 预计时间 |
|------|------|--------|---------|---------|---------|
| 1 | 运行实现清单 | 架构组 | `node scripts/generate-implementation-inventory.js` | JSON文件更新，无重复资源 | 30分钟 |
| 2 | 临时治理巡检 | 架构组+QA | `bash scripts/check-temporary-tags.sh` | 输出无新增临时标记 | 15分钟 |
| 3 | 创建 golangci-lint 配置 | 架构组 | `cat <<'EOF' > .golangci.yml`（示例见下） | `.golangci.yml` 已提交且含 CQRS import 限制 | 15分钟 |
| 4 | 工作量复核 | 架构组+PM | 团队会议确认30%工作量可投入 | 会议纪要 | 1小时 |
| 5 | Git标签基线 | 架构组 | `git tag plan16-phase0-baseline && git push origin plan16-phase0-baseline` | 标签创建成功 | 5分钟 |
| 6 | 更新进展日志 | 计划Owner | 在06号日志填写Phase 0完成时间与责任人 | 模板填写完整 | 10分钟 |

#### 验收产物
- [ ] `reports/implementation-inventory.json` 已更新
- [ ] `scripts/check-temporary-tags.sh` 执行结果截图（无告警）
- [ ] `.golangci.yml` 已创建并纳入版本控制（含 CQRS import 校验规则）
- [ ] 工作量复核会议纪要（确认团队承诺）
- [ ] Git标签 `plan16-phase0-baseline` 已推送至远程仓库
- [ ] 弱类型统计结果已记录（与基线报告一致）
- [ ] 06号日志 "Plan 16代码异味治理进展跟踪" 章节已填写Phase 0完成信息

### Phase 1启动前检查（Phase 0完成后执行）
```bash
# 一键检查脚本（Phase 0验收）
echo "=== Phase 0 验收检查 ==="
echo ""

echo "1. 检查基线报告..."
test -f reports/iig-guardian/code-smell-baseline-20250929.md && echo "✅ 基线报告存在" || echo "❌ 基线报告缺失"

echo ""
echo "2. 检查实现清单..."
test -f reports/implementation-inventory.json && echo "✅ 实现清单已更新" || echo "❌ 实现清单缺失"

echo ""
echo "3. 检查Git标签..."
git tag | grep -q "plan16-phase0-baseline" && echo "✅ 基线标签已创建" || echo "❌ 基线标签缺失"

echo ""
echo "4. 检查 golangci-lint 配置..."
test -f .golangci.yml && echo "✅ golangci 配置存在" || echo "❌ golangci 配置缺失"

echo ""
echo "5. 检查监控脚本..."
test -x scripts/code-smell-check-quick.sh && echo "✅ 监控脚本可执行" || echo "❌ 监控脚本不可执行"

echo ""
echo "6. 检查进展日志..."
grep -q "Phase 0 完成时间" docs/development-plans/06-integrated-teams-progress-log.md && echo "✅ 进展日志已更新" || echo "❌ 进展日志未更新"

echo ""
echo "=== 如所有检查通过，可启动Phase 1 ==="
```

### 首次周五同步提醒（2025-10-04）
在首次周五同步会议上，请更新以下内容至06号日志：
- 填写进展表格第一行（W1完成任务、红灯文件数、阻塞项、风险变化）
- 确认Phase 1任务分配（main.go拆分负责人、预计完成日期）
- 识别并记录任何新风险或阻塞项

### 评审支持材料
- **基线数据可视化**: 运行 `bash scripts/code-smell-check-quick.sh` 查看当前红灯文件
- **历史对比**: 参考 `docs/development-plans/06-integrated-teams-progress-log.md` 了解过往Plan执行情况
- **架构原则**: 参考 `CLAUDE.md` 第2节核心开发指导原则
- **开发规范**: 参考 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 了解团队工作流

---

*本计划遵循Go语言工程实践的行业标准，致力于建立平衡质量与实用性的高质量代码库。通过分级管理策略，既保证了代码质量提升，又考虑了团队开发效率和适应性。*

*评审完成后，请在批准人签名区域勾选并记录批准日期，然后执行"批准后立即行动"清单启动Phase 0。*
