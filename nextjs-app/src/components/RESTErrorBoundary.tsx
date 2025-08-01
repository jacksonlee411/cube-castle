// src/components/RESTErrorBoundary.tsx
import React from 'react';
import { Button } from '@/components/ui/button';
import { AlertTriangle, RefreshCw, Bug, Network } from 'lucide-react';
import { toast } from 'react-hot-toast';

interface RESTErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
  errorId: string;
  errorType: 'network' | 'data' | 'render' | 'unknown';
}

interface RESTErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; retry: () => void }>;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
  resetOnPropsChange?: boolean;
  resetKeys?: (string | number)[];
}

class RESTErrorBoundary extends React.Component<
  RESTErrorBoundaryProps,
  RESTErrorBoundaryState
> {
  private resetTimeoutId: NodeJS.Timeout | null = null;

  constructor(props: RESTErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
      errorType: 'unknown',
    };
  }

  static getDerivedStateFromError(error: Error): Partial<RESTErrorBoundaryState> {
    // Enhanced error classification
    let errorType: 'network' | 'data' | 'render' | 'unknown' = 'unknown';
    
    if (error.message.includes('fetch') || 
        error.message.includes('network') || 
        error.message.includes('Failed to') ||
        error.message.includes('服务器') ||
        error.name === 'TypeError') {
      errorType = 'network';
    } else if (error.message.includes('JSON') ||
               error.message.includes('data') ||
               error.message.includes('employees') ||
               error.message.includes('数据')) {
      errorType = 'data';
    } else if (error.message.includes('Maximum update depth') ||
               error.message.includes('Too many re-renders') ||
               error.message.includes('Cannot update')) {
      errorType = 'render';
    }

    // Don't catch React rendering loops - let them bubble up
    if (errorType === 'render') {
      console.warn('🚨 RESTErrorBoundary: Ignoring React render error to prevent masking:', error.message);
      throw error; // Re-throw to let it be handled elsewhere
    }

    // Generate unique error ID
    const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    console.log('🛡️ RESTErrorBoundary: Catching error:', {
      type: errorType,
      message: error.message,
      errorId
    });

    return {
      hasError: true,
      error,
      errorType,
      errorId,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // Enhanced error logging
    const errorDetails = {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      errorBoundary: 'RESTErrorBoundary',
      timestamp: new Date().toISOString(),
      url: typeof window !== 'undefined' ? window.location.href : 'unknown',
      userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : 'unknown',
      errorType: this.state.errorType,
      errorId: this.state.errorId,
    };

    console.error('🛡️ RESTErrorBoundary: Caught error details:', errorDetails);

    // Call optional error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // Show user-friendly error notification
    const errorMessage = this.state.errorType === 'network' 
      ? '网络连接出现问题，正在尝试恢复...'
      : this.state.errorType === 'data'
      ? '数据处理出现错误，正在尝试恢复...'
      : '应用出现错误，正在尝试恢复...';

    toast.error(errorMessage, {
      duration: 5000,
      position: 'top-right',
    });

    // Auto-recovery mechanism
    this.resetTimeoutId = setTimeout(() => {
      this.handleReset();
    }, 3000);
  }

  componentDidUpdate(prevProps: RESTErrorBoundaryProps) {
    const { resetKeys, resetOnPropsChange } = this.props;
    const { hasError } = this.state;
    
    if (hasError && prevProps.resetKeys !== resetKeys) {
      if (resetKeys && resetKeys.some((key, idx) => prevProps.resetKeys?.[idx] !== key)) {
        this.handleReset();
      }
    }
    
    if (hasError && resetOnPropsChange && prevProps.children !== this.props.children) {
      this.handleReset();
    }
  }

  componentWillUnmount() {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId);
    }
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
      errorType: 'unknown',
    });

    toast.success('应用已恢复正常', {
      duration: 3000,
      position: 'top-right',
    });
  };

  render() {
    if (this.state.hasError) {
      const { fallback: FallbackComponent } = this.props;
      const { error, errorType, errorId } = this.state;

      // Use custom fallback component if provided
      if (FallbackComponent && error) {
        return <FallbackComponent error={error} retry={this.handleReset} />;
      }

      // Enhanced error UI with type-specific styling and icons
      const getErrorIcon = () => {
        switch (errorType) {
          case 'network': return <Network className="h-5 w-5 text-red-400" />;
          case 'data': return <Bug className="h-5 w-5 text-orange-400" />;
          default: return <AlertTriangle className="h-5 w-5 text-red-400" />;
        }
      };

      const getErrorColor = () => {
        switch (errorType) {
          case 'network': return 'red';
          case 'data': return 'orange';
          default: return 'red';
        }
      };

      const getErrorTitle = () => {
        switch (errorType) {
          case 'network': return '网络连接失败';
          case 'data': return '数据处理错误';
          default: return '数据加载失败';
        }
      };

      const getErrorDescription = () => {
        switch (errorType) {
          case 'network': return '无法连接到服务器。请检查网络连接并重试。';
          case 'data': return '数据格式异常或处理失败。系统正在尝试恢复。';
          default: return '无法从服务器加载数据。请检查网络连接并重试。';
        }
      };

      const color = getErrorColor();

      return (
        <div className="p-6">
          <div className={`rounded-lg border border-${color}-200 bg-${color}-50 p-4 shadow-sm`}>
            <div className="flex items-start">
              <div className="flex-shrink-0">
                {getErrorIcon()}
              </div>
              <div className="ml-3 flex-1">
                <h3 className={`text-sm font-medium text-${color}-800`}>
                  {getErrorTitle()}
                </h3>
                <div className={`mt-2 text-sm text-${color}-700`}>
                  <p>{getErrorDescription()}</p>
                  {error && (
                    <p className="mt-1 text-xs">
                      错误信息: {error.message}
                    </p>
                  )}
                </div>
                <div className="mt-4">
                  <div className="flex gap-2">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={this.handleReset}
                      className={`bg-white text-${color}-800 border-${color}-300 hover:bg-${color}-50`}
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      重试
                    </Button>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={() => window.location.reload()}
                      className={`bg-white text-${color}-800 border-${color}-300 hover:bg-${color}-50`}
                    >
                      刷新页面
                    </Button>
                  </div>
                </div>
              </div>
            </div>
            
            {errorId && (
              <div className="mt-3 text-xs text-gray-500 text-center">
                错误ID: {errorId}
              </div>
            )}
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
                      <pre className="mt-2 text-xs bg-gray-100 p-3 rounded overflow-auto max-h-64">
                        错误类型: {errorType}
                        错误ID: {errorId}
                        
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

// Enhanced SWR-specific error boundary component
interface SWRErrorBoundaryProps {
  children: React.ReactNode;
  onError?: (error: Error) => void;
  resetKeys?: (string | number)[];
}

export function SWRErrorBoundary({ children, onError, resetKeys }: SWRErrorBoundaryProps) {
  return (
    <RESTErrorBoundary
      resetOnPropsChange={true}
      resetKeys={resetKeys}
      onError={(error, errorInfo) => {
        // Enhanced SWR error logging
        console.error('🛡️ SWR Error Boundary triggered:', {
          error: error.message,
          stack: error.stack,
          componentStack: errorInfo.componentStack,
          type: 'SWR_ERROR',
          timestamp: new Date().toISOString(),
        });

        if (onError) {
          onError(error);
        }
      }}
      fallback={({ error, retry }) => (
        <div className="min-h-[200px] flex items-center justify-center">
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 max-w-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.25 3.09c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800">
                  SWR数据同步出现问题
                </h3>
                <p className="mt-1 text-sm text-yellow-700">
                  数据获取或同步过程中发生错误。请稍后重试。
                </p>
                <div className="mt-3">
                  <Button size="sm" onClick={retry} className="bg-yellow-100 text-yellow-800 hover:bg-yellow-200">
                    <RefreshCw className="h-4 w-4 mr-2" />
                    重新加载数据
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    >
      {children}
    </RESTErrorBoundary>
  );
}

export default RESTErrorBoundary;