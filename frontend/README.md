# 🏰 Cube Castle Frontend - 企业级React应用

## 🚀 统一配置架构 ⭐ **S级架构成果 (2025-09-07)**

基于React 19 + Canvas Kit v13 + TypeScript的现代化前端应用，采用统一配置管理和企业级架构标准。

### ✅ **企业级端口配置管理**
**权威配置源**: `src/shared/config/ports.ts`
```typescript
export const SERVICE_PORTS = {
  FRONTEND_DEV: 3000,           // 开发服务器
  FRONTEND_PREVIEW: 3001,       // 预览服务器  
  REST_COMMAND_SERVICE: 9090,   // CQRS命令服务
  GRAPHQL_QUERY_SERVICE: 8090,  // CQRS查询服务
  POSTGRESQL: 5432,
  REDIS: 6379
} as const;
```

### ✅ **重复代码消除完成**
- **Hook统一**: 7→2个Hook实现 (71%重复消除)
- **API客户端统一**: 6→1个客户端 (83%重复消除)  
- **类型系统重构**: 90+→8个核心接口 (80%+重复消除)
- **端口配置集中**: 15+文件→1个统一配置 (95%+硬编码消除)

### 🔧 技术栈
- **构建工具**: Vite 7.0+ (统一配置支持)
- **UI框架**: React 19 + Canvas Kit v13 + TypeScript 5.8+
- **状态管理**: TanStack Query + Zustand
- **测试**: Playwright + Vitest

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      ...tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      ...tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      ...tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
