import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { Card } from '@workday/canvas-kit-react/card'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import { TimelineComponent, type TimelineVersion } from '@/features/temporal/components'
import { AuditHistorySection } from '@/features/audit/components/AuditHistorySection'
import {
  PositionAssignmentsPanel,
  PositionOverviewCard,
  PositionTimelinePanel,
  PositionTransfersPanel,
} from './components/PositionDetails'
import type {
  PositionAssignmentRecord,
  PositionRecord,
  PositionTimelineEvent,
  PositionTransferRecord,
} from '@/shared/types/positions'
import { SimpleStack } from './components/SimpleStack'
import { PositionForm } from './components/PositionForm'
import { PositionVersionList, PositionVersionToolbar, buildVersionsCsv } from './components/versioning'
import {
  buildPositionVersionKey,
  createTimelineVersion,
  sortPositionVersions,
} from './timelineAdapter'
import { usePositionDetail } from '@/shared/hooks/useEnterprisePositions'
import { logger } from '@/shared/utils/logger'

const POSITION_CODE_PATTERN = /^P\d{7}$/

const EmptyStateCard: React.FC<{ message: string }> = ({ message }) => (
  <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
    <Text color={colors.licorice400}>{message}</Text>
  </Card>
)

type DetailTab = 'overview' | 'assignments' | 'transfers' | 'timeline' | 'versions' | 'audit'

const DETAIL_TABS: Array<{ key: DetailTab; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'assignments', label: '任职记录' },
  { key: 'transfers', label: '调动记录' },
  { key: 'timeline', label: '时间线' },
  { key: 'versions', label: '版本历史' },
  { key: 'audit', label: '审计历史' },
]

interface VersionEntry {
  key: string
  version: PositionRecord
  timeline: TimelineVersion
}

type DetailQueryResult = ReturnType<typeof usePositionDetail>

