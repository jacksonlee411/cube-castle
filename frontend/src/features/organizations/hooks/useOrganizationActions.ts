import { useState, useCallback } from 'react';
import { useToggleOrganizationStatus, useCreateOrganization, useUpdateOrganization } from '../../../shared/hooks/useOrganizationMutations';
import type { OrganizationUnit, OrganizationStatus } from '../../../shared/types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../shared/hooks/useOrganizationMutations';

export const useOrganizationActions = () => {
  const [selectedOrg, setSelectedOrg] = useState<OrganizationUnit | undefined>();
  const [isFormOpen, setIsFormOpen] = useState(false);
  
  const toggleStatusMutation = useToggleOrganizationStatus();
  const createMutation = useCreateOrganization();
  const updateMutation = useUpdateOrganization();

  const handleCreate = useCallback(() => {
    setSelectedOrg(undefined);
    setIsFormOpen(true);
  }, []);

  const handleEdit = useCallback((org: OrganizationUnit) => {
    setSelectedOrg(org);
    setIsFormOpen(true);
  }, []);

  const handleToggleStatus = useCallback(async (code: string, currentStatus: OrganizationStatus) => {
    const newStatus: OrganizationStatus = currentStatus === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE';
    const confirmMessage = newStatus === 'INACTIVE' 
      ? '确定要停用这个组织单元吗？' 
      : '确定要启用这个组织单元吗？';
      
    if (window.confirm(confirmMessage)) {
      try {
        await toggleStatusMutation.mutateAsync({ code, status: newStatus });
      } catch (error) {
        console.error('Toggle status operation failed:', error);
      }
    }
  }, [toggleStatusMutation]);

  const handleFormClose = useCallback(() => {
    setIsFormOpen(false);
    setSelectedOrg(undefined);
  }, []);

  const handleFormSubmit = useCallback(async (data: CreateOrganizationInput | UpdateOrganizationInput) => {
    try {
      if (selectedOrg) {
        // Edit mode
        const { code: _, ...updateData } = data as UpdateOrganizationInput;
        await updateMutation.mutateAsync({ 
          code: selectedOrg.code, 
          ...updateData 
        });
      } else {
        // Create mode
        await createMutation.mutateAsync(data as CreateOrganizationInput);
      }
      handleFormClose();
    } catch (error) {
      console.error('Form submission failed:', error);
    }
  }, [selectedOrg, createMutation, updateMutation, handleFormClose]);

  // Get the currently toggling item ID for UI state
  const togglingId = toggleStatusMutation.isPending ? toggleStatusMutation.variables?.code : undefined;

  return {
    // State
    selectedOrg,
    isFormOpen,
    togglingId,
    
    // Mutation state
    isToggling: toggleStatusMutation.isPending,
    
    // Actions
    handleCreate,
    handleEdit,
    handleToggleStatus,
    handleFormClose,
    handleFormSubmit,
  };
};