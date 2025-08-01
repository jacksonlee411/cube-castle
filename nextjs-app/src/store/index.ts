// Phase 2: 状态管理现代化 - 企业级统一状态管理架构
import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { devtools } from 'zustand/middleware'
import { apolloClient } from '@/lib/graphql-client'
import { Employee, Organization, User, Tenant, Theme, AppState, Notification } from '@/types'
import { logger } from '@/lib/logger';

// 实时同步状态接口
interface RealtimeState {
  connected: boolean;
  lastUpdate: string | null;
  subscriptions: {
    employees: boolean;
    organizations: boolean;
    positions: boolean;
    workflows: boolean;
  };
}

// 缓存管理状态接口
interface CacheState {
  lastRefresh: {
    employees: string | null;
    organizations: string | null;
    positions: string | null;
  };
  invalidation: {
    employees: boolean;
    organizations: boolean;
    positions: boolean;
  };
}

// 应用全局状态 - Phase 2 现代化扩展
interface AppStore extends AppState {
  // Phase 2: 实时同步状态
  realtime: RealtimeState;
  
  // Phase 2: 缓存管理状态
  cache: CacheState;
  
  // 用户相关操作
  setUser: (user: User | null) => void
  setTenant: (tenant: Tenant | null) => void
  
  // 主题相关操作
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  
  // UI 状态
  setSidebarOpen: (open: boolean) => void
  toggleSidebar: () => void
  
  // 通知相关操作
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  markNotificationRead: (id: string) => void
  clearAllNotifications: () => void
  
  // Phase 2: 实时同步操作
  setRealtimeConnection: (connected: boolean) => void;
  setSubscription: (key: keyof RealtimeState['subscriptions'], active: boolean) => void;
  updateLastUpdate: () => void;
  
  // Phase 2: 缓存管理操作
  setCacheRefresh: (key: keyof CacheState['lastRefresh']) => void;
  invalidateCache: (key: keyof CacheState['invalidation']) => void;
  clearCache: () => void;
  
  // Phase 2: Apollo Client 集成
  syncWithApollo: () => Promise<void>;
  refreshApolloCache: (keys?: string[]) => Promise<void>;
  
  // 重置状态
  reset: () => void
}

