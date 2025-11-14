# Plan 240B – 职位详情 数据装载链路与等待治理

编号: 240B  
上游: Plan 240（职位管理页面重构） · 依赖 240A 完成  
状态: 待启动

—

## 目标
- 统一职位详情的数据装载链路，避免白屏、竞态与重复请求；建立可取消与重试的稳定等待机制。

## 范围
- 路由层与详情层的请求合并、AbortController 取消、错误边界与重试策略。
- 不更改契约字段；优先封装在现有 Hook 外层，逐步与 `useTemporalEntityDetail` 对齐。

## 任务清单
1) Suspense-aware loading manager：聚合并发请求、首屏 skeleton 与错误边界。  
2) 请求取消与重试：路由切换/Tab 切换触发取消；指数退避重试仅对幂等读请求生效。  
3) 租户切换与 cache：React Query key 精确、失效策略明确；命令链路后失效刷新。

## 验收标准
- Vitest 单测覆盖：错误态/租户切换/重复 fetch 抑制/取消成功。
- E2E 用例 `position-lifecycle.spec.ts` 在 Chromium/Firefox 通过（连续两次）。

## 证据与落盘
- 日志：`logs/plan240/B/*.log`；报告：`reports/plan240/baseline/` 补充前后对比。

