# 36号文档：组织命令处理器质量分析

## 背景与单一事实来源
- 本文聚焦命令服务组织业务相关的五个处理器文件：
  - `cmd/organization-command-service/internal/handlers/organization_routes.go`
  - `cmd/organization-command-service/internal/handlers/organization_create.go`
  - `cmd/organization-command-service/internal/handlers/organization_update.go`
  - `cmd/organization-command-service/internal/handlers/organization_events.go`
  - `cmd/organization-command-service/internal/handlers/organization_history.go`
- 分析仅依据上述源码，不引入额外事实来源；与 `internal/middleware`、`internal/repository`、`internal/audit` 的调用契约均以当前实现为准，确保跨层一致性。

## 现状问题
1. **响应结构松散**：多个端点（版本创建、状态变更、事件处理、历史更新）以 `map[string]interface{}` 组装响应，字段类型与命名缺乏编译期保障，未来拓展或字段重构时极易产生运行时异常。
2. **错误判定依赖字符串匹配**：`CreateOrganizationVersion`、`changeOrganizationStatusWithTimeline` 等通过 `strings.Contains(err.Error())` 区分业务错误（如父级缺失、时态冲突），对底层实现高度耦合且脆弱，一旦仓储返回信息调整即失效。
3. **OperationReason 指针语义错误**：`CreateOrganizationVersion` 无条件返回 `&req.OperationReason`，即便调用方未填写也会持久化空字符串，破坏“未提供理由”的语义，并与 `CreateOrganization` 的处理方式不一致。
4. **时间轴返回缺乏防御式编程**：状态变更后直接对 `*timeline` 取长度和索引，若时态管理器因内部条件返回 `nil`（例如跳过更新）会触发 panic；同样问题存在于事件处理返回的时间线构造中。
5. **输入校验分散**：部分端点使用 `len(code) != 7` 做快速校验，未复用 `utils.ValidateOrganizationCode`，与其它路径的规则存在漂移风险；历史记录更新对 `ParentCode` 的自引用校验与主更新逻辑重复，缺乏统一入口。

## 改进建议
1. **引入结构化 DTO**：为各类响应定义显式结构体（含时间轴条目、审计元数据），统一通过共享写入函数输出，确保字段名称/类型一致并便于 IDE 与测试覆盖。
2. **采用强类型错误**：在仓储与服务层定义具备哨兵值的错误类型（如 `ErrParentNotFound`、`ErrTemporalConflict`），在 handler 中使用 `errors.Is` 或 `errors.As` 判定，替换当前字符串匹配逻辑。
3. **修正操作原因处理**：仅当 `strings.TrimSpace(req.OperationReason)` 非空时返回指针，否则置为 `nil`；同时与审计日志写入保持一致，避免产生空字符串记录。
4. **增强时间轴守护**：在状态变更、事件处理路径上对 `timeline` 为空或 `nil` 的情况提前处理，必要时回退到 `repo.GetByCode` 获取最新版本，保证响应与日志不会触发 panic。
5. **统一校验入口**：将组织代码、父级校验、循环引用检测等通用逻辑收敛到共享辅助函数，减少在多个 handler 中复制粘贴的判定分支，降低规则漂移风险。

## 验收标准
- [ ] 所有组织命令端点返回结构化 DTO，并通过单元测试验证 JSON 序列化字段与契约一致。
- [ ] 业务错误通过显式错误类型判定（`errors.Is/As`），移除字符串包含判断；相关路径新增测试覆盖常见错误分支。
- [ ] 时态操作与事件处理在 `timeline == nil` 或空列表时稳健返回，并记录可观测日志，保证不会 panic。
- [ ] `CreateOrganizationVersion` 操作原因仅在有效值时持久化，审计日志与数据库记录保持一致；其他 handler 也复用统一逻辑。
- [ ] 组织代码、父子关系等验证逻辑集中到复用函数，并由测试确认新旧入口行为一致。

## 一致性校验说明
- 以上结论与建议严格来源于五个目标文件的当前实现；执行改进前需对照 `docs/api/openapi.yaml` 与审计事件契约，确保数据结构与字段命名保持一致。
- 本分析文件保存于 `docs/archive/development-plans/`，持续维护唯一事实来源与跨层一致性。
