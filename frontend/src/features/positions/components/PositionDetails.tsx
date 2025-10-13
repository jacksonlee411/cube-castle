import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { StatusBadge } from '../../../shared/components/StatusBadge'
import type { PositionMock } from '../mockData'
import type { PositionLifecycleEvent } from '../types'
import { SimpleStack } from './SimpleStack'

interface PositionDetailsProps {
  position?: PositionMock
}

const TimelineItem: React.FC<{ event: PositionLifecycleEvent }> = ({ event }) => (
  <Box borderLeft={`2px solid ${colors.blueberry400}`} paddingLeft={space.m} marginBottom={space.m}>
    <Text fontWeight="bold">{event.label}</Text>
    <Text as="p" marginY="xxs" color={colors.licorice500}>
      {event.summary}
    </Text>
    <Text fontSize="12px" color={colors.licorice400}>
      {event.occurredAt} · {event.operator}
    </Text>
  </Box>
)

export const PositionDetails: React.FC<PositionDetailsProps> = ({ position }) => {
  if (!position) {
    return (
      <Card data-testid="position-detail-card" padding={space.l} height="100%" backgroundColor={colors.frenchVanilla100}>
        <Text color={colors.licorice400}>请选择左侧职位查看详情</Text>
      </Card>
    )
  }

  const availableCount = Math.max(position.headcountCapacity - position.headcountInUse, 0)

  return (
    <Card data-testid="position-detail-card" padding={space.l} height="100%" backgroundColor={colors.frenchVanilla100}>
      <SimpleStack gap={space.m}>
        <Flex alignItems="center" justifyContent="space-between">
          <Heading level="3">{position.title}</Heading>
          <StatusBadge status={position.status} size="medium" />
        </Flex>
        <Text fontSize="14px" color={colors.licorice500}>
          {position.organization.name} · 汇报给 {position.supervisor.name}
        </Text>
        <DividerLine />
        <SimpleStack gap={space.xs}>
          <Heading level="4">岗位信息</Heading>
          <Text>
            职类 / 职种：{position.jobFamilyGroup} · {position.jobFamily}
          </Text>
          <Text>
            职务 / 职级：{position.jobRole} · {position.jobLevel}
          </Text>
          <Text>工作地点：{position.location}</Text>
          {position.shiftPattern && <Text>排班安排：{position.shiftPattern}</Text>}
          <Text>
            编制：{position.headcountInUse} / {position.headcountCapacity}（可用 {availableCount}）
          </Text>
          <Text>生效日期：{position.effectiveDate}</Text>
          {position.notes && (
            <Text as="p" color={colors.licorice500} marginTop={space.xs}>
              {position.notes}
            </Text>
          )}
        </SimpleStack>
        <DividerLine />
        <SimpleStack gap={space.s}>
          <Heading level="4">最近事件</Heading>
          {position.lifecycle.length === 0 ? (
            <Text color={colors.licorice400}>暂无事件记录</Text>
          ) : (
            position.lifecycle.slice(0, 4).map(item => <TimelineItem key={item.id} event={item} />)
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
