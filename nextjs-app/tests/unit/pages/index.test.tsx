// tests/unit/pages/index.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/router';
import HomePage from '../../../src/pages/index';
import '@testing-library/jest-dom';

// Mock next/router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockRouter = {
  push: jest.fn(),
  pathname: '/',
  route: '/',
  query: {},
  asPath: '/',
};

describe('HomePage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
  });

  it('should render the main title correctly', () => {
    render(<HomePage />);
    
    expect(screen.getByText('员工模型管理系统')).toBeInTheDocument();
    expect(screen.getByText('Employee Model Management System v2.0')).toBeInTheDocument();
  });

  it('should display all feature cards', () => {
    render(<HomePage />);
    
    // Check all feature cards are present
    expect(screen.getByText('员工管理')).toBeInTheDocument();
    expect(screen.getByText('职位历史')).toBeInTheDocument();
    expect(screen.getByText('组织架构')).toBeInTheDocument();
    expect(screen.getByText('SAM 智能分析')).toBeInTheDocument();
    expect(screen.getByText('工作流管理')).toBeInTheDocument();

    // Check feature descriptions
    expect(screen.getByText('全面的员工信息管理，支持CRUD操作和实时数据更新')).toBeInTheDocument();
    expect(screen.getByText('时态数据查询，追踪员工职位变更历史和时间线')).toBeInTheDocument();
    expect(screen.getByText('可视化组织结构，Neo4j图数据库支持的层级关系')).toBeInTheDocument();
    expect(screen.getByText('AI驱动的组织分析，实时监控和智能决策支持')).toBeInTheDocument();
    expect(screen.getByText('Temporal.io驱动的业务流程自动化和审批管理')).toBeInTheDocument();
  });

  it('should display statistics correctly', () => {
    render(<HomePage />);
    
    expect(screen.getByText('89%')).toBeInTheDocument();
    expect(screen.getByText('优秀')).toBeInTheDocument();
    expect(screen.getByText('95%')).toBeInTheDocument();
    expect(screen.getByText('100%')).toBeInTheDocument();

    expect(screen.getByText('测试覆盖率')).toBeInTheDocument();
    expect(screen.getByText('性能指标')).toBeInTheDocument();
    expect(screen.getByText('生产就绪')).toBeInTheDocument();
    expect(screen.getByText('功能完整性')).toBeInTheDocument();
  });

  it('should navigate to employees page when clicking "开始使用" button', async () => {
    render(<HomePage />);
    
    const startButton = screen.getByText('开始使用');
    fireEvent.click(startButton);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/employees');
    });
  });

  it('should navigate to SAM dashboard when clicking "AI 分析" button', async () => {
    render(<HomePage />);
    
    // Get the AI 分析 button specifically 
    const buttons = screen.getAllByText('AI 分析');
    const aiButton = buttons[0]; // Get the first one (from the hero section)
    fireEvent.click(aiButton);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/sam/dashboard');
    });
  });

  it('should navigate to correct paths when clicking feature cards', async () => {
    render(<HomePage />);
    
    // Test employee management card
    const employeeCard = screen.getByText('员工管理').closest('.ant-card');
    expect(employeeCard).toBeInTheDocument();
    fireEvent.click(employeeCard!);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/employees');
    });

    // Test organization chart card
    const orgCard = screen.getByText('组织架构').closest('.ant-card');
    expect(orgCard).toBeInTheDocument();
    fireEvent.click(orgCard!);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/organization/chart');
    });

    // Test SAM dashboard card
    const samCard = screen.getByText('SAM 智能分析').closest('.ant-card');
    expect(samCard).toBeInTheDocument();
    fireEvent.click(samCard!);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/sam/dashboard');
    });

    // Test workflow demo card
    const workflowCard = screen.getByText('工作流管理').closest('.ant-card');
    expect(workflowCard).toBeInTheDocument();
    fireEvent.click(workflowCard!);
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/workflows/demo');
    });
  });

  it('should display technology stack information', () => {
    render(<HomePage />);
    
    expect(screen.getByText('技术架构')).toBeInTheDocument();
    expect(screen.getByText('后端技术')).toBeInTheDocument();
    expect(screen.getByText('数据层')).toBeInTheDocument();
    expect(screen.getByText('前端技术')).toBeInTheDocument();

    // Check technology stack items
    expect(screen.getByText('Go + Ent ORM')).toBeInTheDocument();
    expect(screen.getByText('GraphQL API')).toBeInTheDocument();
    expect(screen.getByText('Temporal.io 工作流')).toBeInTheDocument();
    expect(screen.getByText('PostgreSQL 时态表')).toBeInTheDocument();
    expect(screen.getByText('Neo4j 图数据库')).toBeInTheDocument();
    expect(screen.getByText('React + Next.js')).toBeInTheDocument();
    expect(screen.getByText('Apollo GraphQL')).toBeInTheDocument();
  });

  it('should display highlight tags for features', () => {
    render(<HomePage />);
    
    expect(screen.getByText('核心功能')).toBeInTheDocument();
    expect(screen.getByText('时态数据')).toBeInTheDocument();
    expect(screen.getByText('图数据库')).toBeInTheDocument();
    // Check for all AI 分析 tags (there should be 2: button and highlight tag)
    expect(screen.getAllByText('AI 分析')).toHaveLength(2);
    expect(screen.getByText('工作流')).toBeInTheDocument();
  });

  it('should have proper accessibility attributes', () => {
    render(<HomePage />);
    
    // Check for heading levels
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(1);
    expect(screen.getAllByRole('heading', { level: 3 })).toHaveLength(2); // subtitle + tech section
    // 3 tech sections + 5 feature cards = 8 total
    expect(screen.getAllByRole('heading', { level: 4 })).toHaveLength(8);
    
    // Check for buttons
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThanOrEqual(7); // 2 hero buttons + 5 feature card buttons
    
    // Check that feature cards are clickable
    const featureCards = screen.getAllByText('访问模块');
    expect(featureCards).toHaveLength(5);
  });

  it('should render footer information', () => {
    render(<HomePage />);
    
    expect(screen.getByText('Employee Model Management System v2.0 - 生产就绪版本')).toBeInTheDocument();
    expect(screen.getByText('基于现代化技术栈构建，支持大规模企业级应用')).toBeInTheDocument();
  });

  it('should have responsive layout structure', () => {
    render(<HomePage />);
    
    // Check for main container - find the parent div that contains the title
    const titleElement = screen.getByText('员工模型管理系统');
    const mainContainer = titleElement.closest('div[style*="padding: 24px"]');
    expect(mainContainer).toBeInTheDocument();
  });

  it('should handle hover effects on feature cards', async () => {
    render(<HomePage />);
    
    const employeeCard = screen.getByText('员工管理').closest('.ant-card');
    expect(employeeCard).toBeInTheDocument();
    
    // The card should have hoverable property which adds hover effects
    expect(employeeCard).toHaveClass('ant-card-hoverable');
  });
});