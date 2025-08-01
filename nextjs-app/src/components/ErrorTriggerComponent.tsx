import React from 'react';

export function ErrorTriggerComponent() {
  const [apiResponse, setApiResponse] = React.useState<any>(null);
  const [isLoading, setIsLoading] = React.useState(false);
  
  React.useEffect(() => {
    // Test error API on component mount
    const testErrorAPI = async () => {
      setIsLoading(true);
      try {
        console.log('🧪 Testing error API endpoint');
        const response = await fetch('/api/test-error');
        const data = await response.json();
        
        console.log('🧪 Error API response:', {
          status: response.status,
          ok: response.ok,
          data
        });
        
        setApiResponse({
          status: response.status,
          ok: response.ok,
          data
        });
        
        // If response is not ok, throw error to trigger error boundary
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${data.message || response.statusText}`);
        }
        
      } catch (error) {
        console.log('🚨 Caught error in ErrorTriggerComponent:', error);
        setApiResponse({ error: error instanceof Error ? error.message : error });
        // Re-throw to trigger error boundary
        throw error;
      } finally {
        setIsLoading(false);
      }
    };
    
    // Delay to allow component to render first
    const timer = setTimeout(testErrorAPI, 500);
    return () => clearTimeout(timer);
  }, []);
  
  return (
    <div className="p-4 border border-blue-200 bg-blue-50 rounded">
      <h3 className="font-medium text-blue-800 mb-2">错误边界测试组件</h3>
      <div className="text-sm text-blue-700 mb-3">
        {isLoading && <p>正在测试错误API...</p>}
        {apiResponse && (
          <div className="mt-2 p-2 bg-gray-100 rounded text-xs">
            <strong>API响应:</strong>
            <pre className="mt-1 overflow-auto">{JSON.stringify(apiResponse, null, 2)}</pre>
          </div>
        )}
      </div>
    </div>
  );
}