// src/components/RESTErrorBoundary.tsx
import React from 'react';
import { Button } from '@/components/ui/button';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface RESTErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

interface RESTErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; retry: () => void }>;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

class RESTErrorBoundary extends React.Component<
  RESTErrorBoundaryProps,
  RESTErrorBoundaryState
> {
  constructor(props: RESTErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<RESTErrorBoundaryState> {
    // Only catch errors that are likely network/data related, not React rendering loops
    const isNetworkError = error.message.includes('fetch') || 
                          error.message.includes('network') || 
                          error.message.includes('Failed to') ||
                          error.name === 'TypeError';
    
    const isReactRenderError = error.message.includes('Maximum update depth') ||
                              error.message.includes('Too many re-renders') ||
                              error.message.includes('Cannot update');

    // Don't catch React rendering loops - let them bubble up
    if (isReactRenderError) {
      console.warn('🚨 RESTErrorBoundary: Ignoring React render error to prevent masking:', error.message);
      throw error; // Re-throw to let it be handled elsewhere
    }

    if (isNetworkError) {
      console.log('🛡️ RESTErrorBoundary: Catching network error:', error.message);
      return {
        hasError: true,
        error,
      };
    }

    // For other errors, don't catch them - let parent boundaries handle
    console.warn('🚨 RESTErrorBoundary: Not catching error:', error.message);
    throw error;
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

    console.error('🛡️ RESTErrorBoundary: Caught error:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
    
    // Force page reload to reset SWR cache
    if (typeof window !== 'undefined') {
      window.location.reload();
    }
  };

  render() {
    if (this.state.hasError) {
      const { fallback: FallbackComponent } = this.props;
      const { error } = this.state;

      // Use custom fallback component if provided
      if (FallbackComponent && error) {
        return <FallbackComponent error={error} retry={this.handleRetry} />;
      }

      // Default error UI for REST/SWR errors
      return (
        <div className="p-6">
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 shadow-sm">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <AlertTriangle className="h-5 w-5 text-red-400" aria-hidden="true" />
              </div>
              <div className="ml-3 flex-1">
                <h3 className="text-sm font-medium text-red-800">
                  数据加载失败
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>无法从服务器加载数据。请检查网络连接并重试。</p>
                </div>
                <div className="mt-4">
                  <div className="flex gap-2">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={this.handleRetry}
                      className="bg-white text-red-800 border-red-300 hover:bg-red-50"
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
            <div className="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4 shadow-sm">
              <div className="flex items-start">
                <div className="ml-3 flex-1">
                  <h3 className="text-sm font-medium text-gray-800">
                    开发环境错误详情
                  </h3>
                  <div className="mt-2">
                    <details className="text-sm text-gray-700">
                      <summary className="cursor-pointer font-medium hover:text-gray-900">
                        点击查看错误详情
                      </summary>
                      <pre className="mt-2 text-xs bg-gray-100 p-3 rounded overflow-auto">
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

export default RESTErrorBoundary;