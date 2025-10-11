# 50号文档：组织命令处理器质量回顾补充

## 背景与单一事实来源
- 聚焦命令服务下列处理器源码：`devtools.go`、`operational.go`、`organization_create.go`、`organization_events.go`、`organization_history.go`、`organization_routes.go`、`organization_update.go`，均位于 `cmd/organization-command-service/internal/handlers/`。
- 综述仅基于上述源码及其依赖的本地工具函数（例如 `cmd/organization-command-service/internal/utils/validation.go`），未引入第二事实来源；已与 34/35/36 号计划文档交叉验证，不存在矛盾结论，保持跨层一致性。

## 代码质量
- **复杂度聚集**：`CreateOrganization`、`CreateOrganizationVersion`、`changeOrganizationStatusWithTimeline` 等函数超过 150 行，混杂校验、层级计算、审计记录和响应拼装，难以测试与扩展。
- **松散类型**：多数成功响应与时间轴回传采用 `map[string]interface{}` 组装，缺乏编译期约束，也不利于契约演化。
- **错误识别脆弱**：多处通过 `strings.Contains(err.Error())` 区分业务错误（父级不存在、版本冲突等），对仓储返回文案高度耦合，易在国际化或驱动更新后失效。
- **统计缺陷**：`operational.go` 的 `GetTaskStatus` 未递增 `runningCount`，导致运维视图始终显示 0。
- **校验分歧**：部分 handler 仍使用 `len(code) != 7` 之类早期规则，与 `ValidateOrganizationCode`（3-10 位、允许大写字母/数字/下划线）不一致，存在行为漂移隐患。
- **重复逻辑**：父级变更检测、层级刷新、审计写入在 `organization_update.go` 与 `organization_history.go` 等文件重复实现；时间轴 DTO 构造在状态变更与事件处理间多处复制。
- **运维/调试过载**：`DevToolsHandler` 将 JWT 生成、数据库巡检、HTTP 代理等集中在命令服务进程，虽受 `devMode` 控制但仍增加攻击面；`OperationalHandler` 暴露 cutover / consistency-check 端点却仅写日志，真实执行缺失。

## 建议
1. **拆分逻辑与类型化响应**：将 handler 分解为请求解析、业务执行、响应构造三个单元，引入结构化 DTO 并统一使用 `utils.WriteSuccess/WriteError` 输出。
2. **统一校验与错误语义**：复用 `ValidateOrganizationCode`、`NormalizeParentCodePointer` 等函数；在仓储/时间轴管理器层定义显式错误（如 `ErrParentNotFound`、`ErrTemporalConflict`、`repository.ErrOrganizationHasChildren`），handler 端改用 `errors.Is/As` 判定。
3. **修复运维统计与占位实现**：迭代 `GetTaskStatus` 统计逻辑，并让 `executeCutover`、`executeConsistencyCheck` 真正调用调度接口（或暂停对外暴露），确保响应符合实际执行情况。
4. **强化 DevTools 安全边界**：对白名单路径、HTTP 方法、头部及请求体大小做限制，必要时将 `/dev/test-api` 下沉至独立开发二进制或 CLI 工具。
5. **抽取通用时间轴/审计构造**：集中封装时间轴条目与审计事件构造，减少跨 handler 复制，降低遗漏字段或响应不一致的风险。

## 验收标准
- [ ] handler 拆分后核心逻辑可被独立测试，成功响应改用结构体，序列化结果与契约一致。
- [ ] 所有组织代码、父级、循环引用校验统一调用 utils/validators，删除硬编码长度检查。
- [ ] 业务错误通过强类型判定覆盖常见冲突场景，并新增测试验证父级缺失、版本冲突、子节点存在等路径。
- [ ] 运维接口返回真实统计且支持失败路径，cutover/一致性检查端点具备可观测执行结果。
- [ ] DevTools 端点启用安全限制，文档声明默认仅在开发模式公开，并与 `docs/api/openapi.yaml` 契约保持一致或注明非公开端点。

## 一致性校验
- 建议与结论均源自以上七个源码文件及 `cmd/organization-command-service/internal/utils/validation.go`，未引入外部数据；执行改动前需交叉核对 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 的字段命名。
- 改动完成后请按流程将本计划迁移至 `docs/archive/development-plans/`，并在验收记录中引用本文，维持唯一事实来源链路。
