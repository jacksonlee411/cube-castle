import { useMemo } from 'react'
import {
  useJobFamilyGroups,
  useJobFamilies,
  useJobRoles,
  useJobLevels,
} from './useJobCatalog'

export interface PositionCatalogOption {
  value: string
  label: string
}

export interface PositionCatalogOptionsParams {
  groupCode?: string
  familyCode?: string
  roleCode?: string
  levelCode?: string
}

export interface PositionCatalogOptionsResult {
  groupOptions: PositionCatalogOption[]
  familyOptions: PositionCatalogOption[]
  roleOptions: PositionCatalogOption[]
  levelOptions: PositionCatalogOption[]
  isLoading: boolean
  hasError: boolean
}

const buildLabel = (name?: string | null, code?: string) => {
  if (!code) {
    return name ?? ''
  }
  if (name && name !== code) {
    return `${name} (${code})`
  }
  return code
}

const ensureCurrentOption = (
  options: PositionCatalogOption[],
  currentValue?: string,
  currentLabel?: string | null,
) => {
  if (!currentValue || options.some(option => option.value === currentValue)) {
    return options
  }
  return [...options, { value: currentValue, label: buildLabel(currentLabel, currentValue) }]
}

export const usePositionCatalogOptions = (
  params: PositionCatalogOptionsParams,
): PositionCatalogOptionsResult => {
  const groupCode = params.groupCode ?? ''
  const familyCode = params.familyCode ?? ''
  const roleCode = params.roleCode ?? ''
  const levelCode = params.levelCode ?? ''

  const groupsQuery = useJobFamilyGroups()
  const familiesQuery = useJobFamilies(groupCode)
  const rolesQuery = useJobRoles(familyCode)
  const levelsQuery = useJobLevels(roleCode)

  const groupOptions = useMemo<PositionCatalogOption[]>(() => {
    const base: PositionCatalogOption[] = [{ value: '', label: '请选择职类' }]
    const mapped = (groupsQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], groupCode)
  }, [groupsQuery.data, groupCode])

  const familyOptions = useMemo<PositionCatalogOption[]>(() => {
    const base: PositionCatalogOption[] = [{ value: '', label: '请选择职种' }]
    if (!groupCode) {
      return base
    }
    const mapped = (familiesQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], familyCode)
  }, [familiesQuery.data, familyCode, groupCode])

  const roleOptions = useMemo<PositionCatalogOption[]>(() => {
    const base: PositionCatalogOption[] = [{ value: '', label: '请选择职务' }]
    if (!familyCode) {
      return base
    }
    const mapped = (rolesQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], roleCode)
  }, [rolesQuery.data, familyCode, roleCode])

  const levelOptions = useMemo<PositionCatalogOption[]>(() => {
    const base: PositionCatalogOption[] = [{ value: '', label: '请选择职级' }]
    if (!roleCode) {
      return base
    }
    const mapped = (levelsQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], levelCode)
  }, [levelsQuery.data, levelCode, roleCode])

  const isLoading =
    groupsQuery.isLoading || familiesQuery.isLoading || rolesQuery.isLoading || levelsQuery.isLoading

  const hasError = groupsQuery.isError || familiesQuery.isError || rolesQuery.isError || levelsQuery.isError

  return {
    groupOptions,
    familyOptions,
    roleOptions,
    levelOptions,
    isLoading,
    hasError,
  }
}

