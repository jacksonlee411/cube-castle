# 开发测试技术规范

## 服务进程管理规范

### 禁用危险命令

**严格禁止使用以下命令**：
```bash
# ❌ 禁用 - 可能终止无关进程
pkill -f server
pkill server

# ❌ 禁用 - 过于宽泛的进程匹配
killall server
```

### 推荐的安全替代方案

**1. 使用精确的PID管理**：
```bash
# ✅ 推荐 - 精确停止特定进程
kill $(cat server.pid)  # 使用PID文件
kill $(pgrep -f "./bin/server")  # 精确匹配可执行文件路径
```

**2. 使用进程名称精确匹配**：
```bash
# ✅ 推荐 - 精确匹配进程名
ps aux | grep "./bin/server" | grep -v grep | awk '{print $2}' | xargs kill
```

**3. 使用系统服务管理**：
```bash
# ✅ 推荐 - 使用systemd服务管理
systemctl stop cube-castle-api
systemctl restart cube-castle-api
```

### 理由说明

1. **安全性**: `pkill -f server` 可能意外终止其他包含"server"字样的进程
2. **精确性**: 应该只停止目标服务，不影响其他服务
3. **可控性**: 使用PID文件或精确路径匹配确保操作可控
4. **可维护性**: 明确的命令便于调试和维护

### 违规检查

开发团队应在代码审查中检查是否存在危险命令使用，确保遵循本规范。

---

*最后更新: 2025-08-04*
*版本: v1.0*