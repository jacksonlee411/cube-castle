import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { Card } from '@workday/canvas-kit-react/card'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { useEnterprisePositions } from '@/shared/hooks/useEnterprisePositions'
import type { PositionRecord } from '@/shared/types/positions'
import { PositionSummaryCards, PositionVacancyBoard, PositionHeadcountDashboard } from './components/dashboard'
import { PositionList } from './components/list'
import { SimpleStack } from './components/layout'

const statusOptions: Array<{ label: string; value: string }> = [
  { label: '全部状态', value: 'ALL' },
  { label: '在编', value: 'ACTIVE' },
  { label: '已填充', value: 'FILLED' },
  { label: '空缺', value: 'VACANT' },
  { label: '规划中', value: 'PLANNED' },
  { label: '已结束', value: 'INACTIVE' },
]

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
  const navigate = useNavigate()
  const isMockMode = import.meta.env.VITE_POSITIONS_MOCK_MODE !== 'false'

  const positionsQuery = useEnterprisePositions({
    status: statusFilter !== 'ALL' ? statusFilter : undefined,
    jobFamilyGroupCode: jobFamilyGroupFilter !== 'ALL' ? jobFamilyGroupFilter : undefined,
    page: 1,
    pageSize: 100,
  })

  const positionsData = positionsQuery.data?.positions
  const positions = useMemo(() => positionsData ?? [], [positionsData])
  const hasError = positionsQuery.isError
  const isLoading = positionsQuery.isLoading
  const hasNoData = !isLoading && !hasError && positions.length === 0

  const filteredPositions = useMemo(
    () => applyFilters(positions, searchText, statusFilter, jobFamilyGroupFilter),
    [positions, searchText, statusFilter, jobFamilyGroupFilter],
  )

  const jobFamilyGroupOptions = useMemo(() => {
    const map = new Map<string, string>()
    positions.forEach(item => {
      if (!map.has(item.jobFamilyGroupCode)) {
        map.set(item.jobFamilyGroupCode, item.jobFamilyGroupName ?? item.jobFamilyGroupCode)
      }
    })
    return Array.from(map.entries()).map(([code, label]) => ({ code, label }))
  }, [positions])

  const listData = filteredPositions
  const summaryData = filteredPositions
  const headcountOrganizationCode = filteredPositions[0]?.organizationCode ?? undefined

  return (
    <Box padding={space.l} data-testid="position-dashboard">
      <SimpleStack gap={space.l}>
        <SimpleStack gap={space.xs}>
          <Heading size="medium">职位管理（Stage 1 数据接入）</Heading>
          <Text color={colors.licorice500}>
            当前页面依赖 GraphQL 查询服务与 REST 命令服务，请确保后端接口可用。
          </Text>
          {isMockMode && (
            <Card
              padding={space.m}
              backgroundColor={colors.cinnamon100}
              data-testid="position-dashboard-mock-banner"
              style={{ borderLeft: `4px solid ${colors.cinnamon600}` }}
            >
              <SimpleStack gap={space.xs}>
                <Text color={colors.cinnamon600} fontWeight="bold">
                  ⚠️ 当前处于 Mock 模式，仅支持查看数据。
                </Text>
                <Text fontSize="12px" color={colors.cinnamon600}>
                  若需执行创建、编辑或版本操作，请将环境变量 `VITE_POSITIONS_MOCK_MODE=false` 并确保后端服务运行后再刷新页面。
                </Text>
              </SimpleStack>
            </Card>
          )}
          {isLoading && (
            <Text fontSize="12px" color={colors.licorice300}>
              正在加载最新职位数据...
            </Text>
          )}
          {hasError && (
            <Box data-testid="position-dashboard-error">
              <Text fontSize="12px" color={colors.cinnamon500}>
                无法加载职位数据，请刷新页面或联系系统管理员。
              </Text>
            </Box>
          )}
          {hasNoData && (
            <Text fontSize="12px" color={colors.licorice400}>
              暂无职位记录，如果这是异常情况，请检查数据同步或后端服务状态。
            </Text>
          )}
        </SimpleStack>

        <Flex justifyContent="flex-end">
          <PrimaryButton
            onClick={() => navigate('/positions/new')}
            data-testid="position-create-button"
            disabled={isMockMode || hasError}
          >
            创建职位
          </PrimaryButton>
        </Flex>

        <PositionSummaryCards positions={summaryData} />

        <PositionVacancyBoard />
        <PositionHeadcountDashboard organizationCode={headcountOrganizationCode} />

        <CardLikeContainer>
          <SimpleStack gap={space.m}>
            <Heading size="small">筛选条件</Heading>
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

        <PositionList positions={listData} onSelect={code => navigate(`/positions/${code}`)} />
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
