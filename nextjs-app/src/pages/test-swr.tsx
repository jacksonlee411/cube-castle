import React from 'react';
import { useEmployeesSWR } from '@/hooks/useEmployeesSWR';

const TestSWRPage: React.FC = () => {
  const { employees, totalCount, isLoading, isError, error } = useEmployeesSWR({
    pageSize: 10
  });

  if (isError) {
    return (
      <div className="p-6">
        <h1 className="text-2xl mb-4">SWR错误测试</h1>
        <div className="bg-red-100 p-4 rounded">
          <p className="text-red-700">错误: {error?.message}</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl mb-4">SWR加载测试</h1>
        <div className="bg-blue-100 p-4 rounded">
          <p className="text-blue-700">正在加载数据...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl mb-4">SWR成功测试</h1>
      <div className="bg-green-100 p-4 rounded mb-4">
        <p className="text-green-700">成功加载 {employees.length} / {totalCount} 个员工</p>
      </div>
      
      <div className="space-y-2">
        {employees.slice(0, 5).map((emp) => (
          <div key={emp.id} className="bg-white p-3 rounded border">
            <p className="font-semibold">{emp.legalName}</p>
            <p className="text-sm text-gray-600">{emp.employeeId} • {emp.email}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TestSWRPage;