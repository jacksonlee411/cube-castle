import { useQuery, type UseQueryOptions, type UseQueryResult } from '@tanstack/react-query'
import { graphqlEnterpriseAdapter } from '../api/graphql-enterprise-adapter'
import { createQueryError } from '../api/queryClient'

interface JobFamilyGroupNode {
  code: string
  name: string
  status?: string | null
}

interface JobFamilyNode {
  code: string
  name: string
  groupCode: string
  status?: string | null
}

interface JobRoleNode {
  code: string
  name: string
  familyCode: string
  status?: string | null
}

interface JobLevelNode {
  code: string
  name: string
  roleCode: string
  levelRank?: number | null
  status?: string | null
}

interface JobFamilyGroupsResponse {
  jobFamilyGroups: JobFamilyGroupNode[]
}

interface JobFamiliesResponse {
  jobFamilies: JobFamilyNode[]
}

interface JobRolesResponse {
  jobRoles: JobRoleNode[]
}

interface JobLevelsResponse {
  jobLevels: JobLevelNode[]
}

const JOB_FAMILY_GROUPS_QUERY = /* GraphQL */ `
  query JobFamilyGroups($includeInactive: Boolean, $asOfDate: Date) {
    jobFamilyGroups(includeInactive: $includeInactive, asOfDate: $asOfDate) {
      code
      name
      status
    }
  }
`

const JOB_FAMILIES_QUERY = /* GraphQL */ `
  query JobFamilies($groupCode: JobFamilyGroupCode!, $includeInactive: Boolean, $asOfDate: Date) {
    jobFamilies(groupCode: $groupCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
      code
      name
      groupCode
      status
    }
  }
`

const JOB_ROLES_QUERY = /* GraphQL */ `
  query JobRoles($familyCode: JobFamilyCode!, $includeInactive: Boolean, $asOfDate: Date) {
    jobRoles(familyCode: $familyCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
      code
      name
      familyCode
      status
    }
  }
`

const JOB_LEVELS_QUERY = /* GraphQL */ `
  query JobLevels($roleCode: JobRoleCode!, $includeInactive: Boolean, $asOfDate: Date) {
    jobLevels(roleCode: $roleCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
      code
      name
      roleCode
      levelRank
      status
    }
  }
`

interface JobCatalogQueryOptions {
  includeInactive?: boolean
  asOfDate?: string
}

type QueryContext<T> = UseQueryOptions<T, Error, T, readonly unknown[]>

const handleResponse = <T>(
  response: Awaited<ReturnType<typeof graphqlEnterpriseAdapter.request<T>>>,
  fallbackMessage: string,
): T => {
  if (!response.success || !response.data) {
    throw createQueryError(response.error?.message ?? fallbackMessage, {
      code: response.error?.code,
      requestId: response.requestId,
      details: response.error?.details,
    })
  }
  return response.data
}

export const useJobFamilyGroups = (
  options: JobCatalogQueryOptions = {},
): UseQueryResult<JobFamilyGroupNode[], Error> => {
  const queryOptions: QueryContext<JobFamilyGroupNode[]> = {
    queryKey: ['jobCatalog', 'groups', options.includeInactive ?? false, options.asOfDate ?? null],
    queryFn: async ({ signal }) => {
      const variables: Record<string, unknown> = {}
      if (typeof options.includeInactive === 'boolean') {
        variables.includeInactive = options.includeInactive
      }
      if (options.asOfDate) {
        variables.asOfDate = options.asOfDate
      }

      const response = await graphqlEnterpriseAdapter.request<JobFamilyGroupsResponse>(
        JOB_FAMILY_GROUPS_QUERY,
        variables,
        { signal },
      )

      const data = handleResponse(response, '获取职类列表失败')
      return data.jobFamilyGroups ?? []
    },
    staleTime: 5 * 60 * 1000,
  }

  return useQuery(queryOptions)
}

export const useJobFamilies = (
  groupCode?: string,
  options: JobCatalogQueryOptions & { enabled?: boolean } = {},
): UseQueryResult<JobFamilyNode[], Error> => {
  const enabled = options.enabled ?? Boolean(groupCode)

  const queryOptions: QueryContext<JobFamilyNode[]> = {
    queryKey: ['jobCatalog', 'families', groupCode ?? null, options.includeInactive ?? false, options.asOfDate ?? null],
    queryFn: async ({ signal }) => {
      if (!groupCode) {
        return []
      }

      const variables: Record<string, unknown> = { groupCode }
      if (typeof options.includeInactive === 'boolean') {
        variables.includeInactive = options.includeInactive
      }
      if (options.asOfDate) {
        variables.asOfDate = options.asOfDate
      }

      const response = await graphqlEnterpriseAdapter.request<JobFamiliesResponse>(
        JOB_FAMILIES_QUERY,
        variables,
        { signal },
      )

      const data = handleResponse(response, '获取职种列表失败')
      return data.jobFamilies ?? []
    },
    enabled,
    staleTime: 5 * 60 * 1000,
  }

  return useQuery(queryOptions)
}

export const useJobRoles = (
  familyCode?: string,
  options: JobCatalogQueryOptions & { enabled?: boolean } = {},
): UseQueryResult<JobRoleNode[], Error> => {
  const enabled = options.enabled ?? Boolean(familyCode)

  return useQuery({
    queryKey: ['jobCatalog', 'roles', familyCode ?? null, options.includeInactive ?? false, options.asOfDate ?? null],
    queryFn: async ({ signal }) => {
      if (!familyCode) {
        return []
      }

      const variables: Record<string, unknown> = { familyCode }
      if (typeof options.includeInactive === 'boolean') {
        variables.includeInactive = options.includeInactive
      }
      if (options.asOfDate) {
        variables.asOfDate = options.asOfDate
      }

      const response = await graphqlEnterpriseAdapter.request<JobRolesResponse>(
        JOB_ROLES_QUERY,
        variables,
        { signal },
      )

      const data = handleResponse(response, '获取职务列表失败')
      return data.jobRoles ?? []
    },
    enabled,
    staleTime: 5 * 60 * 1000,
  })
}

export const useJobLevels = (
  roleCode?: string,
  options: JobCatalogQueryOptions & { enabled?: boolean } = {},
): UseQueryResult<JobLevelNode[], Error> => {
  const enabled = options.enabled ?? Boolean(roleCode)

  return useQuery({
    queryKey: ['jobCatalog', 'levels', roleCode ?? null, options.includeInactive ?? false, options.asOfDate ?? null],
    queryFn: async ({ signal }) => {
      if (!roleCode) {
        return []
      }

      const variables: Record<string, unknown> = { roleCode }
      if (typeof options.includeInactive === 'boolean') {
        variables.includeInactive = options.includeInactive
      }
      if (options.asOfDate) {
        variables.asOfDate = options.asOfDate
      }

      const response = await graphqlEnterpriseAdapter.request<JobLevelsResponse>(
        JOB_LEVELS_QUERY,
        variables,
        { signal },
      )

      const data = handleResponse(response, '获取职级列表失败')
      return data.jobLevels ?? []
    },
    enabled,
    staleTime: 5 * 60 * 1000,
  })
}

export type { JobFamilyGroupNode, JobFamilyNode, JobRoleNode, JobLevelNode }
