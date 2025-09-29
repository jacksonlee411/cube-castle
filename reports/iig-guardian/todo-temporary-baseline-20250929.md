# TODO-TEMPORARY 治理基线（2025-09-29）

- **生成时间**: 2025-09-29
- **执行人**: IIG 守护代理
- **命令**: `rg "TODO-TEMPORARY" -n`

## 发现列表

| 所在文件 | 行号 | 截止日期 | 描述摘要 |
| --- | --- | --- | --- |
| `docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md` | 56 | 2025-09-30 | 评审是否启用 419 状态码，当前实现仍统一返回 401。 |

## 结论

- 代码目录中不存在 `TODO-TEMPORARY` 标注，唯一待处理项位于参考文档，需要在 2025-09-30 前决策。
- 现有治理脚本 `scripts/check-temporary-tags.sh` 排除了 Markdown 文件，无法自动捕获上述条目，后续阶段需扩展扫描范围。
