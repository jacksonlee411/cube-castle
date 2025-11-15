# Docker权限问题快速修复指南

## 症状
运行 `docker ps` 或 `make run-dev` 时出现：
```
permission denied while trying to connect to the Docker daemon socket 
at unix:///var/run/docker.sock
```

## 一句话解决（推荐）
```bash
sudo usermod -aG docker $(whoami) && newgrp docker && docker ps
```

## 详细步骤

### 步骤1：诊断当前状态
```bash
bash -c 'echo "=== Docker 权限诊断 ==="; echo "1. Docker daemon:"; docker ps 2>&1 | head -2; echo "2. socket信息:"; ls -la /var/run/docker.sock 2>/dev/null || echo "不存在"; echo "3. 当前用户:"; whoami; echo "4. 用户所在组:"; groups; echo "5. docker组成员:"; getent group docker || echo "docker组不存在"'
```

### 步骤2：执行权限修复

**方案A（推荐）：将用户加入docker组**
```bash
# 创建docker组（如果不存在）
sudo groupadd docker 2>/dev/null || true

# 将当前用户加入docker组
sudo usermod -aG docker $(whoami)

# 立即生效（无需重启）
newgrp docker

# 验证
docker ps
```

**方案B（如果A失败）：修改socket权限**
```bash
# 临时方案（重启后失效）
sudo chmod 666 /var/run/docker.sock

# 永久方案：编辑daemon配置
sudo nano /etc/docker/daemon.json
```

添加以下内容：
```json
{
  "unix-socket-group": "docker",
  "unix-socket-permissions": "0660"
}
```

重启daemon：
```bash
sudo systemctl restart docker
newgrp docker
docker ps
```

### 步骤3：验证修复成功
```bash
# 应能成功列出容器（即使为空）
docker ps

# 应能成功列出本地镜像
docker images | head -5

# 启动全栈服务（可选）
make run-dev
```

## 常见问题

| 问题 | 解决方法 |
|------|--------|
| 执行 `usermod` 后仍 permission denied | 运行 `newgrp docker` 或关闭终端重新打开 |
| 新终端仍报 permission denied | 关闭所有终端窗口，完全重启终端应用 |
| WSL中仍不工作 | `sudo service docker status` / `sudo service docker restart` |
| 显示docker组不存在 | 运行 `sudo groupadd docker` 创建组 |

## 相关文档
- 完整配置指南：`docs/development-plans/06-integrated-teams-progress-log.md`
- 原始阻塞说明：`logs/219E/BLOCKERS-2025-11-06.md`
- 后续E2E测试：`scripts/e2e/org-lifecycle-smoke.sh`
- 性能基准：`scripts/perf/rest-benchmark.sh`
