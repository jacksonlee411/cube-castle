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
    canDelete: isAdmin && (organization.status === 'SUSPENDED' || organization.status === 'DELETED'),
    canActivate: isManager && organization.status !== 'ACTIVE',
    canDeactivate: isManager && organization.status === 'ACTIVE',
    canViewHistory: true,
    canCreateChild: isManager,
    canMove: isAdmin,
    canViewTimeline: true
  };
  
  // TODO-TEMPORARY: Child organization check disabled; implement API integration for child count in v4.3 by 2025-09-20.
  // if (organization.childCount && organization.childCount > 0) {
  //   permissions.canDelete = false;
  //   permissions.reason = '存在子组织，无法删除';
  // }
  
  return permissions;
};

/**
 * 基于 scopes 的权限计算（不替换现有 role 逻辑，供迁移期使用）
 */
export function getOperationPermissionsByScopes(
  organization: Organization,
  scopesInput: string[] | Set<string>
): OrganizationOperationContext {
  const scopes = Array.isArray(scopesInput)
    ? new Set(scopesInput)
    : scopesInput;

  const has = (s: string) => scopes.has(s);

  const canEdit = has('org:update');
  const canDelete = has('org:delete') && (organization.status === 'SUSPENDED' || organization.status === 'DELETED');
  const canActivate = has('org:activate') && organization.status !== 'ACTIVE';
  const canDeactivate = has('org:suspend') && organization.status === 'ACTIVE';
  const canViewHistory = has('org:read:history') || has('org:read');
  const canCreateChild = has('org:create:child') || has('org:create');
  const canMove = has('org:move');
  const canViewTimeline = has('org:read:timeline') || has('org:read');

  const ctx: OrganizationOperationContext = {
    canEdit,
    canDelete,
    canActivate,
    canDeactivate,
    canViewHistory,
    canCreateChild,
    canMove,
    canViewTimeline,
  };

  if (!canDelete && has('org:delete') && organization.status === 'ACTIVE') {
    ctx.reason = 'ACTIVE 状态下不允许删除';
  }

  return ctx;
}
