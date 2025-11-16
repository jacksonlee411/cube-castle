# Cube Castle 参考文档目录

> 说明：参考文档仅用于长期稳定的技术索引与操作指引；项目原则、约束与权威链接以仓库根目录 `AGENTS.md` 为唯一事实来源。若本目录与 `AGENTS.md` 或 `docs/api/*` 存在不一致，请以 `AGENTS.md` 为准。

此目录包含项目的核心参考文档，为开发者提供完整的技术文档支持。

---

## 📚 文档清单

### ⚡ [01. 开发者快速参考](./01-DEVELOPER-QUICK-REFERENCE.md)
**用途**: 开发过程中的快速查阅手册  
**包含内容**:
- 开发前必检清单
- 常用命令速查表
- 端口配置参考
- API端点速查
- 前端组件速查
- 错误排查指南
- 代码规范速查

**使用场景**:
- ✅ 日常开发快速查阅
- ✅ 新团队成员快速上手
- ✅ 问题排查和调试

### 🏗️ [02. 实现清单](./02-IMPLEMENTATION-INVENTORY.md)
**用途**: 项目现有实现索引，避免重复造轮子  
**包含内容**（以脚本输出为准，避免在文档中固化数量以致偏离）:
- REST/GraphQL 契约项与端点索引
- Go 后端关键组件与导出项索引
- 前端导出项与工具/配置索引
- 重复实现风险提示与避免指引

**使用场景**: 
- ✅ 开发新功能前必检，避免重复实现
- ✅ 了解项目现有能力和架构
- ✅ 寻找可复用的组件和工具

### 📖 [03. API与质量工具指南](./03-API-AND-TOOLS-GUIDE.md)
**用途**: API使用与质量工具统一指南
**包含内容**:
- CQRS架构使用和核心原则
- REST命令API和GraphQL查询API使用
- 质量工具操作 (IIG护卫、P3防控系统)
- 开发前检查和最佳实践
- 错误处理和故障排除

**使用场景**:
- ✅ 日常API开发和使用
- ✅ 质量工具操作和问题排查
- ✅ 新功能开发前的完整指导

### 🗂️ [Job Catalog 二级导航使用指南](./job-catalog-navigation-guide.md)
**用途**: 职位管理二级导航与 Job Catalog 操作说明（长期稳定参考）  
**包含内容**:
- 侧栏导航路径与基线截图位置
- Job Catalog 列表/详情操作流程
- Scope 权限映射与回归脚本
- 关联实现代码索引与设计文档链接

**使用场景**:
- ✅ 业务操作或支持团队了解 Job Catalog 入口
- ✅ 验证导航视觉是否符合基线
- ✅ 规划权限与回归测试脚本

### 🧭 [Temporal Entity Experience Guide](./temporal-entity-experience-guide.md)
**用途**: 时态实体（组织/职位等）多页签详情页面的统一设计与交互规范  
**包含内容**:
- 页面架构（版本导航 + 六个页签）与响应式规则
- Mock 模式、审计页签等交互准则
- 可访问性要求与配色/状态标签说明
- 配套截图存放位置及更新指引

**使用场景**:
- ✅ 设计/前端确认多页签布局一致性
- ✅ 规划新增页签或状态时的约束依据
- ✅ QA 验证 Mock 模式或响应式表现

### 🛡️ [04. Docker 最佳实践](./04-DOCKER-BEST-PRACTICES.md)
**用途**: 汇总容器化开发的强制原则与常见操作  
**包含内容**:
- Docker Compose 启停流程与日志查看
- `localhost` 端口映射与环境变量说明
- 可选的 Air 热重载配置入口
- 违规时的回滚与排障建议

