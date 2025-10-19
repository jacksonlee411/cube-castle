import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type {
  PositionAssignmentRecord,
  PositionRecord,
  PositionTimelineEvent,
  PositionTransferRecord,
} from '@/shared/types/positions'
import { getPositionStatusMeta } from '../statusMeta'
import { SimpleStack } from './SimpleStack'
import { PositionTransferDialog } from './PositionTransferDialog'

interface PositionDetailsProps {
  position?: PositionRecord
  timeline: PositionTimelineEvent[]
  assignments?: PositionAssignmentRecord[]
  currentAssignment?: PositionAssignmentRecord | null
  transfers?: PositionTransferRecord[]
  isLoading?: boolean
}

const StatusPill: React.FC<{ status: string }> = ({ status }) => {
  const meta = getPositionStatusMeta(status)
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '6px 12px',
        borderRadius: 16,
        fontSize: 14,
        fontWeight: 600,
        color: meta.color,
        backgroundColor: meta.background,
        border: `1px solid ${meta.border}`,
      }}
    >
      {meta.label}
    </span>
  )
}

const formatDateRange = (start: string, end?: string | null) => {
  if (!start && !end) {
    return '未提供'
  }
  if (!end) {
    return `${start} 起`
  }
  return `${start} - ${end}`
}

const AssignmentItem: React.FC<{ assignment: PositionAssignmentRecord; highlight?: boolean }> = ({
  assignment,
  highlight = false,
}) => (
  <Box
    border={`1px solid ${highlight ? colors.blueberry400 : colors.soap400}`}
    borderRadius="12px"
    padding={space.s}
    backgroundColor={highlight ? colors.blueberry50 : colors.frenchVanilla100}
  >
    <SimpleStack gap={space.xxxs}>
      <Flex justifyContent="space-between" alignItems="baseline">
        <Text fontWeight="bold">
          {assignment.employeeName}
          {assignment.employeeNumber ? `（${assignment.employeeNumber}）` : ''}
        </Text>
        <Text fontSize="12px" color={colors.licorice400}>
          {assignment.assignmentStatus} · {assignment.assignmentType}
        </Text>
      </Flex>
      <Text fontSize="12px" color={colors.licorice400}>
        任职时间：{formatDateRange(assignment.startDate, assignment.endDate)} · FTE：{assignment.fte.toFixed(2)}
      </Text>
      {assignment.notes && (
        <Text fontSize="12px" color={colors.licorice500}>
          备注：{assignment.notes}
        </Text>
      )}
    </SimpleStack>
  </Box>
)

const TransferItem: React.FC<{ transfer: PositionTransferRecord }> = ({ transfer }) => (
  <Box borderLeft={`2px solid ${colors.blueberry400}`} paddingLeft={space.m} marginBottom={space.m}>
    <Text fontWeight="bold">
      {transfer.fromOrganizationCode || '未记录'} → {transfer.toOrganizationCode}
    </Text>
    <Text fontSize="12px" color={colors.licorice400}>
      生效：{transfer.effectiveDate} · 发起人：{transfer.initiatedBy.name || transfer.initiatedBy.id}
    </Text>
    {transfer.operationReason && (
      <Text fontSize="12px" color={colors.licorice500}>
        原因：{transfer.operationReason}
      </Text>
    )}
  </Box>
)

const TimelineItem: React.FC<{ event: PositionTimelineEvent }> = ({ event }) => (
  <Box borderLeft={`2px solid ${colors.blueberry400}`} paddingLeft={space.m} marginBottom={space.m}>
    <Text fontWeight="bold">
      {event.title}{' '}
      <Text as="span" fontSize="12px" color={colors.licorice400}>
        （{getPositionStatusMeta(event.status).label}）
      </Text>
    </Text>
    <Text fontSize="12px" color={colors.licorice400}>
      生效：{event.effectiveDate}
      {event.endDate ? ` · 结束：${event.endDate}` : ''}
    </Text>
    {event.changeReason && (
      <Text as="p" marginY="xxs" color={colors.licorice500}>
        {event.changeReason}
      </Text>
    )}
  </Box>
)

