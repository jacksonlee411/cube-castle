# Phase1 Day6-7 架构审查准备材料

**更新时间**：2025-11-03 20:00 CST  
**执行负责人**：Codex（全栈）  
**参考依据**：`docs/development-plans/06-integrated-teams-progress-log.md`、`docs/development-plans/211-phase1-module-unification-plan.md`

---

## 1. 基线快照
- 分支与提交：`feature/204-phase1-unify@1873757df59ff11d5cdbfbc266f54627710c635d`
- Go 版本：`go 1.24.0`（`go env GOTOOLDIR` 与 `go.mod` 一致）
- 关键校验（2025-11-03）：
  - `npm run lint` ✅
  - `go test ./...` ✅
  - Docker/Compose 未改动，命令/查询服务入口保持 `cmd/hrms-server/{command,query}`

---

## 2. 共享代码抽取清单

| 包路径 | 位置 | 核心职责 | 当前消费者 | 备注 |
|--------|------|----------|------------|------|
| `cube-castle/internal/auth` | `internal/auth` | PBAC/JWT/JWKS 验证、GraphQL 中间件 | `cmd/hrms-server/query/internal/app/app.go:18`、`scripts/cmd/generate-dev-jwt/main.go:18` | command 仍保留独立版本，列为 Day7 合并候选 |
| `cube-castle/internal/cache` | `internal/cache` | 统一缓存管理器（Redis + L1 缓存） | 暂未接入（依赖清单为空） | 保留自 Plan204 预研，需在审查会上决定是否纳入 Phase1 |
| `cube-castle/internal/config` | `internal/config` | JWT 相关配置加载（环境变量 + secrets） | `cmd/hrms-server/query/internal/app/app.go:19`、`scripts/cmd/generate-dev-jwt/main.go:18` | command 仍使用本地 `internal/config`；建议 Day7 评估合并 |
| `cube-castle/internal/graphql` | `internal/graphql` | GraphQL Schema 装载与热更新工具 | `cmd/hrms-server/query/internal/app/app.go:20` | 已完全替换查询服务原实现 |
| `cube-castle/internal/middleware` | `internal/middleware` | GraphQL envelope、请求追踪中间件 | `cmd/hrms-server/query/internal/app/app.go:21`、`cmd/hrms-server/query/internal/auth/graphql_middleware.go:10` | 计划 Day7 评估 command REST middleware 是否能复用 |
| `cube-castle/internal/types` | `internal/types` | 跨服务标准响应体定义 | `cmd/hrms-server/query/internal/auth/graphql_middleware.go:11`、`cmd/hrms-server/query/internal/middleware/graphql_envelope.go:7` | 对齐 `docs/api/openapi.yaml`，需保持唯一事实来源 |
| `cube-castle/shared/config` | `shared/config` | 租户/端口等跨语言配置源 | `scripts/cmd/generate-dev-jwt/main.go:19`、前端 `frontend/src/shared/config/**` | 调和后端工具与前端共享常量 |
| `cube-castle/pkg/health` | `pkg/health` | 健康检查与告警抽象 | 当前未通过 import 使用 | Day7 需决定保留/合并至 `internal/monitoring` |

> 审查关注：command 服务仍保留一套 `internal/*` 包（如 `cmd/hrms-server/command/internal/auth`、`cmd/hrms-server/command/internal/config`）。需要在 Day7 前给出“复用还是保持独立”的决策，并记录于 `reports/phase1-module-unification.md`。

---

## 3. 依赖矩阵（Go 包级别）

> 生成命令：`python3 - <<'PY' …`（详见附录脚本）

| from \\ to | command | query | internal | shared | pkg | tests | other |
|-----------|---------|-------|----------|--------|-----|-------|-------|
| **command** | 35 | 0 | 0 | 0 | 0 | 0 | 0 |
| **query** | 0 | 5 | 9 | 1 | 0 | 0 | 0 |
| **internal** | 0 | 0 | 4 | 0 | 0 | 0 | 0 |
| **shared** | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| **pkg** | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| **tests** | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| **other** | 0 | 0 | 1 | 1 | 0 | 0 | 0 |

