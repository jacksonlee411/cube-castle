import React, { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Card } from '@workday/canvas-kit-react/card'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { Checkbox } from '@workday/canvas-kit-react/checkbox'
import { PositionDetails } from './components/PositionDetails'
import { PositionVersionList } from './components/PositionVersionList'
import { SimpleStack } from './components/SimpleStack'
import { PositionVersionDiff } from './components/PositionVersionDiff'
import { POSITION_VERSION_FIELDS } from './components/positionVersionFields'
import { usePositionDetail } from '@/shared/hooks/useEnterprisePositions'
import type {
  PositionAssignmentRecord,
  PositionRecord,
  PositionTimelineEvent,
  PositionTransferRecord,
} from '@/shared/types/positions'
import { mockPositions } from './mockData'
import { PositionForm } from './components/PositionForm'

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

const getVersionIdentifier = (version: PositionRecord): string =>
  version.recordId ?? `${version.code}-${version.effectiveDate}-${version.updatedAt}`

const formatVersionLabel = (version: PositionRecord): string => {
  const markers = [
    version.isCurrent ? '当前' : undefined,
    version.isFuture ? '计划' : undefined,
    version.status === 'DELETED' ? '已删除' : undefined,
  ].filter(Boolean)

  const markerText = markers.length ? ` · ${markers.join('/')}` : ''
  return `${version.effectiveDate}${markerText}`
}

const formatCsvValue = (value: unknown): string => {
  if (value === null || value === undefined) {
    return ''
  }
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value.toString() : ''
  }
  return String(value)
}

const buildVersionsCsv = (versions: PositionRecord[]): string => {
  const header = POSITION_VERSION_FIELDS.map(field => field.label)
  const rows = versions.map(version =>
    POSITION_VERSION_FIELDS.map(field => formatCsvValue((version as Record<string, unknown>)[field.key])),
  )

  return [header, ...rows]
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

const SELECT_STYLE: React.CSSProperties = {
  minWidth: '220px',
  padding: '8px 12px',
  borderRadius: 8,
  border: `1px solid ${colors.soap500}`,
  fontSize: '14px',
  backgroundColor: colors.frenchVanilla100,
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
  const [baseVersionId, setBaseVersionId] = useState<string | null>(null)
  const [compareVersionId, setCompareVersionId] = useState<string | null>(null)

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

  useEffect(() => {
    if (!versions.length) {
      if (baseVersionId !== null) {
        setBaseVersionId(null)
      }
      if (compareVersionId !== null) {
        setCompareVersionId(null)
      }
      return
    }

    const nextBaseId =
      baseVersionId && versions.some(version => getVersionIdentifier(version) === baseVersionId)
        ? baseVersionId
        : getVersionIdentifier(versions[0])

    const availableForCompare = versions.filter(
      version => getVersionIdentifier(version) !== nextBaseId,
    )

    const nextCompareId = availableForCompare.length
      ? (compareVersionId &&
        availableForCompare.some(version => getVersionIdentifier(version) === compareVersionId)
          ? compareVersionId
          : getVersionIdentifier(availableForCompare[0]))
      : null

    if (nextBaseId !== baseVersionId) {
      setBaseVersionId(nextBaseId)
    }
    if (nextCompareId !== compareVersionId) {
      setCompareVersionId(nextCompareId)
    }
  }, [baseVersionId, compareVersionId, versions])

  const versionOptions = useMemo(
    () =>
      versions.map(version => {
        const id = getVersionIdentifier(version)
        return {
          id,
          label: formatVersionLabel(version),
        }
      }),
    [versions],
  )

  const selectedBaseVersion = useMemo(
    () => versions.find(version => getVersionIdentifier(version) === baseVersionId) ?? null,
    [baseVersionId, versions],
  )

  const selectedCompareVersion = useMemo(
    () => versions.find(version => getVersionIdentifier(version) === compareVersionId) ?? null,
    [compareVersionId, versions],
  )

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

  const handleIncludeDeletedChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setIncludeDeleted(event.target.checked)
  }

  const handleExportVersions = () => {
    if (!versions.length) {
      return
    }
    const csv = buildVersionsCsv(versions)
    const filename = `position-versions-${code}-${Date.now()}.csv`
    downloadCsv(csv, filename)
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
          <SimpleStack gap={space.l}>
            <Flex
              justifyContent="space-between"
              alignItems="flex-start"
              flexWrap="wrap"
              gap={space.s}
            >
              <Flex gap={space.s} alignItems="center">
                <Checkbox
                  label="显示已删除版本"
                  checked={includeDeleted}
                  onChange={handleIncludeDeletedChange}
                  data-testid="position-include-deleted"
                />
                <SecondaryButton
                  type="button"
                  onClick={handleExportVersions}
                  disabled={!versions.length}
                  data-testid="position-versions-export"
                >
                  导出版本 CSV
                </SecondaryButton>
              </Flex>

              <Flex gap={space.m} flexWrap="wrap">
                <Box minWidth="220px">
                  <SimpleStack gap={space.xxxs}>
                    <label
                      htmlFor="position-base-version-select"
                      style={{ fontSize: '12px', color: colors.licorice500 }}
                    >
                      基准版本
                    </label>
                    <select
                      id="position-base-version-select"
                      value={baseVersionId ?? ''}
                      onChange={event => setBaseVersionId(event.target.value || null)}
                      disabled={!versions.length}
                      data-testid="position-base-version-select"
                      style={{ ...SELECT_STYLE, width: '100%' }}
                    >
                      {versionOptions.map(option => (
                        <option key={option.id} value={option.id}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </SimpleStack>
                </Box>

                <Box minWidth="220px">
                  <SimpleStack gap={space.xxxs}>
                    <label
                      htmlFor="position-compare-version-select"
                      style={{ fontSize: '12px', color: colors.licorice500 }}
                    >
                      对比版本
                    </label>
                    <select
                      id="position-compare-version-select"
                      value={compareVersionId ?? ''}
                      onChange={event => setCompareVersionId(event.target.value || null)}
                      disabled={versionOptions.length <= 1}
                      data-testid="position-compare-version-select"
                      style={{ ...SELECT_STYLE, width: '100%' }}
                    >
                      <option value="">（无）</option>
                      {versionOptions
                        .filter(option => option.id !== baseVersionId)
                        .map(option => (
                          <option key={option.id} value={option.id}>
                            {option.label}
                          </option>
                        ))}
                    </select>
                  </SimpleStack>
                </Box>
              </Flex>
            </Flex>

            <PositionVersionList versions={versions} isLoading={detailQuery.isLoading} />

            <PositionVersionDiff
              baseVersion={selectedBaseVersion}
              compareVersion={selectedCompareVersion}
              isLoading={detailQuery.isLoading}
            />
          </SimpleStack>
        )}
      </SimpleStack>
    </Box>
  )
}

export default PositionTemporalPage
