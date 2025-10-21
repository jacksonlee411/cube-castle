import React, { useMemo, useState } from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Flex } from '@workday/canvas-kit-react/layout'
import { PrimaryButton, TertiaryButton } from '@workday/canvas-kit-react/button'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { Checkbox } from '@workday/canvas-kit-react/checkbox'
import { Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { SimpleStack } from '../layout/SimpleStack'
import { AssignmentItem } from './PositionDetails'
import { PaginationControls } from '@/features/organizations/PaginationControls'
import {
  usePositionAssignments,
  fetchPositionAssignmentAudit,
  type PositionAssignmentsQueryParams,
  type PositionAssignmentAuditParams,
  type PositionAssignmentsQueryResult,
} from '@/shared/hooks/useEnterprisePositions'
import type { PositionAssignmentAuditRecord, PositionAssignmentRecord } from '@/shared/types/positions'

interface PositionAssignmentHistoryProps {
  positionCode?: string
  currentAssignment?: PositionAssignmentRecord | null
}

interface AssignmentFilterState {
  assignmentTypes: string[]
  status: string
  dateFrom: string
  dateTo: string
  includeHistorical: boolean
  includeActingOnly: boolean
  page: number
  pageSize: number
}

const ASSIGNMENT_TYPES = [
  { value: 'PRIMARY', label: '主任职' },
  { value: 'SECONDARY', label: '副任职' },
  { value: 'ACTING', label: '代理任职' },
]

const STATUS_OPTIONS = [
  { value: 'ALL', label: '全部状态' },
  { value: 'ACTIVE', label: '进行中' },
  { value: 'PENDING', label: '待生效' },
  { value: 'ENDED', label: '已结束' },
]

const DEFAULT_FILTER_STATE: AssignmentFilterState = {
  assignmentTypes: [],
  status: 'ALL',
  dateFrom: '',
  dateTo: '',
  includeHistorical: true,
  includeActingOnly: false,
  page: 1,
  pageSize: 20,
}

const buildQueryParams = (filters: AssignmentFilterState): PositionAssignmentsQueryParams => ({
  page: filters.page,
  pageSize: filters.pageSize,
  assignmentTypes: filters.assignmentTypes.length > 0 ? filters.assignmentTypes : undefined,
  status: filters.status === 'ALL' ? undefined : filters.status,
  dateFrom: filters.dateFrom || undefined,
  dateTo: filters.dateTo || undefined,
  includeHistorical: filters.includeHistorical,
  includeActingOnly: filters.includeActingOnly,
})

const buildAuditParams = (filters: AssignmentFilterState): PositionAssignmentAuditParams => ({
  dateFrom: filters.dateFrom || undefined,
  dateTo: filters.dateTo || undefined,
})

const buildCsvContent = (records: PositionAssignmentAuditRecord[]): string => {
  const header = ['assignmentId', 'eventType', 'effectiveDate', 'endDate', 'actor', 'createdAt', 'changes']
  const rows = records.map(record => [
    record.assignmentId,
    record.eventType,
    record.effectiveDate,
    record.endDate ?? '',
    record.actor,
    record.createdAt,
    record.changes ? JSON.stringify(record.changes) : '',
  ])

  return [header, ...rows]
    .map(columns =>
      columns
        .map(column => {
          const cell = String(column ?? '')
          return `"${cell.replace(/"/g, '""')}"`
        })
        .join(','),
    )
    .join('\n')
}

const downloadCsv = (filename: string, content: string) => {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.style.display = 'none'
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

const getPaginationSnapshot = (
  result: PositionAssignmentsQueryResult | undefined,
  fallback: AssignmentFilterState,
) =>
  result?.pagination ?? {
    total: result?.totalCount ?? 0,
    page: fallback.page,
    pageSize: fallback.pageSize,
    hasNext: false,
    hasPrevious: fallback.page > 1,
  }

export const PositionAssignmentHistory: React.FC<PositionAssignmentHistoryProps> = ({
  positionCode,
  currentAssignment = null,
}) => {
  const [filters, setFilters] = useState<AssignmentFilterState>(DEFAULT_FILTER_STATE)
  const [isExporting, setIsExporting] = useState(false)
  const [exportError, setExportError] = useState<string | null>(null)

  const queryParams = useMemo(() => buildQueryParams(filters), [filters])
  const assignmentQuery = usePositionAssignments(positionCode, queryParams)

  const assignments = assignmentQuery.data?.data ?? []
  const totalCount = assignmentQuery.data?.totalCount ?? 0
  const pagination = getPaginationSnapshot(assignmentQuery.data, filters)

  const handleAssignmentTypeToggle = (value: string) => {
    setFilters(prev => {
      const exists = prev.assignmentTypes.includes(value)
      return {
        ...prev,
        assignmentTypes: exists ? prev.assignmentTypes.filter(item => item !== value) : [...prev.assignmentTypes, value],
        page: 1,
      }
    })
  }

  const handleStatusChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const value = event.target.value
    setFilters(prev => ({ ...prev, status: value, page: 1 }))
  }

  const handleDateChange = (key: 'dateFrom' | 'dateTo') => (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value
    setFilters(prev => ({ ...prev, [key]: value, page: 1 }))
  }

  const handleToggle = (key: 'includeHistorical' | 'includeActingOnly') => (event: React.ChangeEvent<HTMLInputElement>) => {
    const checked = event.target.checked
    setFilters(prev => ({ ...prev, [key]: checked, page: 1 }))
  }

  const handlePageChange = (page: number) => {
    setFilters(prev => ({ ...prev, page }))
  }

  const handleResetFilters = () => {
    setFilters(DEFAULT_FILTER_STATE)
  }

  const handleExport = async () => {
    if (!positionCode) {
      return
    }
    setIsExporting(true)
    setExportError(null)

    try {
      const auditParams = buildAuditParams(filters)
      const pageSize = 200
      let currentPage = 1
      const allRecords: PositionAssignmentAuditRecord[] = []

      while (true) {
        const result = await fetchPositionAssignmentAudit(positionCode, {
          ...auditParams,
          page: currentPage,
          pageSize,
        })
        allRecords.push(...result.records)

        if (!result.pagination.hasNext || result.records.length === 0) {
          break
        }

        currentPage += 1
      }

      const csv = buildCsvContent(allRecords)
      const fileName = `${positionCode}-assignments-${new Date().toISOString().slice(0, 10)}.csv`
      downloadCsv(fileName, csv)
    } catch (error) {
      const message = error instanceof Error ? error.message : '导出任职审计记录失败'
      setExportError(message)
    } finally {
      setIsExporting(false)
    }
  }

  if (!positionCode) {
    return (
      <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>请选择职位查看任职记录。</Text>
      </Card>
    )
  }

  return (
    <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.m}>
        <Flex gap={space.l} flexWrap="wrap">
          <SimpleStack gap={space.xxxs}>
            <Text fontSize="12px" color={colors.licorice400}>
              任职类型
            </Text>
            <Flex gap="s">
              {ASSIGNMENT_TYPES.map(option => (
                <Checkbox
                  key={option.value}
                  label={option.label}
                  checked={filters.assignmentTypes.includes(option.value)}
                  onChange={() => handleAssignmentTypeToggle(option.value)}
                />
              ))}
            </Flex>
          </SimpleStack>

          <SimpleStack gap={space.xxxs}>
            <Text fontSize="12px" color={colors.licorice400}>
              任职状态
            </Text>
            <select
              value={filters.status}
              onChange={handleStatusChange}
              style={{ padding: '8px 12px', borderRadius: 6, border: `1px solid ${colors.soap500}` }}
            >
              {STATUS_OPTIONS.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </SimpleStack>

          <SimpleStack gap={space.xxxs}>
            <Text fontSize="12px" color={colors.licorice400}>
              生效起始
            </Text>
            <TextInput type="date" value={filters.dateFrom} onChange={handleDateChange('dateFrom')} />
          </SimpleStack>

          <SimpleStack gap={space.xxxs}>
            <Text fontSize="12px" color={colors.licorice400}>
              生效结束
            </Text>
            <TextInput type="date" value={filters.dateTo} onChange={handleDateChange('dateTo')} />
          </SimpleStack>

          <SimpleStack gap={space.xxxs}>
            <Text fontSize="12px" color={colors.licorice400}>
              选项
            </Text>
            <Flex gap="s">
              <Checkbox
                label="包含历史"
                checked={filters.includeHistorical}
                onChange={handleToggle('includeHistorical')}
              />
              <Checkbox
                label="仅显示代理任职"
                checked={filters.includeActingOnly}
                onChange={handleToggle('includeActingOnly')}
              />
            </Flex>
          </SimpleStack>
        </Flex>

        <Flex gap={space.s} alignItems="center" justifyContent="space-between" flexWrap="wrap">
          <Flex gap={space.s} alignItems="center">
            <PrimaryButton onClick={handleExport} disabled={isExporting || assignmentQuery.isLoading}>
              {isExporting ? '正在导出...' : '导出 CSV'}
            </PrimaryButton>
            <TertiaryButton onClick={handleResetFilters} disabled={assignmentQuery.isFetching}>
              重置过滤条件
            </TertiaryButton>
            {exportError && (
              <Text fontSize="12px" color={colors.cinnamon500}>
                {exportError}
              </Text>
            )}
          </Flex>
          <Text fontSize="12px" color={colors.licorice400}>
            当前显示 {assignments.length} 条，共 {totalCount} 条
          </Text>
        </Flex>

        {assignmentQuery.isLoading ? (
          <Text color={colors.licorice400}>正在加载任职记录...</Text>
        ) : assignmentQuery.error ? (
          <Text color={colors.cinnamon500}>
            {(assignmentQuery.error as Error)?.message ?? '加载任职记录失败'}
          </Text>
        ) : assignments.length === 0 ? (
          <Text color={colors.licorice400}>暂无符合条件的任职记录</Text>
        ) : (
          <SimpleStack gap={space.s}>
            {assignments.map(item => (
              <AssignmentItem
                key={item.assignmentId}
                assignment={item}
                highlight={currentAssignment?.assignmentId === item.assignmentId}
              />
            ))}
          </SimpleStack>
        )}

        <PaginationControls
          currentPage={pagination.page}
          totalCount={pagination.total}
          pageSize={pagination.pageSize}
          onPageChange={handlePageChange}
          disabled={assignmentQuery.isFetching}
        />
      </SimpleStack>
    </Card>
  )
}

export default PositionAssignmentHistory
