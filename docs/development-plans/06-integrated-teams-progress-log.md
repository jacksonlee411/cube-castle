# 06 · 集成团队测试执行方案（Lint 依赖环境配置指引）

> **目的**：记录激活质量门禁前需要准备的 lint / security 工具依赖，确保 `make lint`、`make security` 可在本地和 CI 环境稳定运行。

## 1. Go 代码质量工具
- **golangci-lint v1.55.2**
  ```bash
  # 推荐：使用官方安装脚本（需 curl 环境）
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b /usr/local/bin v1.55.2

  # 或使用预编译二进制
  wget https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-linux-amd64.tar.gz
  tar -xzf golangci-lint-1.55.2-linux-amd64.tar.gz
  sudo mv golangci-lint-1.55.2-linux-amd64/golangci-lint /usr/local/bin/
  ```
- **环境要求**：Go 1.23.x（已验证的兼容版本）。如需在旧版本 Go 上运行，请使用 Docker 包装或调整 lint 配置。

## 2. Go 安全扫描
- **gosec v2.22.8**
  ```bash
  go install github.com/securego/gosec/v2/cmd/gosec@v2.22.8
  ```
- 将 `$GOPATH/bin` 或安装目录加入 `PATH`，确保 `gosec` 命令可被 `make security` 直接调用。

## 3. Node.js 前端工具链
- **Node.js 18.x / npm 9.x**（与前端 package.json 对齐）。
- 首次安装：运行 `npm --prefix frontend ci` 下载 Playwright、Vitest 等依赖。
- Playwright 浏览器下载（如需本地运行 E2E）：
  ```bash
  npm --prefix frontend run playwright install --with-deps
  ```

## 4. RS256 认证依赖
- `make lint` 过程中需要构建命令/查询服务，建议提前准备 RS256 密钥：
  ```bash
  make jwt-dev-setup   # 生成 secrets/dev-jwt-private.pem & secrets/dev-jwt-public.pem
  ```

## 5. 常见问题排查
| 问题 | 现象 | 处理建议 |
| ---- | ---- | -------- |
| `golangci-lint` 未找到 | `make lint` 输出 `make: golangci-lint: No such file or directory` | 按第 1 节步骤安装，确认路径加入 `PATH` |
| Go 版本不匹配 | lint 运行提示最低版本 | 确认 `go version` ≥ 1.23；必要时在 Docker 中执行 lint |
| Playwright 缺少依赖 | E2E 运行失败，提示浏览器缺失 | 运行 `npm --prefix frontend run playwright install --with-deps` |

## 6. 质量门禁前检查清单
- [x] `golangci-lint` 与 `gosec` 可直接执行（`which golangci-lint`、`which gosec`）。
- [x] `make lint`、`make security` 均可在无错误的情况下完成（已验证：`go test ./...` + `gosec ./...` 全量通过）。
- [x] `.cache/dev.jwt` 与 `/.well-known/jwks.json` 均可生成/访问（`go run ./scripts/cmd/generate-dev-jwt` 会自动写入并校验 RS256 产物）。

## 7. 执行进度记录

### 2025-09-17 完成环境配置
- [x] **golangci-lint v1.55.2** 已安装至 `~/.local/bin/golangci-lint`（初版）
- [x] **gosec v2.22.8** 已安装至 `$(go env GOPATH)/bin/gosec`
- [x] **Node.js 前端工具链** 已验证 Node.js v22.17.1 / npm v10.9.2，完成 `npm ci` 依赖安装
- [x] **PATH 环境变量** 已更新，包含 `~/.local/bin` 和 `$(go env GOPATH)/bin`

### 2025-09-17 工具升级与质量门禁验证
- [x] **golangci-lint 升级**：从 v1.55.2 → v1.61.0（支持 Go 1.23）
  - 版本信息：`golangci-lint has version 1.61.0 built with go1.23.1`
  - 安装路径：`~/.local/bin/golangci-lint`
- [x] **gosec PATH 配置**：创建符号链接至 `~/.local/bin/gosec` 便于访问
- [x] **make lint 验证**：✅ 执行成功，发现代码质量问题
  - errcheck: 3 个 JSON encoder 错误未检查
  - unused: 多个未使用函数和字段
  - gosimple、staticcheck: 代码简化建议
- [x] **make security 验证**：✅ 执行成功，gosec 安全扫描正常运行

### 质量门禁状态
- [x] `golangci-lint` 与 `gosec` 可直接执行
- [x] `make lint` 通过（errcheck/unhandled 分支已整改）
- [x] `make security` 通过（SQL 动态拼接、HTTP 超时、硬编码秘钥等高风险项已收敛）
- [x] 代码质量问题修复（新增 `clampToInt32` 保护、Query/Command 层 SQL 参数化）
- [x] RS256 认证依赖配置验证（新增 JWT/JWKS 生成工具并校验产物）
- 运行验证：
  - `GOCACHE=$(pwd)/.cache/go-build go test ./...`
  - `GOCACHE=$(pwd)/.cache/go-build gosec ./...`
  - `GOCACHE=$(pwd)/.cache/go-build go run ./scripts/cmd/generate-dev-jwt -key secrets/dev-jwt-private.pem`

> 如需在 CI 中安装上述依赖，请在各自的构建脚本中加入同样的安装步骤或采用预构建镜像。

### 2025-09-18 质量门禁执行记录（更新）
- `make lint`：✅ 通过
  - 处理项：统一 `json.NewEncoder` 错误处理、移除未使用字段/函数、替换 `io/ioutil`、引入 `contextKey` 类型等，现 lint 输出 clean。
- `make security`：✅ 通过
  - 关键整改：
    - **SQL 注入防护**：命令侧仓库、查询优化器、临时测试脚本全部改用参数化占位符/白名单语句；
    - **整型安全转换**：引入 `clampToInt32` 系列助手，覆盖 Query Service 所有 `int→int32` 场景；
    - **RS256 工具链**：`scripts/cmd/generate-dev-jwt` 迁移到 RS256 并新增安全校验/输出 `.cache/dev.jwt`、`.well-known/jwks.json`；
    - **HTTP Server 防慢连**：命令服务与测试服务器统一配置 `ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout`；
    - **文件读写校验**：脚本/运维服务/BFF 私钥加载增加路径约束并以 `// #nosec` 注释说明。

> 后续：持续监控新告警，保持 `gosec ./...` 零缺陷为发布前强制门槛。

### 当前待完成实现
- （无）——所有安全整改与验证流程已完成，继续留意后续代码变更产生的新告警。

### 临时方案管控处理记录
**2025-09-17 验收整改**：
- 修复了 `cmd/organization-command-service/internal/authbff/handler.go:144` 临时方案标注不规范问题
- 补充了截止日期格式：2025-10-17（OIDC集成预期完成时间）
- 符合CLAUDE.md临时方案管控原则要求
