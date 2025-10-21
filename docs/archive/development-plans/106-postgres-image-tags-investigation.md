# 106 号文档：PostgreSQL 镜像标签一致性调查报告

**创建日期**：2025-10-21（UTC）  
**状态**：已完成（2025-10-21 UTC）  
**维护人**：架构组 · 平台运维协作小组

---

## 1. 背景与目标

- **触发原因**：在本地 Docker 环境中同时存在 `postgres:15-alpine` 与 `postgres:16-alpine` 镜像标签，并在 `docker images` 输出中均显示 `Containers=N/A`，引发对“唯一事实来源”与环境一致性的担忧。
- **调查目标**：核实当前运行容器实际依赖的 PostgreSQL 版本，识别多标签并存的根因，评估与项目约束的差异，并提出整改方案。
- **范围界定**：限定于本项目 Docker Compose 配置及本地镜像标签，无涉宿主机直接安装或云端基础设施。

---

## 2. 权威事实来源

本报告仅引用以下单一事实来源并逐一交叉验证，无引入第二事实来源：

- `docker-compose.dev.yml`：本地开发环境配置当前标注为 `image: postgres:15-alpine`。 
- `docker-compose.yml`、`docker-compose.e2e.yml`：标准/完整及 E2E 环境均指定 `image: postgres:16-alpine`。 
- `README.md` 与 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`：官方要求数据库版本为 PostgreSQL 16+。 
- `docker images postgres --format '{{.Repository}}:{{.Tag}}\t{{.ID}}\t{{.Containers}}'`：显示 `postgres:15-alpine` 与 `postgres:16-alpine` 共享镜像 ID `dca9c7aa70c7` 且 `Containers=N/A`。 
- `docker ps -a --format '{{.Image}}\t{{.Names}}'`：当前运行容器 `cubecastle-postgres` 使用 `postgres:15-alpine` 标签。 
- `docker exec cubecastle-postgres postgres --version`：返回 `PostgreSQL 16.9`。 
- `docker inspect postgres:15-alpine` 环境变量：`PG_MAJOR=16`、`PG_VERSION=16.9`，确认该标签被手动重指向 16 版本。

上述来源经相互印证，确保调查结论沿袭单一事实来源并保持跨层一致。

---

## 3. 调查发现

| 序号 | 现象 | 证据 | 说明 |
| --- | --- | --- | --- |
| F1 | 本地 `docker-compose.dev.yml` 仍引用 `postgres:15-alpine` | 配置文件第 6 行 | 与文档宣称的 16+ 版本不符，存在口径漂移。 |
| F2 | `docker images` 显示两个标签容器数均为 `N/A` | 审核命令输出 | Docker 认为两标签当前无容器绑定，导致“未 in use”误解。 |
| F3 | 实际运行容器镜像标签为 `postgres:15-alpine` | `docker ps -a` | 容器仍绑定旧标签。 |
| F4 | 运行中的 PostgreSQL 版本为 16.9 | `docker exec ... postgres --version` | 功能层面符合 16+ 要求。 |
| F5 | `postgres:15-alpine` 标签内置环境变量指向 PostgreSQL 16 | `docker inspect` 输出 | 该标签已被手动重打，真实版本与标签不一致。 |

---

## 4. 根因分析

1. **标签重打缺乏记录**：`postgres:15-alpine` 与 `postgres:16-alpine` 指向同一镜像 ID，说明曾在本地执行过 `docker tag` 或 `docker pull` 重定向操作，但未在文档或运维记录中备案，导致事实来源不透明。
2. **配置-文档不一致**：开发环境 Compose 文件仍引用旧标签，与 `docker-compose.yml`、`docker-compose.e2e.yml` 以及参考文档中的 PostgreSQL 16+ 要求产生偏差。
3. **工具输出误读**：当同一镜像 ID 被多个标签引用时，`docker images` 对每个标签显示 `Containers=N/A`，若未结合 `docker ps` 实际容器绑定关系，容易误判为“无容器使用”。

综上，表面上的“未使用”实为标签漂移与配置滞后所致，违反了项目的资源唯一性与跨层一致性原则。

---

## 5. 整改建议

1. **统一标签版本**：将 `docker-compose.dev.yml` 的 PostgreSQL 镜像更新为 `postgres:16-alpine`，与其他 Compose 文件保持一致。
2. **清理手工标签**：在执行完统一化配置后，删除本地 `postgres:15-alpine` 标签（`docker rmi postgres:15-alpine`），确保只保留官方发布的 16+ 标签。
3. **补充运维记录**：将本次调查结论和镜像重打历史补录至 `reports/operations` 或对应运维记录，要求后续不得在未备案情况下手动篡改镜像标签。
4. **增设验证步骤**：在 `make status` 或运维检查脚本中加入“Compose 镜像标签与权威版本字段一致性”检查，防止再次出现口径漂移。

---

## 6. 验收标准

- [x] `docker-compose.dev.yml`、`docker-compose.yml`、`docker-compose.e2e.yml` 均引用 `postgres:16-alpine`。 
- [x] 本地 `docker images` 中只保留与 PostgreSQL 16 相关的官方标签，且 `docker ps` 显示的运行容器镜像与文档版本一致。 
- [x] 运维记录更新至最新结论，并在 `docs/development-plans` 或对应归档中引用本报告编号 106。 
- [ ] 后续日常巡检加入镜像标签一致性检查项。

---

## 7. 后续动作与归档

- 完成上述整改后，将本报告归档至 `docs/archive/development-plans/`，并在相关计划文档中引用编号 106 的调查结果。 
- 如需进一步验证镜像来源，可追加对 `docker history`、CI/CD 拉取步骤等的复核，但须保证引用同一事实来源。

---

## 8. 调查确认与可行性评估报告

### 8.1 问题验证结果（全部确认）

1. **配置文件不一致**：
   - `docker-compose.dev.yml:6` 使用 `postgres:15-alpine`
   - `docker-compose.yml:9` 使用 `postgres:16-alpine`
   - `docker-compose.e2e.yml:5` 使用 `postgres:16-alpine`

   **结论**：开发环境配置与标准/E2E 环境不一致，违反“资源唯一性与跨层一致性”原则。

2. **Docker 镜像标签漂移**：
   - `docker images postgres` 显示 `postgres:15-alpine` 与 `postgres:16-alpine` 共享镜像 ID `dca9c7aa70c7`，`Containers=N/A`

   **结论**：两个标签指向同一镜像，且环境变量 `PG_MAJOR=16`、`PG_VERSION=16.9`，说明 `15` 标签被手动重打。

3. **运行容器实际版本**：
   - 容器镜像：`postgres:15-alpine`
   - PostgreSQL 版本：`16.9`
   - 数据卷：`cube-castle_postgres_data`

   **结论**：功能层面符合 16+ 要求，但镜像标签与实际版本不符。

4. **文档要求对比**：
   - `README.md:38`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md:58` 均要求 PostgreSQL 16+

   **结论**：文档口径与开发环境配置存在漂移。

