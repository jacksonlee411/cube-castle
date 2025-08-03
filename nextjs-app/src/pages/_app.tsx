// src/pages/_app.tsx - 应用入口文件，引入Workday风格主题
import React from 'react';
import type { AppProps } from 'next/app';
import { useRouter } from 'next/router';
import { ApolloProvider } from '@apollo/client';
import { Toaster } from 'react-hot-toast';
import { apolloClient } from '@/lib/graphql-client';
import GraphQLErrorBoundary from '@/components/GraphQLErrorBoundary';
import RESTErrorBoundary from '@/components/RESTErrorBoundary';

// 初始化 Immer MapSet 插件
import { enableMapSet } from 'immer';
enableMapSet();

// 引入样式文件
import '@/styles/workday-theme.css';
import '@/styles/animations.css';
import '@/styles/mobile-enhancements.css';
import '@/styles/organization-tree.css';

// Tailwind CSS基础样式 - 修复导入问题
import '../styles/globals.css';

interface CubecastleAppProps extends AppProps {
  Component: AppProps['Component'] & {
    getLayout?: (page: React.ReactElement) => React.ReactNode;
  };
}

export default function CubeCastleApp({ Component, pageProps }: CubecastleAppProps) {
  const router = useRouter();
  
  // 开发环境自动认证配置
  React.useEffect(() => {
    if (typeof window !== 'undefined') {
      // 检查并设置默认认证信息
      const tenantId = localStorage.getItem('tenant_id');
      const authToken = localStorage.getItem('auth_token');
      
      if (!tenantId) {
        localStorage.setItem('tenant_id', '550e8400-e29b-41d4-a716-446655440000');
        console.log('🔧 开发环境: 已设置默认 tenant_id');
      }
      
      if (!authToken) {
        localStorage.setItem('auth_token', 'dev-token');
        console.log('🔧 开发环境: 已设置默认 auth_token');
      }
    }
  }, []);
  
  // 获取页面级布局函数（如果有的话）
  const getLayout = Component.getLayout ?? ((page) => page);

  // 定义哪些页面使用REST/SWR (不需要GraphQL错误边界)
  const restPages = ['/employees', '/api/employees'];
  const isRESTPage = restPages.includes(router.pathname);

  // 选择合适的错误边界
  const ErrorBoundaryComponent = isRESTPage ? RESTErrorBoundary : GraphQLErrorBoundary;

  return (
    <ApolloProvider client={apolloClient}>
      <ErrorBoundaryComponent>
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
      </ErrorBoundaryComponent>
    </ApolloProvider>
  );
}