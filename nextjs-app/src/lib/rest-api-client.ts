// src/lib/rest-api-client.ts
import { notification } from 'antd';

// REST API base configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_ENDPOINT || 'http://localhost:8080/api/v1';

interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

class RestApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;
    
    try {
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      const tenantId = typeof window !== 'undefined' ? localStorage.getItem('tenantId') : null;
      
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
          ...(tenantId && { 'X-Tenant-ID': tenantId }),
          ...options.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return {
        success: true,
        data,
      };
    } catch (error: any) {
      // REST API Error - error handled by caller
      return {
        success: false,
        error: error.message || 'Network error',
      };
    }
  }

  // Employee API methods
  async getEmployees(filters?: any, pagination?: { first?: number; after?: string }) {
    const queryParams = new URLSearchParams();
    
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value) queryParams.append(key, String(value));
      });
    }
    
    if (pagination?.first) {
      queryParams.append('limit', String(pagination.first));
    }
    
    if (pagination?.after) {
      queryParams.append('cursor', pagination.after);
    }

    const endpoint = `/employees${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return this.request(endpoint);
  }

  async getEmployee(id: string) {
    return this.request(`/employees/${id}`);
  }

  async getPositionTimeline(employeeId: string, maxEntries: number = 20) {
    return this.request(`/employees/${employeeId}/position-timeline?limit=${maxEntries}`);
  }

  // Position change API methods
  async createPositionChange(input: any) {
    return this.request('/position-changes', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  }

  async validatePositionChange(employeeId: string, effectiveDate: string) {
    return this.request('/position-changes/validate', {
      method: 'POST',
      body: JSON.stringify({ employeeId, effectiveDate }),
    });
  }

  // Workflow API methods
  async getWorkflowStatus(workflowId: string) {
    return this.request(`/workflows/${workflowId}/status`);
  }

  async approvePositionChange(workflowId: string, comments?: string) {
    return this.request(`/workflows/${workflowId}/approve`, {
      method: 'POST',
      body: JSON.stringify({ comments }),
    });
  }

  async rejectPositionChange(workflowId: string, reason: string) {
    return this.request(`/workflows/${workflowId}/reject`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    });
  }

  // Organization API methods
  async getOrganizationChart(filters?: any) {
    const queryParams = new URLSearchParams();
    
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value) queryParams.append(key, String(value));
      });
    }

    const endpoint = `/organization/chart${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return this.request(endpoint);
  }

  // Meta-Contract Editor API methods
  async getProjects(filters?: { limit?: number; offset?: number }) {
    const queryParams = new URLSearchParams();
    
    if (filters?.limit) {
      queryParams.append('limit', String(filters.limit));
    }
    
    if (filters?.offset) {
      queryParams.append('offset', String(filters.offset));
    }

    const endpoint = `/metacontract/projects${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return this.request(endpoint);
  }

  async createProject(projectData: {
    name: string;
    description?: string;
    content: string;
  }) {
    return this.request('/metacontract/projects', {
      method: 'POST',
      body: JSON.stringify(projectData),
    });
  }

  async getProject(projectId: string) {
    return this.request(`/metacontract/projects/${projectId}`);
  }

  async updateProject(projectId: string, updateData: {
    name?: string;
    description?: string;
    content?: string;
  }) {
    return this.request(`/metacontract/projects/${projectId}`, {
      method: 'PUT',
      body: JSON.stringify(updateData),
    });
  }

  async deleteProject(projectId: string) {
    return this.request(`/metacontract/projects/${projectId}`, {
      method: 'DELETE',
    });
  }

  async compileProject(projectId: string, compileData: {
    content: string;
    preview?: boolean;
  }) {
    return this.request(`/metacontract/projects/${projectId}/compile`, {
      method: 'POST',
      body: JSON.stringify(compileData),
    });
  }

  async getTemplates(category?: string) {
    const queryParams = new URLSearchParams();
    
    if (category) {
      queryParams.append('category', category);
    }

    const endpoint = `/metacontract/templates${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return this.request(endpoint);
  }

  async getUserSettings() {
    return this.request('/metacontract/settings');
  }

  async updateUserSettings(settings: {
    theme?: string;
    fontSize?: number;
    autoSave?: boolean;
    autoCompile?: boolean;
    keyBindings?: string;
    settings?: Record<string, any>;
  }) {
    return this.request('/metacontract/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    });
  }

  // Health check
  async healthCheck() {
    try {
      const response = await fetch(`${this.baseUrl.replace('/api/v1', '')}/health`);
      return {
        success: response.ok,
        data: response.ok ? await response.json() : null,
      };
    } catch (error: any) {
      return {
        success: false,
        error: error.message,
      };
    }
  }
}

// Create and export a singleton instance
export const restApiClient = new RestApiClient();

// Helper function to handle API errors with user-friendly messages
export const handleApiError = (error: any, context: string = 'API调用') => {
  let message = '操作失败';
  let description = '请稍后重试或联系系统管理员';

  if (error?.message) {
    if (error.message.includes('fetch')) {
      message = '网络连接失败';
      description = '请检查网络连接后重试';
    } else if (error.message.includes('401')) {
      message = '认证失败'; 
      description = '请重新登录';
    } else if (error.message.includes('403')) {
      message = '权限不足';
      description = '您没有执行此操作的权限';
    } else if (error.message.includes('404')) {
      message = '资源不存在';
      description = '请求的资源未找到';
    } else if (error.message.includes('500')) {
      message = '服务器错误';
      description = '服务器内部错误，请稍后重试';
    } else {
      description = error.message;
    }
  }

  notification.error({
    message: `${context} - ${message}`,
    description,
    duration: 5,
  });
};

export default RestApiClient;