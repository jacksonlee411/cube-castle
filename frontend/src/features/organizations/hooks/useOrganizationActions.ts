import { useState, useCallback } from 'react';
import { useCreateOrganization, useUpdateOrganization } from '../../../shared/hooks/useOrganizationMutations';
import type { OrganizationUnit } from '../../../shared/types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../shared/hooks/useOrganizationMutations';

export const useOrganizationActions = () => {
  const [selectedOrg, setSelectedOrg] = useState<OrganizationUnit | undefined>();
  const [isFormOpen, setIsFormOpen] = useState(false);
  
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

  return {
    // State
    selectedOrg,
    isFormOpen,
    
    // Actions
    handleCreate,
    handleEdit,
    handleFormClose,
    handleFormSubmit,
  };
};