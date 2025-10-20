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
import { getPositionStatusMeta } from '../../statusMeta'
import { SimpleStack } from '../layout/SimpleStack'
import { PositionTransferDialog } from '../transfer/PositionTransferDialog'

const SectionTitle: React.FC<{ title: string }> = ({ title }) => (
  <Heading size="small" as="h3">
    {title}
  </Heading>
)

const formatDateRange = (start: string, end?: string | null) => {
  if (!start && !end) {
    return '未提供'
  }
  if (!end) {
    return `${start} 起`
  }
  return `${start} - ${end}`
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

export const AssignmentItem: React.FC<{ assignment: PositionAssignmentRecord; highlight?: boolean }> = ({
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
        任职时间：{formatDateRange(assignment.effectiveDate, assignment.endDate)} · FTE：{assignment.fte.toFixed(2)}
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

interface PositionOverviewCardProps {
  position?: PositionRecord
  currentAssignment?: PositionAssignmentRecord | null
  isLoading?: boolean
}

export const PositionOverviewCard: React.FC<PositionOverviewCardProps> = ({
  position,
  currentAssignment = null,
  isLoading = false,
}) => {
  if (isLoading) {
    return (
      <Card data-testid='position-overview-card' padding={space.l} backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>正在加载职位详情...</Text>
      </Card>
    )
  }

  if (!position) {
    return (
      <Card data-testid='position-overview-card' padding={space.l} backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>请选择左侧职位查看详情</Text>
      </Card>
    )
  }

  return (
    <Card data-testid='position-overview-card' padding={space.l} backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.m}>
        <Flex alignItems='center' justifyContent='space-between'>
          <Heading size='small'>{position.title}</Heading>
          <StatusPill status={position.status} />
        </Flex>
        <Text fontSize='14px' color={colors.licorice500}>
          职位编码：{position.code}
        </Text>
        <Text fontSize='14px' color={colors.licorice500}>
          所属组织：{position.organizationName ?? `${position.organizationCode}（未提供名称）`}
        </Text>

        <Flex justifyContent='flex-end'>
          <PositionTransferDialog position={position} />
        </Flex>

        <DividerLine />

        <SimpleStack gap={space.xs}>
          <SectionTitle title='岗位信息' />
          <Text>
            职类 / 职种：{position.jobFamilyGroupName ?? position.jobFamilyGroupCode} ·{' '}
            {position.jobFamilyName ?? position.jobFamilyCode}
          </Text>
          <Text>
            职务 / 职级：{position.jobRoleName ?? position.jobRoleCode} · {position.jobLevelName ?? position.jobLevelCode}
          </Text>
          <Text>
            职位类型 / 雇佣方式：{position.positionType}
            {position.employmentType ? ` · ${position.employmentType}` : ''}
          </Text>
          <Text>
            编制：{position.headcountInUse} / {position.headcountCapacity}（可用 {position.availableHeadcount}）
          </Text>
          <Text>
            生效日期：{position.effectiveDate}
            {position.endDate ? ` · 结束日期：${position.endDate}` : ''}
          </Text>
          <Text>汇报职位：{position.reportsToPositionCode ?? '未设置'}</Text>
        </SimpleStack>

        <DividerLine />

        <SimpleStack gap={space.xs}>
          <SectionTitle title='当前任职' />
          {currentAssignment ? (
            <AssignmentItem assignment={currentAssignment} highlight />
          ) : (
            <Text color={colors.licorice400}>暂无当前任职</Text>
          )}
        </SimpleStack>
      </SimpleStack>
    </Card>
  )
}

interface PositionAssignmentsPanelProps {
  assignments: PositionAssignmentRecord[]
  currentAssignment?: PositionAssignmentRecord | null
}

export const PositionAssignmentsPanel: React.FC<PositionAssignmentsPanelProps> = ({
  assignments,
  currentAssignment = null,
}) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <SimpleStack gap={space.m}>
      <SectionTitle title='任职历史' />
      {assignments.length === 0 ? (
        <Text color={colors.licorice400}>暂无任职记录</Text>
      ) : (
        <SimpleStack gap={space.s}>
          {assignments.map(assignment => (
            <AssignmentItem
              key={assignment.assignmentId}
              assignment={assignment}
              highlight={currentAssignment?.assignmentId === assignment.assignmentId}
            />
          ))}
        </SimpleStack>
      )}
    </SimpleStack>
  </Card>
)

interface PositionTransfersPanelProps {
  transfers: PositionTransferRecord[]
}

export const PositionTransfersPanel: React.FC<PositionTransfersPanelProps> = ({ transfers }) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <SimpleStack gap={space.m}>
      <SectionTitle title='调动记录' />
      {transfers.length === 0 ? (
        <Text color={colors.licorice400}>暂无调动记录</Text>
      ) : (
        transfers.map(transfer => <TransferItem key={transfer.transferId} transfer={transfer} />)
      )}
    </SimpleStack>
  </Card>
)

interface PositionTimelinePanelProps {
  timeline: PositionTimelineEvent[]
}

export const PositionTimelinePanel: React.FC<PositionTimelinePanelProps> = ({ timeline }) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <SimpleStack gap={space.m}>
      <SectionTitle title='时间线事件' />
      {timeline.length === 0 ? (
        <Text color={colors.licorice400}>暂无时间线事件</Text>
      ) : (
        timeline.map(event => <TimelineItem key={event.id} event={event} />)
      )}
    </SimpleStack>
  </Card>
)

interface PositionDetailsProps {
  position?: PositionRecord
  timeline: PositionTimelineEvent[]
  assignments?: PositionAssignmentRecord[]
  currentAssignment?: PositionAssignmentRecord | null
  transfers?: PositionTransferRecord[]
  isLoading?: boolean
}

export const PositionDetails: React.FC<PositionDetailsProps> = ({
  position,
  timeline,
  assignments = [],
  currentAssignment = null,
  transfers = [],
  isLoading = false,
}) => (
  <SimpleStack gap={space.m}>
    <PositionOverviewCard position={position} currentAssignment={currentAssignment} isLoading={isLoading} />
    <PositionAssignmentsPanel assignments={assignments} currentAssignment={currentAssignment} />
    <PositionTransfersPanel transfers={transfers} />
    <PositionTimelinePanel timeline={timeline} />
  </SimpleStack>
)

const DividerLine: React.FC = () => <Box borderBottom={`1px solid ${colors.soap400}`} marginY={space.s} />

export default PositionDetails
