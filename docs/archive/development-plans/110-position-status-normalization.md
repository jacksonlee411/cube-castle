# 110 号方案：职位版本状态与“当前版本”标识异常整改

**版本**：v1.0（执行中）  
**创建日期**：2025-10-22  
**责任团队**：职位领域后端组（主责） · 职位前端组（协同）

---

## 1. 背景与触发
- 业务场景：`职位详情：P1000000` 在版本列表与右侧信息卡均显示 `状态：PLANNED`，并且即使切换历史版本仍显示“当前版本”。
- 期望行为：
  - 实际生效的版本应标记为 `ACTIVE`，已结束的历史版本不应呈现为 `PLANNED`。
  - 页面上的提示语应准确区分“当前版本 / 历史版本 / 计划版本”。
- 影响范围：所有职位时态版本展示（REST 数据 → GraphQL 查询 → 前端 UI），用户无法快速辨别真实状态。

---

## 2. 复现步骤
1. 启动开发环境 `make run-dev`，访问前端职位详情页 `P1000000`。
2. 展开“版本历史”列表，可见三条记录均展示 `状态：PLANNED`。
3. 切换不同版本，右侧信息卡仍提示 `当前版本：…（状态：PLANNED）`。
4. 使用 GraphQL 检查：
   ```bash
   curl -s -X POST http://localhost:8090/graphql \
     -H "Authorization: Bearer $(cat .cache/dev.jwt)" \
     -H "X-Tenant-ID: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9" \
     -d '{"query":"query($code: PositionCode!, $includeDeleted: Boolean!){ positionVersions(code:$code, includeDeleted:$includeDeleted){ recordId status isCurrent effectiveDate endDate }}","variables":{"code":"P1000000","includeDeleted":false}}'
   ```
   - 修复前：所有版本 `status = PLANNED`。

---

## 3. 根因分析
| 层级 | 事实来源 | 结论 |
|------|----------|------|
| 后端命令服务 | `PositionService.buildPositionEntity` | 创建/新增版本默认 `status = PLANNED`，后续未自动转换。 |
| 数据库 | `positions` 表 | 历史记录整齐存在 `end_date`，`is_current` 仅当前版本为 true，但 `status` 始终保留 `PLANNED`。 |
| 查询服务 | `postgres_positions.go#scanPosition` | 直接回传数据库状态，未根据生效日期 / 当前标记进行二次归一。 |
| 前端 UI | `PositionTemporalPage` / `VersionList` | 文案固定写为“当前版本”，且直接展示后台原始状态字符串。 |

根因：缺少“状态派生”逻辑——当版本已生效（effective_date ≤ 今日）却仍处于默认 `PLANNED` 状态，查询服务返回后被 UI 原样展示，造成所有版本看似为“规划中”。

---

## 4. 整改方案与实施
| 编号 | 动作 | 说明 | 状态 |
|------|------|------|------|
| A | 查询服务状态归一化 | `scanPosition` 引入 `normalizePositionStatus`：对于 `PLANNED` 状态，若已生效，则派生为 `ACTIVE`（当前版本）或 `INACTIVE`（历史版本）；未来版本保留 `PLANNED`。 | ✅ 完成 |
| B | 前端提示语调整 | `PositionTemporalPage` 根据 `isCurrent/isFuture` 输出“当前版本 / 历史版本 / 计划版本”，并使用 `statusMeta` 显示本地化标签。 | ✅ 完成 |
| C | 回归验证脚本 | GraphQL 查询与前端手动验证同步执行，确认状态和提示语符合预期。 | ✅ 完成 |

---

## 5. 验证结果
- GraphQL 查询现返回：当前版本 `status=ACTIVE`，历史版本 `status=INACTIVE`。
- 前端 UI：
  - 当前版本 banner：`当前版本：2025-10-04（状态：在编）`。
  - 历史版本 banner：`历史版本：2025-10-01（状态：停用）` 等。
  - 版本列表标签根据状态颜色正确展示。
- 数据库未做直接变更，仅在查询层派生展示值，避免破坏既有审计。

---

## 6. 后续关注
- 若未来引入更多业务状态（如 `VACANT` / `SUSPENDED`），需确认归一逻辑是否继续适用。
- 建议评估命令服务是否在版本生效时同步写入业务状态，以减少查询层规则。
- 观察使用反馈，如仍有歧义可在 UI 增加“数据更新至 xx 日”提示。

---

**记录人**：后端查询服务组 · 王小松  
**最后更新**：2025-10-22 04:12 UTC
