import React, { useEffect, useMemo, useState } from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Checkbox } from '@workday/canvas-kit-react/checkbox'
import { Table } from '@workday/canvas-kit-react/table'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { usePositionHeadcountStats } from '@/shared/hooks/useEnterprisePositions'
import type { PositionHeadcountStats } from '@/shared/types/positions'
import { SimpleStack } from '../layout/SimpleStack'
import temporalEntitySelectors from '@/shared/testids/temporalEntity'

interface PositionHeadcountDashboardProps {
  organizationCode?: string
}

const formatNumber = (value: number, fractionDigits = 1) =>
  Number.isFinite(value) ? value.toFixed(fractionDigits) : '0'

const formatPercent = (value: number) =>
  Number.isFinite(value) ? `${(value * 100).toFixed(1)}%` : '0%'

const buildCsv = (stats: PositionHeadcountStats): string => {
  const lines: string[][] = [
    ['Organization Code', stats.organizationCode],
    ['Organization Name', stats.organizationName],
    ['Total Capacity', stats.totalCapacity.toString()],
    ['Total Filled', stats.totalFilled.toString()],
    ['Total Available', stats.totalAvailable.toString()],
    ['Fill Rate', formatPercent(stats.fillRate)],
    [],
    ['Level Breakdown'],
    ['Job Level Code', 'Capacity', 'Utilized', 'Available'],
    ...stats.byLevel.map(item => [
      item.jobLevelCode,
      item.capacity.toString(),
      item.utilized.toString(),
      item.available.toString(),
    ]),
    [],
    ['Type Breakdown'],
    ['Position Type', 'Capacity', 'Filled', 'Available'],
    ...stats.byType.map(item => [
      item.positionType,
      item.capacity.toString(),
      item.filled.toString(),
      item.available.toString(),
    ]),
    [],
    ['Family Breakdown'],
    ['Job Family Code', 'Job Family Name', 'Capacity', 'Utilized', 'Available'],
    ...stats.byFamily.map(item => [
      item.jobFamilyCode,
      item.jobFamilyName ?? '',
      item.capacity.toString(),
      item.utilized.toString(),
      item.available.toString(),
    ]),
  ]

  return lines
    .map(row =>
      row
        .map(column => {
          const value = column ?? ''
          if (value.includes(',') || value.includes('"') || value.includes('\n')) {
            return `"${value.replace(/"/g, '""')}"`
          }
          return value
        })
        .join(','),
    )
    .join('\n')
}

const downloadCsv = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.rel = 'noopener'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export const PositionHeadcountDashboard: React.FC<PositionHeadcountDashboardProps> = ({
  organizationCode,
}) => {
  const [selectedOrganization, setSelectedOrganization] = useState<string>(organizationCode ?? '')
  const [formInput, setFormInput] = useState<string>(organizationCode ?? '')
  const [includeSubordinates, setIncludeSubordinates] = useState(true)

  useEffect(() => {
    if (organizationCode && organizationCode !== selectedOrganization) {
      setSelectedOrganization(organizationCode)
      setFormInput(organizationCode)
    }
  }, [organizationCode, selectedOrganization])

  const statsQuery = usePositionHeadcountStats({
    organizationCode: selectedOrganization,
    includeSubordinates,
  })

  const stats = statsQuery.data

  const summaryTiles = useMemo(() => {
    if (!stats) {
      return {
        totalCapacity: '0',
        totalFilled: '0',
        totalAvailable: '0',
        fillRate: '0%',
      }
    }
    return {
      totalCapacity: formatNumber(stats.totalCapacity, 0),
      totalFilled: formatNumber(stats.totalFilled, 0),
      totalAvailable: formatNumber(stats.totalAvailable),
      fillRate: formatPercent(stats.fillRate),
    }
  }, [stats])

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    const trimmed = formInput.trim()
    setSelectedOrganization(trimmed)
  }

  const handleExport = () => {
    if (!stats) {
      return
    }
    const csv = buildCsv(stats)
    const filename = `position-headcount-${stats.organizationCode}-${Date.now()}.csv`
    downloadCsv(csv, filename)
  }

