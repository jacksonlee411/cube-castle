/**
 * Canvas Kit v13兼容的Badge组件
 * 使用Canvas Kit v13语义化颜色系统
 */
import React from 'react';
import styled from '@emotion/styled';
import { baseColors, statusColors } from '../utils/colorTokens';

export interface BadgeProps {
  children: React.ReactNode;
  variant?: 'positive' | 'caution' | 'negative' | 'neutral' | 'outline';
  size?: 'small' | 'medium';
  color?: string;
}

const StyledBadge = styled.span<{
  variant: BadgeProps['variant'];
  size: BadgeProps['size'];
  color?: string;
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
  
  ${({ size }) => size === 'small' ? `
    padding: 2px 6px;
    font-size: 11px;
    min-height: 16px;
  ` : `
    padding: 4px 8px;
    font-size: 12px;
    min-height: 20px;
  `}
  
  ${({ variant, color }) => {
    if (color) {
      return `
        background-color: ${color};
        color: white;
        border: 1px solid ${color};
      `;
    }
    
    switch (variant) {
      case 'positive':
        return `
          background-color: ${statusColors.success.lighter};
          color: ${baseColors.greenFresca[600]};
          border: 1px solid ${baseColors.greenFresca[300]};
        `;
      case 'caution':
        return `
          background-color: ${statusColors.warning.lighter};
          color: ${baseColors.cantaloupe[600]};
          border: 1px solid ${baseColors.cantaloupe[300]};
        `;
      case 'negative':
        return `
          background-color: ${statusColors.error.lighter};
          color: ${baseColors.cinnamon[600]};
          border: 1px solid ${baseColors.cinnamon[300]};
        `;
      case 'outline':
        return `
          background-color: transparent;
          color: ${baseColors.licorice[500]};
          border: 1px solid ${baseColors.licorice[300]};
        `;
      default: // neutral
        return `
          background-color: ${statusColors.neutral.lighter};
          color: ${baseColors.licorice[600]};
          border: 1px solid ${baseColors.licorice[300]};
        `;
    }
  }}
`;

export const Badge: React.FC<BadgeProps> = ({ 
  children, 
  variant = 'neutral', 
  size = 'medium', 
  color 
}) => {
  return (
    <StyledBadge variant={variant} size={size} color={color}>
      {children}
    </StyledBadge>
  );
};

export default Badge;