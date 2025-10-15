import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type { PositionRecord, PositionTimelineEvent } from '@/shared/types/positions'
import { getPositionStatusMeta } from '../statusMeta'
import { SimpleStack } from './SimpleStack'

interface PositionDetailsProps {
  position?: PositionRecord
  timeline: PositionTimelineEvent[]
  isLoading?: boolean
  dataSource?: 'api' | 'mock'
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
  isLoading = false,
  dataSource = 'api',
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
          <Heading level="3">{position.title}</Heading>
          <StatusPill status={position.status} />
        </Flex>
        <Text fontSize="14px" color={colors.licorice500}>
          职位编码：{position.code}
        </Text>
        <Text fontSize="14px" color={colors.licorice500}>
          所属组织：{position.organizationName ?? `${position.organizationCode}（未提供名称）`}
        </Text>
        {dataSource === 'mock' && (
          <Text fontSize="12px" color={colors.licorice300}>
            当前展示的是演示数据，后端接口不可用时自动回退。
          </Text>
        )}

        <DividerLine />

        <SimpleStack gap={space.xs}>
          <Heading level="4">岗位信息</Heading>
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

        <SimpleStack gap={space.s}>
          <Heading level="4">职位时间线</Heading>
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
