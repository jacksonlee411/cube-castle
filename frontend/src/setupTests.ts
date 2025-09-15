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
  },
  Flex: ({ children, ...props }: MockComponentProps) => {
    const { alignItems: _alignItems, justifyContent: _justifyContent, gap: _gap, ...cleanProps } = props;
    return React.createElement('div', { 'data-testid': 'canvas-flex', ...cleanProps }, children);
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

// Combobox 简易可交互 mock：渲染输入框与按钮项
vi.mock('@workday/canvas-kit-react/combobox', () => {
  const ReactLocal = React
  return {
    Combobox: Object.assign(
      ({ items = [], onChange, disabled, children }: any) => {
        const safeChildren = ReactLocal.Children.toArray(children).filter((c: any) => typeof c !== 'function')
        return ReactLocal.createElement('div', { 'data-testid': 'combobox' },
          ...safeChildren as any,
          ReactLocal.createElement('div', { 'data-testid': 'combobox-items' },
            items.map((it: string) => ReactLocal.createElement('button', {
              key: it,
              'data-testid': `combobox-item-${it}`,
              disabled,
              onClick: () => onChange && onChange(it)
            }, it))
          )
        )
      },
      {
        Input: ({ value, onChange, placeholder, disabled }: any) => ReactLocal.createElement('input', {
          'data-testid': 'combobox-input', value: value || '', onChange, placeholder, disabled
        }),
        Menu: ({ children }: MockComponentProps) => {
          const safe = ReactLocal.Children.toArray(children).filter((c: any) => typeof c !== 'function')
          return ReactLocal.createElement('div', { 'data-testid': 'combobox-menu' }, ...safe as any)
        },
        MenuList: ({ children }: MockComponentProps) => {
          const safe = ReactLocal.Children.toArray(children).filter((c: any) => typeof c !== 'function')
          return ReactLocal.createElement('div', { 'data-testid': 'combobox-menulist' }, ...safe as any)
        },
        Item: ({ children }: MockComponentProps) => ReactLocal.createElement('div', { 'data-testid': 'combobox-item' }, children)
      }
    )
  }
})

// FormField 简易 mock
vi.mock('@workday/canvas-kit-react/form-field', () => ({
  FormField: Object.assign(
    ({ children, error }: MockComponentProps & { error?: string }) => React.createElement('div', { 'data-testid': 'form-field', 'data-error': error || '' }, children),
    {
      Label: ({ children, required }: MockComponentProps & { required?: boolean }) => React.createElement('label', { 'data-testid': 'form-field-label', 'data-required': !!required }, children),
      Hint: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'form-field-hint' }, children),
      Error: ({ children }: MockComponentProps) => React.createElement('div', { role: 'alert', 'data-testid': 'form-field-error' }, children)
    }
  )
}));
