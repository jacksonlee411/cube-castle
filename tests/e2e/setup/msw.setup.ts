/**
 * Mock Service Worker (MSW) setup for API mocking in tests
 */
import { setupServer } from 'msw/node';
import { rest } from 'msw';

// Mock API handlers
const handlers = [
  // Employee management API
  rest.get('/api/v1/employees', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        employees: [
          {
            id: '1',
            name: 'John Doe',
            email: 'john@example.com',
            department: 'Engineering'
          }
        ]
      })
    );
  }),

  // Position management API
  rest.get('/api/v1/positions', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        positions: [
          {
            id: '1',
            title: 'Software Engineer',
            department: 'Engineering',
            status: 'active'
          }
        ]
      })
    );
  }),

  // Organization API
  rest.get('/api/v1/organization', (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        organization: {
          id: '1',
          name: 'Test Company',
          structure: []
        }
      })
    );
  })
];

// Setup MSW server
export const server = setupServer(...handlers);

// Setup functions for tests
export const setupMSW = () => {
  beforeAll(() => server.listen());
  afterEach(() => server.resetHandlers());
  afterAll(() => server.close());
};