export const PositionDetails: React.FC<PositionDetailsProps> = ({
  position,
  timeline,
  assignments = [],
  currentAssignment = null,
  transfers = [],
  isLoading = false,
}) => {
  if (isLoading) {
    return (
      <Card data-testid="position-detail-card" padding={space.l} height="100%" backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>正在加载职位详情...</Text>
      </Card>
    )
  }

  if (!position) {
    return (
      <Card data-testid="position-detail-card" padding={space.l} height="100%" backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>请选择左侧职位查看详情</Text>
      </Card>
    )
  }

  return (
    <Card data-testid="position-detail-card" padding={space.l} height="100%" backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.m}>
        <Flex alignItems="center" justifyContent="space-between">
          <Heading size="small">{position.title}</Heading>
          <StatusPill status={position.status} />
        </Flex>
        <Text fontSize="14px" color={colors.licorice500}>
          职位编码：{position.code}
        </Text>
        <Text fontSize="14px" color={colors.licorice500}>
          所属组织：{position.organizationName ?? `${position.organizationCode}（未提供名称）`}
        </Text>

        <Flex justifyContent="flex-end">
          <PositionTransferDialog position={position} />
        </Flex>

        <DividerLine />

        <SimpleStack gap={space.xs}>
          <Heading size="small">岗位信息</Heading>
          <Text>
            职类 / 职种：{position.jobFamilyGroupName ?? position.jobFamilyGroupCode} ·{' '}
            {position.jobFamilyName ?? position.jobFamilyCode}
          </Text>
          <Text>
            职务 / 职级：{position.jobRoleName ?? position.jobRoleCode} · {position.jobLevelName ?? position.jobLevelCode}
          </Text>
          <Text>
            职位类型 / 雇佣方式：{position.positionType}{' '}
            {position.employmentType ? `· ${position.employmentType}` : ''}
          </Text>
          <Text>
            编制：{position.headcountInUse} / {position.headcountCapacity}（可用 {position.availableHeadcount}）
          </Text>
          <Text>
            生效日期：{position.effectiveDate}
            {position.endDate ? ` · 结束日期：${position.endDate}` : ''}
          </Text>
          <Text>
            汇报职位：{position.reportsToPositionCode ?? '未设置'}
          </Text>
        </SimpleStack>

        <DividerLine />

        <SimpleStack gap={space.xs}>
          <Heading size="small">当前任职</Heading>
          {currentAssignment ? (
            <AssignmentItem assignment={currentAssignment} highlight />
          ) : (
            <Text color={colors.licorice400}>暂无当前任职</Text>
          )}
        </SimpleStack>

        <DividerLine />

        <SimpleStack gap={space.s}>
          <Heading size="small">任职历史</Heading>
          {assignments.length === 0 ? (
            <Text color={colors.licorice400}>暂无任职记录</Text>
          ) : (
            assignments.map(item => (
              <AssignmentItem key={item.assignmentId} assignment={item} highlight={item.assignmentId === currentAssignment?.assignmentId} />
            ))
          )}
        </SimpleStack>

        <DividerLine />

        <SimpleStack gap={space.s}>
          <Heading size="small">调动记录</Heading>
          {transfers.length === 0 ? (
            <Text color={colors.licorice400}>暂无调动记录</Text>
          ) : (
            transfers.map(item => <TransferItem key={item.transferId} transfer={item} />)
          )}
        </SimpleStack>

        <DividerLine />

        <SimpleStack gap={space.s}>
          <Heading size="small">职位时间线</Heading>
          {timeline.length === 0 ? (
            <Text color={colors.licorice400}>暂无时间线记录</Text>
          ) : (
            timeline.map(item => <TimelineItem key={item.id} event={item} />)
          )}
        </SimpleStack>
      </SimpleStack>
    </Card>
  )
}

const DividerLine: React.FC = () => (
  <Box borderBottom={`1px solid ${colors.soap400}`} marginY={space.s} />
)

export default PositionDetails
