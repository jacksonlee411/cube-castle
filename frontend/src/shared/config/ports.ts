import { env, getEnvVar, getNumberEnvVar } from './environment';

const DEFAULT_SERVICE_HOST = getEnvVar(
  'VITE_SERVICE_HOST',
  'localhost',
);
const inferDefaultProtocol = (): string => {
  if (typeof globalThis !== 'undefined') {
    const locationAccessor = (globalThis as { location?: { protocol?: string } }).location;
    const protocol = locationAccessor?.protocol;
    if (typeof protocol === 'string') {
      return protocol === 'https:' ? 'https' : 'http';
    }
  }
  return 'http';
};

const DEFAULT_PROTOCOL = getEnvVar(
  'VITE_SERVICE_PROTOCOL',
  inferDefaultProtocol(),
);

export const SERVICE_PORTS = {
  FRONTEND_DEV: getNumberEnvVar('VITE_PORT_FRONTEND_DEV', 3000),
  FRONTEND_PREVIEW: getNumberEnvVar('VITE_PORT_FRONTEND_PREVIEW', 3001),
  REST_COMMAND_SERVICE: getNumberEnvVar('VITE_PORT_REST_COMMAND', 9090),
  GRAPHQL_QUERY_SERVICE: getNumberEnvVar('VITE_PORT_GRAPHQL_QUERY', 8090),
  POSTGRESQL: getNumberEnvVar('VITE_PORT_POSTGRESQL', 5432),
  REDIS: getNumberEnvVar('VITE_PORT_REDIS', 6379),
} as const;

export type ServicePortKey = keyof typeof SERVICE_PORTS;

const SERVICE_HOST_OVERRIDES: Partial<Record<ServicePortKey, string>> = {
  REST_COMMAND_SERVICE: getEnvVar('VITE_REST_COMMAND_HOST', ''),
  GRAPHQL_QUERY_SERVICE: getEnvVar('VITE_GRAPHQL_QUERY_HOST', ''),
};

const SERVICE_PROTOCOL_OVERRIDES: Partial<Record<ServicePortKey, string>> = {
  REST_COMMAND_SERVICE: getEnvVar('VITE_REST_COMMAND_PROTOCOL', ''),
  GRAPHQL_QUERY_SERVICE: getEnvVar('VITE_GRAPHQL_QUERY_PROTOCOL', ''),
};

export const getServicePort = (service: ServicePortKey): number =>
  SERVICE_PORTS[service];

const getHostForService = (service: ServicePortKey): string => {
  const override = SERVICE_HOST_OVERRIDES[service];
  return override ? override : DEFAULT_SERVICE_HOST;
};

const getProtocolForService = (service: ServicePortKey): string => {
  const override = SERVICE_PROTOCOL_OVERRIDES[service];
  return override ? override : DEFAULT_PROTOCOL;
};

const buildServiceOrigin = (
  service: ServicePortKey,
  overrides?: { protocol?: string; host?: string; port?: number },
): string => {
  const protocol = overrides?.protocol ?? getProtocolForService(service);
  const host = overrides?.host ?? getHostForService(service);
  const port = overrides?.port ?? getServicePort(service);
  return `${protocol}://${host}:${port}`;
};

const ensurePath = (path: string): string => {
  if (!path) {
    return '';
  }
  return path.startsWith('/') ? path : `/${path}`;
};

export const buildServiceURL = (
  service: ServicePortKey,
  path = '',
  overrides?: { protocol?: string; host?: string; port?: number },
): string => {
  const origin = buildServiceOrigin(service, overrides);
  const resolvedPath = ensurePath(path);
  return `${origin}${resolvedPath}`;
};

const resolveConfiguredEndpoint = (
  value: string,
  service: ServicePortKey,
  fallbackPath = '',
): string => {
  const candidate = value || fallbackPath;
  if (!candidate) {
    return buildServiceOrigin(service);
  }
  if (/^https?:\/\//i.test(candidate)) {
    return candidate.replace(/\/+$/, '');
  }
  const normalized = candidate === '/' ? '' : ensurePath(candidate);
  return `${buildServiceOrigin(service)}${normalized}`;
};

