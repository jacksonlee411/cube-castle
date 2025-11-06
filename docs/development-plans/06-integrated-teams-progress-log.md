# Plan 06 – 集成测试要求（219D2 调度配置集中化）

> 唯一事实来源：`docs/development-plans/219D2-scheduler-config.md`；依赖 219D1 目录迁移与 219A 目录结构约束。

## 1. 测试范围与责任
- **后端 / Scheduler 团队**：执行并记录命令服务端的单元测试、启动验证与失败演练；维护 `logs/219D2/` 下的日志与示例。
- **DevOps**：通过 Docker Compose (`make docker-up`) 提供 Temporal、PostgreSQL、Redis 等依赖服务，确保环境变量与 219D2 配置保持一致。
- **计划 Owner**：汇总测试结论、回滚记录并同步更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 与 `internal/organization/README.md#scheduler`。

## 2. 前置条件
- 219D1 输出（目录迁移、依赖注入）已合入，命令服务能够加载新的 Scheduler Facade。
- `.env` 与 `docker-compose*.yml` 已按 219D2 要求补齐 `SCHEDULER_` 前缀配置，并与 `.env.example` 默认值一致。
- 测试环境严格使用 Docker（`make docker-up`），若端口被宿主服务占用必须先卸载宿主服务，禁止调整容器端口映射。
- 契约文档 (`docs/api/*`) 与计划文件保持一致，避免在测试中引用已废弃的配置键名。

## 3. 强制测试与校验
1. **配置包单元测试** — `go test ./internal/config/...`
   - 覆盖 `SchedulerConfig` 默认值、环境变量覆盖、非法配置报错与 `config.ValidateSchedulerConfig` 检查。
   - 测试输出需写入 `logs/219D2/config-validation.log`（追加模式），作为验收附件。
2. **命令服务启动自检** — `make run-dev`
   - 验证 `SCHEDULER_ENABLED=true` 下命令服务可读取集中化配置并成功启动。
   - 在 `logs/219D2/` 存档启动日志，标注配置来源与注册的任务列表。
3. **配置覆盖验证** — `make run-dev SCHEDULER_ENABLED=false`
   - 确认通过环境变量覆盖可关闭 Scheduler，服务启动过程无 panic、无多余警告。
   - 在 `logs/219D2/` 记录对应日志（文件名自定义），说明关闭行为与后续恢复步骤。
4. **失败与回滚演练**
   - 人为引入无效配置（如清空 `SCHEDULER_TASK_QUEUE` 或提供非法 Cron），再次执行 `make run-dev`。
   - 预期：启动被 `config.ValidateSchedulerConfig` 阻断，错误细节写入 `logs/219D2/config-validation.log`。
   - 恢复默认配置并重新启动，确认系统回到健康状态，同时在日志中附注回滚命令。

## 4. 输出归档要求
- `logs/219D2/` 需保留：单元测试输出、成功启动日志、关闭/失败演练日志与回滚说明。
- 更新后的配置清单（参数名称、默认值、覆盖层级）随 PR 或变更记录提交，并在 `internal/organization/README.md#scheduler` 引用。
- 若测试过程中产生额外脚本或工具，需在 PR 中声明来源并确保与 219D2 文档保持一致。

## 5. 验收记录模板

### 验收记录（2025-11-06）✅ **已完成**

- ✅ go test ./internal/config/... **PASS** （logs/219D2/config-validation.log）
- ✅ make run-dev **PASS** （日志：logs/219D2/startup-success.log）
- ✅ make run-dev SCHEDULER_ENABLED=false **PASS** （日志：logs/219D2/startup-disabled.log）
- ✅ 失败演练与回滚 **PASS** （日志：logs/219D2/failure-test.log）
- ✅ 剩余风险：**无** （总体风险等级：🟢 低）

**完整验收报告：** `logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`
**执行摘要：** `logs/219D2/TEST-SUMMARY.txt`
**文档同步建议：** `logs/219D2/DOCUMENTATION-UPDATE-SUGGESTION.md`

---

## 6. 验收摘要

| 测试项 | 状态 | 日志位置 | 说明 |
|--------|------|--------|------|
| 配置单元测试 | ✅ | config-validation.log | 3/3 测试用例通过 |
| 启动自检 | ✅ | startup-success.log | 全服务健康（REST、GraphQL、Temporal、DB） |
| 配置覆盖验证 | ✅ | startup-disabled.log | 无 panic，恢复正常 |
| 失败演练 | ✅ | failure-test.log | 配置验证生效，恢复有效 |
| **总体通过率** | **100%** | logs/219D2/ | 4/4 强制测试通过 |

---

## 7. 后续行动

- [x] 执行全量测试并记录日志
- [ ] 同步更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`（参考 DOCUMENTATION-UPDATE-SUGGESTION.md）
- [ ] 同步更新 `internal/organization/README.md#scheduler`
- [ ] 更新 `CHANGELOG.md` 记录变更
- [ ] 若所有文档同步完毕，标记计划为 **COMPLETED**

---

如需调整上述要求，必须先更新 `docs/development-plans/219D2-scheduler-config.md` 并同步通知相关责任人。
