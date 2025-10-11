import { CQRS_ENDPOINTS } from './ports';

const ensureRelativePath = (path: string): string => {
  if (!path) {
    return '';
  }
  return path.startsWith('/') ? path : `/${path}`;
};

export const commandBaseUrl = CQRS_ENDPOINTS.COMMAND_BASE;
export const commandApiBaseUrl = CQRS_ENDPOINTS.COMMAND_API;
export const commandMetricsUrl = CQRS_ENDPOINTS.METRICS_COMMAND;

export const queryBaseUrl = CQRS_ENDPOINTS.QUERY_BASE;
export const graphqlEndpoint = CQRS_ENDPOINTS.GRAPHQL_ENDPOINT;
export const graphqlMetricsUrl = CQRS_ENDPOINTS.METRICS_QUERY;

export const authEndpoint = CQRS_ENDPOINTS.AUTH_ENDPOINT;

export const resolveCommandUrl = (path: string): string =>
  `${commandApiBaseUrl}${ensureRelativePath(path)}`;

export const resolveQueryUrl = (path: string): string =>
  `${queryBaseUrl}${ensureRelativePath(path)}`;

export const resolveGraphqlUrl = (path = ''): string =>
  `${graphqlEndpoint}${ensureRelativePath(path)}`;

export const resolveAuthUrl = (path = ''): string =>
  `${authEndpoint}${ensureRelativePath(path)}`;
