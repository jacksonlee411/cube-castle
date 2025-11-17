# Plan 254 - 前端端点与代理整合（归档）

文档编号: 254  
标题: 前端端点与代理整合（来源：202 计划拆分）  
创建日期: 2025-11-15  
关闭日期: 2025-11-17  
版本: v1.2（归档）  
状态: 已完成（已归档）  
关联计划: 202、241（前端框架合流）、AGENTS（Docker 强制）

---

## 概要
- 统一前端对后端的访问端点（REST/GraphQL）与本地代理，保持“单基址”（/api/v1、/graphql）访问；
- 与 250A 运行时合流对齐：/graphql 由单体进程（:9090）提供；Vite 代理与端点 SSoT 对齐。

## 执行要点（索引）
- 端点/端口 SSoT：frontend/src/shared/config/ports.ts（QUERY_* → :9090）；
- Vite 代理：frontend/vite.config.ts（/api/v1、/graphql → COMMAND_BASE=9090）；
- E2E 配置：frontend/playwright.config.ts、frontend/tests/e2e/config/test-environment.ts；
- CI 门禁：.github/workflows/plan-254-gates.yml（DevServer、自启、证据上传）。

## 结果与证据
- 连续三次成功（正式 gate）：
  - run 53：https://github.com/jacksonlee411/cube-castle/actions/runs/19413508340
  - run 59：https://github.com/jacksonlee411/cube-castle/actions/runs/19414115404
  - run 60：https://github.com/jacksonlee411/cube-castle/actions/runs/19414172110
- 工件（artifact）：plan254-logs（compose-logs、playwright-report、test-results）。

## 关闭说明
- Plan‑254 Gate 已恢复为唯一正式门禁并稳定通过；后续纳入受保护分支 Required checks（已配置）。
- 本文档仅作为归档索引；请以 215 执行日志与 CHANGELOG 为后续权威记录。

