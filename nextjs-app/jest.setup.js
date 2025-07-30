// jest.setup.js
import '@testing-library/jest-dom';
import React from 'react';

// Import MSW setup
import { server } from './tests/setup/msw.setup';

// Setup MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Mock Next.js router
jest.mock('next/router', () => ({
  useRouter() {
    return {
      route: '/',
      pathname: '/',
      query: {},
      asPath: '/',
      push: jest.fn(),
      replace: jest.fn(),
      reload: jest.fn(),
      back: jest.fn(),
      prefetch: jest.fn(),
      beforePopState: jest.fn(),
      events: {
        on: jest.fn(),
        off: jest.fn(),
        emit: jest.fn(),
      },
    };
  },
}));

// Mock Next.js Image component
jest.mock('next/image', () => ({
  __esModule: true,
  default: (props) => {
    return React.createElement('img', props);
  },
}));

// Mock Next.js Link component
jest.mock('next/link', () => ({
  __esModule: true,
  default: ({ children, href, ...props }) => {
    return React.createElement('a', { href, ...props }, children);
  },
}));

// Mock Apollo Client
jest.mock('@apollo/client', () => ({
  ...jest.requireActual('@apollo/client'),
  useQuery: jest.fn(() => ({
    data: null,
    loading: false,
    error: null,
    refetch: jest.fn(),
  })),
  useMutation: jest.fn(() => [
    jest.fn(),
    {
      data: null,
      loading: false,
      error: null,
    },
  ]),
  gql: jest.requireActual('@apollo/client').gql,
}));

// Mock dayjs
jest.mock('dayjs', () => {
  const mockDayjs = jest.fn(() => ({
    format: jest.fn(() => '2025-01-27 15:30:00'),
    toISOString: jest.fn(() => '2025-01-27T15:30:00.000Z'),
    valueOf: jest.fn(() => 1706371800000),
  }));
  mockDayjs.extend = jest.fn();
  mockDayjs.utc = jest.fn(() => mockDayjs());
  mockDayjs.tz = jest.fn(() => mockDayjs());
  return mockDayjs;
});

