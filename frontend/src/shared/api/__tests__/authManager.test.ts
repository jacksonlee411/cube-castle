import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../config/environment', () => ({
  env: {
    auth: {
      clientId: 'dev-client',
      clientSecret: 'dev-secret',
      tokenEndpoint: '/auth/dev-token',
      mode: 'dev' as const,
    },
    defaultTenantId: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    isDevelopment: true,
    isProduction: false,
    isTest: false,
  },
}));

vi.mock('../unified-client', () => ({
  unauthenticatedRESTClient: {
    request: vi.fn(),
  },
}));

// Import after mocks
import { AuthManager, type OAuthConfig } from '../auth';

const HS256_TOKEN = 'eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.signature';
const RS256_HEADER = 'eyJhbGciOiJSUzI1NiJ9';
const RS256_TOKEN = `${RS256_HEADER}.eyJzdWIiOiJ1c2VyIn0.signature`;

describe('AuthManager storage migration', () => {
  const config: OAuthConfig = {
    clientId: 'dev-client',
    clientSecret: 'dev-secret',
    tokenEndpoint: '/auth/dev-token',
    grantType: 'client_credentials',
  };

  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('clears legacy HS256 tokens from storage on init', () => {
    const stored = {
      accessToken: HS256_TOKEN,
      tokenType: 'Bearer',
      expiresIn: 3600,
      issuedAt: Date.now(),
    };
    localStorage.setItem('cube_castle_oauth_token', JSON.stringify(stored));

    const manager = new AuthManager(config);

    expect(localStorage.getItem('cube_castle_oauth_token')).toBeNull();
    expect(manager.isAuthenticated()).toBe(false);
  });

  it('retains RS256 tokens that remain valid', () => {
    const stored = {
      accessToken: RS256_TOKEN,
      tokenType: 'Bearer',
      expiresIn: 3600,
      issuedAt: Date.now(),
    };
    localStorage.setItem('cube_castle_oauth_token', JSON.stringify(stored));

    const manager = new AuthManager(config);

    expect(localStorage.getItem('cube_castle_oauth_token')).not.toBeNull();
    expect(manager.isAuthenticated()).toBe(true);
  });

  it('removes raw JWT strings stored under oauth token key', () => {
    localStorage.setItem('cube_castle_oauth_token', HS256_TOKEN);

    const manager = new AuthManager(config);

    expect(localStorage.getItem('cube_castle_oauth_token')).toBeNull();
    expect(manager.isAuthenticated()).toBe(false);
  });
});
