import type { Metadata } from 'next'
import { Inter, JetBrains_Mono } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from 'react-hot-toast'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans',
  display: 'swap',
})

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  display: 'swap',
})

export const metadata: Metadata = {
  title: {
    default: 'Cube Castle - 企业级HR管理平台',
    template: '%s | Cube Castle'
  },
  description: 'Cube Castle 是一个基于城堡模型架构的现代化企业级 HR SaaS 平台，集成了人工智能驱动的自然语言交互、分布式工作流编排、企业级安全架构和全面的系统监控。',
  keywords: [
    'HR管理',
    '企业软件',
    'SaaS平台',
    '人力资源',
    '员工管理',
    '组织架构',
    '工作流',
    '智能AI'
  ],
  authors: [{ name: 'Cube Castle Team' }],
  creator: 'Cube Castle',
  publisher: 'Cube Castle',
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'),
  openGraph: {
    type: 'website',
    locale: 'zh_CN',
    url: '/',
    title: 'Cube Castle - 企业级HR管理平台',
    description: '现代化企业级 HR SaaS 平台，让 HR 管理变得简单而智能',
    siteName: 'Cube Castle',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'Cube Castle - 企业级HR管理平台',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Cube Castle - 企业级HR管理平台',
    description: '现代化企业级 HR SaaS 平台，让 HR 管理变得简单而智能',
    images: ['/og-image.png'],
  },
  robots: {
    index: true,
    follow: true,
  },
  manifest: '/manifest.json',
  icons: {
    icon: '/favicon.ico',
    apple: '/apple-touch-icon.png',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body 
        className={`${inter.variable} ${jetbrainsMono.variable} font-sans antialiased`}
        suppressHydrationWarning
      >
        <Providers>
          {children}
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: 'hsl(var(--card))',
                color: 'hsl(var(--card-foreground))',
                border: '1px solid hsl(var(--border))',
              },
              success: {
                iconTheme: {
                  primary: 'hsl(var(--primary))',
                  secondary: 'hsl(var(--primary-foreground))',
                },
              },
              error: {
                iconTheme: {
                  primary: 'hsl(var(--destructive))',
                  secondary: 'hsl(var(--destructive-foreground))',
                },
              },
            }}
          />
        </Providers>
      </body>
    </html>
  )
}