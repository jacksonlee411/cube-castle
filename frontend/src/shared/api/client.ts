const API_BASE_URL = 'http://localhost:8000/api/v1';

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
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': this.tenantID, // 使用统一的默认租户ID
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`API request failed: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  public get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  public post<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  public put<T>(endpoint: string, data: any): Promise<T> {
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