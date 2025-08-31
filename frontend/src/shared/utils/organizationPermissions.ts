import type { OrganizationUnit as Organization } from '../types/organization';

export interface OrganizationOperationContext {
  canEdit: boolean;
  canDelete: boolean;
  canActivate: boolean;
  canDeactivate: boolean;
  canViewHistory: boolean;
  canCreateChild: boolean;
  canMove: boolean;
  canViewTimeline: boolean;
  reason?: string;
}

/**
 * 获取操作权限的辅助函数
 */
export const getOperationPermissions = (
  organization: Organization,
  userRole?: string
): OrganizationOperationContext => {
  // 这里可以根据用户角色和组织状态计算权限
  const isAdmin = userRole === 'admin';
  const isManager = userRole === 'manager' || isAdmin;
  
  const permissions: OrganizationOperationContext = {
    canEdit: isManager,
    canDelete: isAdmin && organization.status === 'INACTIVE',
    canActivate: isManager && organization.status !== 'ACTIVE',
    canDeactivate: isManager && organization.status === 'ACTIVE',
    canViewHistory: true,
    canCreateChild: isManager,
    canMove: isAdmin,
    canViewTimeline: true
  };
  
  // 如果组织有子组织，不允许删除
  if (organization.childCount && organization.childCount > 0) {
    permissions.canDelete = false;
    permissions.reason = '存在子组织，无法删除';
  }
  
  return permissions;
};