import useSWR from 'swr';
import { logger } from '@/lib/logger';
import { 
  Organization, 
  OrganizationsResponse, 
  OrganizationStats, 
  OrganizationTypeData 
} from '@/types';
import { organizationApi } from '@/lib/api-client';

// SWR fetcher function using real PostgreSQL API
const fetcher = async (url: string) => {
  const startTime = Date.now();
  console.log('🚀 Organization SWR Fetcher: 开始获取组织架构数据', url);
  
  try {
    // 解析URL参数
    const urlObj = new URL(url, 'http://localhost');
    const params = Object.fromEntries(urlObj.searchParams.entries());
    
    // 使用真实的organizationApi连接PostgreSQL
    const response = await organizationApi.getOrganizations({
      page: params.page ? parseInt(params.page) : 1,
      pageSize: params.pageSize ? parseInt(params.pageSize) : 100,
      search: params.search,
      parent_unit_id: params.parent_unit_id
    });
    
    // 转换为SWR期望的格式
    const data: OrganizationsResponse = {
      organizations: response.organizations,
      total_count: response.pagination?.total || response.organizations.length
    };
    
    const duration = Date.now() - startTime;
    
    console.log('✅ Organization SWR Fetcher: 成功获取组织架构数据', data.organizations.length, '个组织');
    console.log('📊 数据来源: PostgreSQL数据库 via organizationApi');
    logger.trackSWRRequest(url, true, duration);
    
    return data;
  } catch (error) {
    const duration = Date.now() - startTime;
    console.error('❌ Organization SWR Fetcher: API调用失败', error);
    logger.trackSWRRequest(url, false, duration, error as Error);
    throw error;
  }
};

// Organization API interfaces - hooks specific interfaces

interface UseOrganizationsOptions {
  type?: 'company' | 'department' | 'team' | 'group';
  parent_unit_id?: string;
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
  const { type, parent_unit_id, isActive, level } = options;
  
  // Build query parameters
  const params = new URLSearchParams();
  if (type) params.append('type', type);
  if (parent_unit_id) params.append('parent_unit_id', parent_unit_id);
  if (isActive !== undefined) params.append('is_active', isActive.toString());
  if (level !== undefined) params.append('level', level.toString());
  
  const url = `/api/organizations?${params.toString()}`;
  
  // Modern SWR configuration aligned with employee management standards
  const { data, error, isLoading, mutate } = useSWR<OrganizationsResponse>(
    url, 
    fetcher,
    {
      // Optimized caching strategy for organizational data
      dedupingInterval: 10000,      // 10s deduplication (aligned with employees)
      focusThrottleInterval: 5000,  // 5s focus throttle
      refreshInterval: 300000,      // 5 minutes refresh (organizations change less frequently)
      
      // Enhanced revalidation strategy
      revalidateOnFocus: true,      // Enable focus revalidation for fresh data
      revalidateOnReconnect: true,  // Revalidate on network reconnect
      revalidateOnMount: true,      // Always validate on component mount
      
      // Production-grade error handling
      errorRetryCount: 3,           // 3 retries (aligned with employees)
      errorRetryInterval: 1000,     // 1s retry interval
      
      // Performance optimization
      revalidateIfStale: true,
      shouldRetryOnError: true,
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      // Enhanced SWR callback hooks with performance monitoring
      onSuccess: (data) => {
        console.log('🎉 Organization SWR Success: 组织架构数据加载成功', data.organizations.length, '个组织');
        console.log('📊 现代化缓存策略: 智能缓存 (5分钟刷新)');
        logger.trackSWRRequest(url, true, Date.now() - (Date.now() % 1000));
      },
      onError: (error) => {
        console.error('💥 Organization SWR Error: 组织架构数据加载失败', error.message);
        logger.trackSWRRequest(url, false, Date.now() - (Date.now() % 1000), error);
        // Don't show toast here - let error boundary handle it
      },
      onLoadingSlow: (key, config) => {
        console.warn('⏳ Organization SWR Loading Slow: 请求响应较慢', key);
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
      
      const stats: OrganizationStats = {
        total: allOrgs.length,
        active: allOrgs.filter(org => org.status === 'ACTIVE').length,
        inactive: allOrgs.filter(org => org.status === 'INACTIVE').length,
        companies: allOrgs.filter(org => org.unit_type === 'COMPANY').length,
        departments: allOrgs.filter(org => org.unit_type === 'DEPARTMENT').length,
        projectTeams: allOrgs.filter(org => org.unit_type === 'PROJECT_TEAM').length,
        costCenters: allOrgs.filter(org => org.unit_type === 'COST_CENTER').length,
        totalEmployees: allOrgs.reduce((sum, org) => sum + (org.employee_count || 0), 0),
        maxLevel: Math.max(...allOrgs.map(org => org.level), 0),
      };
      
      // Type distribution
      const typeData: OrganizationTypeData[] = [
        { label: '公司', value: stats.companies, color: 'hsl(210, 70%, 60%)' },
        { label: '部门', value: stats.departments, color: 'hsl(120, 70%, 60%)' },
        { label: '项目团队', value: stats.projectTeams, color: 'hsl(60, 70%, 60%)' },
        { label: '成本中心', value: stats.costCenters, color: 'hsl(300, 70%, 60%)' },
      ].filter(item => item.value > 0);
      
      return { stats, typeData };
    },
    {
      // Aggressive caching for computed statistics (aligned with modern architecture)
      dedupingInterval: 60000,     // 1 minute deduplication for heavy computation
      refreshInterval: 900000,     // 15 minutes refresh for stats
      revalidateOnFocus: false,    // Don't revalidate computed data on focus
      revalidateOnReconnect: false, // Don't revalidate on reconnect for stats
      revalidateOnMount: true,     // Always compute fresh stats on mount
      
      // Optimized for computation-heavy operations
      errorRetryCount: 1,          // Fewer retries for computed data
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      onSuccess: (data) => {
        console.log('📊 组织架构统计数据计算完成:', data.stats);
        console.log('⚡ 性能优化: 计算结果已缓存 (15分钟有效)');
      },
      onError: (error) => {
        console.error('💥 组织统计计算失败:', error);
        logger.error('OrganizationStats', 'Computation failed', error);
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
      projectTeams: 0,
      costCenters: 0,
      totalEmployees: 0,
      maxLevel: 0,
    } as OrganizationStats,
    typeData: statsData.data?.typeData || [] as OrganizationTypeData[],
    isLoading: isLoading || statsData.isLoading,
    isError: isError || statsData.error,
  };
}