import useSWR from 'swr';
import { logger } from '@/lib/logger';

// Organization interface
export interface Organization {
  id: string;
  name: string;
  type: 'company' | 'department' | 'team' | 'group';
  parentId?: string;
  level: number;
  employeeCount: number;
  managerId?: string;
  managerName?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  description?: string;
  children?: Organization[];
}

// SWR fetcher function with monitoring for organization data
const fetcher = async (url: string) => {
  const startTime = Date.now();
  console.log('🚀 Organization SWR Fetcher: 开始获取组织架构数据', url);
  
  try {
    // For now, return mock data since there's no real API endpoint
    await new Promise(resolve => setTimeout(resolve, 600)); // Simulate API delay
    
    const mockOrganizations: Organization[] = [
      {
        id: '1',
        name: 'Cube Castle',
        type: 'company',
        level: 0,
        employeeCount: 50,
        managerId: 'ceo-001',
        managerName: '张总',
        isActive: true,
        createdAt: '2022-01-01',
        updatedAt: '2024-12-01',
        description: '全栈企业管理解决方案提供商',
        children: [
          {
            id: '2',
            name: '技术部',
            type: 'department',
            parentId: '1',
            level: 1,
            employeeCount: 18,
            managerId: 'tech-manager-001',
            managerName: '李经理',
            isActive: true,
            createdAt: '2022-01-15',
            updatedAt: '2024-11-30',
            description: '负责产品研发和技术架构',
            children: [
              {
                id: '21',
                name: '前端团队',
                type: 'team',
                parentId: '2',
                level: 2,
                employeeCount: 6,
                managerId: 'frontend-lead-001',
                managerName: '王组长',
                isActive: true,
                createdAt: '2022-02-01',
                updatedAt: '2024-11-25',
                description: '负责Web和移动端开发'
              },
              {
                id: '22',
                name: '后端团队',
                type: 'team',
                parentId: '2',
                level: 2,
                employeeCount: 8,
                managerId: 'backend-lead-001',
                managerName: '刘组长',
                isActive: true,
                createdAt: '2022-02-01',
                updatedAt: '2024-11-28',
                description: '负责服务端和数据库开发'
              },
              {
                id: '23',
                name: 'DevOps团队',
                type: 'team',
                parentId: '2',
                level: 2,
                employeeCount: 4,
                managerId: 'devops-lead-001',
                managerName: '陈组长',
                isActive: true,
                createdAt: '2022-03-01',
                updatedAt: '2024-11-20',
                description: '负责基础设施和运维自动化'
              }
            ]
          },
          {
            id: '3',
            name: '产品部',
            type: 'department',
            parentId: '1',
            level: 1,
            employeeCount: 12,
            managerId: 'product-manager-001',
            managerName: '赵经理',
            isActive: true,
            createdAt: '2022-01-20',
            updatedAt: '2024-11-29',
            description: '负责产品规划和用户体验',
            children: [
              {
                id: '31',
                name: '产品策划组',
                type: 'group',
                parentId: '3',
                level: 2,
                employeeCount: 5,
                managerId: 'pm-lead-001',
                managerName: '孙组长',
                isActive: true,
                createdAt: '2022-02-15',
                updatedAt: '2024-11-25',
                description: '负责产品需求分析和规划'
              },
              {
                id: '32',
                name: '用户体验组',
                type: 'group',
                parentId: '3',
                level: 2,
                employeeCount: 4,
                managerId: 'ux-lead-001',
                managerName: '周组长',
                isActive: true,
                createdAt: '2022-02-20',
                updatedAt: '2024-11-22',
                description: '负责用户研究和交互设计'
              },
              {
                id: '33',
                name: 'UI设计组',
                type: 'group',
                parentId: '3',
                level: 2,
                employeeCount: 3,
                managerId: 'ui-lead-001',
                managerName: '吴组长',
                isActive: true,
                createdAt: '2022-03-10',
                updatedAt: '2024-11-18',
                description: '负责视觉设计和品牌形象'
              }
            ]
          },
          {
            id: '4',
            name: '销售部',
            type: 'department',
            parentId: '1',
            level: 1,
            employeeCount: 15,
            managerId: 'sales-manager-001',
            managerName: '郑经理',
            isActive: true,
            createdAt: '2022-01-25',
            updatedAt: '2024-12-01',
            description: '负责市场拓展和客户服务',
            children: [
              {
                id: '41',
                name: '企业销售组',
                type: 'group',
                parentId: '4',
                level: 2,
                employeeCount: 8,
                managerId: 'enterprise-sales-lead-001',
                managerName: '何组长',
                isActive: true,
                createdAt: '2022-02-10',
                updatedAt: '2024-11-30',
                description: '负责企业客户开发和维护'
              },
              {
                id: '42',
                name: '客户成功组',
                type: 'group',
                parentId: '4',
                level: 2,
                employeeCount: 7,
                managerId: 'cs-lead-001',
                managerName: '徐组长',
                isActive: true,
                createdAt: '2022-02-25',
                updatedAt: '2024-11-28',
                description: '负责客户满意度和续费管理'
              }
            ]
          },
          {
            id: '5',
            name: '人事行政部',
            type: 'department',
            parentId: '1',
            level: 1,
            employeeCount: 5,
            managerId: 'hr-manager-001',
            managerName: '冯经理',
            isActive: true,
            createdAt: '2022-01-10',
            updatedAt: '2024-11-27',
            description: '负责人力资源和行政管理'
          }
        ]
      }
    ];
    
    const data = { organizations: mockOrganizations, total_count: mockOrganizations.length };
    const duration = Date.now() - startTime;
    
    console.log('✅ Organization SWR Fetcher: 成功获取组织架构数据', data.organizations.length, '个组织');
    logger.trackSWRRequest(url, true, duration);
    
    return data;
  } catch (error) {
    const duration = Date.now() - startTime;
    logger.trackSWRRequest(url, false, duration, error as Error);
    throw error;
  }
};

