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
import { mockPositions } from './mockData'
import { PositionForm } from './components/PositionForm'
import { PositionVersionList, PositionVersionToolbar, buildVersionsCsv } from './components/versioning'
import { logger } from '@/shared/utils/logger'

const POSITION_CODE_PATTERN = /^P\d{7}$/

const mapLifecycleStatus = (type: string): string => {
	switch (type) {
		case 'CREATE':
			return 'PLANNED'
		case 'FILL':
			return 'FILLED'
		case 'VACATE':
			return 'VACANT'
		case 'SUSPEND':
			return 'INACTIVE'
		case 'REACTIVATE':
			return 'ACTIVE'
		case 'TRANSFER':
			return 'ACTIVE'
		default:
			return type.toUpperCase()
	}
}

const normalizeMockPosition = (code: string) => {
  const record = mockPositions.find(item => item.code === code)
  if (!record) {
    return { position: undefined, timeline: [] as PositionTimelineEvent[], versions: [] as PositionRecord[] }
  }

  const position: PositionRecord = {
    code: record.code,
    recordId: `${record.code}-current`,
    title: record.title,
    jobFamilyGroupCode: record.jobFamilyGroup,
    jobFamilyGroupName: record.jobFamilyGroup,
    jobFamilyCode: record.jobFamily,
    jobFamilyName: record.jobFamily,
    jobRoleCode: record.jobRole,
    jobRoleName: record.jobRole,
    jobLevelCode: record.jobLevel,
    jobLevelName: record.jobLevel,
    organizationCode: record.organization.code,
    organizationName: record.organization.name,
    positionType: 'REGULAR',
    employmentType: 'FULL_TIME',
    headcountCapacity: record.headcountCapacity,
    headcountInUse: record.headcountInUse,
    availableHeadcount: Math.max(record.headcountCapacity - record.headcountInUse, 0),
    status: record.status,
    effectiveDate: record.effectiveDate,
    endDate: undefined,
    isCurrent: record.status !== 'PLANNED',
    isFuture: record.status === 'PLANNED',
    createdAt: `${record.effectiveDate}T00:00:00.000Z`,
    updatedAt: `${record.effectiveDate}T00:00:00.000Z`,
    reportsToPositionCode: record.supervisor.code,
  } as PositionRecord

  const timeline: PositionTimelineEvent[] = record.lifecycle.map(event => ({
    id: event.id,
    status: mapLifecycleStatus(event.type),
    title: event.label,
    effectiveDate: event.occurredAt,
    endDate: undefined,
    changeReason: event.summary,
  }))

  const versions: PositionRecord[] = [
    { ...position },
    ...record.lifecycle
      .slice()
      .sort((a, b) => (a.occurredAt < b.occurredAt ? 1 : a.occurredAt > b.occurredAt ? -1 : 0))
      .map((event, index) => ({
        ...position,
        status: mapLifecycleStatus(event.type),
        effectiveDate: event.occurredAt,
        endDate: undefined,
        isCurrent: index === 0 ? position.isCurrent : false,
        isFuture: false,
        createdAt: `${'{'}event.occurredAt{'}'}T00:00:00.000Z`,
        updatedAt: `${'{'}event.occurredAt{'}'}T00:00:00.000Z`,
      })),
  ]

  return { position, timeline, versions }
}

const EmptyStateCard: React.FC<{ message: string }> = ({ message }) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <Text color={colors.licorice400}>{message}</Text>
  </Card>
)

export const PositionTemporalPage: React.FC = () => {
  const { code: rawCode } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const isMockMode = import.meta.env.VITE_POSITIONS_MOCK_MODE !== 'false'
  const [activeForm, setActiveForm] = useState<'none' | 'edit' | 'version'>('none')
  const [includeDeleted, setIncludeDeleted] = useState(false)

  const code = rawCode ? rawCode.toUpperCase() : ''
  const isCreateMode = code === 'NEW'
  const isValidCode = rawCode ? (isCreateMode || POSITION_CODE_PATTERN.test(code)) : false

  const detailQuery = usePositionDetail(isValidCode && !isCreateMode ? code : undefined, {
    enabled: !isMockMode && isValidCode && !isCreateMode,
    includeDeleted,
  })

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

    if (isMockMode) {
      const mock = normalizeMockPosition(code)
      return {
        position: mock.position,
        timeline: mock.timeline,
        assignments: [] as PositionAssignmentRecord[],
        currentAssignment: null as PositionAssignmentRecord | null,
        transfers: [] as PositionTransferRecord[],
        versions: mock.versions,
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
  }, [code, detailQuery.data, isCreateMode, isMockMode, isValidCode])

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
    if (!isMockMode) {
      detailQuery.refetch()
    }
  }

  const canMutate = !isMockMode && Boolean(position)

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
              数据来源：{isMockMode ? '演示数据（Mock 模式）' : 'GraphQL / REST 实时数据'}
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

        <PositionDetails
          position={position}
          timeline={timeline}
          assignments={assignments}
          currentAssignment={currentAssignment}
          transfers={transfers}
          isLoading={detailQuery.isLoading}
          dataSource={isMockMode ? 'mock' : 'api'}
        />

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

        {!isCreateMode && (
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
