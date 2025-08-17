/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  
  // 开发性能优化
  server: {
    port: 3000,
    hmr: { overlay: false },
    proxy: {
      '/api/metrics': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/metrics/, '/metrics')
      },
      '/api/debezium': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/debezium/, '')
      },
      // 时态管理API路由到9091端口 - 事件驱动端点 (必须在通用路由之前)
      '^/api/v1/organization-units/[^/]+/events': {
        target: 'http://localhost:9091',
        changeOrigin: true,
        secure: false
      },
      '^/api/v1/organization-units/[^/]+/temporal': {
        target: 'http://localhost:9091',
        changeOrigin: true,
        secure: false
      },
      // 其他API路由到9090端口
      '/api/v1': {
        target: 'http://localhost:9090',
        changeOrigin: true,
        secure: false
      },
      '/graphql': {
        target: 'http://localhost:8090',
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
    ]
  }
})