### 8.2 解决方案可行性评估

1. **方案一：统一镜像标签（强烈推荐）**
   - 操作：将 `docker-compose.dev.yml:6` 更新为 `postgres:16-alpine`
   - 可行性：操作简单；实际运行版本已为 16.9；数据卷持久化，无数据风险
   - 风险：容器需重启，预计中断 10–30 秒；回滚难度极低

2. **方案二：清理手工标签（强烈推荐）**
   - 操作：在方案一完成并重启容器后执行 `docker rmi postgres:15-alpine`
   - 可行性：两标签同 ID，删除不会移除真实镜像；可彻底消除歧义

3. **方案三：补充运维记录（必须执行）**
   - 操作：将调查结论与整改时间记录至运维文档/日报
   - 价值：履行诚实原则，防止同类问题再次发生

4. **方案四：增设验证步骤（建议执行）**
   - 操作：在 `make status` 或运维脚本中加入镜像标签一致性检查
   - 可行性：需额外开发 1–2 小时；可作为中期优化项

### 8.3 推荐执行顺序

**阶段一（立即执行）**
1. 更新 `docker-compose.dev.yml` 镜像标签为 `postgres:16-alpine`
2. `docker-compose -f docker-compose.dev.yml down && docker-compose -f docker-compose.dev.yml up -d`
3. 通过 `docker ps` 与 `docker exec cubecastle-postgres postgres --version` 验证服务
4. 执行 `docker rmi postgres:15-alpine`

**阶段二（文档补充）**
5. 在本报告中记录整改时间并更新状态为“已完成”
6. 将结论同步至相关运维记录

**阶段三（长期优化）**
7. 评估并实施镜像标签一致性自动化检查

### 8.4 预期影响与回滚方案

- **服务中断**：仅发生在容器重启阶段（10–30 秒）
- **数据影响**：无（版本一致，数据卷持久化）
- **配置一致性**：由“不一致”提升至“跨层统一”
- **回滚步骤**：如遇异常，执行 `git checkout docker-compose.dev.yml`，随后 `docker-compose -f docker-compose.dev.yml down && docker-compose -f docker-compose.dev.yml up -d`

### 8.5 最终评估结论

- ✅ 问题全部确认
- ✅ 解决方案可行、风险可控、收益显著
- ✅ 建议立即执行阶段一与阶段二整改
- ⚠️ 阶段三属优化项，可待后续评审安排

风险评级：🟢 低风险  
收益评级：🟢 高收益  
执行建议：✅ 立即执行

---

## 9. 整改执行记录

- **2025-10-21 02:16 UTC**：更新 `docker-compose.dev.yml` PostgreSQL 镜像至 `postgres:16-alpine` 并重新部署本地开发容器。
- **2025-10-21 02:18 UTC**：使用 `docker exec cubecastle-postgres postgres --version` 验证数据库版本为 `PostgreSQL 16.9`。
- **2025-10-21 02:19 UTC**：执行 `docker rmi postgres:15-alpine` 清除冗余标签，仅保留官方 `16-alpine`。
- **2025-10-21 02:20 UTC**：在 `reports/operations/postgres-image-tag-rectification-20251021.md` 登记运维记录并更新本报告状态。
