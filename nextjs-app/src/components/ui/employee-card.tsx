import React from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';

export interface EmployeeCardProps {
  /** 员工信息 */
  employee: {
    id: string;
    name: string;
    employeeId: string;
    email?: string;
    phone?: string;
    department?: string;
    position?: string;
    status: 'active' | 'inactive' | 'pending';
    avatar?: string;
    hireDate?: string;
  };
  /** 是否处于选择模式 */
  selectable?: boolean;
  /** 是否被选中 */
  selected?: boolean;
  /** 选择状态变化回调 */
  onSelectionChange?: (selected: boolean) => void;
  /** 点击卡片事件 */
  onClick?: () => void;
  /** 快速操作菜单项 */
  actions?: Array<{
    label: string;
    icon?: React.ReactNode;
    onClick: () => void;
    variant?: 'default' | 'destructive';
  }>;
  /** 自定义样式 */
  className?: string;
}

export function EmployeeCard({
  employee,
  selectable = false,
  selected = false,
  onSelectionChange,
  onClick,
  actions = [],
  className
}: EmployeeCardProps) {
  const statusConfig = {
    active: {
      label: '在职',
      variant: 'default' as const,
      color: 'bg-success-light text-success border-success/20'
    },
    inactive: {
      label: '离职',
      variant: 'secondary' as const,
      color: 'bg-gray-100 text-gray-600 border-gray-200'
    },
    pending: {
      label: '待入职',
      variant: 'outline' as const,
      color: 'bg-warning-light text-warning border-warning/20'
    }
  };

  const currentStatus = statusConfig[employee.status];

  // 生成头像显示
  const renderAvatar = () => {
    if (employee.avatar) {
      return (
        <img
          src={employee.avatar}
          alt={employee.name}
          className="w-12 h-12 rounded-full object-cover"
        />
      );
    }
    
    // 使用姓名首字母作为默认头像
    const initials = employee.name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
      
    return (
      <div className="w-10 h-10 sm:w-12 sm:h-12 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold text-xs sm:text-sm transition-colors duration-200">
        {initials}
      </div>
    );
  };

  const handleCardClick = (e: React.MouseEvent) => {
    // 如果点击的是选择框或操作按钮区域，不触发卡片点击
    const target = e.target as HTMLElement;
    if (target.closest('[data-checkbox]') || target.closest('[data-actions]')) {
      return;
    }
    onClick?.();
  };

  return (
    <Card className={cn(
      'card-workday relative p-4 sm:p-6 space-y-4 transition-all duration-200 hover:shadow-md',
      selectable && selected && 'ring-2 ring-primary/20 border-primary/30 shadow-primary/10',
      onClick && 'cursor-pointer',
      className
    )}>
      {/* 选择框 */}
      {selectable && (
        <div 
          className="absolute top-3 right-3 sm:top-4 sm:right-4 z-10" 
          data-checkbox
          onClick={(e) => e.stopPropagation()}
        >
          <input
            type="checkbox"
            checked={selected}
            onChange={(e) => onSelectionChange?.(e.target.checked)}
            className="w-4 h-4 text-primary focus:ring-primary/20 border-border rounded transition-colors duration-200"
          />
        </div>
      )}

      <div onClick={handleCardClick}>
        {/* 头部：头像 + 基本信息 */}
        <div className="flex items-start gap-3 sm:gap-4">
          {renderAvatar()}
          
          <div className="flex-1 min-w-0">
            {/* 姓名和工号 */}
            <div className="flex items-center gap-2 mb-1">
              <h3 className="text-base sm:text-lg font-semibold text-foreground truncate">
                {employee.name}
              </h3>
              <Badge 
                variant="outline" 
                className="text-xs text-gray-500 font-mono"
              >
                {employee.employeeId}
              </Badge>
            </div>
            
            {/* 部门和职位 */}
            <div className="space-y-1">
              {employee.department && (
                <p className="text-body-medium text-gray-600 truncate">
                  {employee.department}
                </p>
              )}
              {employee.position && (
                <p className="text-body-small text-gray-500 truncate">
                  {employee.position}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* 联系信息 */}
        {(employee.email || employee.phone) && (
          <div className="space-y-1 pt-2 border-t border-gray-100">
            {employee.email && (
              <p className="text-body-small text-gray-500 truncate flex items-center gap-2">
                <span className="text-gray-400">📧</span>
                {employee.email}
              </p>
            )}
            {employee.phone && (
              <p className="text-body-small text-gray-500 flex items-center gap-2">
                <span className="text-gray-400">📱</span>
                {employee.phone}
              </p>
            )}
          </div>
        )}

        {/* 底部：状态 + 入职时间 */}
        <div className="flex items-center justify-between pt-3 border-t border-gray-100">
          <div className="flex items-center gap-3">
            <Badge className={cn('status-indicator', currentStatus.color)}>
              {currentStatus.label}
            </Badge>
            {employee.hireDate && (
              <span className="text-body-small text-gray-400">
                入职 {new Date(employee.hireDate).toLocaleDateString('zh-CN')}
              </span>
            )}
          </div>

          {/* 快速操作菜单 */}
          {actions.length > 0 && (
            <div data-actions>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button 
                    variant="ghost" 
                    size="sm"
                    className="h-8 w-8 p-0 hover:bg-gray-100"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <span className="text-gray-400">⋯</span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                  {actions.map((action, index) => (
                    <DropdownMenuItem
                      key={index}
                      onClick={() => {
                        action.onClick();
                      }}
                      className={cn(
                        'flex items-center gap-2 cursor-pointer',
                        action.variant === 'destructive' && 'text-destructive focus:text-destructive'
                      )}
                    >
                      {action.icon}
                      {action.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}

// 员工卡片网格容器
export interface EmployeeCardsGridProps {
  children: React.ReactNode;
  className?: string;
  columns?: 2 | 3 | 4;
}

export function EmployeeCardsGrid({ 
  children, 
  className,
  columns = 3
}: EmployeeCardsGridProps) {
  const gridCols = {
    2: 'grid-cols-1 md:grid-cols-2',
    3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
  };

  return (
    <div className={cn(
      'grid gap-6',
      gridCols[columns],
      className
    )}>
      {children}
    </div>
  );
}

// 员工卡片加载状态
export function EmployeeCardSkeleton({ className }: { className?: string }) {
  return (
    <Card className={cn('card-workday p-6 space-y-4', className)}>
      {/* 头部骨架 */}
      <div className="flex items-start gap-4">
        <div className="skeleton w-12 h-12 rounded-full"></div>
        <div className="flex-1 space-y-2">
          <div className="skeleton h-5 w-32 rounded"></div>
          <div className="skeleton h-4 w-24 rounded"></div>
          <div className="skeleton h-3 w-20 rounded"></div>
        </div>
      </div>
      
      {/* 联系信息骨架 */}
      <div className="space-y-2 pt-2 border-t border-gray-100">
        <div className="skeleton h-3 w-40 rounded"></div>
        <div className="skeleton h-3 w-32 rounded"></div>
      </div>
      
      {/* 底部骨架 */}
      <div className="flex items-center justify-between pt-3 border-t border-gray-100">
        <div className="skeleton h-6 w-16 rounded-full"></div>
        <div className="skeleton h-3 w-20 rounded"></div>
      </div>
    </Card>
  );
}

export default EmployeeCard;