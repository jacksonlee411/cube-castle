import { useMutation, useQueryClient } from '@tanstack/react-query'
import { unifiedRESTClient } from '@/shared/api'
import { createQueryError } from '@/shared/api/queryClient'
import { logger } from '@/shared/utils/logger'
import type { APIResponse } from '@/shared/types/api'
import type { JobCatalogStatus } from '@/generated/graphql-types'

const ensureSuccess = <T>(response: APIResponse<T>, fallbackMessage: string): T => {
  if (!response.success) {
    throw createQueryError(response.error?.message ?? fallbackMessage, {
      code: response.error?.code,
      details: response.error?.details,
      requestId: response.requestId,
    })
  }
  return response.data as T
}

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

const jsonHeaders = { 'Content-Type': 'application/json' } as const

const withIfMatch = (recordId: string) => ({
  ...jsonHeaders,
  'If-Match': recordId,
})

export const useCreateJobFamilyGroup = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateJobFamilyGroupInput) => {
      logger.mutation('[JobCatalog] create job family group', input)
      const response = await unifiedRESTClient.request<APIResponse<unknown>>('/job-family-groups', {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(input),
      })
      ensureSuccess(response, '创建职类失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-family-groups/${code}`, {
        method: 'PUT',
        headers: withIfMatch(recordId),
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '更新职类失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>('/job-families', {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(input),
      })
      ensureSuccess(response, '创建职种失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-families/${code}`, {
        method: 'PUT',
        headers: withIfMatch(recordId),
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '更新职种失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>('/job-roles', {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(input),
      })
      ensureSuccess(response, '创建职务失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-roles/${code}`, {
        method: 'PUT',
        headers: withIfMatch(recordId),
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '更新职务失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>('/job-levels', {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(input),
      })
      ensureSuccess(response, '创建职级失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-levels/${code}`, {
        method: 'PUT',
        headers: withIfMatch(recordId),
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '更新职级失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-family-groups/${code}/versions`, {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '创建职类版本失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-families/${code}/versions`, {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '创建职种版本失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-roles/${code}/versions`, {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '创建职务版本失败')
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
      const response = await unifiedRESTClient.request<APIResponse<unknown>>(`/job-levels/${code}/versions`, {
        method: 'POST',
        headers: jsonHeaders,
        body: JSON.stringify(payload),
      })
      ensureSuccess(response, '创建职级版本失败')
    },
    onSuccess: () => {
      invalidateLevels(queryClient)
    },
    onError: error => logger.error('[JobCatalog] create job level version failed', error),
  })
}