// Mock Ant Design components
jest.mock('antd', () => {
  const React = require('react');
  
  // Mock Menu component
  const Menu = ({ children }) => React.createElement('ul', { 'data-testid': 'menu' }, children);
  Menu.Item = ({ children, onClick, icon, disabled }) => 
    React.createElement('li', { 
      'data-testid': 'menu-item', 
      onClick: disabled ? undefined : onClick,
      'data-disabled': disabled
    }, icon, children);
  Menu.Divider = () => React.createElement('hr', { 'data-testid': 'menu-divider' });

  // Mock Table component
  const Table = ({ columns, dataSource, loading, pagination, onChange }) => {
    if (loading) {
      return React.createElement('div', { 'data-testid': 'table-loading' }, 'Loading...');
    }
    
    if (!dataSource || dataSource.length === 0) {
      return React.createElement('div', { 'data-testid': 'table-empty' }, 'No data');
    }
    
    return React.createElement('table', { 'data-testid': 'table' },
      React.createElement('thead', {},
        React.createElement('tr', {},
          columns?.map((col, idx) => 
            React.createElement('th', { key: idx }, col.title)
          )
        )
      ),
      React.createElement('tbody', {},
        dataSource?.map((item, idx) => 
          React.createElement('tr', { key: idx },
            columns?.map((col, colIdx) => 
              React.createElement('td', { key: `${idx}-${colIdx}` }, 
                // 简化渲染：只返回基本文本内容
                col.render ? String(item[col.dataIndex] || item.legalName || 'Rendered') : String(item[col.dataIndex] || '')
              )
            )
          )
        )
      )
    );
  };

  // Mock Form components
  const Form = ({ children, onFinish, form, initialValues }) => {
    const handleSubmit = (e) => {
      e.preventDefault();
      if (onFinish) onFinish({});
    };
    return React.createElement('form', { onSubmit: handleSubmit, 'data-testid': 'form' }, children);
  };
  Form.Item = ({ children, label, name, rules }) => 
    React.createElement('div', { 'data-testid': 'form-item' },
      label && React.createElement('label', { htmlFor: name }, label),
      React.cloneElement(children, { id: name, name })
    );
  Form.useForm = () => [{
    getFieldsValue: () => ({}),
    setFieldsValue: () => {},
    resetFields: () => {},
    validateFields: () => Promise.resolve({}),
  }];

  return {
    Card: ({ children, title, extra, ...props }) => 
      React.createElement('div', { 'data-testid': 'card', className: 'ant-card', ...props },
        (title || extra) && React.createElement('div', { 'data-testid': 'card-header', className: 'ant-card-header' }, title, extra),
        React.createElement('div', { 'data-testid': 'card-body', className: 'ant-card-body' }, children)
      ),
    Table,
    Button: ({ children, onClick, type, size, icon, loading, ...props }) => 
      React.createElement('button', { 
        onClick, 
        'data-testid': 'button', 
        type: type || 'button', 
        className: loading ? 'ant-btn-loading' : '', 
        ...props 
      }, icon, children),
    Input: (() => {
      const InputComponent = ({ placeholder, onChange, value, ...props }) => 
        React.createElement('input', { placeholder, onChange, value, 'data-testid': 'input', ...props });
      
      InputComponent.Search = ({ placeholder, onSearch, onChange, value, ...searchProps }) => 
        React.createElement('input', { 
          placeholder, 
          onChange, 
          value, 
          'data-testid': 'input-search',
          onKeyPress: (e) => e.key === 'Enter' && onSearch && onSearch(e.target.value),
          ...searchProps 
        });
      
      InputComponent.TextArea = ({ placeholder, onChange, value, maxLength, ...props }) => 
        React.createElement('textarea', { placeholder, onChange, value, maxLength, 'data-testid': 'textarea', ...props });
      
      return InputComponent;
    })(),
    Select: (() => {
      const SelectComponent = ({ children, onChange, value, placeholder, ...props }) => {
        const handleChange = (event) => {
          if (onChange) onChange(event.target.value);
        };
        
        return React.createElement('select', { 
          onChange: handleChange, 
          value: value || '', 
          'data-testid': 'select',
          ...props 
        }, 
          placeholder && React.createElement('option', { value: '' }, placeholder),
          children
        );
      };
      
      SelectComponent.Option = ({ children, value }) => React.createElement('option', { value }, children);
      
      return SelectComponent;
    })(),
    Space: ({ children, ...props }) => 
      React.createElement('div', { 'data-testid': 'space', style: { display: 'flex', gap: '8px' }, ...props }, children),
    Row: ({ children, gutter, justify, align, ...props }) => 
      React.createElement('div', { 'data-testid': 'row', style: { display: 'flex', flexWrap: 'wrap' }, ...props }, children),
    Col: ({ children, span, xs, sm, md, lg, xl, ...props }) => 
      React.createElement('div', { 'data-testid': 'col', style: { flex: span ? `0 0 ${(span/24)*100}%` : '1' }, ...props }, children),
    Tag: ({ children, color, className, ...props }) => 
      React.createElement('span', { 'data-testid': 'tag', style: { color }, className: `ant-tag ${className || ''}`, ...props }, children),
    Avatar: ({ src, icon, children, ...props }) => 
      React.createElement('div', { 'data-testid': 'avatar', ...props }, 
        src ? React.createElement('img', { src, alt: 'avatar' }) : (icon || children)
      ),
    Modal: ({ children, visible, open, title, onOk, onCancel, ...props }) => {
      const isVisible = visible || open;
      return isVisible ? React.createElement('div', { 'data-testid': 'modal', ...props },
        title && React.createElement('h3', {}, title),
        children,
        React.createElement('div', { 'data-testid': 'modal-footer' },
          React.createElement('button', { onClick: onCancel }, '取消'),
          React.createElement('button', { onClick: onOk }, '确定')
        )
      ) : null;
    },
    Form,
    DatePicker: ({ onChange, value, ...props }) => 
      React.createElement('input', { 
        type: 'date', 
        onChange: (e) => onChange && onChange(e.target.value), 
        value, 
        'data-testid': 'date-picker',
        ...props 
      }),
    Radio: (() => {
      const RadioComponent = ({ children, value, checked, onChange, ...props }) => 
        React.createElement('input', { 
          type: 'radio', 
          value, 
          checked, 
          onChange, 
          'data-testid': 'radio',
          ...props 
        });
      
      RadioComponent.Group = ({ children, value, onChange, ...props }) => 
        React.createElement('div', { 'data-testid': 'radio-group', ...props }, children);
      
      return RadioComponent;
    })(),
    Steps: (() => {
      const StepsComponent = ({ children, current, direction, ...props }) => 
        React.createElement('div', { 'data-testid': 'steps', ...props }, children);
      
      StepsComponent.Step = ({ title, description, status, icon, ...props }) => 
        React.createElement('div', { 'data-testid': 'step', ...props }, icon, title, description);
      
      return StepsComponent;
    })(),
    Timeline: ({ children, ...props }) => 
      React.createElement('ol', { 'data-testid': 'timeline', ...props }, children),
    Descriptions: (() => {
      const DescriptionsComponent = ({ children, title, ...props }) => 
        React.createElement('div', { 'data-testid': 'descriptions', ...props },
          title && React.createElement('div', { 'data-testid': 'descriptions-title' }, title),
          children
        );
      
      DescriptionsComponent.Item = ({ label, children, ...props }) => 
        React.createElement('div', { 'data-testid': 'descriptions-item', ...props },
          React.createElement('span', { 'data-testid': 'descriptions-label' }, label),
          React.createElement('span', { 'data-testid': 'descriptions-content' }, children)
        );
      
      return DescriptionsComponent;
    })(),
    notification: {
      success: jest.fn(),
      error: jest.fn(),
      warning: jest.fn(),
      info: jest.fn(),
    },
    Dropdown: ({ children, overlay, trigger, ...props }) => 
      React.createElement('div', { 'data-testid': 'dropdown', ...props }, children),
    Menu,
    Tooltip: ({ children, title, ...props }) => 
      React.createElement('div', { 'data-testid': 'tooltip', title, ...props }, children),
    Typography: {
      Title: ({ children, level, ...props }) => 
        React.createElement(`h${level || 1}`, { 'data-testid': 'typography-title', ...props }, children),
      Text: ({ children, type, strong, ...props }) => 
        React.createElement(strong ? 'strong' : 'span', { 'data-testid': 'typography-text', ...props }, children),
      Paragraph: ({ children, ...props }) => 
        React.createElement('p', { 'data-testid': 'typography-paragraph', ...props }, children),
    },
    message: {
      success: jest.fn(),
      error: jest.fn(),
      warning: jest.fn(),
      info: jest.fn(),
    },
    Tabs: (() => {
      const TabsComponent = ({ children, defaultActiveKey, onChange, ...props }) => {
        return React.createElement('div', { 'data-testid': 'tabs', ...props }, children);
      };
      
      // 为了兼容旧版本的TabPane，添加TabPane支持
      TabsComponent.TabPane = ({ children, tab, key, ...props }) => 
        React.createElement('div', { 'data-testid': 'tab-pane', 'data-tab': tab, 'data-key': key, ...props }, children);
      
      return TabsComponent;
    })(),
    Progress: ({ percent, status, ...props }) => 
      React.createElement('div', { 'data-testid': 'progress', ...props }, `${percent}%`),
    Alert: ({ message, type, showIcon, ...props }) => 
      React.createElement('div', { 'data-testid': 'alert', className: `ant-alert ant-alert-${type}`, ...props }, message),
    Statistic: ({ title, value, prefix, suffix, ...props }) => 
      React.createElement('div', { 'data-testid': 'statistic', ...props },
        title && React.createElement('div', { className: 'ant-statistic-title' }, title),
        React.createElement('div', { className: 'ant-statistic-content' }, prefix, value, suffix)
      ),
    Timeline: (() => {
      const TimelineComponent = ({ children, ...props }) => 
        React.createElement('ol', { 'data-testid': 'timeline', ...props }, children);
      
      TimelineComponent.Item = ({ children, color, dot, ...props }) => 
        React.createElement('li', { 'data-testid': 'timeline-item', ...props }, dot, children);
      
      return TimelineComponent;
    })(),
    Divider: ({ children, ...props }) => 
      React.createElement('hr', { 'data-testid': 'divider', ...props }, children),
    Spin: ({ children, size, loading, ...props }) => {
      if (loading === false) return children;
      return React.createElement('div', { 'data-testid': 'spin', className: `ant-spin ${size ? `ant-spin-${size}` : ''}`, ...props },
        React.createElement('span', { className: 'ant-spin-dot', role: 'img', 'aria-label': 'loading' }),
        children
      );
    },
  };
});
jest.mock('@ant-design/icons', () => ({
  UserOutlined: () => React.createElement('span', { 'data-testid': 'user-icon' }),
  ApartmentOutlined: () => React.createElement('span', { 'data-testid': 'apartment-icon' }),
  DashboardOutlined: () => React.createElement('span', { 'data-testid': 'dashboard-icon' }),
  HistoryOutlined: () => React.createElement('span', { 'data-testid': 'history-icon' }),
  WorkflowOutlined: () => React.createElement('span', { 'data-testid': 'workflow-icon' }),
  RightOutlined: () => React.createElement('span', { 'data-testid': 'right-icon' }),
  PlusOutlined: () => React.createElement('span', { 'data-testid': 'plus-icon' }),
  EditOutlined: () => React.createElement('span', { 'data-testid': 'edit-icon' }),
  DeleteOutlined: () => React.createElement('span', { 'data-testid': 'delete-icon' }),
  SearchOutlined: () => React.createElement('span', { 'data-testid': 'search-icon' }),
  FilterOutlined: () => React.createElement('span', { 'data-testid': 'filter-icon' }),
  SyncOutlined: () => React.createElement('span', { 'data-testid': 'sync-icon' }),
  LoadingOutlined: () => React.createElement('span', { 'data-testid': 'loading-icon' }),
  CheckCircleOutlined: () => React.createElement('span', { 'data-testid': 'check-circle-icon' }),
  ClockCircleOutlined: () => React.createElement('span', { 'data-testid': 'clock-circle-icon' }),
  CloseCircleOutlined: () => React.createElement('span', { 'data-testid': 'close-circle-icon' }),
  ExclamationCircleOutlined: () => React.createElement('span', { 'data-testid': 'exclamation-icon' }),
  PlayCircleOutlined: () => React.createElement('span', { 'data-testid': 'play-circle-icon' }),
  PauseCircleOutlined: () => React.createElement('span', { 'data-testid': 'pause-circle-icon' }),
  MoreOutlined: () => React.createElement('span', { 'data-testid': 'more-icon' }),
  EyeOutlined: () => React.createElement('span', { 'data-testid': 'eye-icon' }),
  DownloadOutlined: () => React.createElement('span', { 'data-testid': 'download-icon' }),
  ReloadOutlined: () => React.createElement('span', { 'data-testid': 'reload-icon' }),
  InfoCircleOutlined: () => React.createElement('span', { 'data-testid': 'info-icon' }),
  WarningOutlined: () => React.createElement('span', { 'data-testid': 'warning-icon' }),
  TeamOutlined: () => React.createElement('span', { 'data-testid': 'team-icon' }),
  BranchesOutlined: () => React.createElement('span', { 'data-testid': 'branches-icon' }),
}));

