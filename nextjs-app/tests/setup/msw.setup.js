// tests/setup/msw.setup.js
// Mock Service Worker 设置，用于模拟API调用

import { setupServer } from 'msw/node';
import { rest } from 'msw';

// 定义模拟的API响应
const handlers = [
  // 健康检查
  rest.get('/api/health', (req, res, ctx) => {
    return res(ctx.json({ status: 'ok', timestamp: new Date().toISOString() }));
  }),

  // 员工相关API
  rest.get('/api/employees', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: {
          employees: {
            edges: [
              {
                node: {
                  id: 'emp-1',
                  employeeId: 'EMP001',
                  legalName: '张三',
                  email: 'zhangsan@example.com',
                  status: 'ACTIVE',
                  hireDate: '2023-01-01',
                  currentPosition: {
                    positionTitle: '软件工程师',
                    department: '技术部',
                    employmentType: 'FULL_TIME'
                  }
                }
              }
            ],
            pageInfo: {
              hasNextPage: false,
              endCursor: null
            },
            totalCount: 1
          }
        }
      })
    );
  }),

  rest.get('/api/employees/:id', (req, res, ctx) => {
    const { id } = req.params;
    return res(
      ctx.json({
        success: true,
        data: {
          id,
          employeeId: 'EMP001',
          legalName: '张三',
          email: 'zhangsan@example.com',
          status: 'ACTIVE',
          hireDate: '2023-01-01',
          currentPosition: {
            positionTitle: '软件工程师',
            department: '技术部',
            employmentType: 'FULL_TIME'
          }
        }
      })
    );
  }),

  // 组织架构API
  rest.get('/api/organizations', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: [
          {
            id: 'org-1',
            name: '技术部',
            type: 'DEPARTMENT',
            parentId: null,
            employees: [],
            children: []
          }
        ]
      })
    );
  }),

  // GraphQL API
  rest.post('/graphql', (req, res, ctx) => {
    return res(
      ctx.json({
        data: {
          employees: {
            edges: [],
            pageInfo: { hasNextPage: false, endCursor: null },
            totalCount: 0
          }
        }
      })
    );
  }),

  // 错误处理 - 500错误
  rest.get('/api/error', (req, res, ctx) => {
    return res(ctx.status(500), ctx.json({ error: 'Internal Server Error' }));
  }),

  // 错误处理 - 404错误
  rest.get('/api/not-found', (req, res, ctx) => {
    return res(ctx.status(404), ctx.json({ error: 'Not Found' }));
  }),
];

// 创建服务器实例
export const server = setupServer(...handlers);