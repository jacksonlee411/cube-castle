# Plan 252 - 权限一致性与契约对齐

文档编号: 252  
标题: 权限一致性与契约对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、203、OpenAPI/GraphQL 契约、auth/pbac

---

## 1. 目标
- 以 OpenAPI/GraphQL 为唯一事实来源，对齐 scopes/roles 与接口权限；
- 建立 PBAC 校验清单与回归用例（读侧 GraphQL 中间件 + REST 权限中间件）。

## 2. 交付物
- 权限-契约映射表（仅引用 docs/api/*）；
- PBAC 校验点清单与最小用例；
- 证据登记：logs/plan252/*。

## 3. 验收标准
- 合同字段（security/scope）与中间件校验一致；
- 样例用例覆盖关键查询/命令；
- CI 权限契约检查通过（强制门禁）。

---

维护者: 后端（安全合规协同评审）

---

## 4. 自动化门禁（CI）
- OpenAPI x-scopes 覆盖率=100%：扫描 `docs/api/openapi.yaml`，所有受保护 REST 端点必须声明 `x-scopes`
- GraphQL resolver 必经 PBAC：扫描 resolver 代码，校验统一调用 PBAC 校验器（或其门面）
- 权限一致性对比：基于映射表抽样对比 REST 与 GraphQL 对同一业务操作的权限结论
- 失败即阻断；允许临时豁免须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注并在 215/06 登记（一个迭代内收敛）
