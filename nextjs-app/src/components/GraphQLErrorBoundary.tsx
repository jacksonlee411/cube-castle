// src/components/GraphQLErrorBoundary.tsx
import React from 'react';
import { Button } from '@/components/ui/button';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { toast } from 'react-hot-toast';

interface GraphQLErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

interface GraphQLErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; retry: () => void }>;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

class GraphQLErrorBoundary extends React.Component<
  GraphQLErrorBoundaryProps,
  GraphQLErrorBoundaryState
> {
  constructor(props: GraphQLErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<GraphQLErrorBoundaryState> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // Call optional error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // GraphQL Error Boundary caught an error - error details available in error and errorInfo state
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      const { fallback: FallbackComponent } = this.props;
      const { error } = this.state;

      // Use custom fallback component if provided
      if (FallbackComponent && error) {
        return <FallbackComponent error={error} retry={this.handleRetry} />;
      }

      // Default error UI
      return (
        <div className="p-6">
          <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4 shadow-sm">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <AlertTriangle className="h-5 w-5 text-yellow-400" aria-hidden="true" />
              </div>
              <div className="ml-3 flex-1">
                <h3 className="text-sm font-medium text-yellow-800">
                  GraphQL服务连接失败
                </h3>
                <div className="mt-2 text-sm text-yellow-700">
                  <p>无法连接到GraphQL服务器，正在尝试使用本地数据。如果问题持续存在，请联系系统管理员。</p>
                </div>
                <div className="mt-4">
                  <div className="flex gap-2">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={this.handleRetry}
                      className="bg-white text-yellow-800 border-yellow-300 hover:bg-yellow-50"
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      重试
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          {process.env.NODE_ENV === 'development' && (
            <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-4 shadow-sm">
              <div className="flex items-start">
                <div className="ml-3 flex-1">
                  <h3 className="text-sm font-medium text-red-800">
                    开发环境错误详情
                  </h3>
                  <div className="mt-2">
                    <details className="text-sm text-red-700">
                      <summary className="cursor-pointer font-medium hover:text-red-900">
                        点击查看错误详情
                      </summary>
                      <pre className="mt-2 text-xs bg-red-100 p-3 rounded overflow-auto">
                        {error?.toString()}
                        {this.state.errorInfo?.componentStack}
                      </pre>
                    </details>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

export default GraphQLErrorBoundary;