// OAuth 2.0客户端认证管理器
// 实现Client Credentials Flow和JWT Token管理
import { env } from '../config/environment';

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

export class AuthManager {
  private token: OAuthToken | null = null;
  private config: OAuthConfig;
  private refreshPromise: Promise<OAuthToken> | null = null;
  // 生产态：来自 /auth/session 的会话信息（仅内存保存，不落盘）
  private sessionTenantId: string | null = null;

  constructor(config: OAuthConfig) {
    this.config = config;
    this.loadTokenFromStorage();
  }

  /**
   * 获取有效的访问令牌
   */
  async getAccessToken(): Promise<string> {
    // 检查现有token是否有效
    if (this.token && this.isTokenValid(this.token)) {
      return this.token.accessToken;
    }

    // 如果已经有刷新请求在进行中，等待它完成
    if (this.refreshPromise) {
      const newToken = await this.refreshPromise;
      return newToken.accessToken;
    }

    // 获取新的token（按模式）
    this.refreshPromise = env.authConfig.mode === 'oidc'
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
    console.log('[OAuth] 正在获取新的访问令牌...');
    
    // 修复：使用开发令牌端点的JSON格式请求
    const response = await fetch(this.config.tokenEndpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      // 标准 client_credentials 请求体（开发/生产均适用）
      body: JSON.stringify({
        grant_type: this.config.grantType,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`OAuth token request failed: ${response.status} - ${errorText}`);
    }

    const tokenResponse = await response.json();

    // 兼容多种开发/生产返回格式：
    // - { accessToken, tokenType, expiresIn }
    // - { token, tokenType, expiresIn }
    // - { data: { token } }
    const accessToken =
      tokenResponse.accessToken ||
      tokenResponse.token ||
      tokenResponse.data?.token;

    if (!accessToken) {
      throw new Error(
        `OAuth token response missing token field: ${JSON.stringify(tokenResponse).slice(0, 200)}`
      );
    }

    // 计算过期时间：优先后端expiresIn，否则默认1小时
    const expiresIn: number = Number(tokenResponse.expiresIn) || 3600;

    const token: OAuthToken = {
      accessToken,
      tokenType: tokenResponse.tokenType || 'Bearer',
      expiresIn,
      scope: tokenResponse.scope,
      issuedAt: Date.now(),
    };

    this.token = token;
    this.saveTokenToStorage();
    
    console.log('[OAuth] 访问令牌获取成功，有效期:', token.expiresIn, '秒');
    return token;
  }

  /**
   * 从BFF会话获取短期访问令牌（生产态）
   */
  private async obtainFromSession(): Promise<OAuthToken> {
    const resp = await fetch('/auth/session', { credentials: 'include' });
    if (!resp.ok) {
      const text = await resp.text();
      throw new Error(`Session fetch failed: ${resp.status} - ${text}`);
    }
    const body = await resp.json();
    const data = body.data || body; // 兼容直接数据
    const accessToken = data.accessToken;
    const expiresIn = Number(data.expiresIn) || 600;
    // 记录会话租户，供统一客户端注入 X-Tenant-ID
    this.sessionTenantId = typeof data.tenantId === 'string' && data.tenantId ? data.tenantId : null;
    const token: OAuthToken = {
      accessToken,
      tokenType: 'Bearer',
      expiresIn,
      scope: Array.isArray(data.scopes) ? data.scopes.join(' ') : data.scope,
      issuedAt: Date.now(),
    };
    this.token = token;
    // 生产态不持久化到localStorage
    return token;
  }

  

  /**
   * 检查token是否有效（考虑5分钟缓冲时间）
   */
  private isTokenValid(token: OAuthToken): boolean {
    const now = Date.now();
    const expirationTime = token.issuedAt + (token.expiresIn * 1000);
    const bufferTime = 5 * 60 * 1000; // 5分钟缓冲
    
    return now < (expirationTime - bufferTime);
  }

  /**
   * 从localStorage加载token
   */
  private loadTokenFromStorage(): void {
    try {
      const stored = localStorage.getItem('cube_castle_oauth_token');
      if (stored) {
        this.token = JSON.parse(stored);
      }
    } catch (error) {
      console.warn('[OAuth] 无法从存储中加载token:', error);
      this.token = null;
    }
  }

  /**
   * 保存token到localStorage
   */
  private saveTokenToStorage(): void {
    try {
      if (this.token) {
        localStorage.setItem('cube_castle_oauth_token', JSON.stringify(this.token));
      }
    } catch (error) {
      console.warn('[OAuth] 无法保存token到存储:', error);
    }
  }

  /**
   * 清除认证状态
   */
  clearAuth(): void {
    this.token = null;
    try { localStorage.removeItem('cube_castle_oauth_token'); } catch {}
    console.log('[OAuth] 认证状态已清除');
  }

  /**
   * 获取当前认证状态
   */
  isAuthenticated(): boolean {
    return this.token !== null && this.isTokenValid(this.token);
  }

  /**
   * 生产态：强制刷新（POST /auth/refresh），开发态：重新获取开发令牌
   */
  async forceRefresh(): Promise<OAuthToken> {
    if (env.authConfig.mode === 'oidc') {
      const csrf = this.getCookie('csrf');
      const resp = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrf || '' },
        credentials: 'include'
      });
      if (!resp.ok) {
        throw new Error(`Refresh failed: ${resp.status}`);
      }
      const body = await resp.json();
      const data = body.data || body;
      const accessToken = data.accessToken;
      const expiresIn = Number(data.expiresIn) || 600;
      const token: OAuthToken = {
        accessToken,
        tokenType: 'Bearer',
        expiresIn,
        issuedAt: Date.now(),
      };
      this.token = token;
      return token;
    }
    // 开发态：走原始刷新逻辑
    this.token = null;
    return this.obtainNewToken();
  }

  private getCookie(name: string): string | null {
    if (typeof document === 'undefined') return null;
    const match = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()\[\]\\/\+^])/g, '\\$1') + '=([^;]*)'));
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
  clientId: env.authConfig.clientId,
  clientSecret: env.authConfig.clientSecret, 
  tokenEndpoint: env.authConfig.tokenEndpoint,  // 使用环境配置的端点
  grantType: 'client_credentials',
};

// 全局认证管理器实例
export const authManager = new AuthManager(defaultOAuthConfig);
