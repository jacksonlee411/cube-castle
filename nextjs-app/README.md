# Cube Castle Next.js 前端应用

这是 Cube Castle 企业级 HR 管理平台的 Next.js 前端应用。

## ✨ 最新更新 (v1.5.0)

### 🎯 完整CRUD功能已恢复 
- ✅ **员工管理页面** - 完整CRUD操作，支持高级筛选和分页
- ✅ **组织架构管理** - 树形视图和表格视图，层级关系管理
- ✅ **职位管理页面** - FTE预算管理，利用率统计和状态跟踪
- ✅ **员工职位历史** - 时间线视图，工作流跟踪，GraphQL错误已修复
- ✅ **主页导航** - 新增职位管理入口，优化用户体验

### 🔧 技术改进
- **GraphQL依赖移除** - 职位历史页面使用独立实现，提高稳定性
- **ES模块兼容性** - 完全解决Ant Design兼容性问题
- **UAT测试就绪** - 所有核心功能可进行完整测试

## 🚀 技术栈

- **Framework**: Next.js 14.1.4 (稳定版本)
- **Language**: TypeScript
- **UI Library**: Ant Design 5.20.6 (已优化兼容性)
- **Styling**: Tailwind CSS
- **Icons**: @ant-design/icons 5.3.7 + Lucide React + Heroicons
- **State Management**: Zustand
- **Data Fetching**: SWR
- **Forms**: React Hook Form + Zod
- **Testing**: Jest + Testing Library + Playwright
- **Linting**: ESLint + Prettier

> ⚠️ **版本说明**: 当前使用稳定版本组合以确保ES模块兼容性，详见 `ES_MODULE_COMPATIBILITY_FIX_REPORT.md`

## 📦 核心依赖版本 (已验证兼容)

```json
{
  "next": "14.1.4",
  "antd": "5.20.6", 
  "@ant-design/icons": "5.3.7",
  "rc-util": "5.38.2",
  "@rc-component/util": "1.1.0"
}
```

**🔒 版本锁定**: 这些版本已经过完整的ES模块兼容性测试，请勿随意升级。

## 📁 项目结构

```
src/
├── pages/                  # Next.js Pages Router
│   ├── index.tsx          # 主页 (包含所有模块导航)
│   ├── employees/         # 员工管理模块
│   │   ├── index.tsx     # 员工列表页 (完整CRUD)
│   │   └── positions/    # 员工职位历史
│   │       └── [id].tsx  # 职位历史详情 (已修复GraphQL错误)
│   ├── organization/      # 组织架构模块  
│   │   └── chart.tsx     # 组织架构管理 (完整功能)
│   ├── positions/         # 职位管理模块
│   │   └── index.tsx     # 职位管理页 (完整CRUD + FTE管理)
│   ├── sam/              # SAM仪表板
│   ├── workflows/        # 工作流管理
│   └── test-antd.tsx     # Ant Design兼容性测试页
├── components/            # React 组件
│   ├── ui/               # 基础 UI 组件
│   └── providers.tsx     # 全局 Provider
├── lib/                  # 工具库
│   ├── api.ts           # API 客户端
│   └── utils.ts         # 工具函数
├── hooks/               # 自定义 Hooks
│   └── useEmployees.ts  # 员工数据管理Hook
├── store/               # 状态管理
├── types/               # TypeScript 类型定义
│   ├── index.ts         # 核心类型定义
│   └── position.ts      # 职位相关类型
└── api/                 # API 接口定义
```

### 🎯 核心功能页面
- **`/`** - 主页仪表板，包含所有模块快速访问
- **`/employees`** - 员工管理，完整CRUD操作
- **`/organization/chart`** - 组织架构可视化管理
- **`/positions`** - 职位管理，FTE预算和利用率分析
- **`/employees/positions/[id]`** - 员工职位变更历史和工作流跟踪

## 🛠️ 开发指南

### 环境要求

- Node.js 18+
- npm 8+

### 安装依赖

```bash
npm install
```

### 开发服务器

```bash
npm run dev
```

应用将在 [http://localhost:3000](http://localhost:3000) 启动。

### 构建生产版本

```bash
npm run build
npm start
```

### 代码质量

```bash
# 类型检查
npm run type-check

# 代码检查
npm run lint

# 代码格式化
npx prettier --write .
```

### 测试

```bash
# 单元测试
npm test

# E2E 测试
npm run test:e2e

# 可视化回归测试
npm run test:visual
```

## 🏗️ 架构设计

### 组件设计原则

1. **单一职责**: 每个组件只负责一个功能
2. **可复用性**: 组件应该是可复用的
3. **类型安全**: 使用 TypeScript 确保类型安全
4. **可访问性**: 遵循 ARIA 标准

### 状态管理

- **本地状态**: 使用 React hooks (useState, useReducer)
- **服务器状态**: 使用 SWR 进行数据获取和缓存
- **全局状态**: 使用 Zustand 进行轻量级状态管理

### API 通信

- **HTTP 客户端**: Axios 
- **认证**: JWT Token
- **错误处理**: 统一错误处理和用户提示
- **重试机制**: 自动重试失败的请求

## 🎨 设计系统

### 主题配置

支持浅色和深色主题，使用 CSS 变量进行主题切换。

### 组件库

基于 Tailwind CSS 构建的企业级组件库：

- **Button**: 按钮组件
- **Card**: 卡片组件
- **Badge**: 徽章组件
- **Input**: 输入框组件
- **Table**: 表格组件
- **Modal**: 模态框组件

### 响应式设计

遵循移动优先的设计原则，支持桌面、平板和手机端。

## 🔐 安全特性

- **XSS 防护**: 输入验证和输出编码
- **CSRF 防护**: 请求令牌验证
- **内容安全策略**: CSP 头部配置
- **HTTPS**: 强制使用 HTTPS

## 📊 性能优化

- **代码分割**: 自动代码分割和懒加载
- **图片优化**: Next.js Image 组件优化
- **缓存策略**: 合理的缓存策略
- **预加载**: 关键资源预加载

## 🚀 部署

### Vercel 部署 (推荐)

```bash
npm install -g vercel
vercel
```

### Docker 部署

```bash
docker build -t cube-castle-nextjs .
docker run -p 3000:3000 cube-castle-nextjs
```

### 环境变量

```bash
# API 配置
CUBE_CASTLE_API_URL=http://localhost:8080
CUBE_CASTLE_WS_URL=ws://localhost:8080

# 应用配置
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_APP_NAME="Cube Castle"

# 第三方服务
NEXT_PUBLIC_ANALYTICS_ID=your-analytics-id
```

## 📈 监控和分析

- **性能监控**: Web Vitals 监控
- **错误追踪**: 错误边界和错误报告
- **用户分析**: 用户行为分析
- **A/B 测试**: 功能开关和测试

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范

- 使用 TypeScript 进行类型检查
- 遵循 ESLint 和 Prettier 配置
- 组件使用 PascalCase 命名
- 文件和目录使用 kebab-case 命名
- 提交信息遵循 Conventional Commits

## 📄 许可证

MIT License - 查看 [LICENSE](../LICENSE) 文件了解详情。

## 🆘 支持

- 📧 邮箱: frontend@cubecastle.com
- 📖 文档: [前端开发文档](./docs/)
- 🐛 问题反馈: [Issues](../../issues)
- 💬 讨论区: [Discussions](../../discussions)

---

**🏰 让企业级 HR 管理变得智能、安全、高效！**