// Mock Chart.js
jest.mock('chart.js', () => ({
  Chart: {
    register: jest.fn(),
  },
  CategoryScale: jest.fn(),
  LinearScale: jest.fn(),
  PointElement: jest.fn(),
  LineElement: jest.fn(),
  BarElement: jest.fn(),
  ArcElement: jest.fn(),
  Title: jest.fn(),
  Tooltip: jest.fn(),
  Legend: jest.fn(),
}));

// Mock react-chartjs-2
jest.mock('react-chartjs-2', () => ({
  Line: ({ data, options }) => React.createElement('div', { 'data-testid': 'line-chart' }, data?.datasets?.[0]?.label || 'Line Chart'),
  Bar: ({ data, options }) => React.createElement('div', { 'data-testid': 'bar-chart' }, data?.datasets?.[0]?.label || 'Bar Chart'),
  Doughnut: ({ data, options }) => React.createElement('div', { 'data-testid': 'doughnut-chart' }, data?.labels?.join(', ') || 'Doughnut Chart'),
}));

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // Deprecated
    removeListener: jest.fn(), // Deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  observe() {
    return null;
  }
  disconnect() {
    return null;
  }
  unobserve() {
    return null;
  }
};

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  observe() {
    return null;
  }
  disconnect() {
    return null;
  }
  unobserve() {
    return null;
  }
};

