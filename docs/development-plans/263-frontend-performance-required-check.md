# Plan 263 - 前端性能影响分析门禁（构建/TS 报错全绿计划）

**文档编号**: 263  
**标题**: 前端性能影响分析门禁（契约测试构建校验）  
**版本**: v0.1  
**创建日期**: 2025-11-18  
**关联计划**: Plan 244（时间线命名统一）、Plan 257（Facade Coverage）、Plan 258（Contract Drift Gate）、Plan 215（Phase2 日志）

---

## 1. 背景与目标

- 在 `契约测试自动化验证` 工作流中，“性能影响分析” job 会运行 `npm run build:verify`，用于验证 TypeScript 构建与前端契约测试的性能/质量基线。
- 当前该 job 未列入 Required checks，但多处 TS 类型错误导致 job 失败（详见 2025-11-18 CI 日志），阻碍我们将其设为硬门禁。
- 本计划目标：修复性能分析 job 中的所有构建/类型报错，确保 `npm run build:verify` 在 CI 环境稳定通过，并制定将该 job 列为 Required 的时间表／回滚策略。

## 2. 当前已知问题（来源：Run #19450979835 -> job “性能影响分析”）

1. **PositionDetailView tabs**：`temporalEntitySelectors.position.tabId` 可能为 undefined（TS2722）。
2. **PositionTransferDialog / TemporalEntityPage Loader tests**：重复导入 `vi`、未使用 React（TS6133/TS2300）。
3. **Temporal master detail hooks**：缺少 `unifiedGraphQLClient`/`unifiedRESTClient` 引用、compare 函数 `a/b` 未声明类型（TS2304/TS7006）。
4. **StatusBadge**：`TemporalEntityStatusMeta` 中访问了不存在的 `backgroundColor`/`borderColor`（TS2551/TS2339）。
5. **useTemporalEntityDetail**：引用 `PositionTimelineEvent` 未导入，导致 `TemporalDetailResult` 类型错误（TS2304）。
6. **temporalEntitySelectors**：position namespace 缺少 `table` 字段，组织 namespace 不能缺席；导致类型校验失败（TS2741）。
7. **useEnterprisePositions/useTemporalEntityDetail**：查询结果类型与 `UseQueryResult` 泛型不一致（TS2322/TS2769）。
8. **其他**：`invalidation.ts` 里 `cacheKey ?? null` 与预期类型不符、`AppErrorBoundary` 缺少 `override` 修饰符等。

(注：上述列表来自 2025-11-18 最新 CI 日志，后续可能根据修复进展更新。)

## 3. 内容与交付物

1. **类型修复清单**：逐项修复“性能影响分析” job 中的 TypeScript 报错，提交落实。
   - 目标：本计划完成后，`npm run build:verify` 在 CI/本地均无 error/warning。
2. **构建稳定性**：记录并固化 `build:verify` 所执行的命令、性能指标（如构建耗时、bundle 大小、初步性能基线）。
3. **Required Check 切换**：在 `契约测试自动化验证` workflow 中，将“性能影响分析” job 设为 Required，确保 PR 合并必须通过该构建。
4. **文档与回滚策略**：
   - 更新 `docs/development-plans/263-frontend-performance-required-check.md`（本文）记录修复与验证过程。
   - 在 `docs/reference/temporal-entity-experience-guide.md` 或相关文档中补充 `build:verify` 使用说明。
   - 提供将 job 从 Required 改回可选的回滚步骤（例如：复原 workflow 配置、说明风险）。

## 4. 验收标准

- [ ] `npm run build:verify` 在 CI 和本地均 zero-error/zero-warning（TS、Vite 构建皆通过）。
- [ ] `契约测试自动化验证` workflow 中 “性能影响分析” job 被设为 Required check，并至少连续 3 次 PR 运行成功。
- [ ] 修复日志（GitHub Actions run ID、TS 错误列表、修复 commit）记录在计划文档中。
- [ ] 如需回滚，将 Required check 恢复为 optional 的步骤已记录，并有明确触发条件。

## 5. 风险与缓解

| 风险 | 描述 | 缓解措施 |
|------|------|---------|
| 类型修复范围较大 | 涉及 Temporal hooks、StatusBadge、PositionDetail 等多个模块 | 逐步提交，每次修复限定在单一场景；保持完整的 TS/测试运行 |
| 新增 Required check 可能阻塞 PR | 若构建偶发失败会阻碍其他人合并 | 在设为 Required 前确保连续多次绿灯，并定义回滚流程 |
| 构建时间增加 | `build:verify` 耗时可能上升，影响 CI 周期 | 可与 DevOps 协同，必要时拆分 job 或缓存 node_modules |

## 6. 时间计划（建议）

- **Week 1**：修复所有现存 TS 报错（拆分多个 MR/Commit），确保 `npm run build:verify` 本地通过。
- **Week 2**：在 PR 上连续验证 `build:verify`，监控 CI 稳定性；完成文档更新。
- **Week 3**：切换 “性能影响分析” job 为 Required；观察 2-3 个 PR 运行情况，收集反馈。

## 7. 参考资源

- `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`
- `frontend/src/shared/testids/temporalEntity.ts`
- `frontend/src/features/positions/PositionDetailView.tsx`
- GitHub Actions run: https://github.com/jacksonlee411/cube-castle/actions/runs/19450979835

---

**更新记录**  
- 2025-11-18：文档创建，记录性能影响分析 job 的 TS 报错及修复计划。 (BY: Codex)