// Organization API interfaces
interface OrganizationsResponse {
  organizations: Organization[];
  total_count: number;
}

interface UseOrganizationsOptions {
  type?: 'company' | 'department' | 'team' | 'group';
  parentId?: string;
  isActive?: boolean;
  level?: number;
}

interface UseOrganizationsReturn {
  organizations: Organization[];
  totalCount: number;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  mutate: () => Promise<any>;
}

// Main hook for organizations data using SWR with intelligent caching
export function useOrganizationsSWR(options: UseOrganizationsOptions = {}): UseOrganizationsReturn {
  const { type, parentId, isActive, level } = options;
  
  // Build query parameters
  const params = new URLSearchParams();
  if (type) params.append('type', type);
  if (parentId) params.append('parent_id', parentId);
  if (isActive !== undefined) params.append('is_active', isActive.toString());
  if (level !== undefined) params.append('level', level.toString());
  
  const url = `/api/organizations?${params.toString()}`;
  
  // Organization data changes less frequently, use longer cache times
  const { data, error, isLoading, mutate } = useSWR<OrganizationsResponse>(
    url, 
    fetcher,
    {
      // Long-term caching for organizational data (changes infrequently)
      dedupingInterval: 30000,     // 30 seconds deduplication
      refreshInterval: 1800000,    // Refresh every 30 minutes
      revalidateOnFocus: false,    // Don't revalidate on focus
      revalidateOnReconnect: true, // Revalidate on reconnect
      revalidateOnMount: false,    // Use cached data when available
      
      // Error handling
      errorRetryCount: 2,
      errorRetryInterval: 2000,
      
      // Performance optimization
      revalidateIfStale: true,
      shouldRetryOnError: true,
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      // SWR callback hooks with enhanced logging
      onSuccess: (data) => {
        console.log('🎉 Organization SWR Success: 组织架构数据加载成功', data.organizations.length, '个组织');
        console.log('📊 缓存策略: 长期缓存 (30分钟)');
      },
      onError: (error) => {
        console.error('💥 Organization SWR Error: 组织架构数据加载失败', error.message);
      },
      onLoadingSlow: (key, config) => {
        console.warn('⏳ Organization SWR Loading Slow: 请求超时', key);
        logger.warn('SWR', key, 'Organization request taking longer than expected');
      },
    }
  );
  
  return {
    organizations: data?.organizations || [],
    totalCount: data?.total_count || 0,
    isLoading,
    isError: !!error,
    error: error || null,
    mutate,
  };
}

