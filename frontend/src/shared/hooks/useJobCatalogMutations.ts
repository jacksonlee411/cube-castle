import { useMutation, useQueryClient } from '@tanstack/react-query'
import { logger } from '@/shared/utils/logger'
import type { JobCatalogStatus } from '@/generated/graphql-types'
import * as jobCatalog from '@/shared/api/facade/jobCatalog'

// 成功判断与错误封装逻辑已在 Facade 处理

const invalidateGroups = (client: ReturnType<typeof useQueryClient>) => {
  client.invalidateQueries({ queryKey: ['jobCatalog', 'groups'], exact: false })
}

const invalidateFamilies = (client: ReturnType<typeof useQueryClient>) => {
  client.invalidateQueries({ queryKey: ['jobCatalog', 'families'], exact: false })
}

const invalidateRoles = (client: ReturnType<typeof useQueryClient>) => {
  client.invalidateQueries({ queryKey: ['jobCatalog', 'roles'], exact: false })
}

const invalidateLevels = (client: ReturnType<typeof useQueryClient>) => {
  client.invalidateQueries({ queryKey: ['jobCatalog', 'levels'], exact: false })
}

export interface CreateJobFamilyGroupInput {
  code: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface CreateJobFamilyInput {
  code: string
  jobFamilyGroupCode: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface CreateJobRoleInput {
  code: string
  jobFamilyCode: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface CreateJobLevelInput {
  code: string
  jobRoleCode: string
  name: string
  levelRank: number
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface CreateCatalogVersionInput {
  code: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface UpdateJobFamilyGroupInput {
  code: string
  recordId: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface UpdateJobFamilyInput {
  code: string
  recordId: string
  jobFamilyGroupCode?: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface UpdateJobRoleInput {
  code: string
  recordId: string
  jobFamilyCode?: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
}

export interface UpdateJobLevelInput {
  code: string
  recordId: string
  jobRoleCode?: string
  name: string
  status: JobCatalogStatus
  effectiveDate: string
  description?: string | null
  levelRank?: number
}

// 头注入与 If-Match 已在 Facade 实现，无需在 Hook 层处理

export const useCreateJobFamilyGroup = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateJobFamilyGroupInput) => {
      logger.mutation('[JobCatalog] create job family group', input)
      await jobCatalog.createJobFamilyGroup(input)
    },
    onSuccess: () => invalidateGroups(queryClient),
    onError: error => logger.error('[JobCatalog] create job family group failed', error),
  })
}

export const useUpdateJobFamilyGroup = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateJobFamilyGroupInput) => {
      const { code, recordId, ...payload } = input
      logger.mutation('[JobCatalog] update job family group', { code, ...payload, recordId })
      await jobCatalog.updateJobFamilyGroup(input)
    },
    onSuccess: () => invalidateGroups(queryClient),
    onError: error => logger.error('[JobCatalog] update job family group failed', error),
  })
}

export const useCreateJobFamily = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateJobFamilyInput) => {
      logger.mutation('[JobCatalog] create job family', input)
      await jobCatalog.createJobFamily(input)
    },
    onSuccess: () => {
      invalidateFamilies(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job family failed', error),
  })
}

export const useUpdateJobFamily = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateJobFamilyInput) => {
      const { code, recordId, ...payload } = input
      logger.mutation('[JobCatalog] update job family', { code, ...payload, recordId })
      await jobCatalog.updateJobFamily(input)
    },
    onSuccess: () => {
      invalidateFamilies(queryClient)
    },
    onError: error => logger.error('[JobCatalog] update job family failed', error),
  })
}

export const useCreateJobRole = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateJobRoleInput) => {
      logger.mutation('[JobCatalog] create job role', input)
      await jobCatalog.createJobRole(input)
    },
    onSuccess: () => {
      invalidateRoles(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job role failed', error),
  })
}

export const useUpdateJobRole = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateJobRoleInput) => {
      const { code, recordId, ...payload } = input
      logger.mutation('[JobCatalog] update job role', { code, ...payload, recordId })
      await jobCatalog.updateJobRole(input)
    },
    onSuccess: () => {
      invalidateRoles(queryClient)
    },
    onError: error => logger.error('[JobCatalog] update job role failed', error),
  })
}

export const useCreateJobLevel = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateJobLevelInput) => {
      logger.mutation('[JobCatalog] create job level', input)
      await jobCatalog.createJobLevel(input)
    },
    onSuccess: () => {
      invalidateLevels(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job level failed', error),
  })
}

export const useUpdateJobLevel = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateJobLevelInput) => {
      const { code, recordId, ...payload } = input
      logger.mutation('[JobCatalog] update job level', { code, ...payload, recordId })
      await jobCatalog.updateJobLevel(input)
    },
    onSuccess: () => {
      invalidateLevels(queryClient)
    },
    onError: error => logger.error('[JobCatalog] update job level failed', error),
  })
}

export const useCreateJobFamilyGroupVersion = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateCatalogVersionInput) => {
      const { code, ...payload } = input
      logger.mutation('[JobCatalog] create job family group version', { code, ...payload })
      await jobCatalog.createJobFamilyGroupVersion(input)
    },
    onSuccess: () => invalidateGroups(queryClient),
    onError: error => logger.error('[JobCatalog] create job family group version failed', error),
  })
}

export const useCreateJobFamilyVersion = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateCatalogVersionInput) => {
      const { code, ...payload } = input
      logger.mutation('[JobCatalog] create job family version', { code, ...payload })
      await jobCatalog.createJobFamilyVersion(input)
    },
    onSuccess: () => {
      invalidateFamilies(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job family version failed', error),
  })
}

export const useCreateJobRoleVersion = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateCatalogVersionInput) => {
      const { code, ...payload } = input
      logger.mutation('[JobCatalog] create job role version', { code, ...payload })
      await jobCatalog.createJobRoleVersion(input)
    },
    onSuccess: () => {
      invalidateRoles(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job role version failed', error),
  })
}

export const useCreateJobLevelVersion = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateCatalogVersionInput) => {
      const { code, ...payload } = input
      logger.mutation('[JobCatalog] create job level version', { code, ...payload })
      await jobCatalog.createJobLevelVersion(input)
    },
    onSuccess: () => {
      invalidateLevels(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job level version failed', error),
  })
}
