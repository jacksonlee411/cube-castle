import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

type JwtPayload = {
  exp?: number;
  tenant_id?: string;
  tenantId?: string;
};

interface EnsureJwtOptions {
  tenantId?: string;
  roles?: string[];
  duration?: string;
  userId?: string;
  clockSkewSeconds?: number;
}

const DEFAULT_TENANT_ID = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';
const DEFAULT_ROLES = ['ADMIN', 'USER'];
const DEFAULT_DURATION = '8h';
const DEV_COMMAND_SERVICE = process.env.PW_COMMAND_URL?.replace(/\/$/, '') || 'http://localhost:9090';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const DEV_JWT_PATH = path.resolve(__dirname, '../../../..', '.cache', 'dev.jwt');

let cachedJwt: string | null = null;

const sanitizeJwt = (rawToken?: string | null): string | null => {
  if (!rawToken) return null;
  const trimmed = rawToken.trim();
  if (!trimmed) return null;
  const token = trimmed.toLowerCase().startsWith('bearer ')
    ? trimmed.slice(7).trim()
    : trimmed;
  return token.split('.').length === 3 ? token : null;
};

const decodeJwtPayload = (token: string): JwtPayload | null => {
  try {
    const [, payload] = token.split('.');
    if (!payload) return null;
    const json = Buffer.from(payload, 'base64').toString('utf8');
    return JSON.parse(json) as JwtPayload;
  } catch (_error) {
    return null;
  }
};

const isExpired = (token: string, skewSeconds = 300): boolean => {
  const payload = decodeJwtPayload(token);
  if (!payload?.exp) return false;
  const expiresAt = payload.exp * 1000;
  const skewMs = skewSeconds * 1000;
  return Date.now() >= expiresAt - skewMs;
};

const readJwtFromDisk = (): string | null => {
  try {
    if (!fs.existsSync(DEV_JWT_PATH)) {
      return null;
    }
    const fileToken = fs.readFileSync(DEV_JWT_PATH, 'utf8');
    return sanitizeJwt(fileToken);
  } catch (_error) {
    return null;
  }
};

const persistJwtToDisk = (token: string): void => {
  try {
    fs.mkdirSync(path.dirname(DEV_JWT_PATH), { recursive: true });
    fs.writeFileSync(DEV_JWT_PATH, token, 'utf8');
  } catch (error) {
    console.warn('⚠️  无法写入 .cache/dev.jwt:', (error as Error).message);
  }
};

type FetchLike = (input: string, init?: {
  method?: string;
  headers?: Record<string, string>;
  body?: string;
}) => Promise<{
  ok: boolean;
  status: number;
  json(): Promise<any>;
}>;

const requestDevToken = async (options: EnsureJwtOptions): Promise<string | null> => {
  const tenantId = options.tenantId ?? process.env.PW_TENANT_ID ?? DEFAULT_TENANT_ID;
  const roles = options.roles ?? DEFAULT_ROLES;
  const duration = options.duration ?? DEFAULT_DURATION;
  const userId = options.userId ?? 'dev-user';

  try {
    const fetchImpl: FetchLike | undefined = (globalThis as { fetch?: FetchLike }).fetch;
    if (!fetchImpl) {
      console.warn('⚠️  当前运行环境不支持 fetch，无法自动获取开发令牌');
      return null;
    }

    const response = await fetchImpl(`${DEV_COMMAND_SERVICE}/auth/dev-token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ userId, tenantId, roles, duration }),
    });

    if (!response.ok) {
      console.warn(`⚠️  /auth/dev-token 返回 HTTP ${response.status}`);
      return null;
    }

    const body = await response.json() as { data?: { token?: string }; token?: string; accessToken?: string };
    const rawToken = body?.data?.token ?? body?.token ?? body?.accessToken;
    return sanitizeJwt(rawToken ?? null);
  } catch (error) {
    console.warn('⚠️  调用 /auth/dev-token 失败:', (error as Error).message);
    return null;
  }
};

const resolveJwt = (): string | null => {
  if (cachedJwt && !isExpired(cachedJwt)) {
    return cachedJwt;
  }

  const fromEnv = sanitizeJwt(process.env.PW_JWT);
  if (fromEnv && !isExpired(fromEnv)) {
    cachedJwt = fromEnv;
    return cachedJwt;
  }

  const fromDisk = readJwtFromDisk();
  if (fromDisk && !isExpired(fromDisk)) {
    cachedJwt = fromDisk;
    process.env.PW_JWT = cachedJwt;
    return cachedJwt;
  }

  cachedJwt = null;
  return null;
};

export function getPwJwt(): string | null {
  return resolveJwt();
}

export function hasPwJwt(): boolean {
  return getPwJwt() !== null;
}

export function requirePwJwt(): string {
  const token = getPwJwt();
  if (!token) {
    throw new Error('缺少有效的 RS256 JWT，请先运行 make run-dev && make jwt-dev-mint');
  }
  return token;
}

export async function ensurePwJwt(options: EnsureJwtOptions = {}): Promise<string | null> {
  const skewSeconds = options.clockSkewSeconds ?? 300;
  const current = resolveJwt();
  if (current && !isExpired(current, skewSeconds)) {
    return current;
  }

  const fresh = await requestDevToken(options);
  if (fresh) {
    cachedJwt = fresh;
    process.env.PW_JWT = fresh;
    persistJwtToDisk(fresh);
    return fresh;
  }

  return null;
}

export function isJwtNearlyExpired(token: string, skewSeconds = 300): boolean {
  return isExpired(token, skewSeconds);
}

export function updateCachedJwt(token: string): void {
  const sanitized = sanitizeJwt(token);
  if (!sanitized) {
    return;
  }
  cachedJwt = sanitized;
  process.env.PW_JWT = sanitized;
  persistJwtToDisk(sanitized);
}
