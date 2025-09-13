# Cube Castle 监控系统彻底卸载清理方案

**文档编号**: 06  
**最后更新**: 2025-09-13 21:20  
**维护团队**: 架构团队 + 开发团队  

## 🎯 **监控系统彻底卸载清理方案**

基于对项目监控系统的完整评估，提供以下彻底卸载和清理方案，确保移除所有监控相关组件。

## 📋 **监控系统组件清单**

### 🐳 **Docker基础设施组件**
- **监控容器**: cube-castle-prometheus, cube-castle-grafana, cube-castle-alertmanager, cube-castle-node-exporter
- **Docker卷**: prometheus-data, grafana-data, alertmanager-data
- **Docker网络**: monitoring network
- **配置文件**: docker-compose.monitoring.yml

### 📁 **文件系统组件**
```
monitoring/                          # 主监控目录
├── docker-compose.monitoring.yml    # Docker Compose配置
├── prometheus.yml                   # Prometheus配置
├── prometheus-rules.yml             # 告警规则
├── alertmanager.yml                 # 告警管理配置  
├── slo-config.yml                   # SLO配置
├── grafana/                         # Grafana配置目录
│   ├── provisioning/
│   │   ├── datasources/prometheus.yml
│   │   └── dashboards/
│   └── dashboards/
└── data/                           # 持久化数据目录
    ├── prometheus/
    ├── grafana/
    └── alertmanager/
```

### 🎯 **前端集成组件**
- **监控页面**: `frontend/src/features/monitoring/MonitoringDashboard.tsx`
- **路由配置**: App.tsx中的 `/monitoring` 路由
- **侧边栏导航**: Sidebar.tsx中的"系统监控"菜单项
- **端口配置**: `frontend/src/shared/config/ports.ts` 中的 MONITORING_ENDPOINTS

### ⚙️ **后端集成组件**
- **Prometheus客户端**: 两个服务都集成了prometheus Go客户端
- **Metrics端点**: `/metrics` 端点在两个服务中都存在
- **指标收集器**: 
  - `cmd/organization-command-service/internal/metrics/collector.go`
  - `cmd/organization-query-service/internal/metrics/metrics.go`

### 🛠️ **开发工具组件**
- **启动脚本**: `scripts/start-monitoring.sh`
- **测试脚本**: `scripts/test-monitoring.sh`, `scripts/tests/test-monitoring-integration.sh`
- **Makefile目标**: `monitoring-up`, `monitoring-down`, `monitoring-test`

## 🗑️ **彻底卸载清理方案**

### **阶段1: 停止并清理Docker组件**
```bash
# 停止监控服务
make monitoring-down
# 或
docker compose -f monitoring/docker-compose.monitoring.yml down -v

# 删除监控容器（如果存在）
docker rm -f cube-castle-prometheus cube-castle-grafana cube-castle-alertmanager cube-castle-node-exporter 2>/dev/null || true

# 删除监控数据卷
docker volume rm monitoring_prometheus-data monitoring_grafana-data monitoring_alertmanager-data 2>/dev/null || true

# 删除监控网络
docker network rm monitoring_monitoring 2>/dev/null || true

# 清理无用镜像（可选）
docker image rm prom/prometheus:v2.40.0 grafana/grafana:9.5.0 prom/alertmanager:v0.25.0 prom/node-exporter:v1.5.0 2>/dev/null || true
```

### **阶段2: 删除监控文件系统**
```bash
# 删除整个监控目录
rm -rf monitoring/

# 删除监控相关脚本
rm -f scripts/start-monitoring.sh
rm -f scripts/test-monitoring.sh
rm -f scripts/tests/test-monitoring-integration.sh

# 删除归档文档
rm -f docs/archive/project-reports/16-monitoring-infrastructure-report.md
rm -rf docs/monitoring/
```

### **阶段3: 清理前端监控集成**
```bash
# 删除监控功能模块
rm -rf frontend/src/features/monitoring/

# 从App.tsx中移除监控路由
# 需要编辑删除：
# - MonitoringDashboard import
# - /monitoring 路由定义

# 从Sidebar.tsx中移除监控菜单
# 需要编辑删除侧边栏中的"系统监控"菜单项

# 从端口配置中清理监控端点
# 编辑 frontend/src/shared/config/ports.ts：
# - 删除监控相关端口配置
# - 删除 MONITORING_ENDPOINTS 导出
```

### **阶段4: 清理后端监控集成**
```bash
# 保留/metrics端点（Prometheus标准），但可移除自定义指标
# 如需完全清理：

# 1. 移除Prometheus依赖包
# 编辑 go.mod 文件，移除 prometheus 相关依赖

# 2. 清理指标收集器
rm -f cmd/organization-command-service/internal/metrics/collector.go
rm -f cmd/organization-query-service/internal/metrics/metrics.go

# 3. 从main.go中移除指标初始化代码
# 4. 从处理器中移除指标记录调用
```

