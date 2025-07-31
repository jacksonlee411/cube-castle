// tests/integration/api/rest-api-client.test.ts
import { restApiClient } from '@/lib/rest-api-client';

// 模拟fetch
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('REST API Client 集成测试', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // 设置默认的成功响应
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: jest.fn().mockResolvedValue({ message: 'success' }),
    });
  });

  describe('员工API调用', () => {
    it('成功获取员工列表', async () => {
      const mockEmployees = {
        employees: [{ id: 'emp-1', legalName: '张三', email: 'zhangsan@example.com' }],
        totalCount: 1
      };
      
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: jest.fn().mockResolvedValue(mockEmployees),
      });

      const result = await restApiClient.getEmployees();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockEmployees);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/employees'),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });

    it('成功获取单个员工', async () => {
      const mockEmployee = { id: 'emp-1', legalName: '张三', email: 'zhangsan@example.com' };
      
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: jest.fn().mockResolvedValue(mockEmployee),
      });

      const result = await restApiClient.getEmployee('emp-1');

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockEmployee);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/employees/emp-1'),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });

    it('处理HTTP错误响应', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: jest.fn().mockResolvedValue({ error: 'Employee not found' }),
      });

      const result = await restApiClient.getEmployee('non-existent');

      expect(result.success).toBe(false);
      expect(result.error).toContain('404');
    });

    it('处理网络错误', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'));

      const result = await restApiClient.getEmployees();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Network error');
    });
  });

  describe('健康检查', () => {
    it('成功的健康检查', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: jest.fn().mockResolvedValue({
          status: 'ok',
          timestamp: new Date().toISOString()
        }),
      });

      const result = await restApiClient.healthCheck();

      expect(result.success).toBe(true);
      expect(result.data.status).toBe('ok');
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/health')
      );
    });

    it('失败的健康检查', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: jest.fn().mockResolvedValue({}),
      });

      const result = await restApiClient.healthCheck();

      expect(result.success).toBe(false);
    });
  });

  describe('Meta-Contract项目API', () => {
    it('成功获取项目列表', async () => {
      const mockProjects = {
        projects: [{ id: 'proj-1', name: '测试项目', description: '测试描述' }],
        totalCount: 1
      };
      
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: jest.fn().mockResolvedValue(mockProjects),
      });

      const result = await restApiClient.getProjects();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockProjects);
    });

    it('成功创建项目', async () => {
      const newProject = { name: '新项目', description: '新项目描述', content: 'contract TestContract { }' };
      const mockCreatedProject = { id: 'proj-new', ...newProject };
      
      mockFetch.mockResolvedValue({
        ok: true,
        status: 201,
        json: jest.fn().mockResolvedValue(mockCreatedProject),
      });

      const result = await restApiClient.createProject(newProject);

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockCreatedProject);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/metacontract/projects'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(newProject),
        })
      );
    });
  });

  describe('错误处理', () => {
    it('处理JSON解析错误', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: jest.fn().mockRejectedValue(new Error('Invalid JSON')),
      });

      const result = await restApiClient.getEmployees();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Invalid JSON');
    });

    it('处理超时错误', async () => {
      mockFetch.mockRejectedValue(new Error('Timeout'));

      const result = await restApiClient.getProjects();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Timeout');
    });
  });
});