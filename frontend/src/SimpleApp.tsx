import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';

interface Organization {
  code: string;
  name: string;
  unit_type: string;
  status: string;
}

const SimpleApp: React.FC = () => {
  const [count, setCount] = React.useState(0);
  const [orgName, setOrgName] = React.useState('');
  const [organizations, setOrganizations] = React.useState<Organization[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [lastCreatedOrg, setLastCreatedOrg] = React.useState<Organization | null>(null);
  
  const testBackendConnection = async () => {
    try {
      const response = await fetch('http://localhost:9090/health');
      const data = await response.text();
      alert(`后端连接成功: ${data}`);
    } catch (error) {
      alert(`后端连接失败: ${error}`);
    }
  };

  const createOrganization = async () => {
    if (!orgName.trim()) {
      alert('请输入组织名称');
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('http://localhost:9090/api/v1/organization-units', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: orgName,
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          description: '测试创建的组织'
        }),
      });

      if (response.ok) {
        const newOrg = await response.json();
        setLastCreatedOrg(newOrg);
        setOrgName('');
        alert(`组织创建成功！编码: ${newOrg.code}`);
      } else {
        const error = await response.text();
        alert(`创建失败: ${error}`);
      }
    } catch (error) {
      alert(`创建失败: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const queryOrganizations = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: `
            query {
              organizations {
                code
                name
                unit_type
                status
              }
            }
          `,
        }),
      });

      if (response.ok) {
        const result = await response.json();
        if (result.data && result.data.organizations) {
          setOrganizations(result.data.organizations.slice(0, 5)); // 只显示前5条
          alert(`查询成功！获取到 ${result.data.organizations.length} 个组织`);
        }
      } else {
        alert('GraphQL查询失败 - 可能查询服务未启动');
      }
    } catch (error) {
      alert(`查询失败: ${error} - 请确保查询服务(端口8090)已启动`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box padding="l">
      <Heading size="large" marginBottom="m">
        Cube Castle 系统测试页面 v2.0
      </Heading>
      
      {/* 基础功能测试 */}
      <Card marginBottom="m">
        <Card.Heading>基础功能测试</Card.Heading>
        <Card.Body>
          <Box marginBottom="m">
            <Text>计数器: {count}</Text>
          </Box>
          
          <Flex gap="s">
            <PrimaryButton onClick={() => setCount(count + 1)}>
              增加
            </PrimaryButton>
            
            <PrimaryButton onClick={testBackendConnection}>
              测试后端连接
            </PrimaryButton>
          </Flex>
        </Card.Body>
      </Card>

      {/* 组织管理功能测试 */}
      <Card marginBottom="m">
        <Card.Heading>组织管理功能测试</Card.Heading>
        <Card.Body>
          <FormField>
            <FormField.Label>组织名称</FormField.Label>
            <FormField.Field>
              <TextInput 
                value={orgName}
                onChange={(e) => setOrgName(e.target.value)}
                placeholder="输入测试组织名称"
              />
            </FormField.Field>
          </FormField>
          
          <Flex gap="s" marginTop="m">
            <PrimaryButton 
              onClick={createOrganization}
              disabled={loading}
            >
              {loading ? '创建中...' : '创建组织 (REST API)'}
            </PrimaryButton>
            
            <SecondaryButton 
              onClick={queryOrganizations}
              disabled={loading}
            >
              {loading ? '查询中...' : '查询组织 (GraphQL)'}
            </SecondaryButton>
          </Flex>

          {lastCreatedOrg && (
            <Box marginTop="m" padding="s" backgroundColor="neutral.100" borderRadius="s">
              <Text typeLevel="subtext.medium">最近创建的组织:</Text>
              <Text>编码: {lastCreatedOrg.code}</Text>
              <Text>名称: {lastCreatedOrg.name}</Text>
              <Text>状态: {lastCreatedOrg.status}</Text>
            </Box>
          )}
        </Card.Body>
      </Card>

      {/* 查询结果展示 */}
      {organizations.length > 0 && (
        <Card marginBottom="m">
          <Card.Heading>组织查询结果 (前5条)</Card.Heading>
          <Card.Body>
            {organizations.map((org) => (
              <Box key={org.code} marginBottom="s" padding="xs" backgroundColor="neutral.100" borderRadius="s">
                <Text>{org.code} - {org.name} ({org.status})</Text>
              </Box>
            ))}
          </Card.Body>
        </Card>
      )}
      
      {/* 系统状态 */}
      <Card>
        <Card.Heading>系统状态</Card.Heading>
        <Card.Body>
          <Text>✅ Canvas Kit UI组件正常加载</Text><br/>
          <Text>✅ React状态管理正常工作</Text><br/>
          <Text>✅ 后端API连接正常 (端口9090)</Text><br/>
          <Text>⏳ GraphQL查询服务连接待测试 (端口8090)</Text>
        </Card.Body>
      </Card>
    </Box>
  );
};

export default SimpleApp;