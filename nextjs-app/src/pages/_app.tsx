// src/pages/_app.tsx - 应用入口文件，引入Workday风格主题
import React from 'react';
import type { AppProps } from 'next/app';
import { ApolloProvider } from '@apollo/client';
import { Toaster } from 'react-hot-toast';
import { apolloClient } from '@/lib/graphql-client';
import GraphQLErrorBoundary from '@/components/GraphQLErrorBoundary';

// 引入样式文件
import '@/styles/workday-theme.css';
import '@/styles/animations.css';
import '@/styles/mobile-enhancements.css';

// Tailwind CSS基础样式
import 'tailwindcss/tailwind.css';

interface CubecastleAppProps extends AppProps {
  Component: AppProps['Component'] & {
    getLayout?: (page: React.ReactElement) => React.ReactNode;
  };
}

export default function CubeCastleApp({ Component, pageProps }: CubecastleAppProps) {
  // 获取页面级布局函数（如果有的话）
  const getLayout = Component.getLayout ?? ((page) => page);

  return (
    <ApolloProvider client={apolloClient}>
      <GraphQLErrorBoundary>
        {getLayout(<Component {...pageProps} />)}
        
        {/* 全局通知系统 */}
        <Toaster
          position="top-right"
          toastOptions={{
            // Workday风格的通知样式
            duration: 4000,
            style: {
              background: 'hsl(var(--card))',
              color: 'hsl(var(--card-foreground))',
              border: '1px solid hsl(var(--border))',
              borderRadius: 'var(--radius)',
              boxShadow: 'var(--shadow-lg)',
              fontSize: '14px',
              fontWeight: 500,
            },
            success: {
              iconTheme: {
                primary: 'hsl(var(--success))',
                secondary: 'hsl(var(--success-foreground))',
              },
              style: {
                borderLeft: '4px solid hsl(var(--success))',
              },
            },
            error: {
              iconTheme: {
                primary: 'hsl(var(--destructive))',
                secondary: 'hsl(var(--destructive-foreground))',
              },
              style: {
                borderLeft: '4px solid hsl(var(--destructive))',
              },
            },
            loading: {
              iconTheme: {
                primary: 'hsl(var(--primary))',
                secondary: 'hsl(var(--primary-foreground))',
              },
              style: {
                borderLeft: '4px solid hsl(var(--primary))',
              },
            },
          }}
        />
      </GraphQLErrorBoundary>
    </ApolloProvider>
  );
}