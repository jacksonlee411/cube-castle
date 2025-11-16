// E2E endpoints helper to avoid hardcoded ports and align with Plan 254/255 single-base proxy

const RAW_BASE = process.env.PW_BASE_URL || '';
export const BASE_URL = RAW_BASE.replace(/\/+$/, '');

const join = (a: string, b: string) => {
  if (!a) return b;
  if (!b) return a;
  return `${a}${b.startsWith('/') ? b : `/${b}`}`;
};

export const commandApi = (path = ''): string => {
  return join(BASE_URL, join('/api/v1', path));
};

export const graphqlEndpoint = (): string => {
  return join(BASE_URL, '/graphql');
};

export const direct = (path = ''): string => {
  return join(BASE_URL, path);
};

