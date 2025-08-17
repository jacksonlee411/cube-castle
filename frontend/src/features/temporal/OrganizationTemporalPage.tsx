/**
 * 组织详情页面
 * 集成了主从视图设计的完整组织详情体验
 */
import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { TemporalMasterDetailView } from './components/TemporalMasterDetailView';

export const OrganizationTemporalPage: React.FC = () => {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();

  if (!code) {
    return (
      <div>
        <h2>错误：缺少组织编码</h2>
        <button onClick={() => navigate('/organization-units')}>
          返回组织列表
        </button>
      </div>
    );
  }

  return (
    <TemporalMasterDetailView
      organizationCode={code}
      onBack={() => navigate('/organization-units')}
      readonly={false}
    />
  );
};

export default OrganizationTemporalPage;