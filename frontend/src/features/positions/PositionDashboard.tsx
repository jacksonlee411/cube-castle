import React, { useEffect, useMemo, useState } from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { useEnterprisePositions, usePositionDetail } from '@/shared/hooks/useEnterprisePositions'
import type { PositionRecord, PositionTimelineEvent } from '@/shared/types/positions'
import { PositionSummaryCards } from './components/PositionSummaryCards'
import { PositionList } from './components/PositionList'
import { PositionDetails } from './components/PositionDetails'
import { PositionVacancyBoard } from './components/PositionVacancyBoard'
import { PositionHeadcountDashboard } from './components/PositionHeadcountDashboard'
import { SimpleStack } from './components/SimpleStack'
import { mockPositions } from './mockData'
import type { PositionMock } from './mockData'

const statusOptions: Array<{ label: string; value: string }> = [
  { label: '全部状态', value: 'ALL' },
  { label: '在编', value: 'ACTIVE' },
  { label: '已填充', value: 'FILLED' },
  { label: '空缺', value: 'VACANT' },
  { label: '规划中', value: 'PLANNED' },
  { label: '停用', value: 'INACTIVE' },
]

const lifecycleTypeToStatus = (type: string): string => {
  switch (type) {
    case 'CREATE':
      return 'PLANNED'
    case 'FILL':
      return 'FILLED'
    case 'VACATE':
      return 'VACANT'
    case 'SUSPEND':
      return 'INACTIVE'
    case 'REACTIVATE':
      return 'ACTIVE'
    case 'TRANSFER':
      return 'ACTIVE'
    default:
      return type.toUpperCase()
  }
}

const mapMockPositionToRecord = (item: PositionMock): PositionRecord => ({
  code: item.code,
  title: item.title,
  jobFamilyGroupCode: item.jobFamilyGroup,
  jobFamilyGroupName: item.jobFamilyGroup,
  jobFamilyCode: item.jobFamily,
  jobFamilyName: item.jobFamily,
  jobRoleCode: item.jobRole,
  jobRoleName: item.jobRole,
  jobLevelCode: item.jobLevel,
  jobLevelName: item.jobLevel,
  organizationCode: item.organization.code,
  organizationName: item.organization.name,
  positionType: 'REGULAR',
  employmentType: 'FULL_TIME',
  headcountCapacity: item.headcountCapacity,
  headcountInUse: item.headcountInUse,
  availableHeadcount: Math.max(item.headcountCapacity - item.headcountInUse, 0),
  gradeLevel: undefined,
  reportsToPositionCode: item.supervisor.code,
  status: item.status,
  effectiveDate: item.effectiveDate,
  endDate: undefined,
  isCurrent: item.status !== 'PLANNED',
  isFuture: item.status === 'PLANNED',
  createdAt: `${item.effectiveDate}T00:00:00.000Z`,
  updatedAt: `${item.effectiveDate}T00:00:00.000Z`,
})

const mapMockTimeline = (item: PositionMock): PositionTimelineEvent[] =>
  item.lifecycle.map(event => ({
    id: event.id,
    status: lifecycleTypeToStatus(event.type),
    title: event.label,
    effectiveDate: event.occurredAt,
    changeReason: event.summary,
  }))

const applyFilters = (
  positions: PositionRecord[],
  keyword: string,
  statusFilter: string,
  jobFamilyGroupFilter: string,
): PositionRecord[] => {
  const search = keyword.trim().toLowerCase()
  return positions.filter(item => {
    const matchesKeyword =
      search.length === 0 ||
      item.code.toLowerCase().includes(search) ||
      item.title.toLowerCase().includes(search) ||
      item.organizationCode.toLowerCase().includes(search) ||
      (item.organizationName ?? '').toLowerCase().includes(search)

    const matchesStatus = statusFilter === 'ALL' || item.status === statusFilter
    const matchesFamily = jobFamilyGroupFilter === 'ALL' || item.jobFamilyGroupCode === jobFamilyGroupFilter

    return matchesKeyword && matchesStatus && matchesFamily
  })
}

