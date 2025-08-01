import useSWR from 'swr';
import { logger } from '@/lib/logger';

// Position interface
export interface Position {
  id: string;
  title: string;
  department: string;
  jobLevel: string;
  employeeCount: number;
  maxCapacity: number;
  minSalary: number;
  maxSalary: number;
  currency: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  description?: string;
  requirements?: string;
  benefits?: string;
}

// SWR fetcher function with monitoring for positions
const fetcher = async (url: string) => {
  const startTime = Date.now();
  console.log('🚀 Positions SWR Fetcher: 开始获取职位数据', url);
  
  try {
    // For now, return mock data since there's no real API endpoint
    // In a real application, this would be: const response = await fetch(url);
    await new Promise(resolve => setTimeout(resolve, 500)); // Simulate API delay
    
    const mockPositions: Position[] = [
      {
        id: '1',
        title: '高级软件工程师',
        department: '技术部',
        jobLevel: 'P6',
        employeeCount: 3,
        maxCapacity: 5,
        minSalary: 18000,
        maxSalary: 30000,
        currency: 'CNY',
        isActive: true,
        createdAt: '2023-01-15',
        updatedAt: '2024-12-01',
        description: '负责核心业务系统的开发和维护',
        requirements: '3年以上React/Node.js开发经验',
        benefits: '五险一金，年终奖，股权激励'
      },
      {
        id: '2', 
        title: '产品经理',
        department: '产品部',
        jobLevel: 'P5',
        employeeCount: 2,
        maxCapacity: 3,
        minSalary: 15000,
        maxSalary: 25000,
        currency: 'CNY',
        isActive: true,
        createdAt: '2023-03-20',
        updatedAt: '2024-11-15',
        description: '负责产品规划和需求分析',
        requirements: '2年以上产品管理经验，有B端产品经验优先',
        benefits: '弹性工作制，培训机会，健康体检'
      },
      {
        id: '3',
        title: '前端工程师',
        department: '技术部', 
        jobLevel: 'P4',
        employeeCount: 1,
        maxCapacity: 4,
        minSalary: 12000,
        maxSalary: 20000,
        currency: 'CNY',
        isActive: true,
        createdAt: '2022-08-10',
        updatedAt: '2024-10-30',
        description: '负责前端页面开发和用户体验优化',
        requirements: 'Vue/React框架熟练，有移动端开发经验',
        benefits: '技术津贴，学习基金，团建活动'
      },
      {
        id: '4',
        title: 'UI设计师',
        department: '设计部',
        jobLevel: 'P4',
        employeeCount: 0,
        maxCapacity: 2,
        minSalary: 10000,
        maxSalary: 18000,
        currency: 'CNY',
        isActive: false,
        createdAt: '2024-01-08',
        updatedAt: '2024-08-20',
        description: '负责产品界面设计和用户体验设计',
        requirements: 'Figma/Sketch熟练使用，有B端产品设计经验',
        benefits: '创意奖金，设计工具报销，作品展示机会'
      }
    ];
    
    const data = { positions: mockPositions, total_count: mockPositions.length };
    const duration = Date.now() - startTime;
    
    console.log('✅ Positions SWR Fetcher: 成功获取职位数据', data.positions.length, '个职位');
    logger.trackSWRRequest(url, true, duration);
    
    return data;
  } catch (error) {
    const duration = Date.now() - startTime;
    logger.trackSWRRequest(url, false, duration, error as Error);
    throw error;
  }
};

// Positions API interfaces
interface PositionsResponse {
  positions: Position[];
  total_count: number;
}

interface UsePositionsOptions {
  department?: string;
  jobLevel?: string;
  isActive?: boolean;
  search?: string;
}

interface UsePositionsReturn {
  positions: Position[];
  totalCount: number;
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  mutate: () => Promise<any>;
}

// Main hook for positions data using SWR with intelligent caching
export function usePositionsSWR(options: UsePositionsOptions = {}): UsePositionsReturn {
  const { department, jobLevel, isActive, search } = options;
  
  // Build query parameters
  const params = new URLSearchParams();
  if (department) params.append('department', department);
  if (jobLevel) params.append('job_level', jobLevel);
  if (isActive !== undefined) params.append('is_active', isActive.toString());
  if (search) params.append('search', search);
  
  const url = `/api/positions?${params.toString()}`;
  
  // Intelligent caching strategy based on data characteristics
  const getCachingStrategy = () => {
    // Real-time data for searches and filters (shorter cache)
    if (search || department || jobLevel || isActive !== undefined) {
      return {
        dedupingInterval: 3000,     // 3 seconds deduplication
        refreshInterval: 60000,     // Refresh every 1 minute for filtered data
        revalidateOnFocus: true,    // Revalidate on focus for search results
        revalidateOnMount: true,    // Always fresh data for searches
      };
    }
    
    // Static data for main positions list (longer cache)
    return {
      dedupingInterval: 15000,    // 15 seconds deduplication
      refreshInterval: 600000,    // Refresh every 10 minutes for static data
      revalidateOnFocus: false,   // Don't revalidate on focus for static data
      revalidateOnMount: false,   // Use cached data when available
    };
  };
  
  const cachingStrategy = getCachingStrategy();
  
  // Use SWR for data fetching with optimized configuration
  const { data, error, isLoading, mutate } = useSWR<PositionsResponse>(
    url, 
    fetcher,
    {
      // Dynamic caching strategy
      ...cachingStrategy,
      
      // Common configuration
      revalidateOnReconnect: true,  // Refetch when reconnecting
      errorRetryCount: 2,           // Fewer retries for positions
      errorRetryInterval: 1500,     // Wait 1.5s between retries
      
      // Performance optimization: Use background refresh
      revalidateIfStale: true,      // Revalidate stale data
      shouldRetryOnError: true,     // Retry on network errors
      
      // SWR callback hooks with enhanced logging
      onSuccess: (data) => {
        console.log('🎉 Positions SWR Success: 职位数据加载成功', data.positions.length, '个职位');
        console.log('📊 缓存策略:', search || department || jobLevel ? '实时刷新' : '长期缓存');
      },
      onError: (error) => {
        console.error('💥 Positions SWR Error: 职位数据加载失败', error.message);
      },
      onLoadingSlow: (key, config) => {
        console.warn('⏳ Positions SWR Loading Slow: 请求超时', key);
        logger.warn('SWR', key, 'Positions request taking longer than expected');
      },
      
      // Performance optimization: Smart refresh
      refreshWhenHidden: false,     // Don't refresh when tab is hidden
      refreshWhenOffline: false,    // Don't refresh when offline
    }
  );
  
  return {
    positions: data?.positions || [],
    totalCount: data?.total_count || 0,
    isLoading,
    isError: !!error,
    error: error || null,
    mutate,
  };
}