export const PositionTemporalPage: React.FC = () => {
  const { code: rawCode } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const [activeForm, setActiveForm] = useState<'none' | 'edit' | 'version'>('none')
  const [includeDeleted, setIncludeDeleted] = useState(false)
  const [activeTab, setActiveTab] = useState<DetailTab>('overview')
  const [selectedVersionKey, setSelectedVersionKey] = useState<string | null>(null)
  const [isCompactLayout, setIsCompactLayout] = useState(false)
  const [isVersionDrawerOpen, setIsVersionDrawerOpen] = useState(false)
  const isMockMode = import.meta.env.VITE_POSITIONS_MOCK_MODE !== 'false'

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
        position: undefined as PositionRecord | undefined,
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

  const versionEntries: VersionEntry[] = useMemo(() => {
    const sorted = sortPositionVersions(versions)
    return sorted.map((version, index) => ({
      version,
      key: buildPositionVersionKey(version, index),
      timeline: createTimelineVersion(version, index),
    }))
  }, [versions])

  const timelineVersions = useMemo(() => versionEntries.map(entry => entry.timeline), [versionEntries])
  const versionKeys = useMemo(() => versionEntries.map(entry => entry.key), [versionEntries])

  const selectedVersion = useMemo(() => {
    if (versionEntries.length === 0) {
      return null
    }

    if (!selectedVersionKey) {
      return versionEntries[0].version
    }

    return versionEntries.find(entry => entry.key === selectedVersionKey)?.version ?? versionEntries[0].version
  }, [versionEntries, selectedVersionKey])

  const selectedTimelineVersion = useMemo(() => {
    if (versionEntries.length === 0) {
      return null
    }
    if (!selectedVersionKey) {
      return timelineVersions[0] ?? null
    }
    return timelineVersions.find(item => item.recordId === selectedVersionKey) ?? timelineVersions[0] ?? null
  }, [timelineVersions, versionEntries.length, selectedVersionKey])

  useEffect(() => {
    if (versionEntries.length === 0) {
      setSelectedVersionKey(null)
      return
    }

    if (!selectedVersionKey) {
      setSelectedVersionKey(versionEntries[0].key)
      return
    }

    if (!versionEntries.some(entry => entry.key === selectedVersionKey)) {
      setSelectedVersionKey(versionEntries[0].key)
    }
  }, [versionEntries, selectedVersionKey])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    const evaluateLayout = () => {
      setIsCompactLayout(window.innerWidth < 960)
    }
    evaluateLayout()
    window.addEventListener('resize', evaluateLayout)
    return () => window.removeEventListener('resize', evaluateLayout)
  }, [])

  useEffect(() => {
    if (!isCompactLayout) {
      setIsVersionDrawerOpen(false)
    }
  }, [isCompactLayout])

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

  const handleFormSuccess = () => {
    setActiveForm('none')
    detailQuery.refetch()
  }

  const handleVersionSelect = useCallback(
    (timelineVersion: TimelineVersion) => {
      setSelectedVersionKey(timelineVersion.recordId)
      if (isCompactLayout) {
        setIsVersionDrawerOpen(false)
      }
    },
    [isCompactLayout],
  )

  const handleVersionRowSelect = useCallback(
    (version: PositionRecord, key: string) => {
      setSelectedVersionKey(key)
      if (isCompactLayout) {
        setIsVersionDrawerOpen(false)
      }
      if (activeTab !== 'overview') {
        setActiveTab('overview')
      }
    },
    [activeTab, isCompactLayout],
  )

  const overviewRecord = selectedVersion ?? position
  const canMutate = Boolean(position) && !detailQuery.isError && !isMockMode

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
          {isMockMode ? (
            <Card
              padding={space.l}
              backgroundColor={colors.cinnamon100}
              data-testid="position-mock-banner"
              style={{ borderLeft: `4px solid ${colors.cinnamon600}` }}
            >
              <SimpleStack gap={space.s}>
                <Text color={colors.cinnamon600} fontWeight="bold">
                  ⚠️ Mock 模式下无法创建职位。
                </Text>
                <Text fontSize="12px" color={colors.cinnamon600}>
                  请将环境变量 `VITE_POSITIONS_MOCK_MODE=false` 并启动后端服务后再尝试创建职位。
                </Text>
              </SimpleStack>
            </Card>
          ) : (
            <PositionForm
              mode="create"
              onCancel={handleBack}
              onSuccess={({ code: createdCode }) => navigate(`/positions/${createdCode}`)}
            />
          )}
        </SimpleStack>
      </Box>
    )
  }

  return (
    <Box padding={space.l} data-testid="position-temporal-page">
      <SimpleStack gap={space.l}>
        {isMockMode && (
          <Card
            padding={space.m}
            backgroundColor={colors.cinnamon100}
            data-testid="position-mock-banner"
            style={{ borderLeft: `4px solid ${colors.cinnamon600}` }}
          >
            <SimpleStack gap={space.xs}>
              <Text color={colors.cinnamon600} fontWeight="bold">
                ⚠️ 当前处于 Mock 模式，仅支持浏览职位数据。
              </Text>
              <Text fontSize="12px" color={colors.cinnamon600}>
                编辑与版本操作已禁用。请设置 `VITE_POSITIONS_MOCK_MODE=false` 并确保后端服务正常后再进行写入操作。
              </Text>
            </SimpleStack>
          </Card>
        )}

        <Flex justifyContent="space-between" alignItems="center">
          <Flex alignItems="center" gap={space.s}>
            <SecondaryButton onClick={handleBack} size="small">
              ← 返回职位列表
            </SecondaryButton>
            <Heading size="small">{isCreateMode ? '创建新职位' : `职位详情：${code}`}</Heading>
          </Flex>
          <Flex alignItems="center" gap={space.s}>
            <Text fontSize="12px" color={colors.licorice400}>
              数据来源：{isMockMode ? '演示环境（只读）' : 'GraphQL / REST 实时数据'}
            </Text>
            {canMutate && (
              <>
                <PrimaryButton
                  size="small"
                  variant={activeForm === 'edit' ? 'inverse' : undefined}
                  onClick={() => setActiveForm(prev => (prev === 'edit' ? 'none' : 'edit'))}
                  data-testid="position-edit-button"
                >
                  {activeForm === 'edit' ? '收起编辑' : '编辑职位'}
                </PrimaryButton>
                <SecondaryButton
                  size="small"
                  variant={activeForm === 'version' ? 'inverse' : undefined}
                  onClick={() => setActiveForm(prev => (prev === 'version' ? 'none' : 'version'))}
                  data-testid="position-version-button"
                >
                  {activeForm === 'version' ? '收起版本表单' : '新增时态版本'}
                </SecondaryButton>
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
          <Flex
            gap={space.l}
            alignItems="flex-start"
            flexWrap={isCompactLayout ? 'wrap' : 'nowrap'}
            data-testid="position-detail-layout"
          >
            {versionEntries.length > 0 && (
              <Box
                flex={isCompactLayout ? '1 1 100%' : '0 0 320px'}
                maxWidth={isCompactLayout ? '100%' : '360px'}
                width={isCompactLayout ? '100%' : '320px'}
              >
                {isCompactLayout ? (
                  <SimpleStack gap={space.s}>
                    <Flex justifyContent="space-between" alignItems="center">
                      <Heading size="small">版本导航</Heading>
                      <SecondaryButton size="small" onClick={() => setIsVersionDrawerOpen(prev => !prev)}>
                        {isVersionDrawerOpen ? '收起版本列表' : '选择其他版本'}
                      </SecondaryButton>
                    </Flex>
                    {isVersionDrawerOpen && (
                      <Card padding={space.m} backgroundColor={colors.frenchVanilla100}>
                        <TimelineComponent
                          versions={timelineVersions}
                          selectedVersion={selectedTimelineVersion}
                          onVersionSelect={handleVersionSelect}
                          isLoading={detailQuery.isLoading && timelineVersions.length === 0}
                          readonly={!canMutate}
                          height="auto"
                        />
                      </Card>
                    )}
                  </SimpleStack>
                ) : (
                  <Card padding={space.m} backgroundColor={colors.frenchVanilla100}>
                    <TimelineComponent
                      versions={timelineVersions}
                      selectedVersion={selectedTimelineVersion}
                      onVersionSelect={handleVersionSelect}
                      isLoading={detailQuery.isLoading && timelineVersions.length === 0}
                      readonly={!canMutate}
                      height="calc(100vh - 220px)"
                      width="100%"
                    />
                  </Card>
                )}
              </Box>
            )}

            <Box flex="1" minWidth={0}>
              <SimpleStack gap={space.l}>
                <TabsNavigation activeTab={activeTab} onTabChange={setActiveTab} />

                {selectedVersion && (
                  <Card padding={space.m} backgroundColor={colors.soap100}>
                    <Text fontSize="14px" color={colors.licorice500}>
                      当前版本：{selectedVersion.effectiveDate}（状态：{selectedVersion.status}）
                    </Text>
                  </Card>
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

                {renderTabContent({
                  activeTab,
                  overviewRecord,
                  currentAssignment,
                  assignments,
                  transfers,
                  timeline,
                  detailQuery,
                  includeDeleted,
                  onIncludeDeletedChange: setIncludeDeleted,
                  onExportVersions: handleExportVersions,
                  versionsForList: versionEntries.map(entry => entry.version),
                  versionKeys,
                  selectedVersionKey,
                  onVersionSelect: handleVersionRowSelect,
                  selectedVersion,
                })}
              </SimpleStack>
            </Box>
          </Flex>
        )}
      </SimpleStack>
    </Box>
  )
}

export default PositionTemporalPage

const TabsNavigation: React.FC<{ activeTab: DetailTab; onTabChange: (tab: DetailTab) => void }> = ({
  activeTab,
  onTabChange,
}) => (
  <Flex borderBottom={`2px solid ${colors.soap300}`}>
    {DETAIL_TABS.map(tab => {
      const isActive = tab.key === activeTab
      return (
        <Box
          key={tab.key}
          padding={`${space.s} ${space.l}`}
          marginBottom="-2px"
          style={{
            cursor: 'pointer',
            borderBottom: isActive ? `3px solid ${colors.blueberry600}` : '3px solid transparent',
            transition: 'all 0.2s ease-in-out',
          }}
          onClick={() => onTabChange(tab.key)}
        >
          <Text
            typeLevel="body.medium"
            fontWeight={isActive ? 'medium' : 'regular'}
            color={isActive ? colors.blueberry600 : colors.licorice600}
          >
            {tab.label}
          </Text>
        </Box>
      )
    })}
  </Flex>
)

interface TabContentProps {
  activeTab: DetailTab
  overviewRecord: PositionRecord | null | undefined
  currentAssignment: PositionAssignmentRecord | null
  assignments: PositionAssignmentRecord[]
  transfers: PositionTransferRecord[]
  timeline: PositionTimelineEvent[]
  detailQuery: DetailQueryResult
  includeDeleted: boolean
  onIncludeDeletedChange: (checked: boolean) => void
  onExportVersions: () => void
  versionsForList: PositionRecord[]
  versionKeys: string[]
  selectedVersionKey: string | null
  onVersionSelect: (version: PositionRecord, key: string) => void
  selectedVersion: PositionRecord | null
}

const renderTabContent = ({
  activeTab,
  overviewRecord,
  currentAssignment,
  assignments,
  transfers,
  timeline,
  detailQuery,
  includeDeleted,
  onIncludeDeletedChange,
  onExportVersions,
  versionsForList,
  versionKeys,
  selectedVersionKey,
  onVersionSelect,
  selectedVersion,
}: TabContentProps) => {
  switch (activeTab) {
    case 'overview':
      return (
        <PositionOverviewCard
          position={overviewRecord ?? undefined}
          currentAssignment={currentAssignment}
          isLoading={detailQuery.isLoading}
        />
      )
    case 'assignments':
      return <PositionAssignmentsPanel assignments={assignments} currentAssignment={currentAssignment} />
    case 'transfers':
      return <PositionTransfersPanel transfers={transfers} />
    case 'timeline':
      return <PositionTimelinePanel timeline={timeline} />
    case 'versions':
      return (
        <SimpleStack gap={space.m}>
          <PositionVersionToolbar
            includeDeleted={includeDeleted}
            onIncludeDeletedChange={onIncludeDeletedChange}
            onExportCsv={onExportVersions}
            isBusy={detailQuery.isFetching}
            hasVersions={versionsForList.length > 0}
          />
          <PositionVersionList
            versions={versionsForList}
            isLoading={detailQuery.isLoading}
            versionKeys={versionKeys}
            selectedVersionKey={selectedVersionKey}
            onSelectVersion={onVersionSelect}
          />
        </SimpleStack>
      )
    case 'audit':
      if (!selectedVersion?.recordId) {
        return (
          <Card padding={space.l} backgroundColor={colors.frenchVanilla100}>
            <Text color={colors.licorice400}>
              当前版本缺少 recordId，无法加载审计历史。请选择其他版本或联系后端补齐审计链路。
            </Text>
          </Card>
        )
      }
      return (
        <AuditHistorySection recordId={selectedVersion.recordId} />
      )
    default:
      return null
  }
}
