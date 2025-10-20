import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { useAuth } from '@/shared/auth/hooks'
import { useJobFamilyGroups, useJobFamilies, useJobRoles, useJobLevels } from '@/shared/hooks/useJobCatalog'
import { useCreateJobLevel } from '@/shared/hooks/useJobCatalogMutations'
import { CatalogFilters } from '../shared/CatalogFilters'
import { CatalogTable, type CatalogTableColumn } from '../shared/CatalogTable'
import { StatusBadge } from '../shared/StatusBadge'
import { formatISODate } from '../types'
import { JobLevelForm } from './JobLevelForm'
import { SimpleStack } from '@/features/positions/components'
import { CardContainer } from '@/shared/components/CardContainer'

const inlineSelectStyle: React.CSSProperties = {
  minWidth: '160px',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
}

export const JobLevelList: React.FC = () => {
  const { hasPermission } = useAuth()
  const navigate = useNavigate()
  const [groupCode, setGroupCode] = useState('')
  const [familyCode, setFamilyCode] = useState('')
  const [roleCode, setRoleCode] = useState('')
  const [searchText, setSearchText] = useState('')
  const [includeInactive, setIncludeInactive] = useState(false)
  const [asOfDate, setAsOfDate] = useState<string | undefined>(undefined)
  const [isFormOpen, setFormOpen] = useState(false)

  const groupQuery = useJobFamilyGroups({ includeInactive: false })
  const familyQuery = useJobFamilies(groupCode || undefined, { includeInactive: false })
  const roleQuery = useJobRoles(familyCode || undefined, { includeInactive: false })
  const levelsQuery = useJobLevels(roleCode || undefined, { includeInactive, asOfDate })

  const createMutation = useCreateJobLevel()

  const filteredLevels = useMemo(() => {
    const data = levelsQuery.data ?? []
    if (!searchText.trim()) {
      return data
    }
    const keyword = searchText.trim().toLowerCase()
    return data.filter(item =>
      item.code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword),
    )
  }, [levelsQuery.data, searchText])

  const columns: CatalogTableColumn<(typeof filteredLevels)[number]>[] = [
    { key: 'code', label: '职级编码', width: '120px' },
    { key: 'name', label: '职级名称' },
    { key: 'roleCode', label: '归属职务', width: '220px' },
    {
      key: 'levelRank',
      label: '等级序号',
      width: '120px',
      render: item => item.levelRank,
    },
    {
      key: 'status',
      label: '状态',
      width: '120px',
      render: item => <StatusBadge status={item.status} />,
    },
    {
      key: 'effectiveDate',
      label: '生效日期',
      width: '140px',
      render: item => formatISODate(item.effectiveDate),
    },
    {
      key: 'endDate',
      label: '结束日期',
      width: '140px',
      render: item => formatISODate(item.endDate),
    },
  ]

  const groupOptions = useMemo(() => {
    const base = [{ value: '', label: '全部职类' }]
    const groups = groupQuery.data ?? []
    return [
      ...base,
      ...groups.map(item => ({ value: item.code, label: `${item.name}（${item.code}）` })),
    ]
  }, [groupQuery.data])

  const familyOptions = useMemo(() => {
    const base = [{ value: '', label: '全部职种' }]
    const families = familyQuery.data ?? []
    return [
      ...base,
      ...families.map(item => ({ value: item.code, label: `${item.name}（${item.code}）` })),
    ]
  }, [familyQuery.data])

  const roleOptions = useMemo(() => {
    const base = [{ value: '', label: '全部职务' }]
    const roles = roleQuery.data ?? []
    return [
      ...base,
      ...roles.map(item => ({ value: item.code, label: `${item.name}（${item.code}）` })),
    ]
  }, [roleQuery.data])

  const handleCreate = async (values: Parameters<typeof createMutation.mutateAsync>[0]) => {
    await createMutation.mutateAsync(values)
    setFormOpen(false)
  }

  return (
    <Box padding={space.l}>
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="center">
          <Heading size="large">职级管理</Heading>
          {hasPermission('job-catalog:create') && (
            <PrimaryButton onClick={() => setFormOpen(true)} disabled={!roleCode}>
              新增职级
            </PrimaryButton>
          )}
        </Flex>

        <CardContainer>
          <CatalogFilters
            searchPlaceholder="搜索职级编码或名称"
            searchValue={searchText}
            onSearchChange={setSearchText}
            includeInactive={includeInactive}
            onIncludeInactiveChange={setIncludeInactive}
            asOfDate={asOfDate}
            onAsOfDateChange={setAsOfDate}
            extraFilters={
              <Flex gap="s">
                <select
                  value={groupCode}
                  onChange={event => {
                    const value = event.target.value
                    setGroupCode(value)
                    setFamilyCode('')
                    setRoleCode('')
                  }}
                  style={inlineSelectStyle}
                >
                  {groupOptions.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <select
                  value={familyCode}
                  onChange={event => {
                    const value = event.target.value
                    setFamilyCode(value)
                    setRoleCode('')
                  }}
                  disabled={!groupCode}
                  style={{
                    ...inlineSelectStyle,
                    backgroundColor: !groupCode ? colors.soap100 : colors.frenchVanilla100,
                    color: !groupCode ? colors.licorice400 : undefined,
                  }}
                >
                  {familyOptions.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <select
                  value={roleCode}
                  onChange={event => setRoleCode(event.target.value)}
                  disabled={!familyCode}
                  style={{
                    ...inlineSelectStyle,
                    backgroundColor: !familyCode ? colors.soap100 : colors.frenchVanilla100,
                    color: !familyCode ? colors.licorice400 : undefined,
                  }}
                >
                  {roleOptions.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </Flex>
            }
            onReset={() => {
              setSearchText('')
              setIncludeInactive(false)
              setAsOfDate(undefined)
              setRoleCode('')
              setFamilyCode('')
              setGroupCode('')
            }}
          />
        </CardContainer>

        <CardContainer>
          <SimpleStack gap={space.s}>
            {roleCode === '' && (
              <Text color="licorice400">请依次选择职类、职种、职务以查看关联职级。</Text>
            )}

            <CatalogTable
              data={filteredLevels}
              columns={columns}
              isLoading={levelsQuery.isLoading}
              onRowClick={item =>
                navigate(`/positions/catalog/levels/${item.code}`, {
                  state: { roleCode: item.roleCode },
                })
              }
              emptyMessage={roleCode ? '暂无职级数据' : '请选择职类、职种与职务后查看'}
            />
          </SimpleStack>
        </CardContainer>

        <JobLevelForm
          isOpen={isFormOpen}
          onClose={() => setFormOpen(false)}
          onSubmit={handleCreate}
          isSubmitting={createMutation.isPending}
          roleCode={roleCode}
        />
      </SimpleStack>
    </Box>
  )
}
