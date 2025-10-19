import React from 'react';
import { vi } from 'vitest';
import '@testing-library/jest-dom';

vi.mock('@/shared/utils/logger', async () => {
  const actual = await vi.importActual<typeof import('@/shared/utils/logger')>(
    '@/shared/utils/logger'
  );

  const mockLogger = {
    debug: vi.fn(),
    info: vi.fn(),
    log: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    group: vi.fn(),
    groupEnd: vi.fn(),
    mutation: vi.fn()
  } as const;

  return {
    ...actual,
    logger: mockLogger
  };
});

// 定义通用的React组件props类型
type MockComponentProps = React.PropsWithChildren<Record<string, unknown>>;

// Mock Canvas Kit components to avoid CSS issues in tests
vi.mock('@workday/canvas-kit-react/layout', () => {
  const stripLayoutProps = (props: Record<string, unknown>) => {
    const {
      marginBottom: _marginBottom,
      marginTop: _marginTop,
      marginY: _marginY,
      marginX: _marginX,
      margin: _margin,
      padding: _padding,
      paddingTop: _paddingTop,
      paddingBottom: _paddingBottom,
      paddingLeft: _paddingLeft,
      paddingRight: _paddingRight,
      paddingY: _paddingY,
      paddingX: _paddingX,
      border: _border,
      borderRadius: _borderRadius,
      borderBottom: _borderBottom,
      borderLeft: _borderLeft,
      borderRight: _borderRight,
      borderColor: _borderColor,
      backgroundColor: _backgroundColor,
      flex: _flex,
      flexWrap: _flexWrap,
      flexDirection: _flexDirection,
      alignItems: _alignItems,
      justifyContent: _justifyContent,
      gap: _gap,
      minWidth: _minWidth,
      width: _width,
      height: _height,
      color: _color,
      as: asProp,
      style,
      ...cleanProps
    } = props
    const component = (asProp as keyof HTMLElementTagNameMap | undefined) ?? 'div'
    return { component, cleanProps: { ...cleanProps, style } }
  }

  const Box = ({ children, ...props }: MockComponentProps & { as?: keyof HTMLElementTagNameMap }) => {
    const { component, cleanProps } = stripLayoutProps(props)
    return React.createElement(component, { 'data-testid': 'canvas-box', ...cleanProps }, children)
  }

  const Flex = ({ children, ...props }: MockComponentProps & { as?: keyof HTMLElementTagNameMap }) => {
    const { component, cleanProps } = stripLayoutProps(props)
    return React.createElement(component, { 'data-testid': 'canvas-flex', ...cleanProps }, children)
  }

  const Stack = ({ children, ...props }: MockComponentProps & { as?: keyof HTMLElementTagNameMap }) => {
    const { component, cleanProps } = stripLayoutProps(props)
    return React.createElement(component, { 'data-testid': 'canvas-stack', ...cleanProps }, children)
  }

  return { Box, Flex, Stack }
})

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

vi.mock('@workday/canvas-kit-react/modal', () => {
  const ModalComponent = ({ children, onClose: _onClose, model: _model, ...rest }: MockComponentProps & { onClose?: () => void; model?: unknown }) => (
    React.createElement('div', { 'data-testid': 'canvas-modal', ...rest }, children)
  )

  ModalComponent.Overlay = ({ children }: MockComponentProps) => (
    React.createElement('div', { 'data-testid': 'canvas-modal-overlay' }, children)
  )

  ModalComponent.Card = ({ children, ...props }: MockComponentProps) => (
    React.createElement('div', { 'data-testid': 'canvas-modal-card', ...props }, children)
  )

  ModalComponent.CloseIcon = ({ onClick, ...props }: { onClick?: () => void }) => (
    React.createElement('button', { type: 'button', 'data-testid': 'canvas-modal-close', onClick, ...props }, 'close')
  )

  ModalComponent.Heading = ({ children }: MockComponentProps) => (
    React.createElement('h2', { 'data-testid': 'canvas-modal-heading' }, children)
  )

  ModalComponent.Body = ({ children }: MockComponentProps) => (
    React.createElement('div', { 'data-testid': 'canvas-modal-body' }, children)
  )

  const useModalModel = () => {
    const [visibility, setVisibility] = React.useState<'hidden' | 'visible'>('hidden')

    return {
      state: { visibility },
      events: {
        show: () => setVisibility('visible'),
        hide: () => setVisibility('hidden'),
      },
    }
  }

  return { Modal: ModalComponent, useModalModel }
})

vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children }: MockComponentProps) => React.createElement('h1', { 'data-testid': 'canvas-heading' }, children),
  Text: ({ children }: MockComponentProps) => React.createElement('span', { 'data-testid': 'canvas-text' }, children)
}));

vi.mock('@workday/canvas-kit-react/text-input', () => ({
  TextInput: ({ children, ...props }: MockComponentProps) => React.createElement('input', { 'data-testid': 'canvas-text-input', ...props }, children),
}));

vi.mock('@workday/canvas-kit-react/text-area', () => ({
  TextArea: ({ children, ...props }: MockComponentProps) => React.createElement('textarea', { 'data-testid': 'canvas-text-area', ...props }, children),
}));

vi.mock('@workday/canvas-kit-react/select', () => ({
  Select: ({ children, ...props }: MockComponentProps) => React.createElement('select', { 'data-testid': 'canvas-select', ...props }, children),
}));

vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: Object.assign(
    ({ children, ...props }: MockComponentProps) => {
      const {
        backgroundColor: _backgroundColor,
        borderTop: _borderTop,
        border: _border,
        padding: _padding,
        paddingTop: _paddingTop,
        paddingBottom: _paddingBottom,
        paddingLeft: _paddingLeft,
        paddingRight: _paddingRight,
        paddingX: _paddingX,
        paddingY: _paddingY,
        ...rest
      } = props
      return React.createElement('div', { 'data-testid': 'canvas-card', ...rest }, children)
    },
    {
      Heading: ({ children, ...props }: MockComponentProps) =>
        React.createElement('div', { 'data-testid': 'card-heading', ...props }, children),
      Body: ({ children, ...props }: MockComponentProps) =>
        React.createElement('div', { 'data-testid': 'card-body', ...props }, children)
    }
  )
}))

vi.mock('@workday/canvas-kit-react/table', () => ({
  Table: Object.assign(
    ({ children, ...props }: MockComponentProps) =>
      React.createElement('table', { 'data-testid': 'canvas-table', ...props }, children),
    {
      Head: ({ children, ...props }: MockComponentProps) =>
        React.createElement('thead', { 'data-testid': 'table-head', ...props }, children),
      Body: ({ children, ...props }: MockComponentProps) =>
        React.createElement('tbody', { 'data-testid': 'table-body', ...props }, children),
      Row: ({ children, ...props }: MockComponentProps) =>
        React.createElement('tr', { 'data-testid': 'table-row', ...props }, children),
      Header: ({ children, ...props }: MockComponentProps) =>
        React.createElement('th', { 'data-testid': 'table-header', ...props }, children),
      Cell: ({ children, ...props }: MockComponentProps) =>
        React.createElement('td', { 'data-testid': 'table-cell', ...props }, children)
    }
  )
}))

vi.mock('@workday/canvas-kit-react/side-panel', () => {
  const SidePanelMock = ({ children, ...props }: MockComponentProps) => {
    const {
      open: _open,
      openWidth: _openWidth,
      backgroundColor: _backgroundColor,
      padding: _padding,
      header: _header,
      onToggleClick: _onToggleClick,
      onBreakpointChange: _onBreakpointChange,
      openDirection: _openDirection,
      closeNavigationAriaLabel: _closeNavigationAriaLabel,
      openNavigationAriaLabel: _openNavigationAriaLabel,
      ...rest
    } = props;
    return React.createElement('div', { 'data-testid': 'side-panel', ...rest }, children);
  };

  SidePanelMock.OpenDirection = { Left: 0, Right: 1 };
  SidePanelMock.BackgroundColor = { White: 'white', Transparent: 'transparent', Gray: 'gray' };

  return { SidePanel: SidePanelMock };
});

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
