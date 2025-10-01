# Code Smell Type Usage Report (TypeScript)

- **Report Date**: 2025-10-07
- **Plan**: Plan 16 – 代码异味分析与改进
- **Scope**: TypeScript `any`/`unknown` usage baseline
- **Command**: `rg "\bany\b|\bunknown\b" frontend/src --stats`

## Summary
- **Total Matches**: 173 occurrences across 38 files (166 distinct lines)
- **Key Hotspots**:
  - Temporal 功能（`frontend/src/features/temporal/**`）集中出现 `Record<string, unknown>` 用法
  - 审计组件、共享类型转换器与组织权限工具中存在多处 `unknown`
  - 测试与设定文件（`__tests__`, `setupTests.ts`）保留 `any` 以模拟 UI 组件
- **Next Actions**:
  - Phase 2 将按模块划分批次迁移至显式类型定义
  - 对测试专用 `any` 评估是否列入豁免清单以减少噪音

## Raw Output Snapshot

```text
173 matches
166 matched lines
38 files contained matches
119 files searched
```

> 详细匹配列表可通过执行上述命令实时复现，报告作为 Plan 16 Phase 2 的唯一事实来源基线。

## 2025-10-08 Batch A 复核
- **目标范围**: `frontend/src/shared/api/**`
- **匹配结果**（排除测试文件后）:

```bash
cd frontend && rg -g '*.ts*' -e '\bany\b|\bunknown\b' -c src/shared/api
```

```
src/shared/api/error-messages.ts:2
src/shared/api/type-guards.ts:3
src/shared/api/error-handling.ts:1
```

- **说明**:
  - 通过引入 `JsonValue` 类型系统、补全 OAuth 响应类型、封装错误处理别名，将 `shared/api` 模块中的弱类型使用从 **74 处降至 6 处（不含测试）**。
  - 保留的 6 处均为必要的类型守卫入口（`formatErrorForUser` 入参、错误别名定义等），其余 `unknown`/`any` 已替换为强类型定义。
