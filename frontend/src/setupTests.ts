import React from 'react';
import { vi } from 'vitest';
import '@testing-library/jest-dom';

// 定义通用的React组件props类型
type MockComponentProps = React.PropsWithChildren<Record<string, unknown>>;

// Mock Canvas Kit components to avoid CSS issues in tests
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: ({ children, ...props }: MockComponentProps) => {
    const { marginBottom: _marginBottom, paddingY: _paddingY, borderBottom: _borderBottom, borderColor: _borderColor, paddingTop: _paddingTop, ...cleanProps } = props;
    return React.createElement('div', { 'data-testid': 'canvas-box', ...cleanProps }, children);
  }
}));

vi.mock('@workday/canvas-kit-react/button', () => ({
  PrimaryButton: ({ children, ...props }: MockComponentProps) => {
    const { marginRight: _marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'primary-button', ...cleanProps }, children);
  },
  SecondaryButton: ({ children, ...props }: MockComponentProps) => {
    const { marginRight: _marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'secondary-button', ...cleanProps }, children);
  },
  TertiaryButton: ({ children, ...props }: MockComponentProps) => {
    const { marginRight: _marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'tertiary-button', ...cleanProps }, children);
  }
}));

vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children }: MockComponentProps) => React.createElement('h1', { 'data-testid': 'canvas-heading' }, children),
  Text: ({ children }: MockComponentProps) => React.createElement('span', { 'data-testid': 'canvas-text' }, children)
}));

vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: Object.assign(
    ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'canvas-card' }, children),
    {
      Heading: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'card-heading' }, children),
      Body: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'card-body' }, children)
    }
  )
}));

vi.mock('@workday/canvas-kit-react/table', () => ({
  Table: Object.assign(
    ({ children }: MockComponentProps) => React.createElement('table', { 'data-testid': 'canvas-table' }, children),
    {
      Head: ({ children }: MockComponentProps) => React.createElement('thead', { 'data-testid': 'table-head' }, children),
      Body: ({ children }: MockComponentProps) => React.createElement('tbody', { 'data-testid': 'table-body' }, children),
      Row: ({ children }: MockComponentProps) => React.createElement('tr', { 'data-testid': 'table-row' }, children),
      Header: ({ children }: MockComponentProps) => React.createElement('th', { 'data-testid': 'table-header' }, children),
      Cell: ({ children }: MockComponentProps) => React.createElement('td', { 'data-testid': 'table-cell' }, children)
    }
  )
}));

vi.mock('@workday/canvas-kit-react/side-panel', () => ({
  SidePanel: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'side-panel' }, children)
}));

vi.mock('@workday/canvas-kit-react/avatar', () => ({
  Avatar: ({ children, altText }: MockComponentProps & { altText?: string }) => React.createElement('div', { 'data-testid': 'avatar', 'aria-label': altText }, children)
}));

vi.mock('@workday/canvas-kit-react/common', () => ({
  CanvasProvider: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'canvas-provider' }, children)
}));