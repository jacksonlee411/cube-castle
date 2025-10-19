import React, { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Card } from '@workday/canvas-kit-react/card'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { PositionDetails } from './components/PositionDetails'
import { SimpleStack } from './components/SimpleStack'
import { usePositionDetail } from '@/shared/hooks/useEnterprisePositions'
import type {
  PositionAssignmentRecord,
  PositionRecord,
  PositionTimelineEvent,
  PositionTransferRecord,
} from '@/shared/types/positions'
import { PositionForm } from './components/PositionForm'
import { PositionVersionList, PositionVersionToolbar, buildVersionsCsv } from './components/versioning'
import { logger } from '@/shared/utils/logger'

const POSITION_CODE_PATTERN = /^P\d{7}$/

const EmptyStateCard: React.FC<{ message: string }> = ({ message }) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <Text color={colors.licorice400}>{message}</Text>
  </Card>
)

export const PositionTemporalPage: React.FC = () => {
  const { code: rawCode } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const [activeForm, setActiveForm] = useState<'none' | 'edit' | 'version'>('none')
  const [includeDeleted, setIncludeDeleted] = useState(false)

  const code = rawCode ? rawCode.toUpperCase() : ''
  const isCreateMode = code === 'NEW'
  const isValidCode = rawCode ? (isCreateMode || POSITION_CODE_PATTERN.test(code)) : false

  const detailQuery = usePositionDetail(isValidCode && !isCreateMode ? code : undefined, {
    enabled: isValidCode && !isCreateMode,
    includeDeleted,
  })

  const detailErrorMessage = detailQuery.error instanceof Error ? detailQuery.error.message : undefined

  const { position, timeline, assignments, currentAssignment, transfers, versions } = useMemo(() => {
    if (isCreateMode || !isValidCode) {
      return {
        position: undefined,
        timeline: [] as PositionTimelineEvent[],
        assignments: [] as PositionAssignmentRecord[],
        currentAssignment: null as PositionAssignmentRecord | null,
        transfers: [] as PositionTransferRecord[],
        versions: [] as PositionRecord[],
      }
    }

    const graph = detailQuery.data
    return {
      position: graph?.position ?? undefined,
      timeline: graph?.timeline ?? [],
      assignments: graph?.assignments ?? [],
      currentAssignment: graph?.currentAssignment ?? null,
      transfers: graph?.transfers ?? [],
      versions: graph?.versions ?? [],
    }
  }, [detailQuery.data, isCreateMode, isValidCode])

  const handleExportVersions = useCallback(() => {
    if (versions.length === 0 || typeof window === 'undefined') {
      return
    }

    try {
      const csv = buildVersionsCsv(versions)
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
      const url = URL.createObjectURL(blob)

      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `${code || 'position'}-versions.csv`
      anchor.style.display = 'none'
      document.body.appendChild(anchor)
      anchor.click()
      document.body.removeChild(anchor)

      URL.revokeObjectURL(url)
    } catch (error) {
      logger.error('[PositionTemporalPage] 导出职位版本失败', error)
    }
  }, [versions, code])

  const handleBack = () => {
    navigate('/positions')
  }

  if (!rawCode) {
    return (
      <Box padding={space.xl}>
        <EmptyStateCard message="未提供职位编码，请从职位列表进入详情页。" />
      </Box>
    )
  }

  if (!isCreateMode && !isValidCode) {
    return (
      <Box padding={space.xl}>
        <EmptyStateCard message="职位编码格式不正确，请从职位列表页面重新进入。" />
      </Box>
    )
  }

  if (isCreateMode) {
    return (
      <Box padding={space.l} data-testid="position-create-page">
        <SimpleStack gap={space.l}>
          <Flex justifyContent="space-between" alignItems="center">
            <Flex alignItems="center" gap={space.s}>
              <SecondaryButton onClick={handleBack} size="small">
                ← 返回职位列表
              </SecondaryButton>
              <Heading size="small">创建职位</Heading>
            </Flex>
          </Flex>
          <PositionForm
            mode="create"
            onCancel={handleBack}
            onSuccess={({ code: createdCode }) => navigate(`/positions/${createdCode}`)}
          />
        </SimpleStack>
      </Box>
    )
  }

  const handleFormSuccess = () => {
    setActiveForm('none')
    detailQuery.refetch()
  }

  const canMutate = Boolean(position) && !detailQuery.isError

  return (
    <Box padding={space.l} data-testid="position-temporal-page">
      <SimpleStack gap={space.l}>
        <Flex justifyContent="space-between" alignItems="center">
          <Flex alignItems="center" gap={space.s}>
            <SecondaryButton onClick={handleBack} size="small">
              ← 返回职位列表
            </SecondaryButton>
            <Heading size="small">{isCreateMode ? '创建新职位' : `职位详情：${code}`}</Heading>
          </Flex>
          <Flex alignItems="center" gap={space.s}>
            <Text fontSize="12px" color={colors.licorice400}>
              数据来源：GraphQL / REST 实时数据
            </Text>
            {canMutate && (
              <>
                <PrimaryButton
                  size="small"
                  variant={activeForm === 'edit' ? 'inverse' : 'primary'}
                  onClick={() => setActiveForm(prev => (prev === 'edit' ? 'none' : 'edit'))}
                  data-testid="position-edit-button"
                >
                  {activeForm === 'edit' ? '收起编辑' : '编辑职位'}
                </PrimaryButton>
                <PrimaryButton
                  size="small"
                  variant={activeForm === 'version' ? 'inverse' : 'secondary'}
                  onClick={() => setActiveForm(prev => (prev === 'version' ? 'none' : 'version'))}
                  data-testid="position-version-button"
                >
                  {activeForm === 'version' ? '收起版本表单' : '新增时态版本'}
                </PrimaryButton>
              </>
            )}
          </Flex>
        </Flex>

        {detailQuery.isError && (
          <Card padding={space.l} backgroundColor={colors.frenchVanilla100} data-testid="position-detail-error">
            <SimpleStack gap={space.xs}>
              <Text color={colors.cinnamon500}>加载职位详情失败，请稍后重试。</Text>
              {detailErrorMessage && (
                <Text fontSize="12px" color={colors.licorice400}>
                  错误详情：{detailErrorMessage}
                </Text>
              )}
              <Flex>
                <PrimaryButton
                  size="small"
                  onClick={() => detailQuery.refetch()}
                  disabled={detailQuery.isFetching}
                >
                  {detailQuery.isFetching ? '正在重新加载...' : '重新加载'}
                </PrimaryButton>
              </Flex>
            </SimpleStack>
          </Card>
        )}

        {!detailQuery.isError && (
          <PositionDetails
            position={position}
            timeline={timeline}
            assignments={assignments}
            currentAssignment={currentAssignment}
            transfers={transfers}
            isLoading={detailQuery.isLoading}
          />
        )}

        {activeForm === 'edit' && position && (
          <PositionForm
            mode="edit"
            position={position}
            onCancel={() => setActiveForm('none')}
            onSuccess={handleFormSuccess}
          />
        )}

        {activeForm === 'version' && position && (
          <PositionForm
            mode="version"
            position={position}
            onCancel={() => setActiveForm('none')}
            onSuccess={handleFormSuccess}
          />
        )}

        {!isCreateMode && !detailQuery.isError && (
          <SimpleStack gap={space.m}>
            <PositionVersionToolbar
              includeDeleted={includeDeleted}
              onIncludeDeletedChange={checked => setIncludeDeleted(checked)}
              onExportCsv={handleExportVersions}
              isBusy={detailQuery.isFetching}
              hasVersions={versions.length > 0}
            />
            <PositionVersionList versions={versions} isLoading={detailQuery.isLoading} />
          </SimpleStack>
        )}
      </SimpleStack>
    </Box>
  )
}

export default PositionTemporalPage
