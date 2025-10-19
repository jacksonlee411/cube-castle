import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { Select } from '@workday/canvas-kit-react/select'
import { useAuth } from '@/shared/auth/hooks'
import { useJobFamilyGroups, useJobFamilies } from '@/shared/hooks/useJobCatalog'
import { useCreateJobFamily } from '@/shared/hooks/useJobCatalogMutations'
import { CatalogFilters } from '../shared/CatalogFilters'
import { CatalogTable, type CatalogTableColumn } from '../shared/CatalogTable'
import { StatusBadge } from '../shared/StatusBadge'
import { formatISODate } from '../types'
import { JobFamilyForm } from './JobFamilyForm'

export const JobFamilyList: React.FC = () => {
  const { hasPermission } = useAuth()
  const navigate = useNavigate()
  const [groupCode, setGroupCode] = useState('')
  const [searchText, setSearchText] = useState('')
  const [includeInactive, setIncludeInactive] = useState(false)
  const [asOfDate, setAsOfDate] = useState<string | undefined>(undefined)
  const [isFormOpen, setFormOpen] = useState(false)

  const groupQuery = useJobFamilyGroups({ includeInactive: false })
  const familiesQuery = useJobFamilies(groupCode || undefined, { includeInactive, asOfDate })

  const createMutation = useCreateJobFamily()

  const filteredFamilies = useMemo(() => {
    const data = familiesQuery.data ?? []
    if (!searchText.trim()) {
      return data
    }
    const keyword = searchText.trim().toLowerCase()
    return data.filter(item =>
      item.code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword),
    )
  }, [familiesQuery.data, searchText])

  const columns: CatalogTableColumn<(typeof filteredFamilies)[number]>[] = [
    { key: 'code', label: '职种编码', width: '200px' },
    { key: 'name', label: '职种名称' },
    { key: 'groupCode', label: '归属职类', width: '160px' },
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

  const handleCreate = async (values: Parameters<typeof createMutation.mutateAsync>[0]) => {
    await createMutation.mutateAsync(values)
    setFormOpen(false)
  }

  const groupOptions = useMemo(() => {
    const base = [{ value: '', label: '全部职类' }]
    const groups = groupQuery.data ?? []
    return [
      ...base,
      ...groups.map(item => ({
        value: item.code,
        label: `${item.name}（${item.code}）`,
      })),
    ]
  }, [groupQuery.data])

  return (
    <Box padding="l" display="flex" flexDirection="column">
      <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
        <Heading size="large">职种管理</Heading>
        {hasPermission('job-catalog:create') && (
          <PrimaryButton onClick={() => setFormOpen(true)} disabled={!groupCode}>
            新增职种
          </PrimaryButton>
        )}
      </Flex>

      <CatalogFilters
        searchPlaceholder="搜索职种编码或名称"
        searchValue={searchText}
        onSearchChange={setSearchText}
        includeInactive={includeInactive}
        onIncludeInactiveChange={setIncludeInactive}
        asOfDate={asOfDate}
        onAsOfDateChange={setAsOfDate}
        extraFilters={
          <Select value={groupCode} onChange={event => setGroupCode(event.target.value)}>
            {groupOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        }
        onReset={() => {
          setSearchText('')
          setIncludeInactive(false)
          setAsOfDate(undefined)
          setGroupCode('')
        }}
      />

      {groupCode === '' && (
        <Text marginBottom="s" color="licorice400">
          请选择左侧职类以查看其下属职种。
        </Text>
      )}

      <CatalogTable
        data={filteredFamilies}
        columns={columns}
        isLoading={familiesQuery.isLoading}
        onRowClick={item => navigate(`/positions/catalog/families/${item.code}`)}
        emptyMessage={groupCode ? '暂无职种数据' : '请选择职类后查看职种列表'}
      />

      <JobFamilyForm
        isOpen={isFormOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={handleCreate}
        isSubmitting={createMutation.isPending}
        groupCode={groupCode}
      />
    </Box>
  )
}
