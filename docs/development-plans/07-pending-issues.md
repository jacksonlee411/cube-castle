# 07 — 组织层级同步修复记录

最后更新：2025-09-21  
责任团队：架构组（主责）+ 测试组
当前状态：修复已完成，等待复测

---

## 1. 已完成的技术调整
- 新增 `recalculateSelfHierarchy`，在 `Update` / `UpdateByRecordId` 内同步写入 `level`、`path`、`code_path`、`name_path`，日志示例：`recalculateSelfHierarchy: code=..., oldLevel=..., newLevel=..., path=...`。
- `UpdateByRecordId` 及常规更新均空值保护 `name`、`parentCode`，并触发上述重算逻辑。
- 循环引用在 Handler 层前置拦截，输出 `⚠️ circular reference attempt` 日志并直接返回 400。
- `refreshHierarchyPaths` 精简为只刷新子树，当前节点由重算逻辑负责写回。

代码参考：
- `cmd/organization-command-service/internal/repository/organization.go`
- `cmd/organization-command-service/internal/handlers/organization.go`

---

## 2. 待测试项目
1. **父级切换后数据同步**  
   - 接口：`PUT /organization-units/{code}/history/{record_id}`  
   - 期望：响应及数据库中 `level`=新层级、`path`/`code_path`/`name_path` 正确写回；查看 `recalculateSelfHierarchy` 日志确认。
2. **循环引用防护**  
   - 请求 `parentCode = self` 或子孙节点  
   - 期望：HTTP 400 + 错误码 `BUSINESS_RULE_VIOLATION`，日志出现 `circular reference attempt`；库中无残留更新。
3. **前端展示一致性**  
   - 页面：`/organizations/{code}/temporal`  
   - 期望：层级显示与数据库字段一致；路径/面包屑正常。
4. **数据一致性校验（可选）**  
   - SQL 或 `TemporalService.RecomputeTimelineForCode`  
   - 期望：抽样组织（例如 1000009）`level=2`、`code_path=/1000000/1000009`，无异常记录。

---

## 3. 验收标准
- 组织 `1000009` 从 `1000056` 迁至 `1000000` 后：
  - `level` 由 3 降至 2；
  - `code_path`/`name_path` 更新为 `/1000000/1000009` 对应值；
  - 前端层级显示与数据库一致；
  - 循环引用请求返回 400，数据库无变更。
- 技术验证：
  - `recalculateSelfHierarchy` 调用日志在复测中可见；
  - 循环引用校验覆盖直接、间接、自引用场景；
  - 级联刷新在目标规模（可先以现有数据验证）内表现正常。

---

复测完成后请将结果（含日志、请求/响应、数据库快照）同步回本页或对应 Issue，以便归档。
