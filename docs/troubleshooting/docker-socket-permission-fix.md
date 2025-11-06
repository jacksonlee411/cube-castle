# Docker Socket 权限问题诊断与修复指南

> 当前版本：2025-11-07  
> 适用场景：`/var/run/docker.sock` permission denied 错误  
> 权限级别：需要 `sudo` 权限（或具有相应能力的系统管理员协助）

## 快速诊断

执行以下一键诊断脚本，获取当前Docker环境状态：

```bash
bash -c 'echo "=== Docker 权限诊断 ==="; echo "1. Docker daemon:"; docker ps 2>&1 | head -2; echo "2. socket信息:"; ls -la /var/run/docker.sock 2>/dev/null || echo "不存在"; echo "3. 当前用户:"; whoami; echo "4. 用户所在组:"; groups; echo "5. docker组成员:"; getent group docker || echo "docker组不存在"'
```

**输出示例（问题情况）：**
```
=== Docker 权限诊断 ===
1. Docker daemon:
Got permission denied while trying to connect to Docker daemon
2. socket信息:
srw-rw---- 1 root docker 0 Nov  7 10:30 /var/run/docker.sock
3. 当前用户:
shangmeilin
4. 用户所在组:
shangmeilin wheel
5. docker组成员:
docker:x:999:
```

**诊断结果解读：**
- Socket权限为 `660`（仅 `root:docker` 可访问）
- 当前用户 `shangmeilin` 不在 `docker` 组内 → 无法访问
- 解决方案：将用户加入 `docker` 组（方案A）

---

## 修复方案

### **方案A：Linux用户组授权（推荐 - 一次性配置）**

> 适用于：多数Linux环境（包括WSL2）  
> 原理：将用户加入 `docker` 组，获得socket访问权限

**步骤：**

```bash
# 1. 创建 docker 用户组（如果不存在）
sudo groupadd docker 2>/dev/null || true

# 2. 将当前用户加入 docker 组
sudo usermod -aG docker $(whoami)

# 3. 更新用户组成员关系（重新获取组权限）
newgrp docker

# 4. 验证修复成功
docker ps
```

**预期输出：**
```
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```
（列表可能为空，但无 permission denied 错误即为成功）

---

### **方案B：修改Socket权限（备选方案）**

> 适用于：无sudo权限或用户组授权失效的情况  
> 特点：临时性或持久化配置

**临时修复（重启后失效）：**
```bash
sudo chmod 666 /var/run/docker.sock
docker ps  # 立即生效
```

**永久修复（编辑daemon配置）：**
```bash
# 1. 编辑Docker daemon配置
sudo nano /etc/docker/daemon.json

# 2. 添加或修改以下内容：
{
  "unix-socket-group": "docker",
  "unix-socket-permissions": "0660"
}

# 3. 重启Docker daemon
sudo systemctl restart docker

# 4. 重新加入docker组
newgrp docker

# 5. 验证
docker ps
```

---

### **方案C：CI/CD环境（无需本地配置）**

> 适用于：GitHub Actions、GitLab CI 等云端运行环境  
> 优点：避免本地权限问题，由平台统一管理

在 `.github/workflows/` 中使用 `ubuntu-latest` runner，Docker daemon 已预装且运行用户已有访问权限：

```yaml
runs-on:
  - ubuntu-latest  # Docker daemon 已就绪，无需额外配置

steps:
  - name: Run E2E tests
    run: scripts/e2e/org-lifecycle-smoke.sh
```

---

## 常见问题排查

| 问题 | 原因 | 解决方法 |
|------|------|--------|
| 执行 `usermod` 后仍 permission denied | 新组关系需重新加载 | `newgrp docker` 或关闭并重新打开终端 |
| 新终端仍报 permission denied | 新shell未加载更新的组信息 | `exec su -l $USER` 或重启终端 |
| WSL2中重启Docker后仍不工作 | Docker daemon 未正确启动 | `sudo service docker status` / `sudo service docker restart` |
| `usermod: user shangmeilin is already in group docker` | 用户已在docker组（配置生效前的状态） | 重新登录或 `newgrp docker` |
| `sudo: command not found` | 环境无sudo权限（特殊容器/受限环境） | 需系统管理员协助执行上述步骤，或使用方案C（CI/CD runner） |

---

## 完整验证清单

修复后执行以下验证，确保Docker完全可用：

```bash
# 1. 验证基本命令
docker ps

# 2. 验证镜像操作
docker images | grep cube-castle

# 3. 启动全栈服务
make run-dev
# 或
docker compose -f docker-compose.dev.yml up -d

# 4. 验证关键服务运行
docker ps | grep -E 'postgres|redis|rest-service|graphql-service'
```

**成功标准：**
- ✅ 所有命令无 `permission denied` 错误
- ✅ `docker ps` 显示正在运行的容器列表
- ✅ `make run-dev` 启动核心服务（postgres、redis、API服务等）

---

## 背景知识

### Docker Socket权限机制

- `/var/run/docker.sock` 是Docker daemon的Unix socket，权限通常为 `660`（读写权限仅限 `root:docker`）
- 用户需要在 `docker` 组内才能访问，或由系统管理员修改socket权限
- 生产环境推荐使用用户组授权（方案A），确保权限管理清晰且安全

### WSL2环境特殊说明

- WSL2（Linux 5.15.x）中，Docker daemon 通常以 `root:docker` 运行
- 本地 `sudo` 因 `no new privileges` 限制可能不可用，但用户组授权（方案A）通常有效
- 如遇sudo失效，考虑在Windows侧使用具备管理员权限的PowerShell执行等价操作

---

## 获取帮助

- **本地诊断失败**：运行快速诊断脚本，将输出提交至 issue
- **持续报错**：参考"常见问题排查"表格，或在CI/CD中使用方案C（GitHub Actions runner）
- **权限受限环境**：联系系统管理员使用方案A/B，或转向CI/CD pipeline执行（方案C）
