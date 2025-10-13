import React, { useMemo, useState } from 'react'
import { Box, Flex, Stack } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { PositionSummaryCards } from './components/PositionSummaryCards'
import { PositionList } from './components/PositionList'
import { PositionDetails } from './components/PositionDetails'
import { mockPositions } from './mockData'

const jobFamilyGroupOptions = Array.from(new Set(mockPositions.map(item => item.jobFamilyGroup)))
const statusOptions: Array<{ label: string; value: string }> = [
  { label: '全部状态', value: 'ALL' },
  { label: '在编', value: 'ACTIVE' },
  { label: '已填充', value: 'FILLED' },
  { label: '空缺', value: 'VACANT' },
  { label: '规划中', value: 'PLANNED' },
  { label: '停用', value: 'INACTIVE' },
]

const filterPositions = (search: string, status: string, familyGroup: string) => {
  const keyword = search.trim().toLowerCase()
  return mockPositions.filter(item => {
    const matchKeyword =
      !keyword ||
      item.code.toLowerCase().includes(keyword) ||
      item.title.toLowerCase().includes(keyword) ||
      item.organization.name.toLowerCase().includes(keyword)

    const matchStatus = status === 'ALL' || item.status === status
    const matchFamily = familyGroup === 'ALL' || item.jobFamilyGroup === familyGroup

    return matchKeyword && matchStatus && matchFamily
  })
}

export const PositionDashboard: React.FC = () => {
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState('ALL')
  const [jobFamilyGroupFilter, setJobFamilyGroupFilter] = useState('ALL')

  const filteredPositions = useMemo(
    () => filterPositions(searchText, statusFilter, jobFamilyGroupFilter),
    [searchText, statusFilter, jobFamilyGroupFilter]
  )

  const [selectedCode, setSelectedCode] = useState(() => filteredPositions[0]?.code)

  const selectedPosition = useMemo(
    () => filteredPositions.find(item => item.code === selectedCode) ?? filteredPositions[0],
    [filteredPositions, selectedCode]
  )

  const handleSelectPosition = (code: string) => {
    setSelectedCode(code)
  }

  return (
    <Box padding={space.l} data-testid="position-dashboard">
      <Stack space={space.l}>
        <Stack space={space.xs}>
          <Heading level="2">职位管理（Stage 0 Mock）</Heading>
          <Text color={colors.licorice500}>
            当前页面展示的是 Stage 0 布局与交互框架，数据来源于内部 mock，待验收后再接入真实 API。
          </Text>
        </Stack>

        <PositionSummaryCards positions={filteredPositions} />

        <CardLikeContainer>
          <Stack space={space.m}>
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
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </NativeSelect>
            </Flex>
          </Stack>
        </CardLikeContainer>

        <Flex gap={space.l} alignItems="stretch" flexDirection={{ xs: 'column', md: 'row' }}>
          <Box flex="2" minWidth="60%">
            <PositionList positions={filteredPositions} selectedCode={selectedPosition?.code} onSelect={handleSelectPosition} />
          </Box>
          <Box flex="1" minWidth="35%">
            <PositionDetails position={selectedPosition} />
          </Box>
        </Flex>
      </Stack>
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
