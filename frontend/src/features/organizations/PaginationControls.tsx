import React from 'react';
import { Text } from '@workday/canvas-kit-react/text';
import { SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';

interface PaginationControlsProps {
  currentPage: number;
  totalCount: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  disabled?: boolean;
}

export const PaginationControls: React.FC<PaginationControlsProps> = ({
  currentPage,
  totalCount,
  pageSize,
  onPageChange,
  disabled = false,
}) => {
  const totalPages = Math.ceil(totalCount / pageSize);
  const startItem = (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalCount);

  // 如果总页数小于等于1，不显示分页控件
  if (totalPages <= 1) {
    return (
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 0' }}>
        <Text typeLevel="subtext.small" color="hint">
          共 {totalCount} 条记录
        </Text>
      </div>
    );
  }

  // 计算显示的页码范围
  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const maxVisiblePages = 7; // 最多显示7个页码按钮
    
    if (totalPages <= maxVisiblePages) {
      // 如果总页数不超过最大显示数，显示所有页码
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // 复杂的分页逻辑：显示首页、末页、当前页附近的页码，用省略号分隔
      if (currentPage <= 4) {
        // 当前页在前面，显示 1,2,3,4,5...最后一页
        for (let i = 1; i <= 5; i++) {
          pages.push(i);
        }
        pages.push('...');
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 3) {
        // 当前页在后面，显示 1...倒数5,4,3,2,1页
        pages.push(1);
        pages.push('...');
        for (let i = totalPages - 4; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        // 当前页在中间，显示 1...当前页-1,当前页,当前页+1...最后一页
        pages.push(1);
        pages.push('...');
        for (let i = currentPage - 1; i <= currentPage + 1; i++) {
          pages.push(i);
        }
        pages.push('...');
        pages.push(totalPages);
      }
    }
    
    return pages;
  };

  const pageNumbers = getPageNumbers();

  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 0' }}>
      {/* 左侧：记录信息 */}
      <Text typeLevel="subtext.small" color="hint">
        显示第 {startItem} - {endItem} 条，共 {totalCount} 条记录
      </Text>

      {/* 右侧：分页控件 */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        {/* 上一页按钮 */}
        <TertiaryButton
          size="small"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={disabled || currentPage <= 1}
          aria-label="上一页"
        >
          ←
        </TertiaryButton>

        {/* 页码按钮 */}
        {pageNumbers.map((page, index) => {
          if (page === '...') {
            return (
              <Text key={`ellipsis-${index}`} style={{ margin: '0 8px' }}>
                ...
              </Text>
            );
          }

          const pageNum = page as number;
          const isCurrentPage = pageNum === currentPage;

          return (
            <SecondaryButton
              key={pageNum}
              size="small"
              onClick={() => onPageChange(pageNum)}
              disabled={disabled}
              style={{
                backgroundColor: isCurrentPage ? '#1565c0' : undefined,
                color: isCurrentPage ? 'white' : undefined,
                minWidth: '32px',
              }}
            >
              {pageNum}
            </SecondaryButton>
          );
        })}

        {/* 下一页按钮 */}
        <TertiaryButton
          size="small"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={disabled || currentPage >= totalPages}
          aria-label="下一页"
        >
          →
        </TertiaryButton>

        {/* 页码信息 */}
        <Text typeLevel="subtext.small" color="hint" style={{ marginLeft: '16px' }}>
          {currentPage} / {totalPages}
        </Text>
      </div>
    </div>
  );
};