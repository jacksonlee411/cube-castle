import { useTemporalEntityDetail, type TemporalDetailResult } from './useTemporalEntityDetail'

export interface OrganizationDetailOptions {
  enabled?: boolean
  asOfDate?: string
}

/**
 * Thin wrapper for organization detail based on unified temporal hook.
 * Keeps external callsites simple and enforces the single-source entry.
 */
export function useOrganizationDetail(
  code: string | undefined,
  options?: OrganizationDetailOptions,
): TemporalDetailResult {
  return useTemporalEntityDetail('organization', code, {
    enabled: options?.enabled,
    asOfDate: options?.asOfDate,
  })
}

export default useOrganizationDetail