export const PositionDashboard: React.FC = () => {
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState('ALL')
  const [jobFamilyGroupFilter, setJobFamilyGroupFilter] = useState('ALL')

  const positionsQuery = useEnterprisePositions({
    status: statusFilter !== 'ALL' ? statusFilter : undefined,
    jobFamilyGroupCode: jobFamilyGroupFilter !== 'ALL' ? jobFamilyGroupFilter : undefined,
    page: 1,
    pageSize: 100,
  })

  const mockRecords = useMemo(() => mockPositions.map(mapMockPositionToRecord), [])
  const mockTimelineMap = useMemo(() => {
    const entries = mockPositions.map(item => [item.code, mapMockTimeline(item)] as const)
    return new Map(entries)
  }, [])

  const apiPositions = positionsQuery.data?.positions ?? []
  const useMockData = !positionsQuery.isLoading && (positionsQuery.isError || apiPositions.length === 0)
  const basePositions = useMockData ? mockRecords : apiPositions

  const filteredPositions = useMemo(
    () => applyFilters(basePositions, searchText, statusFilter, jobFamilyGroupFilter),
    [basePositions, searchText, statusFilter, jobFamilyGroupFilter],
  )

  const jobFamilyGroupOptions = useMemo(() => {
    const map = new Map<string, string>()
    basePositions.forEach(item => {
      if (!map.has(item.jobFamilyGroupCode)) {
        map.set(item.jobFamilyGroupCode, item.jobFamilyGroupName ?? item.jobFamilyGroupCode)
      }
    })
    return Array.from(map.entries()).map(([code, label]) => ({ code, label }))
  }, [basePositions])

  const [selectedCode, setSelectedCode] = useState<string>()

  useEffect(() => {
    if (filteredPositions.length === 0) {
      setSelectedCode(undefined)
      return
    }
    setSelectedCode(prev =>
      prev && filteredPositions.some(item => item.code === prev) ? prev : filteredPositions[0].code,
    )
  }, [filteredPositions])

  const detailQuery = usePositionDetail(selectedCode, {
    enabled: !useMockData && Boolean(selectedCode),
  })

  const selectedPosition = useMemo(() => {
    if (filteredPositions.length === 0) {
      return undefined
    }
    return filteredPositions.find(item => item.code === selectedCode) ?? filteredPositions[0]
  }, [filteredPositions, selectedCode])

  const detailPosition = useMockData
    ? selectedPosition
    : detailQuery.data?.position ?? selectedPosition

  const timeline: PositionTimelineEvent[] = useMockData
    ? selectedCode
      ? mockTimelineMap.get(selectedCode) ?? []
      : []
    : detailQuery.data?.timeline ?? []

  const assignments = useMockData ? [] : detailQuery.data?.assignments ?? []
  const currentAssignment = useMockData ? null : detailQuery.data?.currentAssignment ?? null
  const transfers = useMockData ? [] : detailQuery.data?.transfers ?? []

  const listData = filteredPositions
  const summaryData = filteredPositions
  const headcountOrganizationCode =
    detailPosition?.organizationCode ?? filteredPositions[0]?.organizationCode ?? undefined

  return (
    <Box padding={space.l} data-testid="position-dashboard">
      <SimpleStack gap={space.l}>
        <SimpleStack gap={space.xs}>
          <Heading level="2">职位管理（Stage 1 数据接入）</Heading>
          <Text color={colors.licorice500}>
            当前页面已接入 GraphQL 查询服务与 REST 命令服务，可进行职位筛选、搜索与时间线查看。
          </Text>
          <Text fontSize="12px" color={colors.licorice400}>
            数据来源：{useMockData ? '本地演示数据（API 不可用时自动回退）' : 'GraphQL / REST 实时数据'}
          </Text>
          {positionsQuery.isLoading && !useMockData && (
            <Text fontSize="12px" color={colors.licorice300}>
              正在加载最新职位数据...
            </Text>
          )}
        </SimpleStack>

        <PositionSummaryCards positions={summaryData} />

        <PositionVacancyBoard />
        <PositionHeadcountDashboard organizationCode={headcountOrganizationCode} />

        <CardLikeContainer>
          <SimpleStack gap={space.m}>
            <Heading level="4">筛选条件</Heading>
            <Flex gap={space.m} flexWrap="wrap">
              <TextInput
                placeholder="搜索职位名称 / 编码 / 组织"
                value={searchText}
                onChange={event => setSearchText(event.target.value)}
                width={320}
                data-testid="position-search-input"
              />
              <NativeSelect
                value={statusFilter}
                onChange={event => setStatusFilter(event.target.value)}
                data-testid="position-status-filter"
              >
                {statusOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </NativeSelect>
              <NativeSelect
                value={jobFamilyGroupFilter}
                onChange={event => setJobFamilyGroupFilter(event.target.value)}
                data-testid="position-fg-filter"
              >
                <option value="ALL">全部职类</option>
                {jobFamilyGroupOptions.map(option => (
                  <option key={option.code} value={option.code}>
                    {option.label}
                  </option>
                ))}
              </NativeSelect>
            </Flex>
          </SimpleStack>
        </CardLikeContainer>

        <Flex gap={space.l} alignItems="stretch" flexDirection={{ xs: 'column', md: 'row' }}>
          <Box flex="2" minWidth="60%">
            <PositionList positions={listData} selectedCode={selectedCode} onSelect={setSelectedCode} />
          </Box>
          <Box flex="1" minWidth="35%">
            <PositionDetails
              position={detailPosition}
              timeline={timeline}
              currentAssignment={currentAssignment ?? undefined}
              assignments={assignments}
              transfers={transfers}
              isLoading={!useMockData && detailQuery.isLoading}
              dataSource={useMockData ? 'mock' : 'api'}
            />
          </Box>
        </Flex>
      </SimpleStack>
    </Box>
  )
}

const CardLikeContainer: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <Box
    padding={space.l}
    borderRadius="16px"
    backgroundColor={colors.frenchVanilla100}
    border={`1px solid ${colors.soap400}`}
  >
    {children}
  </Box>
)

const NativeSelect: React.FC<
  React.SelectHTMLAttributes<HTMLSelectElement> & { 'data-testid'?: string }
> = ({ children, style, ...rest }) => (
  <Box
    as="select"
    padding={space.s}
    borderRadius="12px"
    border={`1px solid ${colors.soap400}`}
    backgroundColor={colors.frenchVanilla100}
    minWidth="200px"
    style={{ height: 44, ...style }}
    {...rest}
  >
    {children}
  </Box>
)

export default PositionDashboard
