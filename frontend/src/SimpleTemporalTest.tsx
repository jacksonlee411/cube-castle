import React, { useState } from 'react';
import { Card } from '@workday/canvas-kit-react/card';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton } from '@workday/canvas-kit-react/button';
import { TemporalDatePicker, validateTemporalDate } from './features/temporal/components/TemporalDatePicker';

const SimpleTemporalTest: React.FC = () => {
  const [effectiveDate, setEffectiveDate] = useState('');
  const [endDate, setEndDate] = useState('');

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <Text as="h2" marginBottom="l">时态管理组件功能验证</Text>
      
      <Flex flexDirection="column" gap="l">
        {/* 时态日期选择器测试 */}
        <Card>
          <Card.Heading>✅ 时态日期选择器组件</Card.Heading>
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
            
            <div style={{ marginTop: '16px', padding: '16px', backgroundColor: '#f5f5f5', borderRadius: '8px' }}>
              <Text><strong>测试结果：</strong></Text>
              <ul>
                <li>生效日期：{effectiveDate || '未选择'}</li>
                <li>结束日期：{endDate || '未选择'}</li>
                {effectiveDate && (
                  <li>格式化显示：{validateTemporalDate.formatDateDisplay(effectiveDate)}</li>
                )}
                {effectiveDate && (
                  <li>日期验证：{validateTemporalDate.isValidDate(effectiveDate) ? '✅ 有效' : '❌ 无效'}</li>
                )}
                {effectiveDate && (
                  <li>未来日期：{validateTemporalDate.isFutureDate(effectiveDate) ? '✅ 是' : '❌ 否'}</li>
                )}
                {effectiveDate && endDate && (
                  <li>日期顺序：{validateTemporalDate.isEndDateAfterStartDate(effectiveDate, endDate) ? '✅ 正确' : '❌ 错误'}</li>
                )}
              </ul>
            </div>

            <Flex gap="s" marginTop="m">
              <PrimaryButton onClick={() => setEffectiveDate('2024-01-01')}>
                设置为过去日期
              </PrimaryButton>
              <PrimaryButton onClick={() => setEffectiveDate('2026-01-01')}>
                设置为未来日期
              </PrimaryButton>
              <PrimaryButton onClick={() => { setEffectiveDate(''); setEndDate(''); }}>
                重置
              </PrimaryButton>
            </Flex>
          </Card.Body>
        </Card>

        {/* API功能测试 */}
        <Card>
          <Card.Heading>🧪 时态API功能测试</Card.Heading>
          <Card.Body>
            <Flex flexDirection="column" gap="m">
              <Text>测试计划组织创建API：</Text>
              
              <PrimaryButton 
                onClick={async () => {
                  try {
                    const response = await fetch('http://localhost:9090/api/v1/organization-units/planned', {
                      method: 'POST',
                      headers: { 'Content-Type': 'application/json' },
                      body: JSON.stringify({
                        name: '前端测试计划组织',
                        unit_type: 'DEPARTMENT',
                        description: '通过前端界面创建的测试计划组织',
                        effective_date: '2026-06-01',
                        end_date: '2026-12-31',
                        change_reason: '前端功能验证测试'
                      })
                    });

                    if (response.ok) {
                      const data = await response.json();
                      alert(`✅ 创建成功！组织代码：${data.code}`);
                    } else {
                      const error = await response.json();
                      alert(`❌ 创建失败：${error.error || error.message}`);
                    }
                  } catch (error) {
                    alert(`❌ 请求失败：${error instanceof Error ? error.message : '未知错误'}`);
                  }
                }}
              >
                测试创建计划组织
              </PrimaryButton>

              <PrimaryButton 
                onClick={async () => {
                  try {
                    const response = await fetch('http://localhost:8090/graphql', {
                      method: 'POST',
                      headers: { 'Content-Type': 'application/json' },
                      body: JSON.stringify({
                        query: `
                          query {
                            organizations {
                              code 
                              name 
                              status 
                              effective_date 
                              end_date
                            }
                          }
                        `
                      })
                    });

                    if (response.ok) {
                      const data = await response.json();
                      console.log('GraphQL查询结果:', data);
                      alert(`✅ GraphQL查询成功！找到 ${data.data?.organizations?.length || 0} 个组织`);
                    } else {
                      alert('❌ GraphQL查询失败');
                    }
                  } catch (error) {
                    alert(`❌ GraphQL请求失败：${error instanceof Error ? error.message : '未知错误'}`);
                  }
                }}
              >
                测试GraphQL查询
              </PrimaryButton>
            </Flex>
          </Card.Body>
        </Card>

        {/* 功能总结 */}
        <Card>
          <Card.Heading>📋 时态管理功能实现总结</Card.Heading>
          <Card.Body>
            <div style={{ lineHeight: '1.6' }}>
              <Text><strong>已实现的组件：</strong></Text>
              <ul>
                <li>✅ <strong>TemporalDatePicker</strong> - 时态日期选择器</li>
                <li>✅ <strong>TemporalStatusSelector</strong> - 时态状态选择器</li>
                <li>✅ <strong>TemporalInfoDisplay</strong> - 时态信息显示组件</li>
                <li>✅ <strong>TemporalStatusBadge</strong> - 时态状态徽章</li>
                <li>✅ <strong>PlannedOrganizationForm</strong> - 计划组织创建表单</li>
              </ul>

              <Text marginTop="m"><strong>支持的时态功能：</strong></Text>
              <ul>
                <li>🗓️ 日期验证和格式化</li>
                <li>📅 未来日期计划组织创建</li>
                <li>🏷️ 时态状态管理 (ACTIVE/PLANNED/INACTIVE)</li>
                <li>📊 时态信息多种显示模式</li>
                <li>🔍 时态数据筛选和查询</li>
                <li>📈 时间范围和历史时点查询</li>
              </ul>

              <Text marginTop="m"><strong>集成状态：</strong></Text>
              <ul>
                <li>🔧 后端API已完全支持时态字段</li>
                <li>🗃️ 数据库已升级支持时态管理</li>
                <li>⚡ CDC数据同步已验证</li>
                <li>🎨 前端组件已基本完成</li>
                <li>🚧 待完成：集成到组织架构页面</li>
              </ul>
            </div>
          </Card.Body>
        </Card>
      </Flex>
    </div>
  );
};

export default SimpleTemporalTest;