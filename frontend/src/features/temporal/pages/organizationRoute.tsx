import React from 'react'
import TemporalEntityPage, {
  type TemporalEntityRouteConfig,
  type TemporalEntityParseResult,
} from './TemporalEntityPage'
import { TemporalMasterDetailView } from '@/features/temporal/components/TemporalMasterDetailView'
import { TemporalEntityLayout } from '@/features/temporal/layout/TemporalEntityLayout'

const parseOrganizationCode = (rawCode?: string): TemporalEntityParseResult => {
  if (!rawCode) {
    return { isCreateMode: false, error: 'missing', rawCode }
  }
  if (rawCode === 'new') {
    return { isCreateMode: true, rawCode }
  }
  if (!/^\d{7}$/.test(rawCode)) {
    return { isCreateMode: false, rawCode, error: 'invalid' }
  }
  return { isCreateMode: false, code: rawCode, rawCode }
}

const organizationConfig: TemporalEntityRouteConfig = {
  entity: 'organization',
  listPath: '/organizations',
  buildDetailPath: code => `/organizations/${code}/temporal`,
  parseCode: parseOrganizationCode,
  invalidMessages: {
    missing: {
      title: '无效的组织编码',
      description: '请从组织列表页面正确访问组织详情功能。',
      actionLabel: '返回组织列表',
    },
    invalid: {
      title: '组织编码格式错误',
      description: '组织编码应为 7 位数字，请从列表页面重新进入。',
      actionLabel: '返回组织列表',
    },
  },
  renderContent: ctx => (
    <TemporalEntityLayout.Shell entity="organization">
      <TemporalMasterDetailView
        organizationCode={ctx.isCreateMode ? null : ctx.code ?? null}
        onBack={ctx.navigateToList}
        onCreateSuccess={createdCode => ctx.navigateToDetail(createdCode, { replace: true })}
        readonly={false}
        isCreateMode={ctx.isCreateMode}
      />
    </TemporalEntityLayout.Shell>
  ),
}

export const OrganizationTemporalEntityRoute: React.FC = () => (
  <TemporalEntityPage config={organizationConfig} />
)

export default OrganizationTemporalEntityRoute
