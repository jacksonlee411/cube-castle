// @vitest-environment jsdom
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';

// Mock Canvas Kit layout/text/tokens to simple shims for JSDOM
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Flex: (props: any) => {
    const { children, ...rest } = props;
    return (<div {...rest}>{children}</div>);
  }
}));

vi.mock('@workday/canvas-kit-react/text', () => ({
  Text: (props: any) => {
    const { children, ...rest } = props;
    return (<span {...rest}>{children}</span>);
  }
}));

vi.mock('@workday/canvas-kit-react/tokens', () => ({
  space: { xs: 4 },
  colors: { licorice500: '#333' }
}));

describe('OrganizationBreadcrumb', () => {
  it('renders namePath as readable breadcrumb', async () => {
    const { OrganizationBreadcrumb } = await import('../../src/shared/components/OrganizationBreadcrumb');
    render(
      <OrganizationBreadcrumb
        codePath="/100/200/300"
        namePath="/集团/技术部/研发组"
      />
    );

    expect(screen.getByText('集团')).toBeInTheDocument();
    expect(screen.getByText('技术部')).toBeInTheDocument();
    expect(screen.getByText('研发组')).toBeInTheDocument();
  });

  it('fires onNavigate with clicked code', async () => {
    const onNavigate = vi.fn();
    const { OrganizationBreadcrumb } = await import('../../src/shared/components/OrganizationBreadcrumb');
    render(
      <OrganizationBreadcrumb
        codePath="/100/200/300"
        namePath="/集团/技术部/研发组"
        onNavigate={onNavigate}
      />
    );

    fireEvent.click(screen.getByText('技术部'));
    expect(onNavigate).toHaveBeenCalledWith('200');
  });
});
