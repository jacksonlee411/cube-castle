import React from 'react';
import { useRouter } from 'next/router';

export default function LoginPage() {
  const router = useRouter();

  const handleDevelopmentLogin = () => {
    // 设置开发环境默认认证信息
    localStorage.setItem('tenant_id', '550e8400-e29b-41d4-a716-446655440000');
    localStorage.setItem('auth_token', 'dev-token');
    
    // 跳转回组织架构页面
    router.push('/organization/chart');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Cube Castle 登录
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            开发环境简化登录
          </p>
        </div>
        <div className="mt-8 space-y-6">
          <button
            onClick={handleDevelopmentLogin}
            className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            开发环境登录
          </button>
          <div className="text-sm text-gray-500 text-center">
            <p>这是开发环境的简化登录页面。</p>
            <p>点击按钮将自动设置默认的租户ID和认证令牌。</p>
          </div>
        </div>
      </div>
    </div>
  );
}