### 🤖 [05. CI/本地一键自动化指引](./05-CI-LOCAL-AUTOMATION-GUIDE.md)
**用途**: 统一门禁（CQRS/端口/禁用端点）与 E2E 的 CI/本地一键化实践  
**包含内容**:
- 唯一门禁工具链（architecture-validator）与“禁止直连 :9090/:8090”规则
- 证据规范（logs/plan<ID>/*）与 Playwright 报告/trace/JSON（可选 HAR）
- CI 门禁与 E2E 工作流（plan-255-gates、frontend-e2e-devserver）与 VS Code 任务建议
- SUMMARY 打印与远程抓取脚本：print-e2e-summary.js、fetch-gh-summary.sh

---

## 🚀 快速开始

### 新开发者上手顺序
1. **先读** [01. 开发者快速参考](./01-DEVELOPER-QUICK-REFERENCE.md) - 了解基本开发流程
2. **再看** [02. 实现清单](./02-IMPLEMENTATION-INVENTORY.md) - 了解项目现有功能
3. **最后** [03. API与质量工具指南](./03-API-AND-TOOLS-GUIDE.md) - API使用与质量工具

### 开发前必做事项
```bash
# 1. 检查现有实现，避免重复造轮子
node scripts/generate-implementation-inventory.js

# 2. 启动开发环境
make docker-up && make run-dev && make frontend-dev

# 3. 生成开发JWT令牌
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER
```

---

## 🎯 核心开发原则提醒

- 原则与黑名单请查阅 `AGENTS.md`（唯一事实来源）；本目录不重复列示，以避免事实漂移。
- 关键摘录（非权威，以 AGENTS.md 为准）：查询用 GraphQL（8090）、命令用 REST（9090）；字段使用 camelCase；端口/服务严格由 Docker Compose 管理，遇端口冲突须卸载宿主机同名服务，禁止修改容器端口映射。


## 📊 质量与规模

- 质量与规模等度量请以 CI 产出的 `reports/` 下最新报告为准；为避免事实漂移，本页不固化具体数值。

---

## 🔗 相关资源

### 核心规范文档
- [API契约规范](../api/) - OpenAPI和GraphQL Schema
- [开发计划文档](../development-plans/) - 项目架构和规划
- [项目原则与索引（唯一）](../../AGENTS.md) - 开发规范与约束

### 开发工具
> ⚠️ `localhost` 端口均由 Docker 容器对外暴露，禁止在宿主机安装同名服务占用端口；若遇冲突请卸载宿主服务，切勿修改容器映射。
- GraphiQL调试界面: http://localhost:8090/graphiql（容器 `graphql-service` 映射）
- 实现清单生成器: `node scripts/generate-implementation-inventory.js`

---

## 📌 目录边界声明与交叉链接

- 本目录仅包含“长期稳定、对外可依赖”的参考资料（快速参考、实现清单、API 使用与质量手册）。
- 计划/路线/进展/阶段报告类文档不在本目录，统一放置于 `../development-plans/`。
- 建议流程：
  - 开始新功能前 → 先查 [实现清单](./02-IMPLEMENTATION-INVENTORY.md) 与 [API与质量工具指南](./03-API-AND-TOOLS-GUIDE.md)
  - 确认需要新增能力 → 前往 [开发计划目录使用指南](../development-plans/00-README.md) 建立/更新计划与进展，并按规范归档。

---

## 📝 文档维护

### 更新频率
- **实现清单**: 每次新增功能后自动更新
- **API指南**: API变更时手动更新
- **快速参考**: 月度例行更新

### 维护责任
- **架构组**: 负责文档整体维护和版本管理
- **开发团队**: 负责相关模块文档的及时更新
- **质量团队**: 负责文档质量检查和一致性验证

### 反馈渠道
如发现文档问题或需要补充内容，请：
1. 在项目issue中提出文档改进建议
2. 直接提交文档PR进行修正
3. 在团队会议中讨论文档规范

---

*这些文档是项目开发的重要参考，请保持最新并合理使用！*

*维护团队: Cube Castle 架构组（更新时间以 Git 历史为准）*
