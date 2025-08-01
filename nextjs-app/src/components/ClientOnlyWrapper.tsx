import React from 'react';

interface ClientOnlyWrapperProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * 客户端渲染包装器 - 确保组件只在客户端渲染
 * 解决SWR在SSR/SSG环境中的数据同步问题
 */
export const ClientOnlyWrapper: React.FC<ClientOnlyWrapperProps> = ({ 
  children, 
  fallback = null 
}) => {
  const [hasMounted, setHasMounted] = React.useState(false);

  React.useEffect(() => {
    console.log('🌐 ClientOnlyWrapper: 客户端挂载完成');
    setHasMounted(true);
  }, []);

  if (!hasMounted) {
    console.log('🌐 ClientOnlyWrapper: 等待客户端挂载...');
    return <>{fallback}</>;
  }

  console.log('🌐 ClientOnlyWrapper: 渲染客户端组件');
  return <>{children}</>;
};

export default ClientOnlyWrapper;