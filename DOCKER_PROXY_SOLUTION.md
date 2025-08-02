# 🚨 Docker代理问题解决方案

## 问题确认
- ✅ 代理服务 `127.0.0.1:7890` 正常工作
- ✅ curl可以通过代理访问Docker Registry
- ❌ Docker daemon无法通过代理拉取镜像

## 立即解决方案

### 在Windows端执行以下步骤：

1. **打开Docker Desktop**
   - 右键点击系统托盘中的Docker鲸鱼图标
   - 选择 "Settings" 或 "设置"

2. **进入代理设置**
   - 点击左侧菜单 "Resources" 
   - 点击 "Proxies"

3. **禁用代理配置**
   - 找到 "Manual proxy configuration" 选项
   - **取消勾选** 这个选项
   - 点击 "Apply & Restart"

4. **等待重启**
   - Docker Desktop会重启
   - 等待重启完成（通常1-2分钟）

## 验证步骤

重启完成后，在WSL中执行：

```bash
# 检查代理配置是否已清除
docker system info | grep -i proxy

# 测试镜像拉取
docker pull hello-world

# 如果成功，继续Operation Phoenix
make phoenix-start
```

## 为什么要禁用代理？

1. **代理冲突**: Docker Desktop的代理配置与实际代理服务有协议不匹配
2. **WSL网络模式**: mirrored模式下的网络配置复杂性
3. **直连可行**: 大多数网络环境下Docker可以直接访问Registry

## 后续优化

如果需要代理（如企业网络），可以：
1. 使用国内镜像源
2. 配置Docker daemon.json
3. 或使用企业内部Registry

---

**请现在执行上述步骤，然后告诉我结果！**