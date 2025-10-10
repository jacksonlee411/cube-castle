# 54号文档：查询服务（Query Service）GraphQL中间件质量分析

## 背景与唯一事实来源
- 本文聚焦 `cmd/organization-query-service/internal/auth/graphql_middleware.go`、`cmd/organization-query-service/internal/middleware/graphql_envelope.go`、`cmd/organization-query-service/internal/middleware/request_id.go` 三个文件，所有结论均直接来源于当前源码。
- 已对照 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 与 51 号质量分析文档，确认不存在矛盾结论，确保唯一事实来源与跨层一致性。

## 代码质量评估
- `graphql_middleware.go`：结构清晰但存在冗余逻辑，`handleDevMode` 与 `handleProductionMode` 代码高度重复，`createMockClaims` 属于遗留未使用函数，错误写入函数存在重复实现。
- `graphql_envelope.go`：基础拦截流程明确，但直接以字面量拉取 `requestId` 导致无法复用请求上下文，错误识别对大小写处理使用自建函数，整体缺乏单元测试保障。
- `request_id.go`：职责单一且实现简单，不过自定义 `contextKey` 与查询服务其他模块交互时需保持一致，否则会出现跨包取值失败。

## 主要问题
- **请求 ID 丢失**：`GraphQLEnvelopeMiddleware` 通过 `r.Context().Value("requestId")` 获取标识，但请求上下文使用的是 `RequestIDKey`（自定义类型），导致包装响应始终回退为 `"unknown"`，破坏统一追踪能力。
- **开发/生产分支实质无差异**：所谓“开发模式宽松认证”仍强制校验 JWT 与租户头，仅在日志与错误码上有轻微变动；注释与实际行为不符，易导致误解。
- **错误响应重复实现**：`writeErrorResponse` 与 `WriteEnterpriseErrorResponse` 完全相同，只是导出级别不同，增加维护成本且易造成行为漂移。
- **未使用的 Mock 逻辑**：`createMockClaims` 从未被调用，且默认写死租户 ID 与角色，与权限检查实际流程脱节。
- **日志粒度与安全**：开发模式下将 JWT 验证错误原文拼接返回，可能在前端暴露内部错误细节，违背最小泄露原则。

## 过度设计分析
- `handleDevMode` 与 `handleProductionMode` 基本一致，仅保留冗余函数与注释；设计出开发模式却未提供真正的便捷能力，属于无效复杂度。
- `WriteEnterpriseErrorResponse` 作为导出函数但与私有实现重复，未体现附加价值；更适合统一抽象为单一可复用方法。
- `createMockClaims` 意图支持免验证体验，却仍保留严格 JWT 校验流程，形成“虚设能力”。

## 重复造轮子情况
- `graphql_envelope.go` 手动实现 `containsIgnoreCase`、`stringContains`、`indexOf`，重复标准库 `strings` 包功能且未覆盖 Unicode 大小写等场景。
- 请求 ID 写入与读取使用自定义上下文键而未复用既有 `middleware.GetRequestID`，实质上重新实现了访问手段且造成错配。

## 改进建议
- **统一请求 ID 访问接口**：在 GraphQL 信封中间件中改用 `middleware.GetRequestID` 获取标识，或直接使用 `RequestIDKey`，确保链路追踪一致。
- **收敛错误写入逻辑**：合并 `writeErrorResponse` 与 `WriteEnterpriseErrorResponse`，提供单一导出方法，避免未来差异化更新。
- **理顺开发模式语义**：明确开发模式要么放宽认证（支持本地 Mock）、要么借助配置关闭；若保持严格校验，应更新注释与日志，避免误导使用者。
- **移除或落地 Mock 能力**：若需要离线体验，应调用 `createMockClaims` 并隔离默认租户 ID；若无需求，建议删除该函数并清理相关注释。
- **复用标准库能力**：将字符串大小写匹配改为 `strings.EqualFold`/`strings.Contains` 等现成函数，引入必要测试覆盖权限错误映射分支。
- **加强错误信息保护**：限制返回给客户端的 JWT 校验原文，仅记录在服务器日志，防止泄露内部结构信息。

## 验收标准
- [ ] GraphQL 响应信封能够正确携带 `requestId`，并通过单元测试验证常规查询与错误路径均返回一致的追踪标识。
- [ ] 权限中间件导出的错误写入方法保持唯一实现，开发与生产模式行为差异明确并有文档说明。
- [ ] 字符串匹配逻辑改用标准库，针对权限错误映射提供测试，覆盖大小写及多错误条目场景。
- [ ] 开发模式下的认证策略与注释保持一致，Mock 能力要么可配置可启用，要么彻底移除。

## 一致性校验说明
- 全部建议基于当前源码与既有规范文档，未引入第二事实来源；实施改动时需同步确认 `docs/api/schema.graphql` 中的字段与错误码约定保持 camelCase 与 `{code}` 占位一致。
- 文档位于 `docs/development-plans/`，变更落地后请按流程归档至 `docs/archive/development-plans/`，并在验收记录中引用本文以维持唯一事实来源链路。