// Hook for flattened organization chart (useful for dropdown selects)
export function useOrganizationChartSWR() {
  const { organizations, isLoading, isError, error } = useOrganizationsSWR();
  
  // Flatten the hierarchical structure for easier use
  const flattenOrganizations = (orgs: Organization[]): Organization[] => {
    const flattened: Organization[] = [];
    
    const flatten = (orgList: Organization[]) => {
      orgList.forEach(org => {
        flattened.push(org);
        if (org.children && org.children.length > 0) {
          flatten(org.children);
        }
      });
    };
    
    flatten(orgs);
    return flattened;
  };
  
  return {
    chart: organizations,
    flatChart: flattenOrganizations(organizations),
    isLoading,
    isError,
    error,
  };
}

// Hook for organization statistics using SWR with aggressive caching
export function useOrganizationStatsSWR() {
  const { organizations, isLoading, isError } = useOrganizationsSWR();
  
  // Use SWR for caching computed statistics
  const statsData = useSWR(
    organizations.length > 0 ? 'organization-stats' : null,
    () => {
      console.log('📊 计算组织架构统计数据', organizations.length, '个组织');
      
      // Flatten organizations for statistics
      const flatten = (orgs: Organization[]): Organization[] => {
        const flattened: Organization[] = [];
        orgs.forEach(org => {
          flattened.push(org);
          if (org.children) {
            flattened.push(...flatten(org.children));
          }
        });
        return flattened;
      };
      
      const allOrgs = flatten(organizations);
      
      const stats = {
        total: allOrgs.length,
        active: allOrgs.filter(org => org.isActive).length,
        inactive: allOrgs.filter(org => !org.isActive).length,
        companies: allOrgs.filter(org => org.type === 'company').length,
        departments: allOrgs.filter(org => org.type === 'department').length,
        teams: allOrgs.filter(org => org.type === 'team').length,
        groups: allOrgs.filter(org => org.type === 'group').length,
        totalEmployees: allOrgs.reduce((sum, org) => sum + org.employeeCount, 0),
        maxLevel: Math.max(...allOrgs.map(org => org.level), 0),
      };
      
      // Type distribution
      const typeData = [
        { label: '公司', value: stats.companies, color: 'hsl(210, 70%, 60%)' },
        { label: '部门', value: stats.departments, color: 'hsl(120, 70%, 60%)' },
        { label: '团队', value: stats.teams, color: 'hsl(60, 70%, 60%)' },
        { label: '小组', value: stats.groups, color: 'hsl(300, 70%, 60%)' },
      ].filter(item => item.value > 0);
      
      return { stats, typeData };
    },
    {
      // Aggressive caching for statistics (computed data)
      dedupingInterval: 180000,    // 3 minute deduplication
      refreshInterval: 3600000,    // Refresh every 1 hour
      revalidateOnFocus: false,    // Don't revalidate on focus
      revalidateOnReconnect: false, // Don't revalidate on reconnect
      revalidateOnMount: false,    // Use cached computation
      
      // No retries for computed data
      errorRetryCount: 0,
      
      // Performance optimization
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      onSuccess: (data) => {
        console.log('📊 组织架构统计数据计算完成:', data.stats);
      }
    }
  );
  
  return {
    stats: statsData.data?.stats || {
      total: 0,
      active: 0,
      inactive: 0,
      companies: 0,
      departments: 0,
      teams: 0,
      groups: 0,
      totalEmployees: 0,
      maxLevel: 0,
    },
    typeData: statsData.data?.typeData || [],
    isLoading: isLoading || statsData.isLoading,
    isError: isError || statsData.error,
  };
}