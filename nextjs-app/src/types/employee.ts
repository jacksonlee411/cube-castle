// Employee interface for SWR-based data fetching
export interface Employee {
  id: string;
  employeeId: string;
  legalName: string;
  preferredName?: string | null;
  email: string;
  phone?: string | null;
  status: 'active' | 'inactive' | 'pending';
  hireDate: string;
  department?: string;
  position?: string;
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

// Department data for charts
export interface DepartmentData {
  label: string;
  value: number;
  color: string;
}