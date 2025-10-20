import React, { useMemo, useState } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { useAuth } from '@/shared/auth/hooks'
import { useJobLevels } from '@/shared/hooks/useJobCatalog'
import { useCreateJobLevelVersion, useUpdateJobLevel } from '@/shared/hooks/useJobCatalogMutations'
import { StatusBadge } from '../shared/StatusBadge'
import { CatalogVersionForm, type CatalogVersionFormValues } from '../shared/CatalogVersionForm'
import { formatISODate } from '../types'

interface LocationState {
  roleCode?: string
}

export const JobLevelDetail: React.FC = () => {
  const params = useParams<{ code: string }>()
  const location = useLocation()
  const state = (location.state ?? {}) as LocationState
  const roleCode = state.roleCode
  const code = params.code ?? ''
  const { hasPermission } = useAuth()
  const [isVersionFormOpen, setVersionFormOpen] = useState(false)
  const [isEditFormOpen, setEditFormOpen] = useState(false)

  const levelsQuery = useJobLevels(roleCode, { includeInactive: true })
  const versionMutation = useCreateJobLevelVersion()
  const updateMutation = useUpdateJobLevel()

  const level = useMemo(() => levelsQuery.data?.find(item => item.code === code), [code, levelsQuery.data])

  if (!code) {
    return (
      <Box padding="l">
        <Heading size="medium">未提供职级编码</Heading>
      </Box>
    )
  }

  if (!roleCode) {
    return (
      <Box padding="l">
        <Heading size="medium">缺少职务上下文</Heading>
        <Text marginTop="s">请从职级列表页面重新进入，以便加载完整上下游数据。</Text>
      </Box>
    )
  }

  if (levelsQuery.isLoading) {
    return (
      <Box padding="l">
        <Heading size="medium">加载中...</Heading>
      </Box>
    )
  }

  if (!level) {
    return (
      <Box padding="l">
        <Heading size="medium">未找到职级 {code}</Heading>
        <Text marginTop="s">请确认编码是否正确。</Text>
      </Box>
    )
  }

  const handleCreateVersion = async (values: CatalogVersionFormValues) => {
    await versionMutation.mutateAsync({ code: level.code, ...values })
    setVersionFormOpen(false)
  }

  const handleUpdate = async (values: CatalogVersionFormValues) => {
    await updateMutation.mutateAsync({
      code: level.code,
      recordId: level.recordId,
      jobRoleCode: level.roleCode,
      levelRank: level.levelRank,
      ...values,
    })
    setEditFormOpen(false)
  }

  return (
    <Box padding="l" display="flex" flexDirection="column" gap="l">
      <Flex justifyContent="space-between" alignItems="center">
        <Heading size="large">职级详情</Heading>
        {hasPermission('job-catalog:update') && (
          <Flex gap="s">
            <SecondaryButton onClick={() => setEditFormOpen(true)} disabled={updateMutation.isPending}>
              编辑当前版本
            </SecondaryButton>
            <PrimaryButton onClick={() => setVersionFormOpen(true)} disabled={versionMutation.isPending}>
              新增版本
            </PrimaryButton>
          </Flex>
        )}
      </Flex>

      <Box display="flex" flexDirection="column" gap="m">
        <Flex gap="m" alignItems="center">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              职级编码
            </Text>
            <Text fontSize="18px" fontWeight={600}>
              {level.code}
            </Text>
          </Box>
          <StatusBadge status={level.status} />
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            职级名称
          </Text>
          <Text fontSize="18px" fontWeight={600}>
            {level.name}
          </Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            归属职务
          </Text>
          <Text>{level.roleCode}</Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            等级序号
          </Text>
          <Text>{level.levelRank}</Text>
        </Box>

        <Flex gap="l">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              生效日期
            </Text>
            <Text>{formatISODate(level.effectiveDate)}</Text>
          </Box>
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              结束日期
            </Text>
            <Text>{formatISODate(level.endDate)}</Text>
          </Box>
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            描述
          </Text>
          <Text>{level.description ?? '暂无描述'}</Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            记录标识
          </Text>
          <Text fontFamily="monospace" fontSize="12px">
            {level.recordId}
          </Text>
        </Box>
      </Box>

      <CatalogVersionForm
        title="编辑职级信息"
        isOpen={isEditFormOpen}
        onClose={() => setEditFormOpen(false)}
        onSubmit={handleUpdate}
        isSubmitting={updateMutation.isPending}
        initialName={level.name}
        initialDescription={level.description}
        initialStatus={level.status}
        initialEffectiveDate={level.effectiveDate}
        submitLabel="保存更新"
      />

      <CatalogVersionForm
        title="新增职级版本"
        isOpen={isVersionFormOpen}
        onClose={() => setVersionFormOpen(false)}
        onSubmit={handleCreateVersion}
        isSubmitting={versionMutation.isPending}
        initialName={level.name}
        initialDescription={level.description}
      />
    </Box>
  )
}
