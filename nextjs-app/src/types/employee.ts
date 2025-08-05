// Employee interface for SWR-based data fetching
export interface Employee {
  id: string;
  businessId: string;
  employeeNumber: string;
  personName: string; // 统一姓名字段
  email: string;
  personalEmail?: string | null;
  phoneNumber?: string | null;
  status: 'active' | 'inactive' | 'pending';
  hireDate: string;
  department?: string;
  departmentId?: string;
  position?: string;
  positionId?: string;
  managerId?: string;
  managerName?: string | null;
  avatar?: string;
}

// Employee filters for search and filtering
export interface EmployeeFilters {
  search?: string;
  department?: string;
  status?: 'active' | 'inactive' | 'pending';
  position?: string;
}

// Employee statistics
export interface EmployeeStats {
  total: number;
  active: number;
  inactive: number;
  pending: number;
  departments: number;
}

// Update employee request interface
export interface UpdateEmployeeRequest {
  personName?: string;
  email?: string;
  personalEmail?: string;
  phoneNumber?: string;
  departmentId?: string;
  positionId?: string;
  status?: 'active' | 'inactive' | 'pending';
  hireDate?: string;
  managerId?: string;
}