// OAuth 2.0客户端认证管理器
// 实现Client Credentials Flow和JWT Token管理
import { logger } from '@/shared/utils/logger';
import { env } from '../config/environment';
import { unauthenticatedRESTClient } from './unified-client';
import type { JsonObject, JsonValue } from '../types/json';
import { isJsonObject } from '../types/json';

const camelToSnakeCase = (value: string): string =>
  value.replace(/([A-Z])/g, letter => `_${letter.toLowerCase()}`);

const mapKeysToSnakeCase = (record: Record<string, JsonValue | undefined>): JsonObject =>
  Object.fromEntries(
    Object.entries(record)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]) => [camelToSnakeCase(key), value as JsonValue])
  ) as JsonObject;

const LEGACY_TOKEN_KEY = ['cube', 'castle', 'token'].join('_');
const LEGACY_OAUTH_TOKEN_KEY = ['cube', 'castle', 'oauth', 'token'].join('_');
const LEGACY_OAUTH_TOKEN_RAW_KEY = ['cube', 'castle', 'oauth', 'token', 'raw'].join('_');
export const TOKEN_STORAGE_KEY = 'cubeCastleOauthToken';
const LEGACY_TOKEN_ALIASES = ['cubeCastleToken', 'cube-castle-token'];
const LEGACY_ALL_KEYS = [
  LEGACY_TOKEN_KEY,
  ...LEGACY_TOKEN_ALIASES,
  LEGACY_OAUTH_TOKEN_KEY,
  LEGACY_OAUTH_TOKEN_RAW_KEY
];

export interface OAuthToken {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  scope?: string;
  issuedAt: number;
}

export interface OAuthConfig {
  clientId: string;
  clientSecret: string;
  tokenEndpoint: string;
  grantType: 'client_credentials';
}

type SessionPayload = {
  accessToken: string;
  expiresIn?: number;
  tenantId?: string | null;
  scopes?: string[];
  scope?: string;
};

type OAuthTokenResponse = {
  accessToken?: string | null;
  token?: string | null;
  data?: { token?: string | null } | null;
  tokenType?: string | null;
  expiresIn?: number | string | null;
  scope?: string | null;
};

type SessionApiResponse = {
  data?: JsonValue;
  accessToken?: string | null;
  expiresIn?: number | string | null;
  tenantId?: string | null;
  scopes?: string[];
  scope?: string | null;
};

const parseSessionPayload = (payload: JsonValue): SessionPayload => {
  if (!isJsonObject(payload)) {
    throw new Error('[OAuth] 会话响应格式不正确，缺少主体数据');
  }

  const record = payload;
  const accessToken = record.accessToken;

  if (typeof accessToken !== 'string' || accessToken.length === 0) {
    throw new Error('[OAuth] 会话响应缺少 accessToken 字段');
  }

  const expiresRaw = record.expiresIn;
  let expiresIn: number | undefined;
  if (typeof expiresRaw === 'number' && Number.isFinite(expiresRaw)) {
    expiresIn = expiresRaw;
  } else if (typeof expiresRaw === 'string' && expiresRaw.trim().length > 0) {
    const numericValue = Number(expiresRaw);
    if (Number.isFinite(numericValue)) {
      expiresIn = numericValue;
    }
  }

  const scopes =
    Array.isArray(record.scopes) && record.scopes.every((item) => typeof item === 'string')
      ? (record.scopes as string[])
      : undefined;

  return {
    accessToken,
    expiresIn,
    tenantId:
      typeof record.tenantId === 'string' && record.tenantId.length > 0
        ? record.tenantId
        : null,
    scopes,
    scope: typeof record.scope === 'string' ? record.scope : undefined,
  };
};

export class AuthManager {
  private token: OAuthToken | null = null;
  private config: OAuthConfig;
  private refreshPromise: Promise<OAuthToken> | null = null;
  // 生产态：来自 /auth/session 的会话信息（仅内存保存，不落盘）
  private sessionTenantId: string | null = null;
  private jwksVerified = false;

  constructor(config: OAuthConfig) {
    this.config = config;
    this.loadTokenFromStorage();
  }

