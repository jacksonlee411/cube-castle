# 06 — 集成团队推进记录（RS256 认证与 API 合规治理）

最后更新：2025-09-20 01:25 UTC
维护团队：认证小组（主责）+ 前端工具组
状态：待测试

---

## 1. 本次变更概览
- **2025-09-24**：完成 `/organization-units/temporal` 契约回正专项（计划12），前端回归契约端点 `/api/v1/organization-units/{code}/versions`，后端清理停用处理器并更新实现清单。
- **2025-09-27**：清理仓库残留 `/temporal` 模拟服务、测试脚本和端口代理，Playwright/E2E 改用 GraphQL `organizationVersions`，部署健康检查更新为契约端点。
- **JWT 链路强制 RS256**：命令服务与 BFF 不再接受 HS256，缺省值及回退路径全部改为 RS256；没有私钥时立即失败，避免生成空 JWKS。
- **JWKS 输出修复**：`/.well-known/jwks.json` 现确保携带 `bff-key-1` 公钥条目，查询服务可以稳定拉取并验签。
- **API 合规 Lint CJS 化**：迁移配置为 `frontend/.eslintrc.api-compliance.cjs`，启用 `@typescript-eslint/parser` 与插件以解析所有 `.ts/.tsx` 文件。
- **脚本更新**：根目录 `package.json` 的 `lint:frontend-api` / `compliance:*` 统一个指向新配置文件，后续 CI 可直接复用。

---

## 2. 影响范围
- 命令服务：`cmd/organization-command-service/internal/auth/jwt.go`、`internal/config/jwt.go`、`internal/authbff/handler.go`、`internal/authbff/jwtmint.go` 等模块。
- 查询服务：共享认证库 `internal/auth/jwt.go` 采用 RS256 默认值。
- 开发工具 & 测试：`make jwt-dev-mint`、`make run-auth-rs256-sim`、Playwright E2E 认证依赖。
- 前端工具链：`frontend/.eslintrc.api-compliance.cjs`、根 `package.json`。

---

## 3. 当前状态与已验证项
- ✅ 手工执行 `make run-auth-rs256-sim` 后，命令/查询服务均可启动；JWKS 端点返回有效 RSA key。
- ✅ `NODE_PATH=frontend/node_modules npx eslint@8.57.0 frontend/src/**/*.{ts,tsx} --config frontend/.eslintrc.api-compliance.cjs` 可成功执行并给出告警列表。
- ⚠️ 仍存在 3 项 `camelcase` 错误（`grant_type` 等外部协议字段）和多处 `no-console` 警告，未在本次改动内处理，需要评估是否保留豁免或做封装。
- ⚠️ `npm install` 依赖 `@stoplight/spectral-oasx` 时继续触发 404/网络问题，待工具链仓库替换源或镜像。

---

## 4. 待测试事项（交付测试团队）
1. **认证链路回归（RS256 + JWKS）**
   - 步骤：
     1. 执行 `make run-auth-rs256-sim`（或手动设置 `JWT_ALG=RS256`、`JWT_PRIVATE_KEY_PATH` 等环境变量）。
     2. 调用 `curl http://localhost:9090/.well-known/jwks.json`，确认返回 `"keys"` 数组非空且 `kid=bff-key-1`。
     3. `make jwt-dev-mint` 生成令牌，并以 `curl -H"Authorization: Bearer"` + `X-Tenant-ID` 请求 `http://localhost:8090/graphql` 的任意业务查询，验证响应 200。
   - 预期：令牌签名算法为 RS256（可通过 JWT header 校验），查询服务不再报 `invalid signing method: HS256`。

2. **前端 API 合规 Lint 验证**
   - 步骤：
     1. `cd /home/shangmeilin/cube-castle`
     2. `NODE_PATH=frontend/node_modules npx eslint@8.57.0 frontend/src/**/*.{ts,tsx} --config frontend/.eslintrc.api-compliance.cjs`
   - 预期：命令执行成功，仅剩 3 个 `camelcase` 错误（外部协议字段）及若干 `no-console` 警告；确认无额外解析错误。

3. **Playwright E2E 冒烟**
   - 目的：验证 Playwright 在 RS256 环境下可以获得合法会话。
   - 步骤：
     1. 继承测试 1 准备好的后端环境。
     2. `make jwt-dev-mint && eval $(make jwt-dev-export)` 设置 `PW_JWT`。
     3. `cd frontend && PW_SKIP_SERVER=1 PW_JWT=$JWT_TOKEN PW_TENANT_ID=... npx playwright test --grep "temporal"`（挑选关键场景）。
   - 预期：GraphQL 请求不再因 `invalid signing method` 失败。

4. **风控回归（可选）**
   - 需确认 `cmd/organization-command-service/internal/auth/jwt_test.go` 新增断言在 CI 中通过。
   - 运行：`go test ./cmd/organization-command-service/internal/auth -run TestGenerateTestTokenRS256 -v`。

---

## 5. 后续跟进与风险
- [ ] 决策是否保留 `grant_type` 等字段的 camelCase 告警：若需长期豁免，应在 lint 配置中加入例外，并在文档记录依据。
- [ ] 评估全局移除调试用 `console`，或改写为约定日志工具（影响范围大，建议与前端团队同步节奏）。
- [ ] `@stoplight/spectral-oasx` 拉取失败会阻塞完整 `npm install`，CI 需使用缓存或私有镜像；工具组负责与平台团队协作处理。

---

## 6. 测试完成后需回填的信息
请测试团队在执行上述用例后，更新下表：

| 日期 | 测试项 | 结果 | 备注 |
| ---- | ------ | ---- | ---- |
| 2025-09-20 | RS256 认证链路 | ✅ | JWKS端点正常，令牌验证成功，不再出现HS256错误 |
| 2025-09-20 | API 合规 Lint | ✅ | 配置工作正常，检测到camelcase和no-console问题 |
| 2025-09-20 | Playwright E2E 冒烟 | ⚠️ | JWT认证通过，但业务逻辑权限问题导致测试失败 |
| 2025-09-20 | JWT单元测试 | ✅ | TestGenerateTestTokenRS256 测试通过 | 

---

## 7. 附录：关键命令速查
```bash
# 后端（RS256 + JWKS）一键启动
make run-auth-rs256-sim

# 生成开发令牌
make jwt-dev-mint && cat .cache/dev.jwt

# GraphQL 健康检查（替换 TOKEN/TENANT）
curl -s -X POST http://localhost:8090/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ organizations(pagination:{page:1,pageSize:1}) { data { code name } } }"}'

# 前端 API 合规 Lint（使用固定版本避免 Flat Config 冲突）
NODE_PATH=frontend/node_modules npx eslint@8.57.0 \
  frontend/src/**/*.{ts,tsx} \
  --config frontend/.eslintrc.api-compliance.cjs
```
