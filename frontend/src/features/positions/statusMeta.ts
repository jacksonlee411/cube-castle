import { colors } from '@workday/canvas-kit-react/tokens';
import type { PositionStatus } from '@/shared/types/positions';

export interface PositionStatusMeta {
  label: string;
  color: string;
  background: string;
  border: string;
}

const POSITION_STATUS_META: Record<string, PositionStatusMeta> = {
  ACTIVE: {
    label: '在编',
    color: colors.greenApple600,
    background: colors.greenApple100,
    border: colors.greenApple300,
  },
  FILLED: {
    label: '已填充',
    color: colors.blueberry600,
    background: colors.blueberry100,
    border: colors.blueberry300,
  },
  VACANT: {
    label: '空缺',
    color: colors.cantaloupe600,
    background: colors.cantaloupe100,
    border: colors.cantaloupe300,
  },
  PLANNED: {
    label: '规划中',
    color: colors.soap600,
    background: colors.soap100,
    border: colors.soap300,
  },
  INACTIVE: {
    label: '停用',
    color: colors.peach600,
    background: colors.peach100,
    border: colors.peach300,
  },
};

export const getPositionStatusMeta = (status: PositionStatus | string): PositionStatusMeta => {
  const key = status?.toString().toUpperCase();
  return POSITION_STATUS_META[key] ?? {
    label: key || '未知',
    color: colors.licorice500,
    background: colors.soap100,
    border: colors.soap400,
  };
};
