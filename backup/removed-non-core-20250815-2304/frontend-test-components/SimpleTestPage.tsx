import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton } from '@workday/canvas-kit-react/button';

export const SimpleTestPage: React.FC = () => {
  const [organizations, setOrganizations] = React.useState([]);
  const [loading, setLoading] = React.useState(false);

  const testGraphQLQuery = async () => {
    setLoading(true);
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
                unitType
              }
            }
          `
        })
      });
      const data = await response.json();
      setOrganizations(data.data?.organizations || []);
    } catch (error) {
      console.error('GraphQL查询错误:', error);
    }
    setLoading(false);
  };

  const testRESTCommand = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:9090/api/v1/organization-units', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'MCP测试组织',
          unit_type: 'DEPARTMENT',
          parent_code: null
        })
      });
      const data = await response.json();
      console.log('命令服务响应:', data);
    } catch (error) {
      console.error('REST命令错误:', error);
    }
    setLoading(false);
  };

  return (
    <Box padding="xl" maxWidth="800px">
      <Text as="h1">
        Cube Castle - MCP浏览器验证
      </Text>
      
      <Card margin="m">
        <Card.Heading>CQRS协议验证</Card.Heading>
        <Card.Body>
          <Box display="flex" marginBottom="m" style={{gap: '16px'}}>
            <PrimaryButton 
              onClick={testGraphQLQuery}
              disabled={loading}
            >
              测试GraphQL查询服务
            </PrimaryButton>
            <PrimaryButton 
              onClick={testRESTCommand}
              disabled={loading}
            >
              测试REST命令服务
            </PrimaryButton>
          </Box>
          
          {loading && (
            <Text>加载中...</Text>
          )}
          
          {organizations.length > 0 && (
            <Box marginTop="m">
              <Text as="h3">查询结果 ({organizations.length}个组织):</Text>
              <ul>
                {organizations.slice(0, 3).map((org: any) => (
                  <li key={org.code}>
                    {org.code} - {org.name} ({org.status})
                  </li>
                ))}
              </ul>
            </Box>
          )}
        </Card.Body>
      </Card>
    </Box>
  );
};