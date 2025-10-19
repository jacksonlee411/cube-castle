import React, { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { useAuth } from '@/shared/auth/hooks'
import { useJobRoles } from '@/shared/hooks/useJobCatalog'
import { useCreateJobRoleVersion, useUpdateJobRole } from '@/shared/hooks/useJobCatalogMutations'
import { StatusBadge } from '../shared/StatusBadge'
import { CatalogVersionForm, type CatalogVersionFormValues } from '../shared/CatalogVersionForm'
import { formatISODate } from '../types'

const deriveFamilyCode = (roleCode: string): string | undefined => {
  const segments = roleCode.split('-')
  if (segments.length < 3) {
    return undefined
  }
  return `${segments[0]}-${segments[1]}`
}

export const JobRoleDetail: React.FC = () => {
  const params = useParams<{ code: string }>()
  const code = params.code ?? ''
  const familyCode = deriveFamilyCode(code)
  const { hasPermission } = useAuth()
  const [isVersionFormOpen, setVersionFormOpen] = useState(false)
  const [isEditFormOpen, setEditFormOpen] = useState(false)

  const rolesQuery = useJobRoles(familyCode, { includeInactive: true })
  const versionMutation = useCreateJobRoleVersion()
  const updateMutation = useUpdateJobRole()

  const role = useMemo(() => rolesQuery.data?.find(item => item.code === code), [code, rolesQuery.data])

  if (!code) {
    return (
      <Box padding="l">
        <Heading size="medium">未提供职务编码</Heading>
      </Box>
    )
  }

  if (rolesQuery.isLoading) {
    return (
      <Box padding="l">
        <Heading size="medium">加载中...</Heading>
      </Box>
    )
  }

  if (!role) {
    return (
      <Box padding="l">
        <Heading size="medium">未找到职务 {code}</Heading>
        <Text marginTop="s">请确认编码是否正确。</Text>
      </Box>
    )
  }

  const handleCreateVersion = async (values: CatalogVersionFormValues) => {
    await versionMutation.mutateAsync({ code: role.code, ...values })
    setVersionFormOpen(false)
  }

  const handleUpdate = async (values: CatalogVersionFormValues) => {
    await updateMutation.mutateAsync({
      code: role.code,
      jobFamilyCode: role.familyCode,
      ...values,
    })
    setEditFormOpen(false)
  }

  return (
    <Box padding="l" display="flex" flexDirection="column" gap="l">
      <Flex justifyContent="space-between" alignItems="center">
        <Heading size="large">职务详情</Heading>
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
              职务编码
            </Text>
            <Text fontSize="18px" fontWeight={600}>
              {role.code}
            </Text>
          </Box>
          <StatusBadge status={role.status} />
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            职务名称
          </Text>
          <Text fontSize="18px" fontWeight={600}>
            {role.name}
          </Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            归属职种
          </Text>
          <Text>{role.familyCode}</Text>
        </Box>

        <Flex gap="l">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              生效日期
            </Text>
            <Text>{formatISODate(role.effectiveDate)}</Text>
          </Box>
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              结束日期
            </Text>
            <Text>{formatISODate(role.endDate)}</Text>
          </Box>
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            描述
          </Text>
          <Text>{role.description ?? '暂无描述'}</Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            记录标识
          </Text>
          <Text fontFamily="monospace" fontSize="12px">
            {role.recordId}
          </Text>
        </Box>
      </Box>

      <CatalogVersionForm
        title="编辑职务信息"
        isOpen={isEditFormOpen}
        onClose={() => setEditFormOpen(false)}
        onSubmit={handleUpdate}
        isSubmitting={updateMutation.isPending}
        initialName={role.name}
        initialDescription={role.description}
        initialStatus={role.status}
        initialEffectiveDate={role.effectiveDate}
        submitLabel="保存更新"
      />

      <CatalogVersionForm
        title="新增职务版本"
        isOpen={isVersionFormOpen}
        onClose={() => setVersionFormOpen(false)}
        onSubmit={handleCreateVersion}
        isSubmitting={versionMutation.isPending}
        initialName={role.name}
        initialDescription={role.description}
      />
    </Box>
  )
}
