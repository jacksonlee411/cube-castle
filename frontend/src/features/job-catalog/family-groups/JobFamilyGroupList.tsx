import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading } from '@workday/canvas-kit-react/text'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { space } from '@workday/canvas-kit-react/tokens'
import { useAuth } from '@/shared/auth/hooks'
import { CatalogFilters } from '../shared/CatalogFilters'
import { CatalogTable, type CatalogTableColumn } from '../shared/CatalogTable'
import { StatusBadge } from '../shared/StatusBadge'
import { formatISODate } from '../types'
import { useJobFamilyGroups } from '@/shared/hooks/useJobCatalog'
import { useCreateJobFamilyGroup } from '@/shared/hooks/useJobCatalogMutations'
import { JobFamilyGroupForm } from './JobFamilyGroupForm'
import { SimpleStack } from '@/features/positions/components/SimpleStack'
import { CardContainer } from '@/shared/components/CardContainer'

export const JobFamilyGroupList: React.FC = () => {
  const navigate = useNavigate()
  const { hasPermission } = useAuth()
  const [searchText, setSearchText] = useState('')
  const [includeInactive, setIncludeInactive] = useState(false)
  const [asOfDate, setAsOfDate] = useState<string | undefined>(undefined)
  const [isFormOpen, setFormOpen] = useState(false)
  const {
    data: groups = [],
    isLoading,
  } = useJobFamilyGroups({ includeInactive, asOfDate })

  const createMutation = useCreateJobFamilyGroup()

  const filtered = useMemo(() => {
    if (!searchText) {
      return groups
    }
    const keyword = searchText.trim().toLowerCase()
    return groups.filter(item =>
      item.code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword),
    )
  }, [groups, searchText])

  const columns: CatalogTableColumn<(typeof filtered)[number]>[] = [
    { key: 'code', label: '职类编码', width: '160px' },
    { key: 'name', label: '职类名称' },
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

  const handleCreate = async (input: Parameters<typeof createMutation.mutateAsync>[0]) => {
    await createMutation.mutateAsync(input)
    setFormOpen(false)
  }

  return (
    <Box padding={space.l}>
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="center">
          <Heading size="large">职类管理</Heading>
          {hasPermission('job-catalog:create') && (
            <PrimaryButton onClick={() => setFormOpen(true)}>新增职类</PrimaryButton>
          )}
        </Flex>

        <CardContainer>
          <CatalogFilters
            searchValue={searchText}
            onSearchChange={setSearchText}
            includeInactive={includeInactive}
            onIncludeInactiveChange={setIncludeInactive}
            asOfDate={asOfDate}
            onAsOfDateChange={setAsOfDate}
            onReset={() => {
              setSearchText('')
              setIncludeInactive(false)
              setAsOfDate(undefined)
            }}
          />
        </CardContainer>

        <CardContainer>
          <CatalogTable
            data={filtered}
            columns={columns}
            isLoading={isLoading}
            onRowClick={item => navigate(`/positions/catalog/family-groups/${item.code}`)}
          />
        </CardContainer>

        <JobFamilyGroupForm
          isOpen={isFormOpen}
          onClose={() => setFormOpen(false)}
          onSubmit={handleCreate}
          isSubmitting={createMutation.isPending}
        />
      </SimpleStack>
    </Box>
  )
}