### **阶段5: 清理开发工具集成**
```bash
# 从Makefile中移除监控目标
# 编辑Makefile，删除：
# - monitoring-up 目标
# - monitoring-down 目标  
# - monitoring-test 目标
# - help中的监控相关说明
```

### **阶段6: 更新项目文档**
```bash
# 清理CLAUDE.md中的监控相关内容
# 删除以下章节：
# - 🔍 监控系统配置章节
# - 监控端点配置相关内容
# - 开发环境配置中的监控端口

# 更新README.md
# 移除监控系统相关的使用说明
```

## ⚠️ **清理注意事项**

### **🔒 需要保留的组件**
- **基础/metrics端点**: 这是Go服务的标准做法，建议保留基础健康检查
- **端口配置框架**: 保留统一端口配置系统，仅删除监控相关端口
- **日志记录**: 保留应用日志，仅移除监控指标收集

### **🎯 推荐清理策略**
1. **渐进式清理**: 先停止Docker服务，再逐步清理代码集成
2. **备份重要数据**: 如果监控数据有价值，先导出Grafana仪表板配置
3. **测试验证**: 清理后运行完整测试确保核心功能不受影响
4. **文档更新**: 及时更新项目文档，避免误导后续开发

### **💾 清理后的项目状态**
- ✅ **核心功能完整**: 组织管理API功能完全不受影响
- ✅ **端口配置简化**: 移除监控相关端口，配置更简洁
- ✅ **前端界面清理**: 移除监控入口，界面更聚焦业务功能
- ✅ **部署简化**: 无需维护复杂的监控基础设施
- ✅ **资源优化**: 减少Docker容器和持久化存储占用

## 📁 **具体清理文件列表**

### **需要删除的文件和目录**
```
monitoring/                                              # 整个监控目录
scripts/start-monitoring.sh                             # 监控启动脚本
scripts/test-monitoring.sh                              # 监控测试脚本
scripts/tests/test-monitoring-integration.sh           # 监控集成测试
frontend/src/features/monitoring/                       # 前端监控功能
docs/archive/project-reports/16-monitoring-infrastructure-report.md
docs/monitoring/
cmd/organization-command-service/internal/metrics/collector.go
cmd/organization-query-service/internal/metrics/metrics.go
```

### **需要编辑的文件**
```
frontend/src/App.tsx                                   # 移除监控路由
frontend/src/layout/Sidebar.tsx                        # 移除监控菜单
frontend/src/shared/config/ports.ts                    # 移除监控端点配置
Makefile                                               # 移除监控目标
CLAUDE.md                                              # 移除监控相关配置说明
README.md                                              # 移除监控使用说明
cmd/organization-command-service/main.go              # 移除监控初始化
cmd/organization-query-service/main.go                # 移除监控初始化
go.mod (两个服务)                                      # 移除prometheus依赖
```

## 🎯 **清理验证检查清单**

### **Docker环境验证**
- [ ] 确认所有监控容器已停止和删除
- [ ] 确认监控数据卷已删除
- [ ] 确认监控网络已删除
- [ ] 确认监控镜像已清理（可选）

### **文件系统验证**
- [ ] 确认monitoring/目录已完全删除
- [ ] 确认监控脚本已删除
- [ ] 确认监控文档已删除
- [ ] 确认后端监控代码已删除

### **前端集成验证**
- [ ] 确认监控功能目录已删除
- [ ] 确认App.tsx中监控路由已移除
- [ ] 确认Sidebar.tsx中监控菜单已移除
- [ ] 确认端口配置中监控端点已移除
- [ ] 确认前端应用正常启动无错误

### **后端集成验证**
- [ ] 确认监控指标收集器已删除
- [ ] 确认main.go中监控初始化已移除
- [ ] 确认go.mod中prometheus依赖已移除
- [ ] 确认后端服务正常启动无错误

### **开发工具验证**
- [ ] 确认Makefile中监控目标已移除
- [ ] 确认make help不显示监控相关命令
- [ ] 确认项目文档已更新

## ✅ **清理完成标准**

清理完成后，项目应达到以下状态：

1. **Docker环境**: 无任何监控相关容器、卷或网络
2. **文件系统**: 无监控相关文件或目录
3. **前端应用**: 正常启动，无监控页面或菜单，无相关错误
4. **后端服务**: 正常启动，仅保留基础/metrics端点（可选）
5. **开发工具**: make命令不包含监控相关目标
6. **项目文档**: 已更新，无监控相关说明

这个方案确保监控系统的彻底清理，同时保持项目核心功能完整性和开发环境的简洁性。

---

**方案制定者**: Claude Code架构分析专家  
**分析时间**: 2025-09-13 21:20  
**分析范围**: 完整监控系统组件评估  
**清理目标**: 彻底移除所有监控相关组件


## 🔎 可行性评估与补充执行要点（追加）

基于当前仓库实际结构与依赖关系复核，本方案可行且基本全面。但如需“完全移除 Prometheus 监控”，仍需补充以下精确变更以避免编译失败与文档残留。

