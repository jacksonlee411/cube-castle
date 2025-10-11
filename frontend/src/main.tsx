import React from 'react'
import ReactDOM from 'react-dom/client'
import { CanvasProvider } from '@workday/canvas-kit-react/common'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from './shared/auth/AuthProvider'
import { fonts } from '@workday/canvas-kit-react-fonts'
import { system } from '@workday/canvas-tokens-web'
import { injectGlobal } from '@emotion/css'
import { cssVar } from '@workday/canvas-kit-styling'
import { queryClient } from '@/shared/api'

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

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <CanvasProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AuthProvider>
            <App />
          </AuthProvider>
        </BrowserRouter>
        {import.meta.env.DEV ? (
          <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-right" />
        ) : null}
      </QueryClientProvider>
    </CanvasProvider>
  </React.StrictMode>
)
