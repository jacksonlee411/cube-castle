import React, { useMemo, useState } from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Select } from '@workday/canvas-kit-react/select'
import { Table } from '@workday/canvas-kit-react/table'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { SecondaryButton } from '@workday/canvas-kit-react/button'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { useVacantPositions, type VacantPositionsQueryParams } from '@/shared/hooks/useEnterprisePositions'
import type { VacantPositionRecord } from '@/shared/types/positions'
import { SimpleStack } from './SimpleStack'

const MINIMUM_VACANT_OPTIONS: Array<{ label: string; value?: number }> = [
  { label: '全部空缺', value: undefined },
  { label: '空缺 ≥ 30 天', value: 30 },
  { label: '空缺 ≥ 60 天', value: 60 },
  { label: '空缺 ≥ 90 天', value: 90 },
]

const formatDays = (since: string) => {
  const start = new Date(since)
  if (Number.isNaN(start.getTime())) {
    return '未知'
  }
  const diff = Math.floor((Date.now() - start.getTime()) / (1000 * 60 * 60 * 24))
  return diff <= 0 ? '当天' : `${diff} 天`
}

const summaryFromRecords = (records: VacantPositionRecord[]) => {
  const capacity = records.reduce((sum, item) => sum + item.headcountCapacity, 0)
  const available = records.reduce((sum, item) => sum + item.headcountAvailable, 0)
  return { totalPositions: records.length, totalCapacity: capacity, totalAvailable: available }
}

const buildQueryParams = (minimumVacantDays?: number): VacantPositionsQueryParams => ({
  minimumVacantDays,
  sortField: 'HEADCOUNT_AVAILABLE',
  sortDirection: 'DESC',
  page: 1,
  pageSize: 25,
})

export const PositionVacancyBoard: React.FC = () => {
  const [minimumVacantDays, setMinimumVacantDays] = useState<number | undefined>(undefined)

  const queryParams = useMemo(() => buildQueryParams(minimumVacantDays), [minimumVacantDays])
  const vacancyQuery = useVacantPositions(queryParams)

  const records = useMemo(() => vacancyQuery.data?.data ?? [], [vacancyQuery.data])
  const summary = useMemo(() => summaryFromRecords(records), [records])

  return (
    <Card padding={space.l} data-testid="position-vacancy-board" backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="flex-start" flexWrap="wrap" rowGap={space.m}>
          <SimpleStack gap={space.xxxs}>
            <Heading size="small">空缺职位看板</Heading>
            <Text color={colors.licorice500} fontSize="14px">
              监控长期空缺职位，优先关注空缺天数较长、可用编制较大的岗位。
            </Text>
          </SimpleStack>

          <Flex gap={space.s} alignItems="center">
            <Select
              value={minimumVacantDays !== undefined ? String(minimumVacantDays) : 'ALL'}
              onChange={event => {
                const value = event.target.value
                setMinimumVacantDays(value === 'ALL' ? undefined : Number(value))
              }}
              aria-label="空缺时长筛选"
            >
              {MINIMUM_VACANT_OPTIONS.map(option => (
                <option key={option.label} value={option.value ?? 'ALL'}>
                  {option.label}
                </option>
              ))}
            </Select>
            <SecondaryButton onClick={() => vacancyQuery.refetch()} disabled={vacancyQuery.isFetching}>
              刷新数据
            </SecondaryButton>
          </Flex>
        </Flex>

        <Flex gap={space.l} flexWrap="wrap">
          <SummaryTile label="空缺职位数" value={summary.totalPositions.toString()} accentColor={colors.blueberry500} />
          <SummaryTile
            label="总编制"
            value={summary.totalCapacity.toFixed(0)}
            accentColor={colors.cantaloupe500}
          />
          <SummaryTile
            label="可用编制"
            value={summary.totalAvailable.toFixed(1)}
            accentColor={colors.cinnamon500}
          />
        </Flex>

        <Box border={`1px solid ${colors.soap400}`} borderRadius="12px" overflow="hidden">
          <Table data-testid="vacant-position-table">
            <Table.Head>
              <Table.Row>
                <Table.Header width="120px">职位编码</Table.Header>
                <Table.Header>所属组织</Table.Header>
                <Table.Header width="140px">职务 / 职级</Table.Header>
                <Table.Header width="120px">空缺天数</Table.Header>
                <Table.Header width="120px">可用编制</Table.Header>
                <Table.Header width="120px">总编制</Table.Header>
                <Table.Header width="120px">历史任职数</Table.Header>
              </Table.Row>
            </Table.Head>
            <Table.Body>
              {vacancyQuery.isLoading ? (
                <Table.Row>
                  <Table.Cell colSpan={7}>
                    <Text textAlign="center" color={colors.licorice400}>
                      正在加载空缺职位数据...
                    </Text>
                  </Table.Cell>
                </Table.Row>
              ) : vacancyQuery.isError ? (
                <Table.Row>
                  <Table.Cell colSpan={7}>
                    <Text textAlign="center" color={colors.cinnamon500}>
                      加载失败：{(vacancyQuery.error as Error)?.message ?? '请稍后重试'}
                    </Text>
                  </Table.Cell>
                </Table.Row>
              ) : records.length === 0 ? (
                <Table.Row>
                  <Table.Cell colSpan={7}>
                    <Text textAlign="center" color={colors.licorice400}>
                      当前筛选条件下没有空缺职位。
                    </Text>
                  </Table.Cell>
                </Table.Row>
              ) : (
                records.map(item => (
                  <Table.Row key={item.positionCode}>
                    <Table.Cell>{item.positionCode}</Table.Cell>
                    <Table.Cell>
                      <Text fontWeight="bold">{item.organizationName ?? '未设置组织名称'}</Text>
                      <Text fontSize="12px" color={colors.licorice400}>
                        {item.organizationCode}
                      </Text>
                    </Table.Cell>
                    <Table.Cell>
                      <Text>{item.jobRoleCode}</Text>
                      <Text fontSize="12px" color={colors.licorice400}>
                        {item.jobLevelCode}
                      </Text>
                    </Table.Cell>
                    <Table.Cell>{formatDays(item.vacantSince)}</Table.Cell>
                    <Table.Cell>{item.headcountAvailable.toFixed(1)}</Table.Cell>
                    <Table.Cell>{item.headcountCapacity.toFixed(1)}</Table.Cell>
                    <Table.Cell>{item.totalAssignments}</Table.Cell>
                  </Table.Row>
                ))
              )}
            </Table.Body>
          </Table>
        </Box>

        <Flex justifyContent="flex-end">
          <Text fontSize="12px" color={colors.licorice300}>
            数据更新时间：{vacancyQuery.data?.fetchedAt ? new Date(vacancyQuery.data.fetchedAt).toLocaleString() : '—'}
          </Text>
        </Flex>
      </SimpleStack>
    </Card>
  )
}

interface SummaryTileProps {
  label: string
  value: string
  accentColor: string
}

const SummaryTile: React.FC<SummaryTileProps> = ({ label, value, accentColor }) => (
  <Box
    minWidth="180px"
    padding={space.m}
    borderRadius="12px"
    backgroundColor={colors.frenchVanilla100}
    border={`1px solid ${colors.soap400}`}
    boxShadow={`inset 4px 0 0 ${accentColor}`}
  >
    <Text fontSize="12px" color={colors.licorice400}>
      {label}
    </Text>
    <Text fontSize="28px" fontWeight="bold" color={colors.licorice500}>
      {value}
    </Text>
  </Box>
)
