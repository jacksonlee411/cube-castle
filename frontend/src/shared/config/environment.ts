import { logger } from '@/shared/utils/logger';

type RawEnv = Record<string, string | boolean | undefined>;

type ImportMetaContainer = { env?: unknown };

const rawImportMeta =
  typeof import.meta !== 'undefined'
    ? ((import.meta as unknown) as ImportMetaContainer)
    : undefined;

let rawEnv: RawEnv =
  rawImportMeta?.env && typeof rawImportMeta.env === 'object'
    ? (rawImportMeta.env as RawEnv)
    : {};

// WORKAROUND: 如果 import.meta.env 为空，从外层 import.meta.env 直接读取
// Vite 的 define 注入的变量在顶层可访问
if (Object.keys(rawEnv).length === 0 && typeof import.meta !== 'undefined') {
  try {
    const metaEnv = (import.meta as {env?: Record<string, unknown>}).env;
    if (metaEnv && typeof metaEnv.VITE_AUTH_MODE === 'string') {
      rawEnv = metaEnv as RawEnv;
    }
  } catch (_error) {
    // 忽略错误
  }
}
const BOOLEAN_TRUE_VALUES = new Set(['true', '1', 'yes', 'on']);

const toOptionalString = (
  value: string | boolean | undefined,
): string | undefined => {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : undefined;
  }
  if (typeof value === 'boolean') {
    return value ? 'true' : 'false';
  }
  return undefined;
};

export const getEnvVar = (key: string, fallback?: string): string => {
  const value = toOptionalString(rawEnv[key]);
  if (value !== undefined) {
    return value;
  }
  if (fallback !== undefined) {
    return fallback;
  }
  return '';
};

export const requireEnvVar = (key: string): string => {
  const value = getEnvVar(key);
  if (!value) {
    throw new Error(`[env] Missing required environment variable "${key}"`);
  }
  return value;
};

export const getBooleanEnvVar = (key: string, fallback?: boolean): boolean => {
  const value = rawEnv[key];
  if (typeof value === 'boolean') {
    return value;
  }
  if (typeof value === 'string' && value.trim()) {
    return BOOLEAN_TRUE_VALUES.has(value.trim().toLowerCase());
  }
  return fallback ?? false;
};

export const getNumberEnvVar = (key: string, fallback?: number): number => {
  const value = rawEnv[key];
  if (typeof value === 'number') {
    return value;
  }
  if (typeof value === 'string' && value.trim()) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  if (fallback !== undefined) {
    return fallback;
  }
  return Number.NaN;
};

const ensureLeadingSlash = (value: string): string => {
  if (!value) {
    return '';
  }
  return value.startsWith('/') ? value : `/${value}`;
};

const stripTrailingSlash = (value: string): string => value.replace(/\/+$/, '');

const resolveEndpointValue = (value: string, fallback: string): string => {
  const candidate = value || fallback;
  if (!candidate) {
    return '';
  }
  if (/^https?:\/\//i.test(candidate)) {
    return stripTrailingSlash(candidate);
  }
  if (candidate === '/') {
    return '';
  }
  return stripTrailingSlash(ensureLeadingSlash(candidate));
};

export interface EnvironmentConfig {
  mode: string;
  isDevelopment: boolean;
  isProduction: boolean;
  isTest: boolean;
  apiBaseUrl: string;
  graphqlEndpoint: string;
  defaultTenantId: string;
  auth: {
    clientId: string;
    clientSecret: string;
    tokenEndpoint: string;
    mode: 'dev' | 'oidc';
  };
  features: {
    queryRefactorEnabled: boolean;
  };
}

const mode = typeof rawEnv.MODE === 'string' ? rawEnv.MODE : 'development';
const devMode = getBooleanEnvVar('DEV', false);
const authModeRaw = getEnvVar(
  'VITE_AUTH_MODE',
  devMode ? 'dev' : 'oidc',
);
const authMode = authModeRaw === 'dev' ? 'dev' : 'oidc';

export const env: EnvironmentConfig = {
  mode,
  isDevelopment: Boolean(rawEnv.DEV),
  isProduction: Boolean(rawEnv.PROD),
  isTest: mode === 'test',
  apiBaseUrl: resolveEndpointValue(getEnvVar('VITE_API_BASE_URL'), '/api/v1'),
  graphqlEndpoint: resolveEndpointValue(
    getEnvVar('VITE_GRAPHQL_ENDPOINT'),
    '/graphql',
  ),
  defaultTenantId: getEnvVar(
    'VITE_DEFAULT_TENANT_ID',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
  ),
  auth: {
    clientId: getEnvVar('VITE_AUTH_CLIENT_ID', 'dev-client'),
    clientSecret: getEnvVar('VITE_AUTH_CLIENT_SECRET', ''),
    tokenEndpoint: resolveEndpointValue(
      getEnvVar('VITE_AUTH_TOKEN_ENDPOINT'),
      '/auth/dev-token',
    ),
    mode: authMode,
  },
  features: {
    queryRefactorEnabled: getBooleanEnvVar(
      'VITE_QUERY_REFACTOR_ENABLED',
      true,
    ),
  },
};

export const validateEnvironmentConfig = (): void => {
  const missing: string[] = [];

  if (!env.defaultTenantId) {
    missing.push('VITE_DEFAULT_TENANT_ID');
  }
  if (!env.auth.clientId) {
    missing.push('VITE_AUTH_CLIENT_ID');
  }

  if (missing.length > 0) {
    throw new Error(
      `[env] Missing required environment variables: ${missing.join(', ')}`,
    );
  }

  if (env.isDevelopment) {
    logger.info('[Environment] 开发环境配置已加载', {
      mode: env.mode,
      apiBaseUrl: env.apiBaseUrl || 'relative:/api/v1',
      graphqlEndpoint: env.graphqlEndpoint || 'relative:/graphql',
      defaultTenantId: env.defaultTenantId.slice(0, 8) + '…',
      authMode: env.auth.mode,
      queryRefactorEnabled: env.features.queryRefactorEnabled,
    });
  }
};

if (env.isDevelopment) {
  validateEnvironmentConfig();
}

export default env;