### ✅ 评估结论
- 可行性：可行。文档中列举的路径、容器名、脚本与前后端集成点均与仓库一致，按既定阶段执行可达成彻底卸载。
- 全面性：基本全面。仍有少量代码引用与文档/脚本提及需一并清理，以下列出补充项。

### 🧩 必要补充（避免编译失败）
1) 后端 Query Service（GraphQL）
- 文件：`cmd/organization-query-service/internal/auth/graphql_middleware.go`
  - 移除对 `internal/metrics` 的导入：`gqlmetrics ".../internal/metrics"`
  - 删除调用：`gqlmetrics.RecordPermissionCheck(...)`
- 文件：`cmd/organization-query-service/main.go`
  - 移除 `/metrics` 路由注册与 `promhttp` 相关 import。
- 文件：`cmd/organization-query-service/internal/metrics/metrics.go`
  - 删除该文件（若完全移除 Prometheus）。
- 依赖整理：在 `cmd/organization-query-service/` 目录执行：
  - `go mod tidy`

2) 后端 Command Service（REST）
- 文件：`cmd/organization-command-service/main.go`
  - 移除中间件：`r.Use(metricsCollector.GetMetricsMiddleware())`
  - 移除 `/metrics` 路由与相关日志输出。
  - 移除 `internal/metrics` 包的 import 与初始化。
- 文件：`cmd/organization-command-service/internal/metrics/collector.go`
  - 删除该文件（若完全移除 Prometheus）。
- 根模块依赖整理：在仓库根目录执行：
  - `go mod tidy`

3) `/metrics` 端点保留策略说明（可选）
- 如需“保留基础 /metrics 端点”但移除自定义指标：保留 promhttp 处理器与最小依赖，删除自定义 `internal/metrics` 中间件与业务指标的注册/调用即可。此模式需同步调整文档声明为“最小指标保留”。

### 🎨 前端改动补充（保持 TS 编译通过）
- 删除：`frontend/src/features/monitoring/` 目录。
- App 路由：`frontend/src/App.tsx` 移除监控懒加载 import 与 `/monitoring` 路由。
- 侧边栏：`frontend/src/layout/Sidebar.tsx` 移除“系统监控”菜单项。
- 端口配置：`frontend/src/shared/config/ports.ts`
  - 删除 `SERVICE_PORTS` 中的监控端口项：`PROMETHEUS/GRAFANA/ALERT_MANAGER/NODE_EXPORTER`（如完全移除）。
  - 删除 `MONITORING_ENDPOINTS` 导出。
  - 在 `generatePortConfigReport()` 中删除“📊 监控服务”段落打印，避免引用被删的常量。
  - 类型导出保持一致（移除与监控相关的键名）。

### 📚 文档与脚本清理扩展范围（避免残留引用）
除已在方案中列出的 `README.md`、`CLAUDE.md`、`docs/monitoring/`、归档报告外，建议补充清理以下引用：
- `docs/guides/PRODUCTION-DEPLOYMENT-GUIDE.md`：包含前端 `/monitoring` 页面与 compose 操作指引。
- `docs/guides/MICROSERVICES_MANAGEMENT.md`：提到 `./scripts/start-monitoring.sh` 与 `./scripts/test-monitoring.sh`。
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md`：涉及监控端点、`MONITORING_ENDPOINTS` 说明与 compose 文件。
- `scripts/README.md`：包含监控脚本使用说明。

如仅“停用而非删除”监控能力，上述文档应改为“可选/默认关闭”的措辞，并保留最少可复原步骤。

### 🔍 扫描与验证（推荐命令）
- 代码/文档残留扫描：
  - `rg -n "prometheus|promhttp|/metrics|MONITORING_ENDPOINTS|/monitoring" cube-castle`
- 后端构建与测试：
  - 仓库根：`go mod tidy && make build && make test`
  - Query Service 子模块：`(cd cmd/organization-query-service && go mod tidy && go build ./...)`
- 前端构建：
  - `cd frontend && npm run build`

### 🧭 执行顺序（精简版）
1. 停止与清理 Docker 监控栈（按原“阶段1”）。
2. 删除监控目录与脚本（按原“阶段2/5”）。
3. 前端移除监控模块、路由与端口配置（“前端改动补充”）。
4. 后端两服务按“必要补充”章节进行代码修改与 `go mod tidy`。
5. 清理并更新文档（含扩展范围）。
6. 运行“扫描与验证”命令，直到无残留与编译错误。

### 🧯 回滚与保留建议
- 若未来可能恢复监控：建议将 `monitoring/` 与相关前端模块压缩归档到 `docs/archive/`，并在 `README.md` “扩展能力”中保留一段“如何恢复监控”的简述（链接到归档）。
- 若仅精简：保留 `/metrics` 最小端点与 `SERVICE_PORTS` 框架，但将监控端口设为注释或以配置开关禁用。

以上补充确保彻底移除监控组件后，代码与文档保持一致，避免构建失败与使用误导。