  /**
   * 获取有效的访问令牌
   */
  async getAccessToken(): Promise<string> {
    if (!this.token) {
      this.loadTokenFromStorage();
    }
    await this.ensureRS256();

    // 检查缓存token是否有效，且必须为RS256
    if (this.token) {
      if (!this.isTokenValid(this.token)) {
        this.clearAuth();
      } else if (!this.isRS256Token(this.token)) {
        logger.warn('[OAuth] 检测到历史 HS256 令牌，已强制清除，请重新获取');
        this.clearAuth();
      } else {
        return this.token.accessToken;
      }
    }

    // 如果已经有刷新请求在进行中，等待它完成
    if (this.refreshPromise) {
      const newToken = await this.refreshPromise;
      return newToken.accessToken;
    }

    // 获取新的token（按模式）
    this.refreshPromise = env.auth.mode === 'oidc'
      ? this.obtainFromSession()
      : this.obtainNewToken();
    try {
      const newToken = await this.refreshPromise;
      return newToken.accessToken;
    } finally {
      this.refreshPromise = null;
    }
  }

  /**
   * 获取新的OAuth令牌
   */
  private async obtainNewToken(): Promise<OAuthToken> {
    await this.ensureRS256();
    logger.info('[OAuth] 正在获取新的访问令牌...');
    
    // 修复：使用开发令牌端点的JSON格式请求
    const tokenResponse = await unauthenticatedRESTClient.request<OAuthTokenResponse>(this.config.tokenEndpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(
        mapKeysToSnakeCase({
          grantType: this.config.grantType,
          clientId: this.config.clientId,
          clientSecret: this.config.clientSecret,
        })
      ),
    });

    // 兼容多种开发/生产返回格式：
    // - { accessToken, tokenType, expiresIn }
    // - { token, tokenType, expiresIn }
    // - { data: { token } }
    const accessToken = tokenResponse.accessToken ?? tokenResponse.token ?? tokenResponse.data?.token;

    if (!accessToken) {
      throw new Error(
        `OAuth token response missing token field: ${JSON.stringify(tokenResponse).slice(0, 200)}`
      );
    }

    // 计算过期时间：优先后端expiresIn，否则默认1小时
    const expiresInRaw = tokenResponse.expiresIn;
    const expiresIn: number =
      typeof expiresInRaw === 'number'
        ? expiresInRaw
        : typeof expiresInRaw === 'string'
          ? Number(expiresInRaw) || 3600
          : 3600;

    const scopeValue = typeof tokenResponse.scope === 'string' ? tokenResponse.scope : undefined;

    const token: OAuthToken = {
      accessToken,
      tokenType: tokenResponse.tokenType || 'Bearer',
      expiresIn,
      scope: scopeValue,
      issuedAt: Date.now(),
    };

    this.assertRS256(token);

    this.token = token;
    this.saveTokenToStorage();
    