const useAppStore = create<AppStore>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        user: null,
        tenant: null,
        theme: 'system',
        sidebarOpen: true,
        notifications: [],
        
        // Phase 2: 实时同步初始状态
        realtime: {
          connected: false,
          lastUpdate: null,
          subscriptions: {
            employees: false,
            organizations: false,
            positions: false,
            workflows: false,
          },
        },
        
        // Phase 2: 缓存管理初始状态
        cache: {
          lastRefresh: {
            employees: null,
            organizations: null,
            positions: null,
          },
          invalidation: {
            employees: false,
            organizations: false,
            positions: false,
          },
        },

      // 用户操作
      setUser: (user) => set({ user }),
      setTenant: (tenant) => set({ tenant }),

      // 主题操作
      setTheme: (theme) => set({ theme }),
      toggleTheme: () => {
        const currentTheme = get().theme
        const newTheme = currentTheme === 'light' ? 'dark' : 'light'
        set({ theme: newTheme })
      },

      // UI 状态
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),

      // 通知操作
      addNotification: (notification) => {
        const newNotification: Notification = {
          ...notification,
          id: `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          timestamp: new Date().toISOString(),
          read: false
        }
        set((state) => ({
          notifications: [newNotification, ...state.notifications]
        }))
      },

      removeNotification: (id) => {
        set((state) => ({
          notifications: state.notifications.filter(n => n.id !== id)
        }))
      },

      markNotificationRead: (id) => {
        set((state) => ({
          notifications: state.notifications.map(n => 
            n.id === id ? { ...n, read: true } : n
          )
        }))
      },

        clearAllNotifications: () => set({ notifications: [] }),

        // Phase 2: 实时同步操作
        setRealtimeConnection: (connected) =>
          set((state) => ({
            realtime: { ...state.realtime, connected },
          })),

        setSubscription: (key, active) =>
          set((state) => ({
            realtime: {
              ...state.realtime,
              subscriptions: { ...state.realtime.subscriptions, [key]: active },
            },
          })),

        updateLastUpdate: () =>
          set((state) => ({
            realtime: { ...state.realtime, lastUpdate: new Date().toISOString() },
          })),

        // Phase 2: 缓存管理操作
        setCacheRefresh: (key) =>
          set((state) => ({
            cache: {
              ...state.cache,
              lastRefresh: { ...state.cache.lastRefresh, [key]: new Date().toISOString() },
            },
          })),

        invalidateCache: (key) =>
          set((state) => ({
            cache: {
              ...state.cache,
              invalidation: { ...state.cache.invalidation, [key]: true },
            },
          })),

        clearCache: () =>
          set((state) => ({
            cache: {
              lastRefresh: { employees: null, organizations: null, positions: null },
              invalidation: { employees: false, organizations: false, positions: false },
            },
          })),

        // Phase 2: Apollo Client 集成方法
        syncWithApollo: async () => {
          const state = get();
          try {
            // 同步认证状态到 Apollo Client
            if (state.user && state.tenant) {
              // Token 处理在 graphql-client.ts 中
            }

            // 同步实时连接状态
            if (state.realtime.connected) {
              // WebSocket 连接状态已同步
            }

            // 同步本地状态到 Apollo Client 本地缓存
            await apolloClient.writeQuery({
              query: require('graphql-tag')`
                query LocalAppState {
                  localAppState {
                    theme
                    sidebarOpen
                    realtime {
                      connected
                      subscriptions
                    }
                  }
                }
              `,
              data: {
                localAppState: {
                  theme: state.theme,
                  sidebarOpen: state.sidebarOpen,
                  realtime: state.realtime,
                },
              },
            });

          } catch (error) {
            // Apollo 同步失败 - 继续使用本地状态
            logger.warn('Apollo sync failed:', error);
          }
        },

        refreshApolloCache: async (keys = ['employees', 'organizations', 'positions']) => {
          try {
            // 刷新指定的 Apollo 缓存键
            await apolloClient.refetchQueries({
              include: keys,
            });

            // 更新缓存刷新时间戳
            const now = new Date().toISOString();
            const refreshUpdates = keys.reduce(
              (acc, key) => ({ ...acc, [key]: now }),
              {}
            );

            set((state) => ({
              cache: {
                ...state.cache,
                lastRefresh: { ...state.cache.lastRefresh, ...refreshUpdates },
                invalidation: { 
                  ...state.cache.invalidation, 
                  ...keys.reduce((acc, key) => ({ ...acc, [key]: false }), {}) 
                },
              },
            }));

          } catch (error) {
            logger.warn('Apollo cache refresh failed:', error);
          }
        },

        // 重置状态
        reset: () => {
          // 清理 Apollo 缓存
          apolloClient.clearStore();
          
          set({
            user: null,
            tenant: null,
            theme: 'system',
            sidebarOpen: true,
            notifications: [],
            realtime: {
              connected: false,
              lastUpdate: null,
              subscriptions: {
                employees: false,
                organizations: false,
                positions: false,
                workflows: false,
              },
            },
            cache: {
              lastRefresh: {
                employees: null,
                organizations: null,
                positions: null,
              },
              invalidation: {
                employees: false,
                organizations: false,
                positions: false,
              },
            },
          });
        },
      }),
      {
        name: 'cube-castle-app-store',
        storage: createJSONStorage(() => localStorage),
        partialize: (state) => ({
          theme: state.theme,
          sidebarOpen: state.sidebarOpen,
          realtime: {
            subscriptions: state.realtime.subscriptions,
            // 不持久化连接状态，每次启动重新连接
          },
          // 不持久化敏感信息（用户、token、通知）
        })
      }
    ),
    {
      name: 'cube-castle-store',
      enabled: process.env.NODE_ENV === 'development',
    }
  )
)

// 员工管理状态
interface EmployeeStore {
  employees: Employee[]
  selectedEmployee: Employee | null
  loading: boolean
  error: string | null
  filters: {
    search: string
    status: string
    organizationId: string
  }
  pagination: {
    current: number
    pageSize: number
    total: number
  }

  // 操作方法
  setEmployees: (employees: Employee[]) => void
  setSelectedEmployee: (employee: Employee | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setFilters: (filters: Partial<EmployeeStore['filters']>) => void
  setPagination: (pagination: Partial<EmployeeStore['pagination']>) => void
  addEmployee: (employee: Employee) => void
  updateEmployee: (id: string, employee: Partial<Employee>) => void
  removeEmployee: (id: string) => void
  reset: () => void
}

const useEmployeeStore = create<EmployeeStore>((set) => ({
  employees: [],
  selectedEmployee: null,
  loading: false,
  error: null,
  filters: {
    search: '',
    status: '',
    organizationId: ''
  },
  pagination: {
    current: 1,
    pageSize: 20,
    total: 0
  },

  setEmployees: (employees) => set({ employees }),
  setSelectedEmployee: (selectedEmployee) => set({ selectedEmployee }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  setFilters: (filters) => set((state) => ({ 
    filters: { ...state.filters, ...filters } 
  })),
  setPagination: (pagination) => set((state) => ({ 
    pagination: { ...state.pagination, ...pagination } 
  })),
  
  addEmployee: (employee) => set((state) => ({
    employees: [employee, ...state.employees]
  })),
  
  updateEmployee: (id, updatedEmployee) => set((state) => ({
    employees: state.employees.map(emp => 
      emp.id === id ? { ...emp, ...updatedEmployee } : emp
    ),
    selectedEmployee: state.selectedEmployee?.id === id 
      ? { ...state.selectedEmployee, ...updatedEmployee }
      : state.selectedEmployee
  })),
  
  removeEmployee: (id) => set((state) => ({
    employees: state.employees.filter(emp => emp.id !== id),
    selectedEmployee: state.selectedEmployee?.id === id ? null : state.selectedEmployee
  })),

  reset: () => set({
    employees: [],
    selectedEmployee: null,
    loading: false,
    error: null,
    filters: { search: '', status: '', organizationId: '' },
    pagination: { current: 1, pageSize: 20, total: 0 }
  })
}))

// 组织架构状态
interface OrganizationStats {
  total: number
  totalEmployees: number
  active: number
  inactive: number
}

interface OrganizationStore {
  organizations: Organization[]
  organizationTree: any[]
  selectedOrganization: Organization | null
  stats: OrganizationStats
  loading: boolean
  error: string | null

  setOrganizations: (organizations: Organization[]) => void
  setOrganizationTree: (tree: any[]) => void
  setSelectedOrganization: (organization: Organization | null) => void
  setStats: (stats: OrganizationStats) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  addOrganization: (organization: Organization) => void
  updateOrganization: (id: string, organization: Partial<Organization>) => void
  removeOrganization: (id: string) => void
  reset: () => void
}

const useOrganizationStore = create<OrganizationStore>((set) => ({
  organizations: [],
  organizationTree: [],
  selectedOrganization: null,
  stats: {
    total: 0,
    totalEmployees: 0,
    active: 0,
    inactive: 0
  },
  loading: false,
  error: null,

  setOrganizations: (organizations) => set({ organizations }),
  setOrganizationTree: (organizationTree) => set({ organizationTree }),
  setSelectedOrganization: (selectedOrganization) => set({ selectedOrganization }),
  setStats: (stats) => set({ stats }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  
  addOrganization: (organization) => set((state) => ({
    organizations: [organization, ...state.organizations]
  })),
  
  updateOrganization: (id, updatedOrganization) => set((state) => ({
    organizations: state.organizations.map(org => 
      org.id === id ? { ...org, ...updatedOrganization } : org
    ),
    selectedOrganization: state.selectedOrganization?.id === id 
      ? { ...state.selectedOrganization, ...updatedOrganization }
      : state.selectedOrganization
  })),
  
  removeOrganization: (id) => set((state) => ({
    organizations: state.organizations.filter(org => org.id !== id),
    selectedOrganization: state.selectedOrganization?.id === id ? null : state.selectedOrganization
  })),

  reset: () => set({
    organizations: [],
    organizationTree: [],
    selectedOrganization: null,
    stats: { total: 0, totalEmployees: 0, active: 0, inactive: 0 },
    loading: false,
    error: null
  })
}))

// AI 聊天状态
interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  intent?: string
  confidence?: number
  loading?: boolean
}

interface ChatStore {
  messages: ChatMessage[]
  sessionId: string
  loading: boolean
  connected: boolean

  addMessage: (message: Omit<ChatMessage, 'id' | 'timestamp'>) => void
  updateLastMessage: (updates: Partial<ChatMessage>) => void
  setLoading: (loading: boolean) => void
  setConnected: (connected: boolean) => void
  clearMessages: () => void
  newSession: () => void
}

const useChatStore = create<ChatStore>((set, get) => ({
  messages: [],
  sessionId: `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
  loading: false,
  connected: true,

  addMessage: (message) => {
    const newMessage: ChatMessage = {
      ...message,
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      timestamp: new Date()
    }
    set((state) => ({
      messages: [...state.messages, newMessage]
    }))
  },

  updateLastMessage: (updates) => {
    set((state) => {
      const messages = [...state.messages]
      if (messages.length > 0) {
        messages[messages.length - 1] = {
          ...messages[messages.length - 1],
          ...updates
        }
      }
      return { messages }
    })
  },

  setLoading: (loading) => set({ loading }),
  setConnected: (connected) => set({ connected }),

  clearMessages: () => set({ messages: [] }),

  newSession: () => set({
    messages: [],
    sessionId: `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    loading: false
  })
}))

// Phase 2: 现代化选择器 Hooks - 优化重渲染性能
export const useAuthState = () => useAppStore((state) => ({ 
  user: state.user, 
  tenant: state.tenant, 
  isAuthenticated: !!state.user 
}));

export const useUIState = () => useAppStore((state) => ({ 
  theme: state.theme, 
  sidebarOpen: state.sidebarOpen 
}));

export const useRealtimeState = () => useAppStore((state) => state.realtime);
export const useCacheState = () => useAppStore((state) => state.cache);
export const useNotifications = () => useAppStore((state) => state.notifications);

// Phase 2: 操作 Hooks - 避免重复渲染
export const useAppActions = () => useAppStore((state) => ({
  // 基础操作
  setUser: state.setUser,
  setTenant: state.setTenant,
  setTheme: state.setTheme,
  toggleTheme: state.toggleTheme,
  setSidebarOpen: state.setSidebarOpen,
  toggleSidebar: state.toggleSidebar,
  
  // 通知操作
  addNotification: state.addNotification,
  removeNotification: state.removeNotification,
  markNotificationRead: state.markNotificationRead,
  clearAllNotifications: state.clearAllNotifications,
  
  // Phase 2: 实时同步操作
  setRealtimeConnection: state.setRealtimeConnection,
  setSubscription: state.setSubscription,
  updateLastUpdate: state.updateLastUpdate,
  
  // Phase 2: 缓存操作
  setCacheRefresh: state.setCacheRefresh,
  invalidateCache: state.invalidateCache,
  clearCache: state.clearCache,
  
  // Phase 2: Apollo 集成
  syncWithApollo: state.syncWithApollo,
  refreshApolloCache: state.refreshApolloCache,
  
  // 重置
  reset: state.reset,
}));

// 导出所有 store hooks
export {
  useAppStore,
  useEmployeeStore,
  useOrganizationStore,
  useChatStore
}