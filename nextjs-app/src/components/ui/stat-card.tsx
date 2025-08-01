import React from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';

export interface StatCardProps {
  /** 统计标题 */
  title: string;
  /** 主要数值 */
  value: string | number;
  /** 变化趋势百分比，正数为增长，负数为下降 */
  change?: number;
  /** 变化时间周期描述 */
  changeLabel?: string;
  /** 图标组件 */
  icon?: React.ReactNode;
  /** 自定义样式类名 */
  className?: string;
  /** 是否显示加载状态 */
  loading?: boolean;
  /** 点击事件处理 */
  onClick?: () => void;
  /** 自定义颜色主题 */
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'destructive';
}

export function StatCard({
  title,
  value,
  change,
  changeLabel = '较上周',
  icon,
  className,
  loading = false,
  onClick,
  variant = 'default'
}: StatCardProps) {
  const isPositiveChange = change && change > 0;
  const isNegativeChange = change && change < 0;
  const hasChange = change !== undefined && change !== null;

  const cardVariants = {
    default: 'border-border hover:border-gray-300 bg-card',
    primary: 'border-primary/20 bg-gradient-to-br from-primary/5 to-primary/10 hover:border-primary/30 hover:shadow-primary/10',
    success: 'border-success/20 bg-gradient-to-br from-success/5 to-success/10 hover:border-success/30 hover:shadow-success/10',
    warning: 'border-warning/20 bg-gradient-to-br from-warning/5 to-warning/10 hover:border-warning/30 hover:shadow-warning/10',
    destructive: 'border-destructive/20 bg-gradient-to-br from-destructive/5 to-destructive/10 hover:border-destructive/30 hover:shadow-destructive/10'
  };

  const valueColors = {
    default: 'text-foreground',
    primary: 'text-primary',
    success: 'text-success',
    warning: 'text-warning',
    destructive: 'text-destructive'
  };

  if (loading) {
    return (
      <Card className={cn(
        'card-workday p-6 space-y-4',
        cardVariants[variant],
        className
      )}>
        {/* 加载骨架屏 */}
        <div className="flex items-center justify-between">
          <div className="skeleton h-4 w-24 rounded"></div>
          {icon && <div className="skeleton h-8 w-8 rounded"></div>}
        </div>
        <div className="space-y-2">
          <div className="skeleton h-10 w-32 rounded"></div>
          <div className="skeleton h-3 w-20 rounded"></div>
        </div>
      </Card>
    );
  }

  return (
    <Card
      className={cn(
        'card-workday card-hover p-6 space-y-4 cursor-pointer transition-all duration-200',
        cardVariants[variant],
        onClick && 'hover:shadow-md',
        className
      )}
      onClick={onClick}
    >
      {/* 标题和图标行 */}
      <div className="flex items-center justify-between">
        <p className="text-body-medium font-medium text-muted-foreground uppercase tracking-wide">
          {title}
        </p>
        {icon && (
          <div className={cn(
            "flex-shrink-0 opacity-80 transition-colors duration-200",
            variant === 'default' ? 'text-muted-foreground' : valueColors[variant]
          )}>
            {icon}
          </div>
        )}
      </div>

      {/* 数值和趋势行 */}
      <div className="space-y-3">
        <div className="stat-number">
          <span className={cn('text-2xl sm:text-3xl font-bold leading-none', valueColors[variant])}>
            {typeof value === 'number' ? value.toLocaleString() : value}
          </span>
        </div>

        {/* 趋势指示器 */}
        {hasChange && (
          <div className="flex items-center gap-2">
            <div className={cn(
              'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors duration-200',
              isPositiveChange && 'bg-success/10 text-success border border-success/20',
              isNegativeChange && 'bg-destructive/10 text-destructive border border-destructive/20',
              !isPositiveChange && !isNegativeChange && 'bg-muted text-muted-foreground border border-border'
            )}>
              {/* 趋势箭头 */}
              <span className="text-xs">
                {isPositiveChange && '↗'}
                {isNegativeChange && '↘'}
                {!isPositiveChange && !isNegativeChange && '→'}
              </span>
              <span>
                {Math.abs(change)}%
              </span>
            </div>
            <span className="text-body-small text-gray-500">
              {changeLabel}
            </span>
          </div>
        )}
      </div>
    </Card>
  );
}

// 统计卡片网格容器组件
export interface StatCardsGridProps {
  children: React.ReactNode;
  className?: string;
  columns?: 2 | 3 | 4;
}

export function StatCardsGrid({ 
  children, 
  className,
  columns = 4 
}: StatCardsGridProps) {
  const gridCols = {
    2: 'grid-cols-1 md:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4'
  };

  return (
    <div className={cn(
      'grid gap-4 sm:gap-6',
      gridCols[columns],
      className
    )}>
      {children}
    </div>
  );
}

// 导出默认统计卡片组件
export default StatCard;