    logger.info('[OAuth] 访问令牌获取成功，有效期:', token.expiresIn, '秒');
    return token;
  }

  /**
   * 从BFF会话获取短期访问令牌（生产态）
   */
  private async obtainFromSession(): Promise<OAuthToken> {
    await this.ensureRS256();
    const body = await unauthenticatedRESTClient.request<SessionApiResponse>('/auth/session', { credentials: 'include' });
    const rawPayload = body.data ?? body;
    const session = parseSessionPayload(rawPayload ?? {});
    const expiresIn = session.expiresIn ?? 600;
    // 记录会话租户，供统一客户端注入 X-Tenant-ID
    this.sessionTenantId = session.tenantId ?? null;
    const token: OAuthToken = {
      accessToken: session.accessToken,
      tokenType: 'Bearer',
      expiresIn,
      scope: session.scopes?.join(' ') ?? session.scope,
      issuedAt: Date.now(),
    };

    this.assertRS256(token);

    this.token = token;
    // 生产态不持久化到localStorage
    return token;
  }

  

  /**
   * 检查token是否有效（考虑5分钟缓冲时间）
   */
  private isTokenValid(token: OAuthToken): boolean {
    const expirationTime = token.issuedAt + (token.expiresIn * 1000);
    const bufferTime = 5 * 60 * 1000; // 5分钟缓冲
    const now = Date.now();
    return now < (expirationTime - bufferTime);
  }

  private async ensureRS256(): Promise<void> {
    if (this.jwksVerified) {
      return;
    }
    try {
      const response = await fetch('/.well-known/jwks.json', { credentials: 'include' });
      if (!response.ok) {
        throw new Error(`JWKS 请求失败，HTTP ${response.status}`);
      }
      const jwks = (await response.json()) as { keys?: JsonValue[] };
      if (!jwks || !Array.isArray(jwks.keys) || jwks.keys.length === 0) {
        throw new Error('JWKS 未返回任何公钥，无法确认 RS256 配置');
      }
      this.jwksVerified = true;
    } catch (error) {
      throw new Error(`[OAuth] 检测JWKS失败：${(error as Error).message}。请确认命令服务已使用 RS256 并暴露 /.well-known/jwks.json。`);
    }
  }

  private assertRS256(token: OAuthToken): void {
    if (!this.isRS256Token(token)) {
      this.clearAuth();
      throw new Error('检测到非 RS256 签名的令牌，已强制清除。请重新生成 RS256 令牌。');
    }
  }

  private isRS256Token(token: OAuthToken): boolean {
    const alg = this.extractTokenAlgorithm(token.accessToken);
    return alg ? alg.toUpperCase() === 'RS256' : false;
  }

  private extractTokenAlgorithm(rawToken: string | undefined): string | undefined {
    if (!rawToken) {
      return undefined;
    }
    const parts = rawToken.split('.');
    if (parts.length < 2) {
      return undefined;
    }
    try {
      const decoded = this.decodeBase64Url(parts[0]);
      const header = JSON.parse(decoded) as { alg?: string };
      return header.alg;
    } catch (error) {
      logger.warn('[OAuth] 无法解析JWT头部:', error);
      return undefined;
    }
  }

  private decodeBase64Url(value: string): string {
    const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
    const padLength = (4 - (normalized.length % 4)) % 4;
    const padded = normalized + '='.repeat(padLength);
    if (typeof atob === 'function') {
      return atob(padded);
    }
    const globalBuffer =
      typeof globalThis !== 'undefined' &&
      typeof (globalThis as { Buffer?: { from: (input: string, encoding: string) => { toString: (encoding: string) => string } } }).Buffer === 'object'
        ? (globalThis as { Buffer?: { from: (input: string, encoding: string) => { toString: (encoding: string) => string } } }).Buffer
        : undefined;
    if (globalBuffer) {
      return globalBuffer.from(padded, 'base64').toString('binary');
    }
    throw new Error('当前运行环境不支持 Base64 解码');
  }

  /**
   * 从localStorage加载token
   */
  private loadTokenFromStorage(): void {
    try {
      // 1. 先尝试从新键读取
      const storedValue = localStorage.getItem(TOKEN_STORAGE_KEY);

      // 2. 如果新键不存在，尝试从旧的 snake_case 键读取（迁移逻辑）
      const legacyStoredValue = storedValue === null ? localStorage.getItem(LEGACY_OAUTH_TOKEN_KEY) : null;

      // 3. 清理其他历史键（但不包括 LEGACY_OAUTH_TOKEN_KEY，它将在迁移后清理）
      const keysToClean = LEGACY_ALL_KEYS.filter(key => key !== LEGACY_OAUTH_TOKEN_KEY);
      keysToClean.forEach((key) => {
        try {
          localStorage.removeItem(key);
        } catch (legacyError) {
          logger.warn('[OAuth] 无法清理历史令牌字段:', key, legacyError);
        }
      });

      const stored = storedValue ?? legacyStoredValue;
      if (!stored) {
        this.token = null;
        return;
      }

      // 4. 如果是从旧键读取的，执行迁移
      if (legacyStoredValue) {
        try {
          localStorage.setItem(TOKEN_STORAGE_KEY, legacyStoredValue);
          localStorage.removeItem(LEGACY_OAUTH_TOKEN_KEY);
        } catch (migrateError) {
          logger.warn('[OAuth] 迁移历史令牌失败，继续使用内存令牌:', migrateError);
        }
      }

      if (stored.trim().startsWith('eyJ')) {
        // 旧版本直接存储原始JWT字符串
        logger.warn('[OAuth] 检测到历史原始JWT存储，已清理');
        localStorage.removeItem(TOKEN_STORAGE_KEY);
        localStorage.removeItem(LEGACY_OAUTH_TOKEN_KEY);
        this.token = null;
        return;
      }

      const parsed = JSON.parse(stored) as OAuthToken;
      this.token = parsed;

      // 迁移：清除 HS256 或已过期令牌，避免后续请求失败
      if (!this.isRS256Token(parsed)) {
        logger.warn('[OAuth] 检测到历史 HS256 令牌，已清理');
        this.clearAuth();
        return;
      }
      if (!this.isTokenValid(parsed)) {
        logger.warn('[OAuth] 检测到过期令牌，已清理');
        this.clearAuth();
      }
    } catch (error) {
      logger.warn('[OAuth] 无法从存储中加载token:', error);
      this.token = null;
      try {
        localStorage.removeItem(TOKEN_STORAGE_KEY);
        localStorage.removeItem(LEGACY_OAUTH_TOKEN_KEY);
      } catch (clearError) {
        logger.warn('[OAuth] 清理损坏的token失败:', clearError);
      }
    }
  }

  /**
   * 保存token到localStorage
   */
  private saveTokenToStorage(): void {
    try {
      if (this.token) {
        localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(this.token));
      }
    } catch (error) {
      logger.warn('[OAuth] 无法保存token到存储:', error);
    }
  }

  /**
   * 清除认证状态
   */
  clearAuth(): void {
    this.token = null;
    try { localStorage.removeItem(TOKEN_STORAGE_KEY); } catch (error) {
      logger.warn('[OAuth] Failed to clear localStorage:', error);
    }
    LEGACY_ALL_KEYS.forEach((key) => {
      try { localStorage.removeItem(key); } catch (legacyError) {
        logger.warn('[OAuth] Failed to clear legacy token key:', key, legacyError);
      }
    });
    logger.info('[OAuth] 认证状态已清除');
  }

  /**
   * 获取当前认证状态
   */
  isAuthenticated(): boolean {
    if (!this.token) {
      this.loadTokenFromStorage();
    }
    return this.token !== null && this.isTokenValid(this.token);
  }

  /**
   * 生产态：强制刷新（POST /auth/refresh），开发态：重新获取开发令牌
   */
  async forceRefresh(): Promise<OAuthToken> {
    if (env.auth.mode === 'oidc') {
      await this.ensureRS256();
      const csrf = this.getCookie('csrf');
      const body = await unauthenticatedRESTClient.request<SessionApiResponse>('/auth/refresh', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrf || '' },
        credentials: 'include'
      });
      const rawPayload = body.data ?? body;
      const session = parseSessionPayload(rawPayload ?? {});
      const accessToken = session.accessToken;
      const expiresIn = session.expiresIn ?? 600;
      const token: OAuthToken = {
        accessToken,
        tokenType: 'Bearer',
        expiresIn,
        scope: session.scopes?.join(' ') ?? session.scope,
        issuedAt: Date.now(),
      };
      this.sessionTenantId = session.tenantId ?? this.sessionTenantId;
      this.assertRS256(token);
      this.token = token;
      return token;
    }
    // 开发态：走原始刷新逻辑
    this.token = null;
    return this.obtainNewToken();
  }

  private getCookie(name: string): string | null {
    if (typeof document === 'undefined') return null;
    const match = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()[\]/+^])/g, '\\$1') + '=([^;]*)'));
    return match ? decodeURIComponent(match[1]) : null;
  }

  /**
   * 获取当前会话的租户ID（生产态从 /auth/session 获得）。
   * 若不可用，返回 null，由调用方决定回退策略。
   */
  getTenantId(): string | null {
    return this.sessionTenantId;
  }
}

// 默认OAuth配置 - 使用环境配置避免硬编码
export const defaultOAuthConfig: OAuthConfig = {
  clientId: env.auth.clientId,
  clientSecret: env.auth.clientSecret, 
  tokenEndpoint: env.auth.tokenEndpoint,  // 使用环境配置的端点
  grantType: 'client_credentials',
};

// 全局认证管理器实例
export const authManager = new AuthManager(defaultOAuthConfig);
