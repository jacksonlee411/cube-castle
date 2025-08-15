import React, { useState } from 'react';
import { Card } from '@workday/canvas-kit-react/card';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton } from '@workday/canvas-kit-react/button';
import { TemporalDatePicker, validateTemporalDate } from './features/temporal/components/TemporalDatePicker';
import { TemporalStatusSelector, type TemporalStatus } from './features/temporal/components/TemporalStatusSelector';
import { TemporalInfoDisplay, TemporalStatusBadge } from './features/temporal/components/TemporalInfoDisplay';
import { PlannedOrganizationForm } from './features/temporal/components/PlannedOrganizationForm';
import { useCreatePlannedOrganization } from './features/temporal/hooks/useTemporalAPI';

const TemporalComponentTest: React.FC = () => {
  const [effectiveDate, setEffectiveDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [temporalStatus, setTemporalStatus] = useState<TemporalStatus>('ACTIVE');
  const [showPlannedForm, setShowPlannedForm] = useState(false);

  const createPlannedMutation = useCreatePlannedOrganization();

  const testTemporalInfo = {
    effective_date: effectiveDate || '2024-01-01',
    end_date: endDate || '2024-12-31',
    status: temporalStatus,
    is_temporal: true,
    change_reason: '测试用途',
    version: 1,
  };

  const handleCreatePlanned = async (data: any) => {
    try {
      await createPlannedMutation.mutateAsync(data);
      console.log('计划组织创建成功');
    } catch (error) {
      console.error('创建失败:', error);
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <Text as="h2" marginBottom="l">时态管理组件测试页面</Text>
      
      <Flex flexDirection="column" gap="l">
        {/* 时态日期选择器测试 */}
        <Card>
          <Card.Heading>时态日期选择器</Card.Heading>
          <Card.Body>
            <Flex gap="m">
              <TemporalDatePicker
                label="生效日期"
                value={effectiveDate}
                onChange={setEffectiveDate}
                helperText="选择组织生效日期"
              />
              <TemporalDatePicker
                label="结束日期"
                value={endDate}
                onChange={setEndDate}
                minDate={effectiveDate}
                helperText="选择组织结束日期"
              />
            </Flex>
            
            <Flex marginTop="m" gap="s">
              <Text>选择的日期范围：</Text>
              <Text color="positive">
                {effectiveDate && validateTemporalDate.formatDateDisplay(effectiveDate)} 
                {effectiveDate && endDate && ' - '}
                {endDate && validateTemporalDate.formatDateDisplay(endDate)}
              </Text>
            </Flex>
          </Card.Body>
        </Card>

        {/* 时态状态选择器测试 */}
        <Card>
          <Card.Heading>时态状态选择器</Card.Heading>
          <Card.Body>
            <Flex gap="m" alignItems="flex-end">
              <TemporalStatusSelector
                value={temporalStatus}
                onChange={setTemporalStatus}
                helperText="选择组织时态状态"
              />
              
              <TemporalStatusBadge status={temporalStatus} size="medium" />
            </Flex>
          </Card.Body>
        </Card>

        {/* 时态信息显示测试 */}
        <Card>
          <Card.Heading>时态信息显示组件</Card.Heading>
          <Card.Body>
            <Flex flexDirection="column" gap="m">
              <div>
                <Text as="span">紧凑模式：</Text>
                <TemporalInfoDisplay 
                  temporalInfo={testTemporalInfo}
                  variant="compact"
                />
              </div>
              
              <div>
                <Text as="span">默认模式：</Text>
                <TemporalInfoDisplay 
                  temporalInfo={testTemporalInfo}
                  variant="default"
                />
              </div>
              
              <div>
                <Text as="span">详细模式：</Text>
                <TemporalInfoDisplay 
                  temporalInfo={testTemporalInfo}
                  variant="detailed"
                  showChangeReason
                  showVersion
                />
              </div>
            </Flex>
          </Card.Body>
        </Card>

        {/* 计划组织创建测试 */}
        <Card>
          <Card.Heading>计划组织创建表单</Card.Heading>
          <Card.Body>
            <Flex gap="m" alignItems="center">
              <PrimaryButton 
                onClick={() => setShowPlannedForm(true)}
                disabled={createPlannedMutation.isPending}
              >
                {createPlannedMutation.isPending ? '创建中...' : '创建计划组织'}
              </PrimaryButton>
              
              {createPlannedMutation.isSuccess && (
                <Text color="positive">✓ 创建成功</Text>
              )}
              
              {createPlannedMutation.isError && (
                <Text color="alert">✗ 创建失败: {createPlannedMutation.error?.message}</Text>
              )}
            </Flex>
          </Card.Body>
        </Card>

        {/* 状态徽章展示 */}
        <Card>
          <Card.Heading>状态徽章展示</Card.Heading>
          <Card.Body>
            <Flex gap="m" flexWrap="wrap">
              <TemporalStatusBadge status="ACTIVE" size="small" />
              <TemporalStatusBadge status="ACTIVE" size="medium" />
              <TemporalStatusBadge status="ACTIVE" size="large" />
              
              <TemporalStatusBadge status="PLANNED" size="small" />
              <TemporalStatusBadge status="PLANNED" size="medium" />
              <TemporalStatusBadge status="PLANNED" size="large" />
              
              <TemporalStatusBadge status="INACTIVE" size="small" />
              <TemporalStatusBadge status="INACTIVE" size="medium" />
              <TemporalStatusBadge status="INACTIVE" size="large" />
            </Flex>
          </Card.Body>
        </Card>
      </Flex>

      {/* 计划组织创建表单 */}
      <PlannedOrganizationForm
        isOpen={showPlannedForm}
        onClose={() => setShowPlannedForm(false)}
        onSubmit={handleCreatePlanned}
        loading={createPlannedMutation.isPending}
      />
    </div>
  );
};

export default TemporalComponentTest;