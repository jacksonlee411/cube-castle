import { useState, useCallback } from 'react';
import { useDeleteOrganization, useCreateOrganization, useUpdateOrganization } from '../../../shared/hooks/useOrganizationMutations';
import type { OrganizationUnit } from '../../../shared/types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../shared/hooks/useOrganizationMutations';

export const useOrganizationActions = () => {
  const [selectedOrg, setSelectedOrg] = useState<OrganizationUnit | undefined>();
  const [isFormOpen, setIsFormOpen] = useState(false);
  
  const deleteMutation = useDeleteOrganization();
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

  const handleDelete = useCallback(async (code: string) => {
    if (window.confirm('确定要删除这个组织单元吗？')) {
      try {
        await deleteMutation.mutateAsync(code);
      } catch (error) {
        console.error('Delete operation failed:', error);
      }
    }
  }, [deleteMutation]);

  const handleFormClose = useCallback(() => {
    setIsFormOpen(false);
    setSelectedOrg(undefined);
  }, []);

  const handleFormSubmit = useCallback(async (data: CreateOrganizationInput | UpdateOrganizationInput) => {
    try {
      if (selectedOrg) {
        // Edit mode
        await updateMutation.mutateAsync({ 
          code: selectedOrg.code, 
          ...data as UpdateOrganizationInput 
        });
      } else {
        // Create mode
        await createMutation.mutateAsync(data as CreateOrganizationInput);
      }
      handleFormClose();
    } catch (error) {
      console.error('Form submission failed:', error);
    }
  }, [selectedOrg, createMutation, updateMutation]);

  // Get the currently deleting item ID for UI state
  const deletingId = deleteMutation.isPending ? deleteMutation.variables : undefined;

  return {
    // State
    selectedOrg,
    isFormOpen,
    deletingId,
    
    // Mutation state
    isDeleting: deleteMutation.isPending,
    
    // Actions
    handleCreate,
    handleEdit,
    handleDelete,
    handleFormClose,
    handleFormSubmit,
  };
};