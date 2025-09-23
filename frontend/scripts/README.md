# 前端脚本目录说明

## 迁移脚本（migrations/）
- `20250921-replace-temporal-validation.ts`
  - 功能：将所有导入/导出 `temporalValidation` 的语句替换为统一适配层 `@/shared/utils/temporal-validation-adapter`。
  - 依赖：`tsx` + `ts-morph`，执行前需安装 `npm install` 同步依赖。
  - 使用：
    - 检查模式（不写入文件）：`npm run migrate:temporal-validation -- --check`
    - 执行模式（写入文件）：`npm run migrate:temporal-validation`
  - 注意：
    - 仅识别以 `temporalValidation` 结尾的模块路径，执行前请确保未手动改名。
    - 批量替换完成后仍需手动审核差异并运行测试。

## 验证脚本（根目录）
- `validate-field-naming*.js`：字段命名约束校验。
- `validate-port-config.ts`：本地端口配置校验（`npm run validate:ports`）。
- 其他质量脚本详见仓库根目录 `scripts/README.md`。

> 所有脚本输出、文档与提交流程请遵循 `CLAUDE.md` 和 `AGENTS.md` 中的规范。
