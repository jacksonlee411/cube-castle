# DevOps 故障排除索引

**维护时间**: 2025年7月27日  
**目录说明**: 本目录包含 Cube Castle 项目的所有 DevOps 相关故障排除文档

---

## 📁 文档分类

### 🔧 服务配置与部署
- [Temporal服务问题完整解决方案报告](./Temporal_Complete_Solution_Report.md) - ✅ 完整解决方案
  - 问题: Temporal UI 500错误 + "Frontend is not healthy yet" + 无限重启循环
  - 解决: 深度根本原因分析 + 简化配置策略 + 数据库驱动修正
  - 内容: 完整的诊断过程、解决方案、配置文件、最佳实践
  - 状态: 生产就绪，知识完整沉淀
- [Temporal服务深度根本原因分析报告](./Deep_Root_Cause_Analysis.md) - ✅ 技术深度分析
  - 专项: 深层技术分析和诊断方法论
  - 发现: Worker服务启动失败 + Auto-setup脚本冲突详细分析
  - 状态: 技术知识沉淀完成
- [服务标准化完成报告](./Service_Standardization_Report.md) - ✅ 运维操作记录
  - 操作: 服务清理与标准化，统一命名规范
  - 结果: 所有服务恢复标准命名，配置集成到主文件
  - 状态: 标准化完成

### 🐳 Docker & 容器化
- Docker网络配置冲突解决方案 (计划中)
- 容器健康检查最佳实践 (计划中)
- 多服务容器编排优化 (计划中)

### 🗄️ 数据库相关
- PostgreSQL连接问题排查 (计划中)
- 数据库驱动配置指南 (计划中)
- 数据迁移与备份策略 (计划中)

### 🌐 网络与通信
- 服务间通信故障排查 (计划中)
- 端口冲突解决方案 (计划中)
- API网关配置问题 (计划中)

### 📊 监控与日志
- 容器日志分析技巧 (计划中)
- 系统监控配置 (计划中)
- 性能瓶颈识别 (计划中)

---

## 🚨 紧急故障处理流程

### 1. 问题识别
- [ ] 确定影响范围
- [ ] 收集错误日志
- [ ] 检查服务状态
- [ ] 验证网络连接

### 2. 快速诊断
- [ ] 查阅相关故障文档
- [ ] 检查最近的配置变更
- [ ] 验证依赖服务状态
- [ ] 分析错误模式

### 3. 解决方案实施
- [ ] 应用已知解决方案
- [ ] 测试修复效果
- [ ] 验证服务恢复
- [ ] 记录解决过程

### 4. 事后总结
- [ ] 更新故障文档
- [ ] 完善监控配置
- [ ] 制定预防措施
- [ ] 团队知识分享

---

## 🛠️ 常用工具与命令

### Docker 诊断
```bash
# 查看容器状态
docker ps -a

# 查看容器日志
docker logs [container_name] --tail 50

# 检查网络配置
docker network ls
docker network inspect [network_name]

# 资源使用情况
docker stats --no-stream
```

### 服务健康检查
```bash
# Temporal服务
curl -s http://localhost:8087/api/v1/cluster-info

# PostgreSQL连接
docker exec [postgres_container] pg_isready -U [username] -d [database]

# 端口占用检查
netstat -tulpn | grep [port]
```

### 日志分析
```bash
# 过滤错误日志
docker logs [container] 2>&1 | grep -i error

# 实时日志监控
docker logs -f [container]

# 特定时间段日志
docker logs [container] --since "2025-07-27T00:00:00" --until "2025-07-27T23:59:59"
```

---

## 📚 相关资源

### 官方文档
- [Temporal官方文档](https://docs.temporal.io/)
- [Docker Compose参考](https://docs.docker.com/compose/)
- [PostgreSQL文档](https://www.postgresql.org/docs/)

### 内部文档
- [开发问题总结与最佳实践](../开发问题总结与最佳实践.md)
- [脚本开发规范](../脚本开发规范.md)
- [开发快速参考卡片](../开发快速参考卡片.md)

### 社区资源
- [Temporal社区论坛](https://community.temporal.io/)
- [Docker故障排除指南](https://docs.docker.com/config/containers/logging/)

---

## 🔄 文档维护

### 更新频率
- 紧急故障: 立即更新
- 常规问题: 每周汇总
- 最佳实践: 每月回顾

### 贡献指南
1. 遇到新问题时，先查阅现有文档
2. 解决问题后，及时更新相关文档
3. 定期review和完善文档内容
4. 与团队分享经验和教训

### 文档模板
创建新故障排除文档时，请使用以下结构：
- 问题描述和症状
- 详细诊断过程
- 解决方案步骤
- 验证方法
- 经验总结
- 预防措施

---

**最后更新**: 2025年7月27日  
**维护人员**: SuperClaude DevOps团队  
**联系方式**: 项目内部沟通渠道