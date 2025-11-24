import type { OrganizationField } from './organizationManifest'
import { getOrganizationFieldMeta } from './organizationManifest'

const meta = (field: OrganizationField) => getOrganizationFieldMeta(field)

export const organizationFieldLabel = (field: OrganizationField, fallback: string) =>
  meta(field)?.label ?? fallback

export const organizationFieldPlaceholder = (field: OrganizationField, fallback?: string) =>
  meta(field)?.placeholder ?? fallback ?? ''

export const organizationFieldDescription = (field: OrganizationField) =>
  meta(field)?.description

export const organizationFieldRequired = (field: OrganizationField, fallback = false) =>
  meta(field)?.required ?? fallback

export const organizationFieldHelperText = (field: OrganizationField, error?: string) =>
  error ?? organizationFieldDescription(field) ?? ''