// Mock fetch
global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve({}),
  })
);

// Suppress console warnings for deprecated antd props
const originalError = console.error;
console.error = (...args) => {
  if (
    typeof args[0] === 'string' &&
    (args[0].includes('bodyStyle') || args[0].includes('Warning:'))
  ) {
    return;
  }
  originalError.call(console, ...args);
};

// Mock Monaco Editor
jest.mock('@monaco-editor/react', () => ({
  Editor: jest.fn(({ onChange, value, onMount }) => {
    const mockEditor = {
      getValue: jest.fn(() => value || ''),
      setValue: jest.fn(),
      setPosition: jest.fn(),
      revealLineInCenter: jest.fn(),
      focus: jest.fn(),
      dispose: jest.fn(),
      onDidChangeModelContent: jest.fn((callback) => {
        // Simulate content change
        setTimeout(() => callback({ changes: [] }), 0);
        return { dispose: jest.fn() };
      }),
      getModel: jest.fn(() => ({
        onDidChangeContent: jest.fn(),
        setValue: jest.fn(),
        getValue: jest.fn(() => value || '')
      }))
    };

    // Call onMount if provided
    if (onMount) {
      setTimeout(() => onMount(mockEditor, {}), 0);
    }

    // Simulate content changes
    if (onChange) {
      setTimeout(() => onChange('updated content'), 100);
    }

    return React.createElement('div', {
      'data-testid': 'monaco-editor',
      children: value || 'Monaco Editor Mock'
    });
  }),
  loader: {
    init: jest.fn(() => Promise.resolve())
  }
}));

