# Atlas 离线使用指南（Plan 214 补充）

为解决外网镜像受限导致 `atlas` CLI 无法直接拉取的问题，仓库现提供离线编译产物与源码补丁，方便在 Docker 环境下继续使用 Atlas + Goose 工作流。

## 可用资源

- 可执行文件：`bin/atlas`
- 源码补丁：`tools/atlas/`（CLI 包装层）与 `tools/atlaslib/`（核心库，兼容 PostgreSQL 16 `pg_settings` 输出）
- 生成的 HCL 示例：`database/schema/schema-inspect.hcl`

## 使用方式

1. **调用 CLI**
   ```bash
   # 检查版本
   ./bin/atlas version

   # 导出当前数据库 Schema（容器内服务）
   ./bin/atlas schema inspect --url "postgres://user:password@localhost:5432/cubecastle?sslmode=disable" --schema public > database/schema/schema-inspect.hcl

   # 基于声明式 Schema 生成 Goose 迁移草案
   ./bin/atlas migrate diff \
     --dir "file://database/migrations" \
     --dev "postgres://user:password@localhost:5432/cubecastle?sslmode=disable" \
     --to "file://database/schema.sql" \
     --format goose
   ```

2. **重新编译（如需升级）**
   ```bash
   # 在具备 Go >=1.24.9 的环境中执行
   cd tools/atlas
   GOWORK=off go build -o ../../bin/atlas
   ```
   > 若需同步 core library，请在 `tools/atlaslib` 内更新并保持与补丁一致。

3. **CI/自动化使用建议**
   - 在 GitHub Actions 中将 `bin/atlas` 加入 `PATH`，无需外网下载。
   - 对应工作流（如 `.github/workflows/ops-scripts-quality.yml`）已启用 Goose round-trip + `go test` 校验，可按需扩展 Atlas diff。

## 注意事项

- 仍需确保所有数据库操作针对 Docker 容器（禁止宿主机直接运行 Postgres）。
- 若后续官方发行版恢复可用，可逐步迁回官方镜像；届时记得移除本地补丁。
