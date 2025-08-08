import React, { useEffect } from 'react';

// 直接测试导入 - 正常工作的组件
import { Box } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { PrimaryButton } from '@workday/canvas-kit-react/button';
import { Table } from '@workday/canvas-kit-react/table';

console.log('✅ 基础组件导入成功');
console.log('Box:', typeof Box);
console.log('Text:', typeof Text);
console.log('Heading:', typeof Heading);
console.log('PrimaryButton:', typeof PrimaryButton);
console.log('Table:', typeof Table);

export default function CanvasKitTest() {
  useEffect(() => {
    console.log('🔍 开始测试有问题的Canvas Kit组件...');
    
    // 测试Modal组件
    import('@workday/canvas-kit-react/modal')
      .then(({ Modal }) => {
        console.log('✅ Modal 组件导入成功:', typeof Modal);
      })
      .catch(e => {
        console.error('❌ Modal 组件导入失败:', e.message);
      });
    
    // 测试Card组件
    import('@workday/canvas-kit-react/card')
      .then(({ Card }) => {
        console.log('✅ Card 组件导入成功:', typeof Card);
      })
      .catch(e => {
        console.error('❌ Card 组件导入失败:', e.message);
      });
    
    // 测试FormField组件
    import('@workday/canvas-kit-react/form-field')
      .then(({ FormField }) => {
        console.log('✅ FormField 组件导入成功:', typeof FormField);
      })
      .catch(e => {
        console.error('❌ FormField 组件导入失败:', e.message);
      });
    
    // 测试TextInput组件
    import('@workday/canvas-kit-react/text-input')
      .then(({ TextInput }) => {
        console.log('✅ TextInput 组件导入成功:', typeof TextInput);
      })
      .catch(e => {
        console.error('❌ TextInput 组件导入失败:', e.message);
      });
    
    // 测试Select组件
    import('@workday/canvas-kit-react/select')
      .then(({ Select }) => {
        console.log('✅ Select 组件导入成功:', typeof Select);
      })
      .catch(e => {
        console.error('❌ Select 组件导入失败:', e.message);
      });
    
    // 测试TextArea组件
    import('@workday/canvas-kit-react/text-area')
      .then(({ TextArea }) => {
        console.log('✅ TextArea 组件导入成功:', typeof TextArea);
      })
      .catch(e => {
        console.error('❌ TextArea 组件导入失败:', e.message);
      });
  }, []);

  return (
    <Box padding="l">
      <Heading size="large">Canvas Kit 组件导入测试</Heading>
      <Text marginTop="m">请打开浏览器控制台查看导入测试结果</Text>
      
      <Box marginTop="l">
        <Heading size="medium" marginBottom="s">基础组件测试（应该正常工作）:</Heading>
        <PrimaryButton marginRight="s">按钮测试</PrimaryButton>
        <Text>文本测试</Text>
      </Box>
      
      <Box marginTop="l">
        <Table>
          <Table.Head>
            <Table.Row>
              <Table.Header>组件</Table.Header>
              <Table.Header>状态</Table.Header>
            </Table.Row>
          </Table.Head>
          <Table.Body>
            <Table.Row>
              <Table.Cell>Box, Text, Button, Table</Table.Cell>
              <Table.Cell>✅ 正常工作</Table.Cell>
            </Table.Row>
            <Table.Row>
              <Table.Cell>Modal, Card, FormField</Table.Cell>
              <Table.Cell>❓ 检查中...</Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table>
      </Box>
    </Box>
  );
}