const renderLevelTable = (data: PositionHeadcountStats['byLevel']) => {
  if (!data.length) {
    return (
      <Table.Row>
        <Table.Cell colSpan={4}>
            <Text color={colors.licorice400} textAlign="center">
              暂无职级维度数据
            </Text>
          </Table.Cell>
        </Table.Row>
      )
    }

    return data.map(item => (
      <Table.Row key={item.jobLevelCode}>
        <Table.Cell>{item.jobLevelCode}</Table.Cell>
        <Table.Cell>{formatNumber(item.capacity, 1)}</Table.Cell>
        <Table.Cell>{formatNumber(item.utilized, 1)}</Table.Cell>
        <Table.Cell>{formatNumber(item.available, 1)}</Table.Cell>
      </Table.Row>
  ))
}

const renderTypeTable = (data: PositionHeadcountStats['byType']) => {
  if (!data.length) {
      return (
        <Table.Row>
          <Table.Cell colSpan={4}>
            <Text color={colors.licorice400} textAlign="center">
              暂无职位类型维度数据
            </Text>
          </Table.Cell>
        </Table.Row>
      )
    }

  return data.map(item => (
    <Table.Row key={item.positionType}>
      <Table.Cell>{item.positionType}</Table.Cell>
      <Table.Cell>{formatNumber(item.capacity, 1)}</Table.Cell>
      <Table.Cell>{formatNumber(item.filled, 1)}</Table.Cell>
      <Table.Cell>{formatNumber(item.available, 1)}</Table.Cell>
    </Table.Row>
  ))
}

