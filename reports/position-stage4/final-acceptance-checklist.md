# Stage 4 + 87 迁移联合验收清单（草案）

- [ ] **数据库**：047 生产迁移执行（含备份、执行日志、回滚预案确认）
- [ ] **数据库**：迁移后数据完整性校验（无 NULL effective_date / 约束生效）
- [ ] **命令服务**：REST 任职 API 冒烟（Fill/Vacate/Assignments）
- [ ] **命令服务**：跨租户 REST 脚本（tests/consolidated/position-assignments-cross-tenant.sh）
- [ ] **命令服务**：代理自动恢复任务手动触发 + 审计日志校验
- [ ] **查询服务**：GraphQL 任职过滤/时间轴（包含跨租户脚本）
- [ ] **前端**：Position Tabbed Experience 验收脚本（含 CSV 导出）
- [ ] **前端**：缓存清理验证（版本号 bump / queryClient.invalidateQueries）
- [ ] **QA**：Playwright 场景（代理创建→到期→恢复→时间轴）
- [ ] **运维**：监控与告警面板检查（30 分钟内无 ERROR）
- [ ] **文档**：06 号日志更新 + 86/87 计划归档 + 80 号方案勾选
- [ ] **外部通知**：Breaking Change 通知发送记录（T-3）

> 草案版本 2025-10-21，待各责任团队补充实际执行结果与验收证据。
