import { defineConfig } from 'vitest/config';
import { resolve } from 'node:path';

export default defineConfig({
  root: resolve(__dirname, '.'),
  resolve: {
    alias: {
      '@': resolve(__dirname, 'frontend/src'),
      '@shared': resolve(__dirname, 'frontend/src/shared'),
      '@layout': resolve(__dirname, 'frontend/src/layout'),
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: resolve(__dirname, 'frontend/src/setupTests.ts'),
    css: true,
    exclude: [
      'node_modules',
      'dist',
      'frontend/dist',
      'tests/e2e/**',
      'frontend/tests/e2e/**',
    ],
    coverage: {
      reporter: ['text', 'html'],
      include: [
        'frontend/src/shared/api/queryClient.ts',
        'frontend/src/shared/api/type-guards.ts',
        'frontend/src/shared/api/graphql-enterprise-adapter.ts',
        'frontend/src/shared/hooks/useEnterpriseOrganizations.ts',
        'frontend/src/shared/hooks/useMessages.ts',
      ],
      exclude: [
        'frontend/src/shared/api/__tests__/**',
        'frontend/src/shared/hooks/__tests__/**',
      ],
    },
  },
});
