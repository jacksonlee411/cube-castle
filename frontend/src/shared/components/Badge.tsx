/**
 * Canvas Kit兼容的Badge组件
 * 解决Canvas Kit 13.x版本中Badge组件变更问题
 */
import React from 'react';
import styled from '@emotion/styled';
import { colors, space } from '@workday/canvas-kit-react/tokens';

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
          background-color: ${colors.greenFresca100};
          color: ${colors.greenFresca600};
          border: 1px solid ${colors.greenFresca300};
        `;
      case 'caution':
        return `
          background-color: ${colors.cantaloupe100};
          color: ${colors.cantaloupe600};
          border: 1px solid ${colors.cantaloupe300};
        `;
      case 'negative':
        return `
          background-color: ${colors.cinnamon100};
          color: ${colors.cinnamon600};
          border: 1px solid ${colors.cinnamon300};
        `;
      case 'outline':
        return `
          background-color: transparent;
          color: ${colors.licorice500};
          border: 1px solid ${colors.licorice300};
        `;
      default: // neutral
        return `
          background-color: ${colors.licorice100};
          color: ${colors.licorice600};
          border: 1px solid ${colors.licorice300};
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