const renderFamilyTable = (data: PositionHeadcountStats['byFamily']) => {
  if (!data.length) {
    return (
      <Table.Row>
        <Table.Cell colSpan={5}>
          <Text color={colors.licorice400} textAlign="center">
            暂无职种维度数据
          </Text>
        </Table.Cell>
      </Table.Row>
    )
  }

  return data.map(item => (
    <Table.Row key={item.jobFamilyCode}>
      <Table.Cell>{item.jobFamilyCode}</Table.Cell>
      <Table.Cell>{item.jobFamilyName ?? '—'}</Table.Cell>
      <Table.Cell>{formatNumber(item.capacity, 1)}</Table.Cell>
      <Table.Cell>{formatNumber(item.utilized, 1)}</Table.Cell>
      <Table.Cell>{formatNumber(item.available, 1)}</Table.Cell>
    </Table.Row>
  ))
}

  return (
    <Card padding={space.l} data-testid={temporalEntitySelectors.position.headcountDashboard} backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="flex-start" flexWrap="wrap" rowGap={space.m}>
          <SimpleStack gap={space.xxxs}>
            <Heading size="small">职位编制统计</Heading>
            <Text color={colors.licorice500} fontSize="14px">
              查看指定组织的编制占用情况、职级分布和职位类型分布，辅助判断招聘与调度策略。
            </Text>
          </SimpleStack>

          <Flex as="form" gap={space.s} onSubmit={handleSubmit}>
            <TextInput
              placeholder="输入 7 位组织编码，例如 1000000"
              value={formInput}
              onChange={event => setFormInput(event.target.value)}
              width={220}
              data-testid={temporalEntitySelectors.position.headcountOrgInput}
            />
            <Checkbox
              label="包含下级组织"
              checked={includeSubordinates}
              onChange={event => setIncludeSubordinates(event.target.checked)}
              data-testid={temporalEntitySelectors.position.headcountIncludeSubordinates}
            />
            <PrimaryButton type="submit" disabled={statsQuery.isFetching}>
              加载统计
            </PrimaryButton>
            <SecondaryButton
              type="button"
              disabled={!stats}
              onClick={handleExport}
              data-testid={temporalEntitySelectors.position.headcountExportButton}
            >
              导出 CSV
            </SecondaryButton>
          </Flex>
        </Flex>

        {!selectedOrganization && (
          <Text fontSize="14px" color={colors.cinnamon500}>
            请先输入组织编码并点击“加载统计”。建议使用根组织（例如 1000000）或职位所属的组织编码。
          </Text>
        )}

        {selectedOrganization && (
          <>
            {statsQuery.isLoading ? (
              <Text fontSize="14px" color={colors.licorice400}>
                正在加载编制数据...
              </Text>
            ) : statsQuery.isError ? (
              <Text fontSize="14px" color={colors.cinnamon500}>
                加载失败：{(statsQuery.error as Error)?.message ?? '请稍后重试或检查组织编码'}
              </Text>
            ) : stats ? (
              <SimpleStack gap={space.l}>
                <Flex gap={space.l} flexWrap="wrap">
                  <SummaryTile label="总编制" value={summaryTiles.totalCapacity} accentColor={colors.blueberry500} />
                  <SummaryTile label="已占用" value={summaryTiles.totalFilled} accentColor={colors.cantaloupe500} />
                  <SummaryTile label="可用编制" value={summaryTiles.totalAvailable} accentColor={colors.cinnamon500} />
                  <SummaryTile label="占用率" value={summaryTiles.fillRate} accentColor={colors.greenApple500} />
                </Flex>

                <Flex gap={space.l} flexDirection="row" flexWrap="wrap">
                  <Box flex="1" minWidth="280px" border={`1px solid ${colors.soap400}`} borderRadius="12px" overflow="hidden">
                    <Table data-testid={temporalEntitySelectors.position.headcountLevelTable}>
                      <Table.Head>
                        <Table.Row>
                          <Table.Header width="120px">职级</Table.Header>
                          <Table.Header>编制</Table.Header>
                          <Table.Header>已占用</Table.Header>
                          <Table.Header>可用</Table.Header>
                        </Table.Row>
                      </Table.Head>
                      <Table.Body>{renderLevelTable(stats.byLevel)}</Table.Body>
                    </Table>
                  </Box>
                  <Box flex="1" minWidth="280px" border={`1px solid ${colors.soap400}`} borderRadius="12px" overflow="hidden">
                    <Table data-testid={temporalEntitySelectors.position.headcountTypeTable}>
                      <Table.Head>
                        <Table.Row>
                          <Table.Header width="160px">职位类型</Table.Header>
                          <Table.Header>编制</Table.Header>
                          <Table.Header>已占用</Table.Header>
                          <Table.Header>可用</Table.Header>
                        </Table.Row>
                      </Table.Head>
                      <Table.Body>{renderTypeTable(stats.byType)}</Table.Body>
                    </Table>
                  </Box>
                  <Box flex="1" minWidth="320px" border={`1px solid ${colors.soap400}`} borderRadius="12px" overflow="hidden">
                    <Table data-testid={temporalEntitySelectors.position.headcountFamilyTable}>
                      <Table.Head>
                        <Table.Row>
                          <Table.Header width="160px">职种编码</Table.Header>
                          <Table.Header>职种名称</Table.Header>
                          <Table.Header>编制</Table.Header>
                          <Table.Header>已占用</Table.Header>
                          <Table.Header>可用</Table.Header>
                        </Table.Row>
                      </Table.Head>
                      <Table.Body>{renderFamilyTable(stats.byFamily)}</Table.Body>
                    </Table>
                  </Box>
                </Flex>

                <Text fontSize="12px" color={colors.licorice300} textAlign="right">
                  数据更新时间：{stats.fetchedAt ? new Date(stats.fetchedAt).toLocaleString() : '—'}
                </Text>
              </SimpleStack>
            ) : (
              <Text fontSize="14px" color={colors.licorice400}>
                未获取到数据，请确认组织编码是否正确。
              </Text>
            )}
          </>
        )}
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
