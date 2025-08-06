import React from 'react';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';

describe('Canvas Kit Integration', () => {
  it('应该成功渲染实际Canvas组件', () => {
    const TestComponent = () => {
      const { CanvasProvider } = require('@workday/canvas-kit-react/common');
      return (
        <CanvasProvider>
          <div data-testid="test-content">Canvas Provider 集成测试</div>
        </CanvasProvider>
      );
    };
    
    render(<TestComponent />);
    
    // 检查Canvas Provider是否正确渲染了内容
    expect(screen.getByTestId('test-content')).toBeInTheDocument();
    
    // 检查Canvas样式是否应用 (通过className)
    const container = screen.getByTestId('test-content').parentElement;
    expect(container).toHaveClass(/css-/);
  });

  it('应该正确渲染Canvas Box组件', () => {
    const TestComponent = () => {
      const { Box } = require('@workday/canvas-kit-react/layout');
      return (
        <Box marginBottom="l" padding="m" data-testid="test-box">
          Canvas Box 测试
        </Box>
      );
    };
    
    render(<TestComponent />);
    
    expect(screen.getByTestId('test-box')).toBeInTheDocument();
    expect(screen.getByText('Canvas Box 测试')).toBeInTheDocument();
    
    // 检查Canvas样式是否应用
    const box = screen.getByTestId('test-box');
    expect(box).toHaveClass(/css-/);
  });

  it('应该支持Canvas设计令牌系统', () => {
    // 测试Canvas设计令牌是否可以正确导入
    expect(() => {
      require('@workday/canvas-tokens-web');
    }).not.toThrow();
    
    // 测试字体是否可以正确导入
    expect(() => {
      require('@workday/canvas-kit-react-fonts');
    }).not.toThrow();
  });

  it('应该支持Canvas按钮组件', () => {
    const TestComponent = () => {
      const { PrimaryButton } = require('@workday/canvas-kit-react/button');
      return (
        <PrimaryButton data-testid="test-button">
          测试按钮
        </PrimaryButton>
      );
    };
    
    render(<TestComponent />);
    
    expect(screen.getByTestId('test-button')).toBeInTheDocument();
    expect(screen.getByText('测试按钮')).toBeInTheDocument();
  });
});