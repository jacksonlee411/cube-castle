import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { Employee, Organization, User, Tenant, Theme, AppState, Notification } from '@/types'

// 应用全局状态
interface AppStore extends AppState {
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
  
  // 重置状态
  reset: () => void
}

const useAppStore = create<AppStore>()(
  persist(
    (set, get) => ({
      // 初始状态
      user: null,
      tenant: null,
      theme: 'system',
      sidebarOpen: true,
      notifications: [],

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

      // 重置状态
      reset: () => set({
        user: null,
        tenant: null,
        theme: 'system',
        sidebarOpen: true,
        notifications: []
      })
    }),
    {
      name: 'cube-castle-app-store',
      partialize: (state) => ({
        theme: state.theme,
        sidebarOpen: state.sidebarOpen,
        // 不持久化用户和通知信息，保证安全性
      })
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

// 导出所有 store hooks
export {
  useAppStore,
  useEmployeeStore,
  useOrganizationStore,
  useChatStore
}