- `query → internal`（9 条边）：GraphQL 服务依赖共享认证、配置、Schema Loader、Envelope。示例：`cmd/hrms-server/query/internal/app/app.go:18-21`
- `query → shared`（1 条边）：开发工具（`scripts/cmd/generate-dev-jwt/main.go:18-19`）引用 `shared/config` 以保持租户基线一致。
- `other → internal/shared`：仅 `scripts/cmd/generate-dev-jwt`（辅助工具）引用；需保持脚本与主服务同步。
- `pkg/health` 当前无消费者，需在 Day7 审查结论中明确处理路径。

---

## 4. 回滚说明（Phase1 内部控制）

Steering 已确认 Phase1 不额外维护独立回滚决策树（参见 `docs/development-plans/06-integrated-teams-progress-log.md:255`）。审查会需同步以下最小回滚路径：
1. **创建安全点**：所有 Day6-7 提交 push 前，在 `feature/204-phase1-unify` 打 `annotated tag`（格式：`plan211-day6-checkpoint`）。
2. **代码回滚**：遇到阻塞问题时，通过 `git reset --hard <checkpoint>` 回退分支，再按正常 PR 流程重新提交；禁止强制推送覆盖主线。
3. **依赖回滚**：保留 Day5 合并前的 `go.mod`、`go.sum` 副本（位于 `reports/phase1-module-unification.md` 附录）。若需恢复至多模块结构，使用 `git checkout <tag> -- go.mod go.sum`，并重新执行 `go mod tidy` 验证。
4. **配置恢复**：保持 `.github/workflows/*`、`docker-compose.dev.yml` 变更 commit 粒度可回退；必要时将 Day5 变更以 PR revert 形式回滚。
5. **数据脚本**：Day8 数据一致性脚本入库前，保留 `scripts/tmp` 副本，回滚时删除新脚本并恢复 `reports/phase1-regression.md` 基线。

---

## 5. 审查会议建议议程
1. **共享代码现状确认**（10 分钟）：逐项确认上表包是否满足复用目标，特别是 command 与 query 是否存在重复实现。
2. **依赖矩阵讨论**（15 分钟）：重点关注 `query → internal` 依赖是否需要向 command 开放，以及 `pkg/health` 的处置方案。
3. **回滚可行性**（10 分钟）：确认所有 Day6-7 操作均可通过 git tag + reset 回滚；记录潜在阻塞点。
4. **后续行动**（10 分钟）：形成 Day7 行动项列表（例如合并 command/config、评估 cache 包、清理未使用共享代码）。

---

## 6. 附录：依赖矩阵生成脚本

```bash
python3 - <<'PY'
import json, subprocess
from collections import defaultdict
text = subprocess.run(['go', 'list', '-json', './...'], capture_output=True, text=True, check=True).stdout
decoder = json.JSONDecoder()
objs, data = [], text
while data:
    data = data.lstrip()
    if not data:
        break
    obj, pos = decoder.raw_decode(data)
    objs.append(obj)
    data = data[pos:]
prefix = 'cube-castle'
def classify(path):
    if path == prefix or not path.startswith(prefix + '/'):
        return 'external'
    suffix = path[len(prefix)+1:]
    for group in ('cmd/hrms-server/command', 'cmd/hrms-server/query', 'internal', 'shared', 'pkg', 'tests'):
        if suffix.startswith(group):
            return group.split('/')[-1] if group.startswith('cmd') else group
    return 'other'
groups = {o['ImportPath']: classify(o['ImportPath']) for o in objs if 'ImportPath' in o}
edges = defaultdict(lambda: defaultdict(set))
for o in objs:
    src = groups.get(o.get('ImportPath'))
    if src in (None, 'external'):
        continue
    deps = set(o.get('Imports') or [])
    for key in ('TestImports', 'XTestImports'):
        deps.update(o.get(key) or [])
    for dep in deps:
        dst = groups.get(dep)
        if dst in (None, 'external'):
            continue
        edges[src][dst].add((o['ImportPath'], dep))
for src in ['command','query','internal','shared','pkg','tests','other']:
    row = [src] + [str(len(edges[src].get(dst, ()))) for dst in ['command','query','internal','shared','pkg','tests','other']]
    print('\\t'.join(row))
PY
```

命令输出与上表一致，并已归档于会议纪要。
