import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { useAuth } from '@/shared/auth/hooks'
import { useJobFamilyGroups, useJobFamilies, useJobRoles } from '@/shared/hooks/useJobCatalog'
import { useCreateJobRole } from '@/shared/hooks/useJobCatalogMutations'
import { CatalogFilters } from '../shared/CatalogFilters'
import { CatalogTable, type CatalogTableColumn } from '../shared/CatalogTable'
import { StatusBadge } from '../shared/StatusBadge'
import { formatISODate } from '../types'
import { JobRoleForm } from './JobRoleForm'
import { SimpleStack } from '@/features/positions/components'
import { CardContainer } from '@/shared/components/CardContainer'

const inlineSelectStyle: React.CSSProperties = {
  minWidth: '180px',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  backgroundColor: colors.frenchVanilla100,
  fontSize: '14px',
}

export const JobRoleList: React.FC = () => {
  const { hasPermission } = useAuth()
  const navigate = useNavigate()
  const [groupCode, setGroupCode] = useState('')
  const [familyCode, setFamilyCode] = useState('')
  const [searchText, setSearchText] = useState('')
  const [includeInactive, setIncludeInactive] = useState(false)
  const [asOfDate, setAsOfDate] = useState<string | undefined>(undefined)
  const [isFormOpen, setFormOpen] = useState(false)

  const groupQuery = useJobFamilyGroups({ includeInactive: false })
  const familyQuery = useJobFamilies(groupCode || undefined, { includeInactive: false })
  const rolesQuery = useJobRoles(familyCode || undefined, { includeInactive, asOfDate })

  const createMutation = useCreateJobRole()

  const filteredRoles = useMemo(() => {
    const data = rolesQuery.data ?? []
    if (!searchText.trim()) {
      return data
    }
    const keyword = searchText.trim().toLowerCase()
    return data.filter(item =>
      item.code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword),
    )
  }, [rolesQuery.data, searchText])

  const columns: CatalogTableColumn<(typeof filteredRoles)[number]>[] = [
    { key: 'code', label: '职务编码', width: '220px' },
    { key: 'name', label: '职务名称' },
    { key: 'familyCode', label: '归属职种', width: '200px' },
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

  const handleCreate = async (values: Parameters<typeof createMutation.mutateAsync>[0]) => {
    await createMutation.mutateAsync(values)
    setFormOpen(false)
  }

  return (
    <Box padding={space.l}>
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="center">
          <Heading size="large">职务管理</Heading>
          {hasPermission('job-catalog:create') && (
            <PrimaryButton onClick={() => setFormOpen(true)} disabled={!familyCode}>
              新增职务
            </PrimaryButton>
          )}
        </Flex>

        <CardContainer>
          <CatalogFilters
            searchPlaceholder="搜索职务编码或名称"
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
                  onChange={event => setFamilyCode(event.target.value)}
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
              </Flex>
            }
            onReset={() => {
              setSearchText('')
              setIncludeInactive(false)
              setAsOfDate(undefined)
              setFamilyCode('')
              setGroupCode('')
            }}
          />
        </CardContainer>

        <CardContainer>
          <SimpleStack gap={space.s}>
            {familyCode === '' && (
              <Text color="licorice400">请先选择职类和职种以查看对应职务列表。</Text>
            )}

            <CatalogTable
              data={filteredRoles}
              columns={columns}
              isLoading={rolesQuery.isLoading}
              onRowClick={item => navigate(`/positions/catalog/roles/${item.code}`)}
              emptyMessage={familyCode ? '暂无职务数据' : '请选择职类和职种后查看'}
            />
          </SimpleStack>
        </CardContainer>

        <JobRoleForm
          isOpen={isFormOpen}
          onClose={() => setFormOpen(false)}
          onSubmit={handleCreate}
          isSubmitting={createMutation.isPending}
          familyCode={familyCode}
        />
      </SimpleStack>
    </Box>
  )
}
