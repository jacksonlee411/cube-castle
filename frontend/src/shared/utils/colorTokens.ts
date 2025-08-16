/**
 * Canvas Kit v13 颜色token工具类
 * 提供向后兼容的颜色token映射和语义化token
 */

// 基础颜色映射 (向后兼容) - 使用硬编码安全颜色值
export const baseColors = {
  // 基础调色板 - 使用安全的硬编码颜色值替代problematic tokens
  cinnamon: {
    100: '#ffefee',
    200: '#fde0dd',
    300: '#f5b5af',
    400: '#ed7f75',
    500: '#de2e21',
    600: '#a31b12',
  },
  
  peach: {
    100: '#fef5f0',
    400: '#ff7a59',
    600: '#c44b2b',
  },
  
  cantaloupe: {
    100: '#ffeed9',
    300: '#ffcc80',
    400: '#ffa126',
    600: '#c06c00',
    700: '#c06c00', // 使用600的值
  },
  
  // greenFresca映射为greenApple的安全颜色值
  greenFresca: {
    100: '#ebfff0',
    300: '#a3e3b4', 
    600: '#319c4c',
    700: '#217a37',
  },
  
  blueberry: {
    50: '#D7EAFC', // 使用100的安全值
    100: '#D7EAFC',
    400: '#0875E1',
    600: '#004387',
  },
  
  licorice: {
    100: '#A1AAB3',
    200: '#7b858f',
    300: '#5E6A75',
    400: '#4a5561',
    500: '#333d47',
    600: '#1f262e',
    700: '#1f262e', // 使用600的值
  },
  
  soap: {
    50: '#f6f7f8', // 使用100的安全值
    100: '#f6f7f8',
    200: '#F0F1F2',
    300: '#e8ebed',
    400: '#DFE2E6',
    500: '#ced3d9',
  },
  
  frenchVanilla: {
    100: '#ffffff',
  },
  
  blackPepper: {
    100: '#f8f8f8',
    200: '#f0f0f0',
    300: '#494949',
    400: '#333333',
    500: '#1e1e1e',
    600: '#000000',
  }
};

// 状态颜色快捷方式
export const statusColors = {
  success: {
    bg: '#ebfff0',  // greenApple100
    text: '#319c4c', // greenApple600
    lighter: '#ebfff0', // greenApple100
  },
  
  warning: {
    bg: '#ffeed9',  // cantaloupe100
    text: '#c06c00', // cantaloupe600
    lighter: '#ffeed9', // cantaloupe100
  },
  
  error: {
    bg: '#ffefee',  // cinnamon100
    text: '#a31b12', // cinnamon600
    lighter: '#ffefee', // cinnamon100
  },
  
  info: {
    bg: '#D7EAFC',  // blueberry100
    text: '#004387', // blueberry600
    lighter: '#D7EAFC', // blueberry100
  },
  
  neutral: {
    bg: '#e8ebed',  // soap300
    text: '#5E6A75', // licorice300
    lighter: '#f6f7f8', // soap100
  }
};

// 向后兼容的颜色映射 (用于迁移)
export const legacyColors = {
  // 保持原有的颜色名称但使用安全的硬编码颜色值
  greenFresca100: '#ebfff0',
  greenFresca600: '#319c4c',
  
  blueberry100: '#D7EAFC',
  blueberry400: '#0875E1',
  blueberry600: '#004387',
  
  cinnamon100: '#ffefee',
  cinnamon600: '#a31b12',
  
  peach600: '#c44b2b',
  
  cantaloupe600: '#c06c00',
  
  licorice400: '#4a5561',
  licorice500: '#333d47',
  licorice600: '#1f262e',
  
  soap100: '#f6f7f8',
  soap200: '#F0F1F2',
  soap300: '#e8ebed',
  soap400: '#DFE2E6',
  
  frenchVanilla100: '#ffffff',
  
  // 常用硬编码颜色的安全替代
  success: '#2ECC71',
  info: '#3498DB', 
  warning: '#F39C12',
  error: '#E74C3C',
  neutral: '#666666',
};

export default {
  base: baseColors,
  status: statusColors,
  legacy: legacyColors,
};