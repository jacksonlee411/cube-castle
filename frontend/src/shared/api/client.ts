import { APIError } from '../types/api';

const API_BASE_URL = 'http://localhost:9090/api/v1';

// 项目默认租户ID - 高谷集团
const DEFAULT_TENANT_ID = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

export class ApiClient {
  private baseURL: string;
  private tenantID: string;

  constructor(baseURL: string = API_BASE_URL, tenantID: string = DEFAULT_TENANT_ID) {
    this.baseURL = baseURL;
    this.tenantID = tenantID;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    try {
      console.log(`[API] ${options.method || 'GET'} ${url}`, options.body ? JSON.parse(options.body as string) : '');
      
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-ID': this.tenantID,
          ...options.headers,
        },
        ...options,
      });

      console.log(`[API] Response: ${response.status} ${response.statusText}`);

      // 检查响应状态，但对于204 No Content等成功状态码不抛出错误
      if (!response.ok && response.status !== 204) {
        let errorMessage = `API request failed: ${response.status} ${response.statusText}`;
        
        // 尝试获取详细错误信息
        try {
          const errorBody = await response.text();
          if (errorBody) {
            errorMessage += ` - ${errorBody}`;
          }
        } catch {
          // 忽略解析错误的错误
        }
        
        const error = new APIError(
          response.status,
          response.statusText,
          errorMessage
        );
        console.error('[API] Error:', error);
        throw error;
      }

      // 处理成功响应
      if (response.status === 204 || response.headers.get('content-length') === '0') {
        console.log('[API] Success: No content');
        return {} as T;
      }

      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const result = await response.json();
        console.log('[API] Success:', result);
        return result;
      } else {
        const result = await response.text();
        console.log('[API] Success (text):', result);
        return result as unknown as T;
      }
    } catch (error) {
      // 检查是否为APIError（有status属性的错误）
      if (error && typeof error === 'object' && 'status' in error) {
        // 重新抛出API错误
        throw error;
      }
      
      // 网络错误或其他异常的详细处理
      console.error('[API] Network/Parse error:', error);
      
      let errorMessage = 'Network connection failed';
      if (error instanceof TypeError && error.message.includes('fetch')) {
        errorMessage = 'Unable to connect to server. Please check if the service is running.';
      } else if (error instanceof Error) {
        errorMessage = `Request failed: ${error.message}`;
      }
      
      throw new Error(errorMessage);
    }
  }

  public get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  public post<T>(endpoint: string, data: Record<string, unknown>): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  public put<T>(endpoint: string, data: Record<string, unknown>): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  public delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

export const apiClient = new ApiClient();