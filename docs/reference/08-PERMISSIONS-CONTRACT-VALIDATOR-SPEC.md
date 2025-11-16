# 权限契约校验器（Plan 252）接口规范

最后更新：2025-11-15

本文定义“权限契约校验器”的接口与输出要求，作为 Plan 252 的执行与验收参考。脚本实现由后续任务完成；本规范不引入新的事实来源，仅读取 OpenAPI 与 GraphQL 契约并对实现进行静态核对。

—

## 目标
- 从唯一事实来源生成并校验权限契约：
  - OpenAPI `security` → scopes 引用与注册一致性（禁止未注册使用）。
  - GraphQL Query → scope 映射由 `docs/api/schema.graphql` 注释生成（SSoT），覆盖率=100%。
  - 读取实现层（resolver/PBAC）以验证鉴权门面调用覆盖率=100%。

## CLI 形态（建议）
```
node scripts/quality/auth-permission-contract-validator.js \
  --openapi docs/api/openapi.yaml \
  --graphql docs/api/schema.graphql \
  --resolver-dirs internal/organization/resolver,cmd/hrms-server/query/internal/auth \
  --out reports/permissions \
  --fail-on missing-scope,unregistered-scope,mapping-missing,resolver-bypass
```

参数说明：
- `--openapi`：OpenAPI 文件路径（YAML）。
- `--graphql`：GraphQL schema 文件路径（.graphql）。
- `--resolver-dirs`：以逗号分隔的源码目录，用于扫描 resolver 授权调用与 PBAC 实现。
- `--out`：输出目录（默认 `reports/permissions`）。
- `--fail-on`：以逗号分隔的失败项类型（见“校验项与失败策略”）。

## 输出制品（报告）
- `openapi-scope-usage.json`：REST 路径 → scopes 使用清单（逐路径、逐方法）。
- `openapi-scope-registry.json`：scopes 注册表导出（来自 `components.securitySchemes.OAuth2ClientCredentials.scopes`）。
- `graphql-query-permissions.json`：从 schema 注释生成的 Query → scope 映射（唯一事实来源衍生物）。
- `resolver-permission-calls.json`：resolver 授权调用覆盖报告（缺失列出具体 Query/文件/行）。
- `summary.txt`：人类可读汇总（总计数、失败项列表与建议）。

生成规则：
- GraphQL 权限注释解析：以 `Permissions Required: <scope>` 文本为锚点，解析 `type Query` 下各字段的权限；支持一对多（逗号/斜杠分隔）但默认一对一；无法解析时记为缺失。
- Resolver 授权检测：静态搜索 `permissions.CheckQueryPermission(ctx, "<queryName>")` 或透传门面调用，校验与 schema 定义的 Query 一致；缺失或 queryName 不一致记为 `resolver-bypass` 或 `mapping-mismatch`。

## 校验项与失败策略
1) OpenAPI 引用→注册一致性（`unregistered-scope`，阻断）
   - 路径 `security` 中的任意 scope 必须存在于注册表。
2) GraphQL 映射覆盖率（`mapping-missing`，阻断）
   - 每个 Query 必须有 `Permissions Required: …` 注释并成功生成映射。
3) Resolver 授权覆盖（`resolver-bypass`，阻断）
   - 每个 Query 入口（resolver 方法）必须调用授权门面；缺失即失败。
4) 注册未引用（`unused-scope`，信息级）
   - 注册表中未被任何路径引用的 scope 仅提示，不阻断。

退出码：
- 0 = 全部通过；非 0 = 存在失败项（按 `--fail-on` 类型统计）。

## 验收度量（最小）
- OpenAPI scope 一致性=100%（0 未注册引用）。
- GraphQL 映射覆盖率=100%（0 缺失）。
- Resolver 授权覆盖率=100%（0 bypass）。
- 报告可复现并落盘至 `reports/permissions/*`。

## 目录与证据约定
- 报告目录：`reports/permissions/*`
- 运行日志：`logs/plan252/*`

## 约束与边界
- 不修改任何契约文件；仅读取并比对。
- 任何“临时兼容”在实现层必须使用 `// TODO-TEMPORARY(YYYY-MM-DD): ...` 标注且登记回收期。

## 后续扩展（可选）
- 跨层权限语义抽样对比：同一业务在 REST 与 GraphQL 的权限是否一致（规则白名单化）。
- 与 Spectral 集成：将未注册 scope 规则下放到 OpenAPI linter。 
