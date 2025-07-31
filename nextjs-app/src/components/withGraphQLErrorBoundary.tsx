// src/components/withGraphQLErrorBoundary.tsx
import React from 'react';
import GraphQLErrorBoundary from './GraphQLErrorBoundary';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { RefreshCw, Home, AlertTriangle } from 'lucide-react';
import { useRouter } from 'next/router';

interface ErrorFallbackProps {
  error: Error;
  retry: () => void;
}

const DefaultErrorFallback: React.FC<ErrorFallbackProps> = ({ error, retry }) => {
  const router = useRouter();

  const handleGoHome = () => {
    router.push('/');
  };

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <Card>
        <CardContent className="pt-6">
          <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4 mb-6 shadow-sm">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <AlertTriangle className="h-5 w-5 text-yellow-400" aria-hidden="true" />
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800">
                  GraphQL服务连接失败
                </h3>
                <div className="mt-1 text-sm text-yellow-700">
                  系统正在尝试使用备用数据源，某些功能可能受限。
                </div>
              </div>
            </div>
          </div>
          
          <div className="text-center">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              服务暂时不可用
            </h3>
            <p className="text-gray-600 mb-6">
              我们正在努力恢复服务。您可以尝试刷新页面或返回首页。
            </p>
            
            <div className="flex gap-3 justify-center">
              <Button onClick={retry} className="inline-flex items-center">
                <RefreshCw className="h-4 w-4 mr-2" />
                重试连接
              </Button>
              <Button variant="outline" onClick={handleGoHome} className="inline-flex items-center">
                <Home className="h-4 w-4 mr-2" />
                返回首页
              </Button>
            </div>
          </div>

          {process.env.NODE_ENV === 'development' && (
            <details className="mt-6">
              <summary className="cursor-pointer font-bold text-gray-900 hover:text-gray-700">
                开发环境错误详情
              </summary>
              <pre className="mt-3 p-3 bg-gray-100 rounded text-xs overflow-auto text-gray-800">
                {error.toString()}
              </pre>
            </details>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

// Higher-order component to wrap pages with GraphQL error boundary
function withGraphQLErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  customFallback?: React.ComponentType<ErrorFallbackProps>
) {
  const WithErrorBoundaryComponent = (props: P) => {
    return (
      <GraphQLErrorBoundary 
        fallback={customFallback || DefaultErrorFallback}
        onError={(error, errorInfo) => {
          // GraphQL Error Boundary error logged - integrating with monitoring services
          
          // You can integrate with error reporting services here
          // Example: Sentry, LogRocket, etc.
          if (typeof window !== 'undefined' && (window as any).gtag) {
            (window as any).gtag('event', 'exception', {
              description: error.toString(),
              fatal: false,
            });
          }
        }}
      >
        <WrappedComponent {...props} />
      </GraphQLErrorBoundary>
    );
  };

  WithErrorBoundaryComponent.displayName = `withGraphQLErrorBoundary(${WrappedComponent.displayName || WrappedComponent.name})`;

  return WithErrorBoundaryComponent;
}

export default withGraphQLErrorBoundary;