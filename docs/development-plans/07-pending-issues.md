# 07 — 组织层级同步修复记录

最后更新：2025-09-21
责任团队：架构组（主责）+ 测试组
当前状态：修复已完成，发现新的认证问题

---

## 1. 已完成的技术调整
- 新增 `recalculateSelfHierarchy`，在 `Update` / `UpdateByRecordId` 内同步写入 `level`、`path`、`code_path`、`name_path`，日志示例：`recalculateSelfHierarchy: code=..., oldLevel=..., newLevel=..., path=...`。
- `UpdateByRecordId` 及常规更新均空值保护 `name`、`parentCode`，并触发上述重算逻辑。
- 循环引用在 Handler 层前置拦截，输出 `⚠️ circular reference attempt` 日志并直接返回 400。
- `refreshHierarchyPaths` 精简为只刷新子树，当前节点由重算逻辑负责写回。

代码参考：
- `cmd/organization-command-service/internal/repository/organization.go`
- `cmd/organization-command-service/internal/handlers/organization.go`

---

## 2. 待测试项目
1. **父级切换后数据同步**  
   - 接口：`PUT /organization-units/{code}/history/{record_id}`  
   - 期望：响应及数据库中 `level`=新层级、`path`/`code_path`/`name_path` 正确写回；查看 `recalculateSelfHierarchy` 日志确认。
2. **循环引用防护**  
   - 请求 `parentCode = self` 或子孙节点  
   - 期望：HTTP 400 + 错误码 `BUSINESS_RULE_VIOLATION`，日志出现 `circular reference attempt`；库中无残留更新。
3. **前端展示一致性**  
   - 页面：`/organizations/{code}/temporal`  
   - 期望：层级显示与数据库字段一致；路径/面包屑正常。
4. **数据一致性校验（可选）**  
   - SQL 或 `TemporalService.RecomputeTimelineForCode`  
   - 期望：抽样组织（例如 1000009）`level=2`、`code_path=/1000000/1000009`，无异常记录。

---

## 3. 验收标准
- 组织 `1000009` 从 `1000056` 迁至 `1000000` 后：
  - `level` 由 3 降至 2；
  - `code_path`/`name_path` 更新为 `/1000000/1000009` 对应值；
  - 前端层级显示与数据库一致；
  - 循环引用请求返回 400，数据库无变更。
- 技术验证：
  - `recalculateSelfHierarchy` 调用日志在复测中可见；
  - 循环引用校验覆盖直接、间接、自引用场景；
  - 级联刷新在目标规模（可先以现有数据验证）内表现正常。

---

## 4. 新发现的问题（2025-09-21测试）

### 4.1 JWT认证失败导致组织修改操作退回登录页面

**问题描述**：
在前端页面点击"修改记录"并提交修改后，请求返回401 Unauthorized错误，页面自动跳转到登录页面。

**复现步骤**：
1. 访问 http://localhost:3000/organizations
2. 点击任意组织的"详情管理"按钮
3. 在组织详情页点击"修改记录"
4. 修改任意字段（包括上级组织）
5. 点击"提交修改"
6. 观察到页面自动跳转到登录页

**错误详情**：
- **API端点**：`PUT /api/v1/organization-units/{code}/history/{record_id}`
- **HTTP状态**：401 Unauthorized
- **服务器日志**：
  ```
  rest_middleware.go:84: Dev mode: JWT validation failed:
  token parsing failed: token is unverifiable:
  error while executing keyfunc: no public key available for RS256
  ```
- **前端日志**：
  ```
  [REST Client] 401 未认证，尝试强制刷新令牌并重试一次
  [OAuth] 正在获取新的访问令牌...
  [OAuth] 访问令牌获取成功，有效期: 3600 秒
  Error: 认证已过期，请刷新页面重新登录
  ```

**网络请求分析**：
1. `POST /auth/dev-token` => 200 OK（令牌获取成功）
2. `PUT /api/v1/organization-units/{code}/history/{record_id}` => 401（第一次失败）
3. `POST /auth/dev-token` => 200 OK（重新获取令牌）
4. `PUT /api/v1/organization-units/{code}/history/{record_id}` => 401（重试仍失败）

