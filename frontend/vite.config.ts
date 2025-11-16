/// <reference types="vitest" />
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'node:path'
import { SERVICE_PORTS, CQRS_ENDPOINTS } from './src/shared/config/ports'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // 显式加载环境变量并注入到客户端
  const env = loadEnv(mode, process.cwd(), '');

  return {
  // 显式定义环境变量以注入到客户端 (解决 Vite 环境变量注入问题)
  define: {
    'import.meta.env.VITE_AUTH_MODE': JSON.stringify(env.VITE_AUTH_MODE || 'oidc'),
  },
  plugins: [
    react(),
    // Dev-only health endpoint for E2E env validation
    {
      name: 'vite-health-endpoint',
      configureServer(server) {
        server.middlewares.use('/health', (_req, res) => {
          res.statusCode = 200;
          res.setHeader('Content-Type', 'application/json');
          res.end(JSON.stringify({ status: 'ok', service: 'frontend-dev', ts: Date.now() }));
        });
      }
    }
  ],
  
  // 开发性能优化
  server: {
    port: SERVICE_PORTS.FRONTEND_DEV,
    strictPort: true,
    hmr: { overlay: false },
    proxy: {
      // 时态管理API路由 - 符合严格CQRS架构
      // 事件驱动端点 (命令操作) 路由到REST命令服务
      '^/api/v1/organization-units/[^/]+/events': {
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      },
      // 历史记录更新端点 (命令操作) 路由到REST命令服务
      '^/api/v1/organization-units/history/[^/]+': {
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      },
      // ❌ 已移除时态查询端点 - 现在使用GraphQL查询
      // '^/api/v1/organization-units/[^/]+/temporal': 现统一使用 /graphql 端点
      // 认证端点代理
      '/auth': {
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      },
      // JWKS端点代理
      '/.well-known/jwks.json': {
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      },
      // 其他API路由到REST命令服务
      '/api/v1': {
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      },
      '/graphql': {
        // 合流到单体进程（:9090）后，GraphQL 由同一进程提供
        target: CQRS_ENDPOINTS.COMMAND_BASE,
        changeOrigin: true,
        secure: false
      }
    },
    // 静态文件服务 - 提供Playwright测试报告访问
    fs: {
      allow: ['..'] // 允许访问上级目录的文件
    }
  },

  // 静态文件配置
  publicDir: 'public',
  
  // 路径别名配置  
  resolve: {
    alias: {
      '@': resolve(process.cwd(), './src'),
      '@shared': resolve(process.cwd(), './src/shared'),
      '@layout': resolve(process.cwd(), './src/layout')
    }
  },
  
  // 预构建优化
  optimizeDeps: {
    include: [
      '@workday/canvas-kit-react',
      '@workday/canvas-tokens-web',
      '@workday/canvas-kit-react-fonts'
    ],
    exclude: [
      'playwright-core',
      '@playwright/test', 
      'chromium-bidi'
    ]
  },
  
  // 大型应用性能优化
  build: {
    target: 'es2015',
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom'],
          'vendor-canvas': ['@workday/canvas-kit-react'],
          'vendor-router': ['react-router-dom'],
          'vendor-state': ['zustand', '@tanstack/react-query']
        }
      }
    },
    chunkSizeWarningLimit: 1000
  },

  // 测试环境配置
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.ts',
    css: true,
    exclude: [
      'node_modules',
      'dist',
      'tests/e2e/**' // 排除E2E测试文件，这些由Playwright处理
    ],
    coverage: {
      reporter: ['text', 'html'],
      include: [
        'src/shared/api/queryClient.ts',
        'src/shared/api/type-guards.ts',
        'src/shared/api/graphql-enterprise-adapter.ts',
        'src/shared/hooks/useEnterpriseOrganizations.ts',
        'src/shared/hooks/useMessages.ts'
      ],
      exclude: ['src/shared/api/__tests__/**', 'src/shared/hooks/__tests__/**']
    }
  }
};
})
