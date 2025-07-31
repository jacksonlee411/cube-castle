// tests/setup/env.setup.js
// 环境变量设置和全局配置

// 设置测试环境变量
process.env.NODE_ENV = 'test';
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:3000/api';
process.env.NEXT_PUBLIC_WS_URL = 'ws://localhost:3000/ws';
process.env.NEXT_PUBLIC_GRAPHQL_URL = 'http://localhost:3000/graphql';

// 模拟 localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// 模拟 sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.sessionStorage = sessionStorageMock;

// 抑制特定的警告信息
const originalWarn = console.warn;
console.warn = (...args) => {
  if (
    typeof args[0] === 'string' &&
    (args[0].includes('componentWillReceiveProps') ||
     args[0].includes('componentWillMount') ||
     args[0].includes('Legacy context API'))
  ) {
    return;
  }
  originalWarn.call(console, ...args);
};