**问题分析结果（2025-09-21）**：
- 复核 `cmd/organization-command-service/internal/auth/jwt.go` 的校验逻辑可知，当RS256模式下既未配置JWKS也未成功解析本地公钥时，将直接返回 `no public key available for RS256` 错误。
- 本地复测 `make run-dev` 启动后，通过 `POST /auth/dev-token` 取得的令牌头部携带 `kid=bff-key-1`，但命令服务运行日志持续输出 `Dev mode: JWT validation failed: token parsing failed: token is unverifiable: error while executing keyfunc: no public key available for RS256`。
- 进一步检查进程环境变量，发现 `JWT_PUBLIC_KEY_PATH` 未被注入（或指向不存在的文件），导致 `ParseRSAPublicKeyFromPEM` 未能写入 `publicKey` 字段，从而在验证阶段报错。

**根因定位**：
1. 命令服务在开发模式下强制 RS256，但 `JWT_PUBLIC_KEY_PATH`/`JWT_PRIVATE_KEY_PATH` 没有随启动脚本正确挂载，或对应 `secrets/dev-jwt-public.pem` 缺失。
2. 由于 `kid` 存在且 JWKS 未开启，`ValidateToken` 只能依赖内置公钥；缺失后所有需要认证的 REST API 均返回 401。
3. 前端刷新令牌仍复用相同签名算法，因后端无法加载公钥，重试无效，触发登录态清除。

**影响范围**：
- 所有组织历史记录修改操作（`PUT /api/v1/organization-units/{code}/history/{record_id}`），及任何命令服务需要认证的 REST 接口。
- 查询服务（GraphQL）在默认 `make run-dev` 下通过 JWKS 校验，不受该问题影响。

**解决方案建议**：
1. **修复环境变量**：在启动命令服务前执行 `make jwt-dev-setup`，确认生成 `secrets/dev-jwt-*.pem`，并通过 `echo $JWT_PUBLIC_KEY_PATH` 验证路径是否注入；必要时在 VSCode/IDE 运行配置中显式添加这两个变量。
2. **启动自检**：为命令服务增加启动期检查日志（或失败即退出），确认 `JWT_PUBLIC_KEY_PATH` 文件存在且成功解析，避免在运行期才暴露认证错误。
3. **提供兜底 JWKS**：启用 `make run-auth-rs256-sim` 或配置 `JWT_JWKS_URL=http://localhost:9090/.well-known/jwks.json`，让校验逻辑在公钥缺失时可回退到在线 JWKS。
4. **前端容错**：在前端的 401 重试逻辑中，捕获 `DEV_INVALID_TOKEN`/`INVALID_TOKEN` 的特定错误提示，引导开发者检查本地密钥配置，减少误判为登录过期。

**后续动作**：
- 修复后需再次执行组织历史记录修改流程，确认 200 响应、页面不再跳转登录，并记录服务日志截图。
- 建议补充脚本 `scripts/dev/check-jwt-env.sh`（或在现有启动脚本内嵌）用于 CI/本地预检查，结果回填至本档案。

### 4.2 验证结果（2025-09-21 14:02）

**验证步骤执行结果**：

1. **获取开发令牌** ✅
   ```bash
   POST /auth/dev-token => 200 OK
   Response: {"success":true, "data":{"token":"eyJ..."}}
   ```

2. **查询历史记录** ✅
   ```sql
   record_id: d06c8e73-e487-4fdc-abc0-1ecfd0129420
   code: 1000009
   name: TEST FINAL STATE
   ```

3. **调用历史记录更新接口** ✅
   ```bash
   PUT /api/v1/organization-units/1000009/history/d06c8e73-e487-4fdc-abc0-1ecfd0129420
   Response: HTTP/1.1 200 OK
   {
     "success": true,
     "data": {
       "code": "1000009",
       "description": "RS256 dev token validation OK",
       "changeReason": "fix rs256 public key",
       "updatedAt": "2025-09-21T06:02:04.427577Z"
     }
   }
   ```

4. **前端验证** ✅
   - 页面成功跳转至 `/organizations/1000009/temporal`
   - 描述字段显示："RS256 dev token validation OK"
   - 最后更新时间："2025/9/21 14:02:04"
   - 控制台日志："[OAuth] 访问令牌获取成功，有效期: 3600 秒"

**验证结论**：
- ✅ RS256令牌认证成功（使用正确的JWT_PUBLIC_KEY_PATH环境变量）
- ✅ 历史记录修改API调用成功
- ✅ 前端保持登录状态且数据更新正确显示
- ✅ 验证了问题根因确实为JWT公钥配置缺失

---

复测完成后请将结果（含日志、请求/响应、数据库快照）同步回本页或对应 Issue，以便归档。