// Mock React DnD
jest.mock('react-dnd', () => ({
  useDrag: jest.fn(() => [{}, jest.fn(), jest.fn()]),
  useDrop: jest.fn(() => [{}, jest.fn()]),
  DndProvider: ({ children }) => children,
}));

jest.mock('react-dnd-html5-backend', () => ({
  HTML5Backend: {}
}));

// Mock React Beautiful DnD
jest.mock('react-beautiful-dnd', () => ({
  DragDropContext: ({ children }) => children,
  Droppable: ({ children }) => children({
    draggableProps: {},
    dragHandleProps: {},
    innerRef: jest.fn(),
  }),
  Draggable: ({ children }) => children({
    draggableProps: {},
    dragHandleProps: {},
    innerRef: jest.fn(),
  }),
}));

// Mock @dnd-kit
jest.mock('@dnd-kit/core', () => ({
  DndContext: ({ children }) => children,
  useDraggable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: jest.fn(),
    transform: null,
  }),
  useDroppable: () => ({
    setNodeRef: jest.fn(),
    isOver: false,
  }),
  DragOverlay: ({ children }) => children,
}));

jest.mock('@dnd-kit/sortable', () => ({
  SortableContext: ({ children }) => children,
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: jest.fn(),
    transform: null,
    transition: null,
  }),
}));

// Mock Framer Motion
jest.mock('framer-motion', () => ({
  motion: {
    div: React.forwardRef(({ children, ...props }, ref) => 
      React.createElement('div', { ref, ...props }, children)
    ),
    button: React.forwardRef(({ children, ...props }, ref) => 
      React.createElement('button', { ref, ...props }, children)
    ),
    span: React.forwardRef(({ children, ...props }, ref) => 
      React.createElement('span', { ref, ...props }, children)
    ),
  },
  AnimatePresence: ({ children }) => children,
  useAnimation: () => ({}),
}));

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    this.CONNECTING = WebSocket.CONNECTING;
    this.OPEN = WebSocket.OPEN;
    this.CLOSING = WebSocket.CLOSING;
    this.CLOSED = WebSocket.CLOSED;
    
    // Simulate connection
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      if (this.onopen) this.onopen(new Event('open'));
    }, 0);
  }
  
  send = jest.fn();
  close = jest.fn(() => {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) this.onclose(new CloseEvent('close'));
  });
  addEventListener = jest.fn();
  removeEventListener = jest.fn();
  onopen = null;
  onclose = null;
  onmessage = null;
  onerror = null;
}

global.WebSocket = MockWebSocket;

// Mock environment variables
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:3000/api';
process.env.NEXT_PUBLIC_WS_URL = 'ws://localhost:3000/ws';

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: jest.fn(() => Promise.resolve()),
    readText: jest.fn(() => Promise.resolve('')),
  },
});

// Mock File API
global.File = class File extends Blob {
  constructor(fileBits, fileName, options) {
    super(fileBits, options);
    this.name = fileName;
    this.lastModified = Date.now();
  }
};

// Mock URL.createObjectURL
global.URL.createObjectURL = jest.fn(() => 'mocked-object-url');
global.URL.revokeObjectURL = jest.fn();

// Mock getBoundingClientRect
Element.prototype.getBoundingClientRect = jest.fn(() => ({
  width: 100,
  height: 100,
  top: 0,
  left: 0,
  bottom: 100,
  right: 100,
  x: 0,
  y: 0,
  toJSON: jest.fn(),
}));

// Mock scrollTo
global.scrollTo = jest.fn();
Element.prototype.scrollTo = jest.fn();

// Increase test timeout for complex operations
jest.setTimeout(10000);