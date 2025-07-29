# 元合约编辑器快速启动指南

## 🚀 快速启动

### 先决条件
- Docker & Docker Compose
- Git
- 8GB+ RAM推荐

### 1. 克隆项目
```bash
cd cube-castle
git pull origin main
```

### 2. 启动开发环境
```bash
# 启动完整开发环境
docker-compose -f docker-compose.editor-dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.editor-dev.yml ps
```

### 3. 等待服务就绪
```bash
# 检查数据库连接
docker-compose -f docker-compose.editor-dev.yml logs postgres

# 检查Go应用启动
docker-compose -f docker-compose.editor-dev.yml logs go-app

# 检查前端应用
docker-compose -f docker-compose.editor-dev.yml logs nextjs-app
```

### 4. 访问应用
- **编辑器界面**: http://localhost:3000/metacontract-editor
- **API文档**: http://localhost:8080/swagger (如果配置)
- **数据库**: localhost:5432 (用户: cube_castle, 密码: dev_password_123)

## 🔧 开发工作流

### 代码修改
- **Go代码**: 修改后自动热重载 (Air)
- **React代码**: 修改后自动刷新 (Next.js Dev Server)
- **数据库**: 迁移文件在容器启动时自动执行

### 日志查看
```bash
# 查看所有服务日志
docker-compose -f docker-compose.editor-dev.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.editor-dev.yml logs -f go-app
docker-compose -f docker-compose.editor-dev.yml logs -f nextjs-app
```

### 数据库操作
```bash
# 连接到PostgreSQL
docker exec -it cube-castle-postgres-dev psql -U cube_castle -d cube_castle_dev

# 运行SQL查询
\dt  # 列出所有表
SELECT * FROM metacontract_editor_projects;
```

## 🐛 故障排除

### 端口冲突
如果端口被占用，修改 docker-compose.editor-dev.yml 中的端口映射：
```yaml
ports:
  - "3001:3000"  # 前端
  - "8081:8080"  # 后端
```

### 数据库连接问题
```bash
# 重启数据库服务
docker-compose -f docker-compose.editor-dev.yml restart postgres

# 检查数据库健康状态
docker-compose -f docker-compose.editor-dev.yml exec postgres pg_isready -U cube_castle
```

### 前端编译错误
```bash
# 清理node_modules并重新安装
docker-compose -f docker-compose.editor-dev.yml exec nextjs-app rm -rf node_modules
docker-compose -f docker-compose.editor-dev.yml exec nextjs-app npm install
```

## 📝 开发提示

### 推荐的开发流程
1. 启动Docker环境
2. 在IDE中打开项目文件夹
3. 修改代码 (自动热重载)
4. 在浏览器中测试
5. 使用Git提交变更

### 调试技巧
- **Go调试**: 在代码中添加 `fmt.Printf()` 或使用专业调试器
- **React调试**: 使用浏览器开发者工具
- **WebSocket调试**: 使用浏览器Network面板查看WS连接
- **数据库调试**: 直接连接PostgreSQL查看数据

### 性能优化
- **编译时间**: 使用增量编译减少等待时间
- **前端性能**: 使用React DevTools分析组件渲染
- **数据库性能**: 查看慢查询日志
- **内存使用**: 监控Docker容器资源使用

## 🔄 更新和维护

### 拉取最新代码
```bash
git pull origin main
docker-compose -f docker-compose.editor-dev.yml down
docker-compose -f docker-compose.editor-dev.yml up -d --build
```

### 清理开发环境
```bash
# 停止所有服务
docker-compose -f docker-compose.editor-dev.yml down

# 清理数据卷 (注意: 会删除所有数据)
docker-compose -f docker-compose.editor-dev.yml down -v

# 清理Docker镜像
docker system prune -a
```

## 📧 支持

如遇到问题，请检查:
1. Docker服务是否正常运行
2. 端口是否被其他应用占用
3. 系统资源是否充足
4. 网络连接是否正常

更多技术细节请参考 `METACONTRACT_EDITOR_IMPLEMENTATION.md`