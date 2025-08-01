import React from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';

export interface ChartDataPoint {
  label: string;
  value: number;
  color?: string;
  percentage?: number;
}

export interface ChartProps {
  data: ChartDataPoint[];
  title?: string;
  description?: string;
  className?: string;
  height?: number;
  showLegend?: boolean;
  loading?: boolean;
}

// 简单饼图组件
function PieChartComponent({ 
  data, 
  title, 
  description, 
  className, 
  height = 200,
  showLegend = true,
  loading = false 
}: ChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0);
  const radius = height / 2 - 20;
  const centerX = height / 2;
  const centerY = height / 2;

  // 计算路径
  const calculatePath = (startAngle: number, endAngle: number) => {
    const x1 = centerX + radius * Math.cos(startAngle);
    const y1 = centerY + radius * Math.sin(startAngle);
    const x2 = centerX + radius * Math.cos(endAngle);
    const y2 = centerY + radius * Math.sin(endAngle);
    
    const largeArcFlag = endAngle - startAngle <= Math.PI ? "0" : "1";
    
    return `M ${centerX} ${centerY} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2} Z`;
  };

  if (loading) {
    return (
      <Card className={cn('chart-container space-y-4', className)}>
        {title && <div className="skeleton h-6 w-32 rounded"></div>}
        <div className="skeleton rounded-full mx-auto" style={{ width: height, height }}></div>
        {showLegend && (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-2">
                <div className="skeleton h-3 w-3 rounded-full"></div>
                <div className="skeleton h-3 w-16 rounded"></div>
              </div>
            ))}
          </div>
        )}
      </Card>
    );
  }

  if (total === 0) {
    return (
      <Card className={cn('chart-container flex items-center justify-center', className)}>
        <div className="text-center text-gray-400">
          <div className="text-4xl mb-2">📊</div>
          <p className="text-body-medium">暂无数据</p>
        </div>
      </Card>
    );
  }

  let currentAngle = -Math.PI / 2; // 从顶部开始

  return (
    <Card className={cn('chart-container space-y-4', className)}>
      {title && (
        <div className="space-y-1">
          <h3 className="text-display-small font-semibold">{title}</h3>
          {description && (
            <p className="text-body-small text-gray-500">{description}</p>
          )}
        </div>
      )}
      
      <div className="flex items-center justify-center">
        <svg width={height} height={height} className="transform -rotate-90">
          {data.map((item, index) => {
            const angle = (item.value / total) * 2 * Math.PI;
            const path = calculatePath(currentAngle, currentAngle + angle);
            const color = item.color ?? `hsl(${(index * 360) / data.length}, 70%, 60%)`;
            
            currentAngle += angle;
            
            return (
              <path
                key={index}
                d={path}
                fill={color}
                className="transition-all duration-200 hover:opacity-80"
              />
            );
          })}
        </svg>
      </div>

      {showLegend && (
        <div className="space-y-2">
          {data.map((item, index) => {
            const percentage = ((item.value / total) * 100).toFixed(1);
            const color = item.color ?? `hsl(${(index * 360) / data.length}, 70%, 60%)`;
            
            return (
              <div key={index} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div 
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: color }}
                  ></div>
                  <span className="text-body-medium">{item.label}</span>
                </div>
                <div className="text-body-medium font-medium">
                  {item.value} ({percentage}%)
                </div>
              </div>
            );
          })}
        </div>
      )}
    </Card>
  );
}

