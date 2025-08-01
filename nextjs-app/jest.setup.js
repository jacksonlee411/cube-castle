// jest.setup.js
import '@testing-library/jest-dom';
import React from 'react';

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


// Modern UI components mocks (shadcn/ui + Radix UI)
jest.mock('@/components/ui/table', () => ({
  Table: ({ children, ...props }) => React.createElement('table', { 'data-testid': 'table', ...props }, children),
  TableBody: ({ children, ...props }) => React.createElement('tbody', { ...props }, children),
  TableCell: ({ children, ...props }) => React.createElement('td', { ...props }, children),
  TableHead: ({ children, ...props }) => React.createElement('thead', { ...props }, children),
  TableHeader: ({ children, ...props }) => React.createElement('tr', { ...props }, children),
  TableRow: ({ children, ...props }) => React.createElement('tr', { ...props }, children),
}));

jest.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, disabled, variant, size, ...props }) => 
    React.createElement('button', { 
      onClick: disabled ? undefined : onClick, 
      disabled, 
      'data-variant': variant,
      'data-size': size,
      'data-testid': 'button',
      ...props 
    }, children),
}));

jest.mock('@/components/ui/card', () => ({
  Card: ({ children, ...props }) => React.createElement('div', { ...props, 'data-testid': 'card' }, children),
  CardContent: ({ children, ...props }) => React.createElement('div', { ...props, 'data-testid': 'card-content' }, children),
  CardHeader: ({ children, ...props }) => React.createElement('div', { ...props, 'data-testid': 'card-header' }, children),
  CardTitle: ({ children, ...props }) => React.createElement('h3', { ...props, 'data-testid': 'card-title' }, children),
}));

jest.mock('@/lib/logger', () => ({
  logger: {
    debug: jest.fn(),
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
  },
  log: {
    debug: jest.fn(),
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
  },
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