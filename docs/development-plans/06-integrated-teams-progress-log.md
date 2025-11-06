# 06号文档：219C2W Job Catalog/Position 校验修复后续动作

## 目标
确保命令服务加载最新校验逻辑，并完成 219C2W 自测脚本及报告的闭环更新。

## 待办清单
1. **重建并重启命令服务**  
   - 执行 `make docker-build`（如近期未重建镜像）。  
   - 执行 `make run-dev`（或等效 `docker compose up`）重启 REST/GraphQL 服务，确认容器日志显示最新二进制已加载。
2. **重新执行 Day24 自测脚本**  
   - 运行 `bash scripts/219C2D-validator-self-test.sh`。  
   - 验证脚本输出中 `jobCatalog.createVersion`、`jobFamily.createVersion`、`position.fill`、`assignment.close` 等场景均返回期望的 `errorCode` / `ruleId` / `severity`。  
   - 将最新结果追加至 `logs/219C2/validation.log`，并覆盖生成 `tests/e2e/organization-validator/report-Day24.json`。
3. **同步唯一事实来源**  
   - 自测通过后，更新 `internal/organization/README.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/development-plans/219C2D-extension-acceptance.md` 勾选项，记录验证完成的证据路径。
4. **记录进度与证据**  
   - 在 `docs/development-plans/219C2W-validation-error-reconciliation.md` 追加完成时间与日志引用。  
   - 如仍遇 `INTERNAL_ERROR`，需导出容器日志与请求 ID，并在 06 文档补充阻塞说明及处理人。

## 负责人
- 命令服务值班工程师：待认领
- 文档同步：219C2W 计划 Owner

## 截止时间
- 建议在 2025-11-06 18:00 前完成，以便 Day24 归档。

## 完成情况（2025-11-06 21:56）
- ✅ `docker compose -f docker-compose.dev.yml up -d --build --force-recreate rest-service graphql-service` 已执行，命令/查询服务运行最新二进制。
- ✅ `scripts/219C2D-validator-self-test.sh` 自测通过，`logs/219C2/validation.log` 与 `tests/e2e/organization-validator/report-Day24.json` 已更新。
- ✅ 文档同步：`docs/development-plans/219C2W-validation-error-reconciliation.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`internal/organization/README.md` 均记录最新证据路径，219C2W 可关闭。
