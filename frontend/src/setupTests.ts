import React from 'react';
import { vi } from 'vitest';
import '@testing-library/jest-dom';

// Mock Canvas Kit components to avoid CSS issues in tests
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: ({ children, ...props }: any) => {
    const { marginBottom, paddingY, borderBottom, borderColor, paddingTop, ...cleanProps } = props;
    return React.createElement('div', { 'data-testid': 'canvas-box', ...cleanProps }, children);
  }
}));

vi.mock('@workday/canvas-kit-react/button', () => ({
  PrimaryButton: ({ children, ...props }: any) => {
    const { marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'primary-button', ...cleanProps }, children);
  },
  SecondaryButton: ({ children, ...props }: any) => {
    const { marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'secondary-button', ...cleanProps }, children);
  },
  TertiaryButton: ({ children, ...props }: any) => {
    const { marginRight, ...cleanProps } = props;
    return React.createElement('button', { 'data-testid': 'tertiary-button', ...cleanProps }, children);
  }
}));

vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children, ...props }: any) => React.createElement('h1', { 'data-testid': 'canvas-heading' }, children),
  Text: ({ children, ...props }: any) => React.createElement('span', { 'data-testid': 'canvas-text' }, children)
}));

vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: Object.assign(
    ({ children }: any) => React.createElement('div', { 'data-testid': 'canvas-card' }, children),
    {
      Heading: ({ children }: any) => React.createElement('div', { 'data-testid': 'card-heading' }, children),
      Body: ({ children }: any) => React.createElement('div', { 'data-testid': 'card-body' }, children)
    }
  )
}));

vi.mock('@workday/canvas-kit-react/table', () => ({
  Table: Object.assign(
    ({ children }: any) => React.createElement('table', { 'data-testid': 'canvas-table' }, children),
    {
      Head: ({ children }: any) => React.createElement('thead', { 'data-testid': 'table-head' }, children),
      Body: ({ children }: any) => React.createElement('tbody', { 'data-testid': 'table-body' }, children),
      Row: ({ children }: any) => React.createElement('tr', { 'data-testid': 'table-row' }, children),
      Header: ({ children }: any) => React.createElement('th', { 'data-testid': 'table-header' }, children),
      Cell: ({ children }: any) => React.createElement('td', { 'data-testid': 'table-cell' }, children)
    }
  )
}));

vi.mock('@workday/canvas-kit-react/side-panel', () => ({
  SidePanel: ({ children }: any) => React.createElement('div', { 'data-testid': 'side-panel' }, children)
}));

vi.mock('@workday/canvas-kit-react/avatar', () => ({
  Avatar: ({ children, altText }: any) => React.createElement('div', { 'data-testid': 'avatar', 'aria-label': altText }, children)
}));

vi.mock('@workday/canvas-kit-react/common', () => ({
  CanvasProvider: ({ children }: any) => React.createElement('div', { 'data-testid': 'canvas-provider' }, children)
}));