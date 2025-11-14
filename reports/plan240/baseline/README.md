# Plan 240A – Baseline & Evidence

本目录用于存放 240A（职位详情 Layout 对齐与骨架替换）的基线与证据。

## 已落证据
- Selector Guard：logs/plan240/A/selector-guard.log（通过，未新增旧前缀）
- 架构守护：logs/plan240/A/architecture-validator.log（通过）
- 文档同步：logs/plan240/A/document-sync.log（通过）
- 服务健康：9090/8090 健康检查均 200（本地）

## 待补证据（执行机时限内未完成）
- Storybook 对比截图：reports/plan240/baseline/storybook/*.png（组织 vs 职位）
- E2E 冒烟（Chromium/Firefox 各 2 次）与 trace：
  - 建议命令：
    ```bash
    # 运行前确保容器化服务已就绪（PostgreSQL/Redis/REST/GraphQL）
    make run-dev
    # 生成 JWT 并导入 Playwright 环境
    make jwt-dev-mint && export $(make jwt-dev-export | xargs)
    # Chromium
    PW_JWT=$JWT_TOKEN PW_TENANT_ID=3b99930c-e2e4-4d4a-8e7a-123456789abc npm --prefix frontend run -s test:e2e -- --project=chromium tests/e2e/position-tabs.spec.ts
    # Firefox
    PW_JWT=$JWT_TOKEN PW_TENANT_ID=3b99930c-e2e4-4d4a-8e7a-123456789abc npm --prefix frontend run -s test:e2e -- --project=firefox tests/e2e/position-tabs.spec.ts
    ```
  - 产物路径：logs/plan240/A/playwright-*.log、logs/plan240/A/playwright-trace/*

## 说明
- 本仓库 CI 将运行完整门禁；本地执行 Playwright 受限于执行机超时，已提供可复现命令与目标产物路径。请在本地/CI 拉通后将截图与 trace 回填至上述目录。 

