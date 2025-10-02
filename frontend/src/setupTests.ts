import React from 'react';
import { vi } from 'vitest';
import '@testing-library/jest-dom';

// 定义通用的React组件props类型
type MockComponentProps = React.PropsWithChildren<React.HTMLAttributes<HTMLElement>>;

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

  type ComboboxOption = string | { code: string; name: string }

  interface ComboboxModelConfig {
    items?: ComboboxOption[]
    getId?: (item: ComboboxOption) => string
    getTextValue?: (item: ComboboxOption) => string
  }

  interface ComboboxModel {
    items: ComboboxOption[]
    events: {
      hide: ReturnType<typeof vi.fn>
      show: ReturnType<typeof vi.fn>
      select: ReturnType<typeof vi.fn>
      setSelectedIds: ReturnType<typeof vi.fn>
      unselectAll: ReturnType<typeof vi.fn>
      setWidth: ReturnType<typeof vi.fn>
    }
    state: {
      items: ComboboxOption[]
      selectedIds: string[]
      value: string
      visibility: 'hidden' | 'visible'
    }
    navigation: { getItem: () => undefined }
    getId: (item: ComboboxOption) => string
    getTextValue: (item: ComboboxOption) => string
  }

  type ComboProps = {
    items?: ComboboxOption[]
    onChange?: (val: string) => void
    disabled?: boolean
    children?: React.ReactNode
    model?: ComboboxModel
  }

  type InputProps = {
    value?: string
    onChange?: React.ChangeEventHandler<HTMLInputElement>
    placeholder?: string
    disabled?: boolean
  }

  const defaultGetId = (item: ComboboxOption) => (typeof item === 'string' ? item : item.code || 'item')
  const defaultGetTextValue = (item: ComboboxOption) =>
    typeof item === 'string' ? item : `${item.code} - ${item.name}`

  return {
    Combobox: Object.assign(
      ({ items = [], onChange, disabled, children, model }: ComboProps) => {
        const safeChildren: React.ReactNode[] = ReactLocal.Children.toArray(children).filter((c) => typeof c !== 'function')
        return ReactLocal.createElement('div', { 'data-testid': 'combobox' },
          ...safeChildren,
          ReactLocal.createElement('div', { 'data-testid': 'combobox-items' },
            (model?.items || items).map((option) => {
              const key = defaultGetId(option)
              const label = defaultGetTextValue(option)
              const code = typeof option === 'string' ? option : option.code
              return ReactLocal.createElement('button', {
                key,
                'data-testid': `combobox-item-${key}`,
                disabled,
                onClick: () => onChange?.(code)
              }, label)
            })
          )
        )
      },
      {
        Input: ({ value, onChange, placeholder, disabled }: InputProps) => ReactLocal.createElement('input', {
          'data-testid': 'combobox-input', value: value || '', onChange, placeholder, disabled
        }),
        Menu: Object.assign(
          ({ children }: MockComponentProps) => {
            const safe: React.ReactNode[] = ReactLocal.Children.toArray(children).filter((c) => typeof c !== 'function')
            return ReactLocal.createElement('div', { 'data-testid': 'combobox-menu' }, ...safe)
          },
          {
            Popper: ({ children }: MockComponentProps) => {
              const safe: React.ReactNode[] = ReactLocal.Children.toArray(children).filter((c) => typeof c !== 'function')
              return ReactLocal.createElement('div', { 'data-testid': 'combobox-menu-popper' }, ...safe)
            },
            Card: ({ children }: MockComponentProps) => {
              const safe: React.ReactNode[] = ReactLocal.Children.toArray(children).filter((c) => typeof c !== 'function')
              return ReactLocal.createElement('div', { 'data-testid': 'combobox-menu-card' }, ...safe)
            },
            List: ({ children }: MockComponentProps) => {
              const safe: React.ReactNode[] = ReactLocal.Children.toArray(children).filter((c) => typeof c !== 'function')
              return ReactLocal.createElement('div', { 'data-testid': 'combobox-menu-list' }, ...safe)
            },
            Item: ({ children }: MockComponentProps) => ReactLocal.createElement('div', { 'data-testid': 'combobox-menu-item' }, children)
          }
        )
      }
    ),
    useComboboxModel: ({ items = [], getId, getTextValue }: ComboboxModelConfig = {}): ComboboxModel => {
      const state = {
        items,
        selectedIds: [] as string[],
        value: '',
        visibility: 'hidden' as 'hidden' | 'visible'
      }
      const events = {
        hide: vi.fn(() => {
          state.visibility = 'hidden'
        }),
        show: vi.fn(() => {
          state.visibility = 'visible'
        }),
        select: vi.fn((data?: { id: string }) => {
          if (data?.id) {
            state.selectedIds = [data.id]
          }
        }),
        setSelectedIds: vi.fn((ids: string[]) => {
          state.selectedIds = ids
        }),
        unselectAll: vi.fn(() => {
          state.selectedIds = []
        }),
        setWidth: vi.fn()
      }

      return {
        items,
        events,
        state,
        navigation: {
          getItem: () => undefined
        },
        getId: getId ?? defaultGetId,
        getTextValue: getTextValue ?? defaultGetTextValue
      }
    }
  }
})

// FormField 简易 mock
vi.mock('@workday/canvas-kit-react/form-field', () => ({
  FormField: Object.assign(
    ({ children, error, model: _model }: MockComponentProps & { error?: string; model?: Record<string, never> }) => React.createElement('div', { 'data-testid': 'form-field', 'data-error': error || '' }, children),
    {
      Label: ({ children, required }: MockComponentProps & { required?: boolean }) => React.createElement('label', { 'data-testid': 'form-field-label', 'data-required': !!required }, children),
      Hint: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'form-field-hint' }, children),
      Error: ({ children }: MockComponentProps) => React.createElement('div', { role: 'alert', 'data-testid': 'form-field-error' }, children),
      Field: ({ children }: MockComponentProps) => React.createElement('div', { 'data-testid': 'form-field-field' }, children)
    }
  ),
  useFormFieldModel: ({ isRequired, error }: { isRequired?: boolean; error?: string }) => ({
    isRequired: !!isRequired,
    error: error ?? undefined,
    state: { error: error ?? undefined }
  })
}));
