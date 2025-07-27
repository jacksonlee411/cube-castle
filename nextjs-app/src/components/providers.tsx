'use client'

import { ThemeProvider } from 'next-themes'
import { ReactNode } from 'react'
import { SWRConfig } from 'swr'
import { fetcher } from '@/lib/api'

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <SWRConfig
        value={{
          fetcher,
          errorRetryCount: 3,
          errorRetryInterval: 5000,
          revalidateOnFocus: false,
          revalidateOnReconnect: true,
          dedupingInterval: 2000,
        }}
      >
        {children}
      </SWRConfig>
    </ThemeProvider>
  )
}