export const CQRS_ENDPOINTS = {
  COMMAND_BASE: buildServiceOrigin('REST_COMMAND_SERVICE'),
  COMMAND_API: resolveConfiguredEndpoint(
    env.apiBaseUrl,
    'REST_COMMAND_SERVICE',
    '/api/v1',
  ),
  AUTH_ENDPOINT: resolveConfiguredEndpoint(
    env.auth.tokenEndpoint,
    'REST_COMMAND_SERVICE',
    '/auth/dev-token',
  ),
  METRICS_COMMAND: buildServiceURL('REST_COMMAND_SERVICE', '/metrics'),
  QUERY_BASE: buildServiceOrigin('GRAPHQL_QUERY_SERVICE'),
  GRAPHQL_ENDPOINT: resolveConfiguredEndpoint(
    env.graphqlEndpoint,
    'GRAPHQL_QUERY_SERVICE',
    '/graphql',
  ),
  GRAPHQL_PLAYGROUND: buildServiceURL(
    'GRAPHQL_QUERY_SERVICE',
    '/graphiql',
  ),
  METRICS_QUERY: buildServiceURL('GRAPHQL_QUERY_SERVICE', '/metrics'),
} as const;

export const FRONTEND_ENDPOINTS = {
  DEV_SERVER: buildServiceURL('FRONTEND_DEV'),
  PREVIEW_SERVER: buildServiceURL('FRONTEND_PREVIEW'),
} as const;

export const INFRASTRUCTURE_ENDPOINTS = {
  DATABASE: buildServiceURL('POSTGRESQL'),
  CACHE: buildServiceURL('REDIS'),
} as const;

export type CQRSEndpointKey = keyof typeof CQRS_ENDPOINTS;

export const validatePortConfiguration = (): {
  isValid: boolean;
  errors: string[];
} => {
  const errors: string[] = [];
  const values = Object.values(SERVICE_PORTS).filter((port) =>
    Number.isFinite(port),
  );
  const duplicates = values.filter(
    (port, index) => values.indexOf(port) !== index,
  );

  if (duplicates.length > 0) {
    errors.push(`端口冲突检测到: ${Array.from(new Set(duplicates)).join(', ')}`);
  }

  const invalidPorts = values.filter(
    (port) => port < 1024 || port > 65535,
  );
  if (invalidPorts.length > 0) {
    errors.push(
      `无效端口范围: ${invalidPorts.join(', ')} (有效范围: 1024-65535)`,
    );
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const generatePortConfigReport = (): string => {
  const validation = validatePortConfiguration();
  const lines: string[] = [
    '🎯 端口配置报告',
    '====================',
    '',
    '🏗️ 核心服务:',
    `  前端开发服务器: ${SERVICE_PORTS.FRONTEND_DEV}`,
    `  REST命令服务: ${SERVICE_PORTS.REST_COMMAND_SERVICE}`,
    `  GraphQL查询服务: ${SERVICE_PORTS.GRAPHQL_QUERY_SERVICE}`,
    '',
    '🛠️ 基础设施:',
    `  PostgreSQL: ${SERVICE_PORTS.POSTGRESQL}`,
    `  Redis: ${SERVICE_PORTS.REDIS}`,
    '',
    '🔍 配置验证:',
    `  状态: ${validation.isValid ? '✅ 通过' : '❌ 失败'}`,
  ];

  validation.errors.forEach((error) => lines.push(`  错误: ${error}`));

  lines.push(
    '',
    '🎯 CQRS端点:',
    `  命令API: ${CQRS_ENDPOINTS.COMMAND_API}`,
    `  GraphQL查询: ${CQRS_ENDPOINTS.GRAPHQL_ENDPOINT}`,
  );

  return `${lines.join('\n')}\n`;
};
