# Code Smell Type Usage Report (TypeScript)

- **Report Date**: 2025-10-09
- **Plan**: Plan 21 – 弱类型治理专项计划
- **Scope**: TypeScript `any`/`unknown` usage follow-up
- **Command**: `rg -g '*.{ts,tsx}' -o -e '\bany\b|\bunknown\b' frontend/src --stats`

## Summary
- **Total Matches**: 0 occurrences across 0 files
- 所有 TypeScript 源码已移除 `any`/`unknown`，剩余测试及脚手架场景通过结构化类型替换实现
- CI 阈值仍配置为 120（后续可按计划降低至 30）

## Notes
- `scripts/code-smell-check-quick.sh --with-types` 已更新，支持零结果输出并在 CI 中生成报告。
- 本报告作为 Plan 16 Phase 2 收敛证据，替换 `code-smell-types-20251007.md` 中的基线统计。

## Raw Command Output

```
0 matches
0 matched lines
0 files contained matches
133 files searched
0 bytes printed
625368 bytes searched
0.000445 seconds spent searching
0.007541 seconds
```

> 运行脚本时 `rg` 在无匹配场景返回非零状态，`scripts/code-smell-check-quick.sh` 已通过 `|| true` 处理确保流水线稳定。
