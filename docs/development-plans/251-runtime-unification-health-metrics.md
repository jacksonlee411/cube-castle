# Plan 251 - 运行时统一与健康指标对齐

文档编号: 251  
标题: 运行时统一与健康指标对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、203、215、218（logger/metrics）

---

## 1. 目标
- 统一运行时配置加载与健康检查接口（/health、/metrics）；
- 建立基础可观测性指标（请求、审计、时态操作、outbox 派发等）；
- 对齐 Docker 化本地/CI 行为（端口、健康、退出策略）。

## 2. 交付物
- 运行时配置与健康文档（仅引用 215/218 与代码事实来源）；
- 健康检查基线（command/query 一致）：返回 JSON，含 service 与 status；
- 指标基线：logger/metrics（pkg/logger、internal/organization/utils/metrics.go）映射说明；
- 验收日志：logs/plan251/*（curl /health、/metrics 样本与规则校验）。

## 3. 验收标准
- 本地/CI 下 /health 全绿；/metrics 暴露基础计数器与直方图；
- 运行时配置来源单一，参数覆盖规则一致；
- 证据登记完整，脚本文档引用权威来源。

---

维护者: 平台与后端联合（与 218 保持一致）

