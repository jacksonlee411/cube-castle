# 时态数据运维脚本指南

这个目录包含了用于维护时态数据一致性的运维脚本和工具。

## 📁 脚本文件

### 核心SQL脚本

| 文件名 | 用途 | 执行频率 | 说明 |
|--------|------|----------|------|
| `daily-cutover.sql` | 每日cutover任务 | 每日凌晨2点 | 维护is_current和is_future状态标志 |
| `data-consistency-check.sql` | 数据一致性检查 | 每4小时或按需 | 检查时态数据的完整性和一致性 |

### 运维工具脚本

| 文件名 | 用途 | 执行方式 | 说明 |
|--------|------|----------|------|
| `setup-cron.sh` | 设置cron定时任务 | 系统管理员手动执行 | 配置系统级定时任务 |

## 🔧 使用指南

### 1. 设置定时任务（一次性操作）

```bash
# 以root权限执行
sudo bash setup-cron.sh
```

这将自动：
- 创建系统级cron任务
- 设置日志目录 `/var/log/cubecastle`
- 复制脚本到 `/opt/cubecastle/scripts/`
- 创建配置文件 `/etc/default/cubecastle`

### 2. 手动执行维护任务

```bash
# 手动执行每日cutover
sudo /usr/local/bin/cubecastle-daily-cutover.sh

# 手动执行一致性检查
sudo /usr/local/bin/cubecastle-consistency-check.sh
```

### 3. 直接执行SQL脚本

```bash
# 设置数据库连接环境变量
export PGPASSWORD="password"

# 执行cutover任务
psql -h localhost -p 5432 -U user -d cubecastle -f daily-cutover.sql

# 执行一致性检查
psql -h localhost -p 5432 -U user -d cubecastle -f data-consistency-check.sql
```

## 📊 监控和日志

### 日志文件位置

- **每日任务日志**: `/var/log/cubecastle/daily-cutover-YYYYMMDD.log`
- **一致性检查日志**: `/var/log/cubecastle/consistency-check-YYYYMMDD-HHMM.log`
- **错误日志**: `/var/log/cubecastle/*-error.log`

### 查看日志

```bash
# 查看最新的每日任务日志
sudo tail -f /var/log/cubecastle/daily-cutover-$(date +%Y%m%d).log

# 查看一致性检查日志
sudo ls -la /var/log/cubecastle/consistency-check-*

# 查看错误日志
sudo tail -20 /var/log/cubecastle/*-error.log
```

### API端点监控

项目还提供了RESTful API端点来监控运维任务状态：

```bash
# 获取系统健康状态
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/health

# 获取详细监控指标
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/metrics

# 获取当前告警
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/alerts

# 获取任务状态
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/tasks/status

# 手动触发cutover任务
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/cutover

# 手动触发一致性检查
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:9090/api/v1/operational/consistency-check
```

## 🚨 告警机制

### 内置告警规则

系统内置了以下告警规则：

| 告警类型 | 阈值 | 级别 | 说明 |
|----------|------|------|------|
| 重复当前记录 | 0 | CRITICAL | 任何重复都是严重问题 |
| 缺失当前记录 | 0 | CRITICAL | 任何缺失都是严重问题 |
| 时间线重叠 | 0 | CRITICAL | 任何重叠都是严重问题 |
| 标志不一致 | 5 | WARNING | 少量不一致可能是时间差导致 |
| 孤立记录 | 10 | WARNING | 少量孤立记录可以接受 |
| 健康分数 | 85 | WARNING | 健康分数低于85%告警 |

### 告警处理

当发现告警时，请按以下步骤处理：

1. **立即检查**: 通过API或日志查看具体错误详情
2. **数据分析**: 运行一致性检查脚本获取详细分析
3. **手动修复**: 根据错误类型进行相应的数据修复
4. **验证修复**: 重新运行检查脚本确认问题已解决

## 🔒 安全注意事项

1. **权限管理**: 运维脚本需要数据库写权限，请严格控制访问权限
2. **密码安全**: 数据库密码存储在 `/etc/default/cubecastle`，权限为600
3. **日志敏感信息**: 日志文件可能包含敏感信息，请适当保护
4. **备份策略**: 执行维护任务前建议先备份相关数据

## 📈 性能优化建议

1. **执行时间**: 选择业务低峰期执行维护任务（如凌晨2点）
2. **监控频率**: 根据数据变化频率调整监控检查间隔
3. **索引优化**: 确保时态查询相关的数据库索引已正确创建
4. **日志清理**: 系统自动清理过期日志，避免磁盘空间不足

## 🆘 故障排查

### 常见问题

**问题1**: Cron任务未执行
- 检查cron服务状态: `sudo systemctl status cron`
- 查看cron日志: `sudo journalctl -u cron`
- 验证任务配置: `sudo cat /etc/cron.d/cubecastle-temporal`

**问题2**: 数据库连接失败
- 检查环境变量配置: `sudo cat /etc/default/cubecastle`
- 测试数据库连接: `PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1;"`
- 检查网络和防火墙设置

**问题3**: 一致性检查失败
- 查看详细错误日志: `sudo tail -50 /var/log/cubecastle/consistency-check-error.log`
- 手动运行检查脚本查看具体错误信息
- 根据错误类型进行针对性修复

### 紧急恢复

如果数据一致性严重受损，请：

1. **停止应用写入**: 临时停止所有写入操作
2. **数据库备份**: 立即创建当前状态的数据库备份
3. **专家支持**: 联系技术专家进行深度分析
4. **分步恢复**: 制定详细的数据恢复计划并分步执行

## 📞 技术支持

如需技术支持，请提供以下信息：
- 错误发生时间
- 相关日志文件内容
- 系统环境信息
- 最近执行的操作记录