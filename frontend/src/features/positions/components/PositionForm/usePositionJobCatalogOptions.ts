import { useMemo } from 'react'
import {
  useJobFamilyGroups,
  useJobFamilies,
  useJobLevels,
  useJobRoles,
} from '@/shared/hooks/useJobCatalog'
import type { SelectOption } from './types'

const buildLabel = (name?: string | null, code?: string) => {
  if (!code) {
    return name ?? ''
  }
  if (name && name !== code) {
    return `${name} (${code})`
  }
  return code
}

const ensureCurrentOption = (options: SelectOption[], currentValue: string, currentLabel?: string | null) => {
  if (!currentValue || options.some(option => option.value === currentValue)) {
    return options
  }
  return [...options, { value: currentValue, label: buildLabel(currentLabel, currentValue) }]
}

export const usePositionJobCatalogOptions = (
  params: {
    groupCode: string
    familyCode: string
    roleCode: string
    levelCode: string
  },
) => {
  const groupsQuery = useJobFamilyGroups()
  const familiesQuery = useJobFamilies(params.groupCode)
  const rolesQuery = useJobRoles(params.familyCode)
  const levelsQuery = useJobLevels(params.roleCode)

  const groupOptions: SelectOption[] = useMemo(() => {
    const base: SelectOption[] = [{ value: '', label: '请选择职类' }]
    const data = groupsQuery.data ?? []
    const mapped = data.map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], params.groupCode)
  }, [groupsQuery.data, params.groupCode])

  const familyOptions: SelectOption[] = useMemo(() => {
    const base: SelectOption[] = [{ value: '', label: '请选择职种' }]
    if (!params.groupCode) {
      return base
    }
    const mapped = (familiesQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], params.familyCode)
  }, [familiesQuery.data, params.familyCode, params.groupCode])

  const roleOptions: SelectOption[] = useMemo(() => {
    const base: SelectOption[] = [{ value: '', label: '请选择职务' }]
    if (!params.familyCode) {
      return base
    }
    const mapped = (rolesQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], params.roleCode)
  }, [rolesQuery.data, params.familyCode, params.roleCode])

  const levelOptions: SelectOption[] = useMemo(() => {
    const base: SelectOption[] = [{ value: '', label: '请选择职级' }]
    if (!params.roleCode) {
      return base
    }
    const mapped = (levelsQuery.data ?? []).map(item => ({
      value: item.code,
      label: buildLabel(item.name, item.code),
    }))
    return ensureCurrentOption([...base, ...mapped], params.levelCode)
  }, [levelsQuery.data, params.levelCode, params.roleCode])

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

export type PositionJobCatalogOptionsResult = ReturnType<typeof usePositionJobCatalogOptions>