// 简单柱状图组件
function BarChartComponent({ 
  data, 
  title, 
  description, 
  className, 
  height = 200,
  loading = false 
}: ChartProps) {
  const maxValue = Math.max(...data.map(item => item.value)) || 1; // 防止除零
  const barWidth = Math.max(30, (300 / Math.max(data.length, 1)) - 10); // 防止负数或NaN

  if (loading) {
    return (
      <Card className={cn('chart-container space-y-4', className)}>
        {title && <div className="skeleton h-6 w-32 rounded"></div>}
        <div className="flex items-end gap-2 justify-center" style={{ height }}>
          {Array.from({ length: 4 }).map((_, i) => (
            <div 
              key={i} 
              className="skeleton rounded-t"
              style={{ 
                width: barWidth, 
                height: Math.random() * (height - 40) + 40 
              }}
            ></div>
          ))}
        </div>
      </Card>
    );
  }

  return (
    <Card className={cn('chart-container space-y-4', className)}>
      {title && (
        <div className="space-y-1">
          <h3 className="text-display-small font-semibold">{title}</h3>
          {description && (
            <p className="text-body-small text-gray-500">{description}</p>
          )}
        </div>
      )}
      
      <div className="flex items-end justify-center gap-2" style={{ height }}>
        {data.map((item, index) => {
          const barHeight = (item.value / maxValue) * (height - 60);
          const color = item.color ?? `hsl(var(--primary))`;
          
          return (
            <div key={index} className="flex flex-col items-center gap-2">
              <div 
                className="rounded-t transition-all duration-300 hover:opacity-80 relative group"
                style={{ 
                  width: barWidth, 
                  height: barHeight,
                  backgroundColor: color
                }}
              >
                {/* 数值提示 */}
                <div className="absolute -top-8 left-1/2 transform -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity bg-gray-800 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                  {item.value}
                </div>
              </div>
              <div className="text-body-small text-center w-16 truncate">
                {item.label}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}

// 进度条图表
export interface ProgressItem {
  label: string;
  value: number;
  total: number;
  color?: string;
}

export interface ProgressChartProps {
  data: ProgressItem[];
  title?: string;
  className?: string;
  loading?: boolean;
}

function ProgressChartComponent({ 
  data, 
  title, 
  className,
  loading = false 
}: ProgressChartProps) {
  if (loading) {
    return (
      <Card className={cn('chart-container space-y-4', className)}>
        {title && <div className="skeleton h-6 w-32 rounded"></div>}
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <div className="skeleton h-4 w-20 rounded"></div>
              <div className="skeleton h-2 w-full rounded-full"></div>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  return (
    <Card className={cn('chart-container space-y-4', className)}>
      {title && (
        <h3 className="text-display-small font-semibold">{title}</h3>
      )}
      
      <div className="space-y-4">
        {data.map((item, index) => {
          const percentage = (item.value / item.total) * 100;
          const color = item.color ?? `hsl(var(--primary))`;
          
          return (
            <div key={index} className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-body-medium font-medium">{item.label}</span>
                <span className="text-body-small text-gray-500">
                  {item.value}/{item.total} ({percentage.toFixed(1)}%)
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="h-2 rounded-full transition-all duration-500 ease-out"
                  style={{ 
                    width: `${percentage}%`,
                    backgroundColor: color
                  }}
                ></div>
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}

// 趋势线图表（简化版）
export interface TrendPoint {
  label: string;
  value: number;
}

export interface TrendChartProps {
  data: TrendPoint[];
  title?: string;
  description?: string;
  className?: string;
  height?: number;
  color?: string;
  loading?: boolean;
}

function TrendChartComponent({ 
  data, 
  title, 
  description,
  className, 
  height = 200,
  color = 'hsl(var(--primary))',
  loading = false 
}: TrendChartProps) {
  const maxValue = Math.max(...data.map(point => point.value));
  const minValue = Math.min(...data.map(point => point.value));
  const range = maxValue - minValue || 1;
  const width = 400;
  const padding = 40;

  if (loading) {
    return (
      <Card className={cn('chart-container space-y-4', className)}>
        {title && <div className="skeleton h-6 w-32 rounded"></div>}
        <div className="skeleton rounded" style={{ width, height }}></div>
      </Card>
    );
  }

  // 计算路径点
  const points = data.map((point, index) => {
    const x = padding + (index * (width - 2 * padding)) / (data.length - 1);
    const y = padding + ((maxValue - point.value) / range) * (height - 2 * padding);
    return { x, y, ...point };
  });

  const pathData = points
    .map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`)
    .join(' ');

  return (
    <Card className={cn('chart-container space-y-4', className)}>
      {title && (
        <div className="space-y-1">
          <h3 className="text-display-small font-semibold">{title}</h3>
          {description && (
            <p className="text-body-small text-gray-500">{description}</p>
          )}
        </div>
      )}
      
      <div className="flex justify-center">
        <svg width={width} height={height}>
          {/* 网格线 */}
          <defs>
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path d="M 40 0 L 0 0 0 40" fill="none" stroke="#f1f5f9" strokeWidth="1"/>
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
          
          {/* 趋势线 */}
          <path
            d={pathData}
            fill="none"
            stroke={color}
            strokeWidth="2"
            className="transition-all duration-300"
          />
          
          {/* 数据点 */}
          {points.map((point, index) => (
            <circle
              key={index}
              cx={point.x}
              cy={point.y}
              r="4"
              fill={color}
              className="transition-all duration-200 hover:r-6"
            >
              <title>{point.label}: {point.value}</title>
            </circle>
          ))}
        </svg>
      </div>
      
      {/* X轴标签 */}
      <div className="flex justify-between text-body-small text-gray-500 px-10">
        {data.map((point, index) => (
          <span key={index} className="text-center">
            {point.label}
          </span>
        ))}
      </div>
    </Card>
  );
}

// 统一导出所有图表组件
export const PieChart = PieChartComponent;
export const BarChart = BarChartComponent;
export const ProgressChart = ProgressChartComponent;
export const TrendChart = TrendChartComponent;