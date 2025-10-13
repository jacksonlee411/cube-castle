import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type { PositionMock } from '../mockData'
import { SimpleStack } from './SimpleStack'

interface PositionSummaryCardsProps {
  positions: PositionMock[]
}

const formatNumber = (value: number) => value.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 1 })

export const PositionSummaryCards: React.FC<PositionSummaryCardsProps> = ({ positions }) => {
  const totalPositions = positions.length
  const totalCapacity = positions.reduce((acc, item) => acc + item.headcountCapacity, 0)
  const totalInUse = positions.reduce((acc, item) => acc + item.headcountInUse, 0)
  const totalAvailable = Math.max(totalCapacity - totalInUse, 0)
  const plannedCount = positions.filter(item => item.status === 'PLANNED').length
  const multiSeatPositions = positions.filter(item => item.headcountCapacity > 1).length

  const metrics = [
    {
      title: '岗位总数',
      value: totalPositions,
      description: `${multiSeatPositions} 个岗位支持一岗多人编制`,
      accent: colors.blueberry500,
    },
    {
      title: '编制容量（FTE）',
      value: formatNumber(totalCapacity),
      description: `当前占用 ${formatNumber(totalInUse)} · 可用 ${formatNumber(totalAvailable)}`,
      accent: colors.orange500,
    },
    {
      title: '规划职位',
      value: plannedCount,
      description: '等待预算批复或启用的岗位数量',
      accent: colors.cantaloupe600,
    },
  ]

  return (
    <Flex gap={space.l} flexWrap="wrap">
      {metrics.map(metric => (
        <Card
          key={metric.title}
          padding={space.l}
          width={320}
          backgroundColor={colors.frenchVanilla100}
          style={{ borderTop: `4px solid ${metric.accent}` }}
        >
          <SimpleStack gap={space.xs}>
            <Heading level="3">{metric.title}</Heading>
            <Text fontSize="32px" fontWeight="bold" color={metric.accent}>
              {metric.value}
            </Text>
            <Text as="p" color={colors.licorice500}>
              {metric.description}
            </Text>
          </SimpleStack>
        </Card>
      ))}
    </Flex>
  )
}

export default PositionSummaryCards
