// OAuth 2.0客户端认证管理器
// 实现Client Credentials Flow和JWT Token管理

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

    // 获取新的token
    this.refreshPromise = this.obtainNewToken();
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
      body: JSON.stringify({
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`OAuth token request failed: ${response.status} - ${errorText}`);
    }

    const tokenResponse = await response.json();
    
    // 修复：适配开发令牌响应格式
    const token: OAuthToken = {
      accessToken: tokenResponse.data?.token || tokenResponse.token,
      tokenType: 'Bearer',
      expiresIn: 3600, // 开发令牌默认1小时
      scope: tokenResponse.scope,
      issuedAt: Date.now(),
    };

    this.token = token;
    this.saveTokenToStorage();
    
    console.log('[OAuth] 访问令牌获取成功，有效期:', token.expiresIn, '秒');
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
    localStorage.removeItem('cube_castle_oauth_token');
    console.log('[OAuth] 认证状态已清除');
  }

  /**
   * 获取当前认证状态
   */
  isAuthenticated(): boolean {
    return this.token !== null && this.isTokenValid(this.token);
  }
}

// 默认OAuth配置 - 使用代理端点避免CORS问题
export const defaultOAuthConfig: OAuthConfig = {
  clientId: 'dev-client',
  clientSecret: 'dev-secret', 
  tokenEndpoint: '/auth/dev-token',  // 使用代理路径，避免CORS
  grantType: 'client_credentials',
};

// 全局认证管理器实例
export const authManager = new AuthManager(defaultOAuthConfig);