// Hook for single position with SWR and optimized caching
export function usePositionSWR(positionId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    positionId ? `/api/positions/${positionId}` : null,
    async (url: string) => {
      // Mock single position data for now
      const startTime = Date.now();
      await new Promise(resolve => setTimeout(resolve, 300));
      
      const mockPosition: Position = {
        id: positionId,
        title: '高级软件工程师',
        department: '技术部',
        jobLevel: 'P6',
        employeeCount: 3,
        maxCapacity: 5,
        minSalary: 18000,
        maxSalary: 30000,
        currency: 'CNY',
        isActive: true,
        createdAt: '2023-01-15',
        updatedAt: '2024-12-01',
        description: '负责核心业务系统的开发和维护',
        requirements: '3年以上React/Node.js开发经验',
        benefits: '五险一金，年终奖，股权激励'
      };
      
      const duration = Date.now() - startTime;
      logger.trackSWRRequest(url, true, duration);
      
      return mockPosition;
    },
    {
      // Long-term caching for individual position data
      dedupingInterval: 60000,      // 1 minute deduplication
      refreshInterval: 1200000,     // Refresh every 20 minutes
      revalidateOnFocus: false,     // Don't revalidate on focus for individual position
      revalidateOnReconnect: true,  // Revalidate on reconnect
      revalidateOnMount: false,     // Use cache when available
      
      // Error handling
      errorRetryCount: 2,
      errorRetryInterval: 2000,
      
      // Performance optimization
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      revalidateIfStale: true,
      
      // Enhanced logging
      onSuccess: (data) => {
        console.log('🎉 Position SWR Success: 单个职位数据加载成功', data.id);
        console.log('📊 缓存策略: 长期缓存 (20分钟)');
      },
      onError: (error) => {
        console.error('💥 Position SWR Error: 单个职位数据加载失败', error.message);
      },
    }
  );
  
  return {
    position: data,
    isLoading,
    isError: !!error,
    error,
    mutate,
  };
}

// Hook for position statistics using SWR with aggressive caching
export function usePositionStatsSWR() {
  const { positions, isLoading, isError } = usePositionsSWR();
  
  // Use SWR for caching computed statistics
  const statsData = useSWR(
    positions.length > 0 ? 'position-stats' : null,
    () => {
      console.log('📊 计算职位统计数据', positions.length, '个职位');
      
      const stats = {
        total: positions.length,
        active: positions.filter(pos => pos.isActive).length,
        inactive: positions.filter(pos => !pos.isActive).length,
        departments: new Set(positions.map(pos => pos.department)).size,
        totalCapacity: positions.reduce((sum, pos) => sum + pos.maxCapacity, 0),
        currentEmployees: positions.reduce((sum, pos) => sum + pos.employeeCount, 0),
      };
      
      // Department distribution
      const departmentData = Array.from(
        positions.reduce((acc, pos) => {
          acc.set(pos.department, (acc.get(pos.department) || 0) + 1);
          return acc;
        }, new Map())
      ).map(([department, count]) => ({
        label: department,
        value: count,
        color: `hsl(${Math.random() * 360}, 70%, 60%)`
      }));
      
      return { stats, departmentData };
    },
    {
      // Aggressive caching for statistics (computed data)
      dedupingInterval: 120000,    // 2 minute deduplication
      refreshInterval: 1800000,    // Refresh every 30 minutes
      revalidateOnFocus: false,    // Don't revalidate on focus
      revalidateOnReconnect: false, // Don't revalidate on reconnect
      revalidateOnMount: false,    // Use cached computation
      
      // No retries for computed data
      errorRetryCount: 0,
      
      // Performance optimization
      refreshWhenHidden: false,
      refreshWhenOffline: false,
      
      onSuccess: (data) => {
        console.log('📊 职位统计数据计算完成:', data.stats);
      }
    }
  );
  
  return {
    stats: statsData.data?.stats || {
      total: 0,
      active: 0,
      inactive: 0,
      departments: 0,
      totalCapacity: 0,
      currentEmployees: 0,
    },
    departmentData: statsData.data?.departmentData || [],
    isLoading: isLoading || statsData.isLoading,
    isError: isError || statsData.error,
  };
}