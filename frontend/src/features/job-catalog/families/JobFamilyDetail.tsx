import React, { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { PrimaryButton } from '@workday/canvas-kit-react/button'
import { useAuth } from '@/shared/auth/hooks'
import { useJobFamilies } from '@/shared/hooks/useJobCatalog'
import { useCreateJobFamilyVersion } from '@/shared/hooks/useJobCatalogMutations'
import { StatusBadge } from '../shared/StatusBadge'
import { CatalogVersionForm, type CatalogVersionFormValues } from '../shared/CatalogVersionForm'
import { formatISODate } from '../types'

const deriveGroupCode = (familyCode: string): string | undefined => {
  const [group] = familyCode.split('-')
  return group && group.length >= 4 ? group : undefined
}

export const JobFamilyDetail: React.FC = () => {
  const params = useParams<{ code: string }>()
  const code = params.code ?? ''
  const groupCode = deriveGroupCode(code)
  const { hasPermission } = useAuth()
  const [isVersionFormOpen, setVersionFormOpen] = useState(false)

  const familiesQuery = useJobFamilies(groupCode, { includeInactive: true })
  const versionMutation = useCreateJobFamilyVersion()

  const family = useMemo(() => familiesQuery.data?.find(item => item.code === code), [code, familiesQuery.data])

  if (!code) {
    return (
      <Box padding="l">
        <Heading size="medium">未提供职种编码</Heading>
      </Box>
    )
  }

  if (familiesQuery.isLoading) {
    return (
      <Box padding="l">
        <Heading size="medium">加载中...</Heading>
      </Box>
    )
  }

  if (!family) {
    return (
      <Box padding="l">
        <Heading size="medium">未找到职种 {code}</Heading>
        <Text marginTop="s">请确认编码是否正确。</Text>
      </Box>
    )
  }

  const handleCreateVersion = async (values: CatalogVersionFormValues) => {
    await versionMutation.mutateAsync({ code: family.code, ...values })
    setVersionFormOpen(false)
  }

  return (
    <Box padding="l" display="flex" flexDirection="column" gap="l">
      <Flex justifyContent="space-between" alignItems="center">
        <Heading size="large">职种详情</Heading>
        {hasPermission('job-catalog:update') && (
          <PrimaryButton onClick={() => setVersionFormOpen(true)}>新增版本</PrimaryButton>
        )}
      </Flex>

      <Box display="flex" flexDirection="column" gap="m">
        <Flex gap="m" alignItems="center">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              职种编码
            </Text>
            <Text fontSize="18px" fontWeight={600}>
              {family.code}
            </Text>
          </Box>
          <StatusBadge status={family.status} />
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            职种名称
          </Text>
          <Text fontSize="18px" fontWeight={600}>
            {family.name}
          </Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            归属职类
          </Text>
          <Text>{family.groupCode}</Text>
        </Box>

        <Flex gap="l">
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              生效日期
            </Text>
            <Text>{formatISODate(family.effectiveDate)}</Text>
          </Box>
          <Box>
            <Text typeLevel="body.small" color="licorice400">
              结束日期
            </Text>
            <Text>{formatISODate(family.endDate)}</Text>
          </Box>
        </Flex>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            描述
          </Text>
          <Text>{family.description ?? '暂无描述'}</Text>
        </Box>

        <Box>
          <Text typeLevel="body.small" color="licorice400">
            记录标识
          </Text>
          <Text fontFamily="monospace" fontSize="12px">
            {family.recordId}
          </Text>
        </Box>
      </Box>

      <CatalogVersionForm
        title="新增职种版本"
        isOpen={isVersionFormOpen}
        onClose={() => setVersionFormOpen(false)}
        onSubmit={handleCreateVersion}
        isSubmitting={versionMutation.isPending}
        initialName={family.name}
        initialDescription={family.description}
      />
    </Box>
  )
}
