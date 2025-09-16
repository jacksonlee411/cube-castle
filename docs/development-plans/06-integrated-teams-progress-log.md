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
- [ ] `golangci-lint` 与 `gosec` 可直接执行（`which golangci-lint`、`which gosec`）。
- [ ] `make lint`、`make security` 均可在无错误的情况下完成（若某项以环境未配置提示失败，请先补齐依赖）。
- [ ] `.cache/dev.jwt` 与 `/.well-known/jwks.json` 均可生成/访问，确保认证中间件可通过构建。

> 如需在 CI 中安装上述依赖，请在各自的构建脚本中加入同样的安装步骤或采用预构建镜像。
