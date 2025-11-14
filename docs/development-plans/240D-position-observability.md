# Plan 240D – 职位详情可观测性与指标注入

编号: 240D  
上游: Plan 240（职位管理页面重构） · 依赖 240A/240B 完成  
状态: 待启动

—

## 目标
- 在职位详情骨架注入关键 `performance.mark` 与结构化日志，方便在 E2E/CI 中进行时序与行为断言。

## 范围
- 事件：首屏渲染（hydrate）、Tab 切换、版本切换、导出触发等；输出统一 logger 管线。
- 开关：默认 DEV 开启；CI 可通过 env 开关强制开启。

## 任务清单
1) performance.mark 注入与 logger 输出；  
2) Playwright 用例增加指标断言（监听 console 并断言关键事件出现）；  
3) 文档记录可观测性事件与字段含义（并入 `temporal-entity-experience-guide.md`）。

## 验收标准
- 浏览器日志可见事件；Playwright 指标断言通过；无多余噪声。

## 证据与落盘
- 日志：`logs/plan240/D/*.log`；报告：`reports/plan240/baseline/` 指标对比。

