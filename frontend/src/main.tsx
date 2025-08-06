import React from 'react'
import ReactDOM from 'react-dom/client'
import { CanvasProvider } from '@workday/canvas-kit-react/common'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { fonts } from '@workday/canvas-kit-react-fonts'
import { system } from '@workday/canvas-tokens-web'
import { injectGlobal } from '@emotion/css'
import { cssVar } from '@workday/canvas-kit-styling'

// Canvas CSS变量导入
import '@workday/canvas-tokens-web/css/base/_variables.css'
import '@workday/canvas-tokens-web/css/brand/_variables.css'
import '@workday/canvas-tokens-web/css/system/_variables.css'

import App from './App'

// Canvas全局样式注入
injectGlobal({
  ...fonts,
  'html, body': {
    fontFamily: cssVar(system.fontFamily.default),
    margin: 0,
    minHeight: '100vh',
    backgroundColor: cssVar(system.color.bg.default)
  },
  '#root': {
    minHeight: '100vh',
    ...system.type.body.medium
  }
})

// 创建 React Query 客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5分钟
      retry: 2,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <CanvasProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </QueryClientProvider>
    </CanvasProvider>
  </React.StrictMode>
)
