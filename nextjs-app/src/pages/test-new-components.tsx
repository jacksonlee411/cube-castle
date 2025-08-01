import React from 'react';
import StatCard, { StatCardsGrid } from '@/components/ui/stat-card';
import EmployeeCard, { EmployeeCardsGrid } from '@/components/ui/employee-card';
import SmartFilter from '@/components/ui/smart-filter';
import { PieChart, BarChart } from '@/components/ui/data-visualization';
import { Users, UserCheck, UserPlus, Building } from 'lucide-react';

export default function TestNewComponents() {
  // 测试数据
  const testEmployee = {
    id: '1',
    name: '张三',
    employeeId: 'EMP001',
    email: 'zhangsan@company.com',
    phone: '13800138001',
    department: '技术部',
    position: '高级软件工程师',
    status: 'active' as const,
    hireDate: '2023-01-15'
  };

  const testChartData = [
    { label: '技术部', value: 15 },
    { label: '产品部', value: 8 },
    { label: '人事部', value: 5 }
  ];

  const testFilterOptions = [
    {
      key: 'department',
      label: '部门',
      type: 'select' as const,
      options: [
        { label: '技术部', value: '技术部' },
        { label: '产品部', value: '产品部' }
      ]
    }
  ];

  return (
    <div className="p-8 space-y-8">
      <h1 className="text-display-large">新组件测试页面</h1>
      
      {/* 统计卡片测试 */}
      <section>
        <h2 className="text-display-medium mb-4">统计卡片组件</h2>
        <StatCardsGrid columns={4}>
          <StatCard
            title="总员工数"
            value={28}
            change={8.5}
            changeLabel="较上月"
            icon={<Users className="w-8 h-8" />}
            variant="primary"
          />
          <StatCard
            title="在职员工"
            value={25}
            change={2.1}
            changeLabel="较上月"
            icon={<UserCheck className="w-8 h-8" />}
            variant="success"
          />
          <StatCard
            title="待入职"
            value={3}
            change={-1.2}
            changeLabel="较上月"
            icon={<UserPlus className="w-8 h-8" />}
            variant="warning"
          />
          <StatCard
            title="部门数量"
            value={4}
            change={0}
            changeLabel="较上月"
            icon={<Building className="w-8 h-8" />}
            variant="default"
          />
        </StatCardsGrid>
      </section>

      {/* 数据可视化测试 */}
      <section>
        <h2 className="text-display-medium mb-4">数据可视化组件</h2>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <PieChart
            data={testChartData}
            title="部门分布"
            description="各部门员工数量分布"
          />
          <BarChart
            data={[
              { label: '在职', value: 25 },
              { label: '离职', value: 2 },
              { label: '待入职', value: 3 }
            ]}
            title="员工状态统计"
            description="不同状态员工数量对比"
          />
        </div>
      </section>

      {/* 智能筛选测试 */}
      <section>
        <h2 className="text-display-medium mb-4">智能筛选组件</h2>
        <SmartFilter
          filterOptions={testFilterOptions}
          activeFilters={[]}
          onFiltersChange={() => {}}
          searchValue=""
          onSearchChange={() => {}}
          searchPlaceholder="搜索员工..."
        />
      </section>

      {/* 员工卡片测试 */}
      <section>
        <h2 className="text-display-medium mb-4">员工卡片组件</h2>
        <EmployeeCardsGrid columns={3}>
          <EmployeeCard
            employee={testEmployee}
            selectable={true}
            selected={false}
            onSelectionChange={() => {}}
            onClick={() => console.log('员工卡片点击')}
            actions={[
              {
                label: '编辑信息',
                onClick: () => console.log('编辑')
              },
              {
                label: '删除员工',
                onClick: () => console.log('删除'),
                variant: 'destructive'
              }
            ]}
          />
          <EmployeeCard
            employee={{
              ...testEmployee,
              id: '2',
              name: '王五',
              employeeId: 'EMP002',
              status: 'pending'
            }}
            selectable={true}
            selected={true}
            onSelectionChange={() => {}}
          />
          <EmployeeCard
            employee={{
              ...testEmployee,
              id: '3',
              name: '李四',
              employeeId: 'EMP003',
              status: 'inactive'
            }}
            selectable={false}
            selected={false}
            onSelectionChange={() => {}}
          />
        </EmployeeCardsGrid>
      </section>
    </div>
  );
}