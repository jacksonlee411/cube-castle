# 测试检查清单

> 参考：`AGENTS.md` 测试章节、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`

- [ ] `go test ./...`、`go test -race ./...`、`go test -cover ./...` 全部通过，覆盖率 ≥80%。
- [ ] 集成测试使用 `//go:build integration` 标签并通过 `make test-integration` 在 Docker 环境执行。
- [ ] GraphQL/REST handler/resolver 各自拥有正向、异常、权限测试（可在 `tests/` 或模块内部 `*_test.go`）。
- [ ] Playwright/E2E 用例存放 `tests/e2e/`，命名遵循业务场景，并通过 `npm run test:e2e` 执行。
- [ ] 性能或长耗时测试需记录输入/输出数据集并写入 `logs/`，便于 Phase2 验证（参考 `docs/archive/development-plans/222-organization-verification.md`）。
