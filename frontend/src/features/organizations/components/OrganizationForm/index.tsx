import React, { useState, useEffect, useCallback } from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { useCreateOrganization, useUpdateOrganization, useCreateOrganizationVersion } from '../../../../shared/hooks/useOrganizationMutations';
// import { useTemporalMode } from '../../../../shared/hooks/useTemporalQuery';
import { FormFields } from './FormFields';
import { validateForm } from '../../../../shared/validation/schemas';
import type { OrganizationFormProps, FormData } from './FormTypes';
import { prepareFormDataForValidation } from './validation';
import { TemporalConverter } from '../../../../shared/utils/temporal-converter';
import { useMessages } from '../../../../shared/hooks/useMessages';
import { normalizeParentCode } from '../../../../shared/utils/organization-helpers';
import { unifiedRESTClient } from '../../../../shared/api';
import type { OrganizationRequest } from '../../../../shared/types';

export const OrganizationForm: React.FC<OrganizationFormProps> = ({
  organization,
  isOpen,
  onClose,
  temporalMode = 'current',
  isHistorical = false,
  enableTemporalFeatures = true
}) => {
  const createMutation = useCreateOrganization();
  const updateMutation = useUpdateOrganization();
  const createVersionMutation = useCreateOrganizationVersion();
  // const { isCurrent, isPlanning } = useTemporalMode();
  const { showError } = useMessages();
  
  const isEditing = !!organization;
  const [isSubmitting, setIsSubmitting] = useState(false);
  const model = useModalModel();

  const [formData, setFormData] = useState<FormData>({
    code: organization?.code || '',
    name: organization?.name || '',
    unitType: organization?.unitType || 'DEPARTMENT',
    status: organization?.status || 'ACTIVE',
    description: organization?.description || '',
    parentCode: normalizeParentCode.forForm(organization?.parentCode),
    level: organization?.level || 1,
    sortOrder: organization?.sortOrder || 0,
    // 时态字段
    isTemporal: false,
    effectiveFrom: '',
    effectiveTo: '',
    changeReason: '',
  });

  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  // Modal state management
  useEffect(() => {
    if (isOpen && model.state.visibility !== 'visible') {
      model.events.show();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.hide();
    }
  }, [isOpen, model]);

  // Reset form data when organization changes
  useEffect(() => {
    setFormData({
      code: organization?.code || '',
      name: organization?.name || '',
      unitType: organization?.unitType || 'DEPARTMENT',
      status: organization?.status || 'ACTIVE',
      description: organization?.description || '',
      parentCode: normalizeParentCode.forForm(organization?.parentCode),
      level: organization?.level || 1,
      sortOrder: organization?.sortOrder || 0,
      // 时态字段重置
      isTemporal: false,
      effectiveFrom: '',
      effectiveTo: '',
      changeReason: '',
    });
    setFormErrors({});
  }, [organization]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Prevent double submission
    if (isSubmitting || createMutation.isPending || updateMutation.isPending || createVersionMutation.isPending) {
      return;
    }
    
    // Validate form
    const normalizedFormData = prepareFormDataForValidation(formData);
    const errors = validateForm(normalizedFormData, isEditing);
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      // 提交前服务器端数据校验（/api/v1/organization-units/validate）
      try {
        const operation = isEditing ? 'update' : 'create';
          const temporalEffectiveDate = formData.isTemporal && formData.effectiveFrom
            ? TemporalConverter.dateToDateString(formData.effectiveFrom as string)
            : undefined;

          const payload = {
            operation,
            data: {
              code: isEditing ? organization!.code : formData.code || undefined,
              name: formData.name,
              unitType: formData.unitType,
              status: (formData.status as 'ACTIVE' | 'INACTIVE'),
              parentCode: normalizeParentCode.forAPI(formData.parentCode),
            effectiveDate: temporalEffectiveDate
          },
          dryRun: true
        };
        const validateResp = await unifiedRESTClient.request('/organization-units/validate', {
          method: 'POST',
          body: JSON.stringify(payload)
        }) as unknown as { success?: boolean; data?: { valid?: boolean; errors?: string[]; warnings?: string[] } };

        if (validateResp && validateResp.success === true && validateResp.data && validateResp.data.valid === false) {
          const errs = validateResp.data.errors || ['服务器校验未通过'];
          const errorsMap: Record<string, string> = {};
          // 仅将第一条错误映射到通用字段；详细错误通过消息提示
          errorsMap['name'] = errs[0];
          setFormErrors(errorsMap);
          showError(errs.join('\n'));
          return; // 阻止后续提交
        }
      } catch (precheckError) {
        // 无权限或校验端点不可用时，不阻塞提交（后端最终裁决）
        const msg = precheckError instanceof Error ? precheckError.message : String(precheckError);
        if (/权限不足|禁止|Unauthorized|Forbidden/i.test(msg)) {
          console.warn('[Validate] 跳过服务器校验（权限不足）：', msg);
        } else {
          console.warn('[Validate] 校验端点不可用或失败，继续提交：', msg);
        }
      }
      const trimmedReason = formData.changeReason?.trim() ?? '';
      const operationReason = trimmedReason.length > 0 ? trimmedReason : undefined;

      if (isEditing) {
        if (formData.isTemporal) {
          if (!formData.effectiveFrom) {
            throw new Error('请填写时态版本的生效日期');
          }

          await createVersionMutation.mutateAsync({
            code: organization!.code,
            name: formData.name,
            unitType: formData.unitType as 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM',
            parentCode: normalizeParentCode.forAPI(formData.parentCode),
            description: formData.description || undefined,
            sortOrder: formData.sortOrder,
            effectiveDate: TemporalConverter.dateToDateString(formData.effectiveFrom as string),
            ...(formData.effectiveTo ? { endDate: TemporalConverter.dateToDateString(formData.effectiveTo as string) } : {}),
            ...(operationReason ? { operationReason } : {}),
          });
        } else {
          const updateData: OrganizationRequest = {
            code: organization!.code,
            name: formData.name,
            unitType: formData.unitType as 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM',
            status: formData.status as 'ACTIVE' | 'INACTIVE',
            description: formData.description,
            sortOrder: formData.sortOrder,
            parentCode: normalizeParentCode.forAPI(formData.parentCode),
            ...(operationReason ? { changeReason: operationReason, operationReason } : {}),
          };

          await updateMutation.mutateAsync(updateData);
        }
      } else {
        const createData: OrganizationRequest = {
          code: formData.code && formData.code.trim() ? formData.code.trim() : undefined,
          name: formData.name,
          unitType: formData.unitType as 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE',
          description: formData.description,
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
          ...(operationReason ? { changeReason: operationReason, operationReason } : {}),
          ...(formData.isTemporal && formData.effectiveFrom
            ? { effectiveDate: TemporalConverter.dateToDateString(formData.effectiveFrom as string) }
            : {}),
          ...(formData.isTemporal && formData.effectiveTo
            ? { endDate: TemporalConverter.dateToDateString(formData.effectiveTo as string) }
            : {}),
        };

        await createMutation.mutateAsync(createData);
      }
      
      // Reset form if creating new
      if (!isEditing) {
        setFormData({
          code: '',
          name: '',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          description: '',
          parentCode: '',
          level: 1,
          sortOrder: 0,
          // 重置时态字段
          isTemporal: false,
          effectiveFrom: '',
          effectiveTo: '',
          changeReason: '',
        });
      }
      
      model.events.hide();
      onClose();
    } catch (error) {
      console.error(`[Form] ${isEditing ? '更新' : '创建'}失败:`, error);
      
      let errorMessage = '操作失败';
      
      if (error && typeof error === 'object' && 'message' in error) {
        const apiError = error as { message: string };
        
        if (apiError.message.includes('duplicate key value violates unique constraint')) {
          if (apiError.message.includes('ukTenantName')) {
            errorMessage = '组织名称已存在，请使用不同的名称';
          } else {
            errorMessage = '数据重复，请检查输入信息';
          }
        } else if (apiError.message.includes('Network connection failed')) {
          errorMessage = '网络连接失败，请检查服务器状态';
        } else {
          errorMessage = apiError.message.split(' - ')[0] || errorMessage;
        }
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }
      
      showError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  }, [isEditing, formData, createMutation, updateMutation, createVersionMutation, isSubmitting, model, onClose, organization, showError]);

  const handleClose = () => {
    setIsSubmitting(false);
    setFormErrors({});
    model.events.hide();
    onClose();
  };

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width={600} data-testid="organization-form">
          <Modal.CloseIcon aria-label="关闭" onClick={handleClose} />
          <Modal.Heading>
            {isEditing ? '编辑组织单元' : '新增组织单元'}
            {formData.isTemporal && (
              <span style={{ 
                marginLeft: '8px', 
                fontSize: '14px', 
                color: '#1890ff',
                fontWeight: 'normal'
              }}>
                设置 组织详情
              </span>
            )}
          </Modal.Heading>
          <Modal.Body>
            <form onSubmit={handleSubmit} data-testid="organization-form-content">
              <FormFields
                formData={formData}
                setFormData={setFormData}
                isEditing={isEditing}
                temporalMode={temporalMode}
                enableTemporalFeatures={enableTemporalFeatures && !isHistorical}
              />

              {Object.keys(formErrors).length > 0 && (
                <div style={{ 
                  marginBottom: '16px', 
                  padding: '8px', 
                  backgroundColor: '#fff2f0', 
                  border: '1px solid #ffccc7',
                  borderRadius: '4px' 
                }}>
                  {Object.entries(formErrors).map(([field, error]) => (
                    <div key={field} style={{ color: '#ff4d4f', fontSize: '14px' }}>
                      {error}
                    </div>
                  ))}
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
                <SecondaryButton 
                  type="button" 
                  onClick={handleClose}
                  data-testid="form-cancel-button"
                >
                  取消
                </SecondaryButton>
                <PrimaryButton 
                  type="submit" 
                  disabled={isSubmitting || createMutation.isPending || updateMutation.isPending || createVersionMutation.isPending}
                  data-testid="form-submit-button"
                >
                  {isSubmitting ? '处理中...' : 
                   formData.isTemporal ? 
                     (isEditing ? '更新时态组织' : '创建计划组织') : 
                     (isEditing ? '更新' : '创建')
                  }
                </PrimaryButton>
              </div>
            </form>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};
