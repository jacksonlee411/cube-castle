import React, { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { useAuth } from '@/shared/auth/hooks'
import { useJobFamilyGroups } from '@/shared/hooks/useJobCatalog'
import { useCreateJobFamilyGroupVersion, useUpdateJobFamilyGroup } from '@/shared/hooks/useJobCatalogMutations'
import { StatusBadge } from '../shared/StatusBadge'
import { CatalogVersionForm, type CatalogVersionFormValues } from '../shared/CatalogVersionForm'
import { formatISODate, getCatalogStatusMeta } from '../types'

export const JobFamilyGroupDetail: React.FC = () => {
  const params = useParams<{ code: string }>()
  const code = params.code ?? ''
  const { hasPermission } = useAuth()
  const [isVersionFormOpen, setVersionFormOpen] = useState(false)
  const [isEditFormOpen, setEditFormOpen] = useState(false)
  const {
    data: groups = [],
    isLoading,
    isError,
  } = useJobFamilyGroups({ includeInactive: true })

  const group = useMemo(() => groups.find(item => item.code === code), [code, groups])

  const versionMutation = useCreateJobFamilyGroupVersion()
  const updateMutation = useUpdateJobFamilyGroup()

  if (!code) {
    return (
      <Box padding="l">
        <Heading size="medium">未提供职类编码</Heading>
      </Box>
    )
  }

  if (isLoading) {
    return (
      <Box padding="l">
        <Heading size="medium">加载中...</Heading>
      </Box>
    )
  }

  if (isError || !group) {
    return (
      <Box padding="l">
        <Heading size="medium">未找到职类 {code}</Heading>
        <Text marginTop="s">请确认编码是否正确或稍后再试。</Text>
      </Box>
    )
  }

  const statusMeta = getCatalogStatusMeta(group.status)

  const handleCreateVersion = async (values: CatalogVersionFormValues) => {
    await versionMutation.mutateAsync({
      code: group.code,
      ...values,
    })
    setVersionFormOpen(false)
  }

  const handleUpdate = async (values: CatalogVersionFormValues) => {
    await updateMutation.mutateAsync({
      code: group.code,
      recordId: group.recordId,
      ...values,
    })
    setEditFormOpen(false)
  }

  return (
    <Box padding="l" display="flex" flexDirection="column" gap="l">
      <Flex justifyContent="space-between" alignItems="center">
        <Heading size="large">职类详情</Heading>
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
        <Flex gap="m">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              职类编码
            </Text>
            <Text fontSize="18px" fontWeight={600}>
              {group.code}
            </Text>
          </Box>
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              状态
            </Text>
            <StatusBadge status={group.status} />
          </Box>
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            职类名称
          </Text>
          <Text fontSize="18px" fontWeight={600}>
            {group.name}
          </Text>
        </Box>

        <Flex gap="l">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              生效日期
            </Text>
            <Text>{formatISODate(group.effectiveDate)}</Text>
          </Box>
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              结束日期
            </Text>
            <Text>{formatISODate(group.endDate)}</Text>
          </Box>
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            描述
          </Text>
          <Text>{group.description ?? '暂无描述'}</Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            记录标识
          </Text>
          <Text fontFamily="monospace" fontSize="12px">
            {group.recordId}
          </Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            当前状态说明
          </Text>
          <Text>
            {statusMeta.label} · 自 {formatISODate(group.effectiveDate)} 起生效
          </Text>
        </Box>
      </Box>

      <CatalogVersionForm
        title="编辑职类信息"
        isOpen={isEditFormOpen}
        onClose={() => setEditFormOpen(false)}
        onSubmit={handleUpdate}
        isSubmitting={updateMutation.isPending}
        initialName={group.name}
        initialDescription={group.description}
        initialStatus={group.status}
        initialEffectiveDate={group.effectiveDate}
        submitLabel="保存更新"
      />

      <CatalogVersionForm
        title="新增职类版本"
        isOpen={isVersionFormOpen}
        onClose={() => setVersionFormOpen(false)}
        onSubmit={handleCreateVersion}
        isSubmitting={versionMutation.isPending}
        initialName={group.name}
        initialDescription={group.description}
      />
    </Box>
  )
}
