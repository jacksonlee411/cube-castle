import { describe, it, expect } from 'vitest';
import { prepareFormDataForValidation } from '../validation';
import { validateForm } from '../../../../../shared/validation/schemas';
import type { FormData } from '../FormTypes';
import { ROOT_PARENT_CODE } from '@/shared/utils/organization-helpers';
import { OrganizationUnitTypeEnum, OrganizationStatusEnum } from '@/shared/types/contract_gen';
import type { JsonValue } from '@/shared/types/json';

describe('OrganizationForm validation normalization', () => {
  it('allows empty code to be treated as optional when validating create payloads', () => {
    const formData: FormData = {
      code: '',
      name: '测试组织',
      unitType: OrganizationUnitTypeEnum.Department,
      status: OrganizationStatusEnum.Active,
      parentCode: ROOT_PARENT_CODE,
      description: '  ',
      sortOrder: 0,
      level: 1,
      isTemporal: false,
      effectiveFrom: '',
      effectiveTo: '',
      changeReason: '  '
    };

    const normalized = prepareFormDataForValidation(formData);
    const errors = validateForm(normalized as JsonValue, false);

    expect(errors.code).toBeUndefined();
  });
});
