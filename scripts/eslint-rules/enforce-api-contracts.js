/**
 * ESLint Rule: enforce-api-contracts
 * 强制执行API契约一致性，确保字段命名符合企业级标准
 * 
 * 检查项：
 * - 强制使用camelCase字段命名
 * - 禁止snake_case字段名
 * - 检查标准字段名词汇表
 * - 验证时态字段命名一致性
 * 
 * 符合API一致性设计规范：统一camelCase命名标准
 */

module.exports = {
  meta: {
    type: 'problem',
    docs: {
      description: '强制执行API契约一致性，确保字段命名符合企业级标准',
      category: 'Architecture',
      recommended: true,
    },
    fixable: 'code',
    schema: [
      {
        type: 'object',
        properties: {
          fieldNamingStyle: {
            type: 'string',
            enum: ['camelCase', 'snake_case'],
            default: 'camelCase'
          },
          allowedFields: {
            type: 'array',
            items: { type: 'string' },
            default: []
          },
          standardFields: {
            type: 'object',
            default: {
              // 核心业务字段 (camelCase)
              identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
              timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'],
              statusFields: ['status', 'isCurrent', 'isFuture', 'isTemporal'],
              operationFields: ['operationType', 'operatedBy', 'operationReason'],
              hierarchyFields: ['level', 'codePath', 'namePath', 'hierarchyDepth'],
              configFields: ['unitType', 'sortOrder', 'description', 'profile']
            }
          },
          deprecatedFields: {
            type: 'array',
            default: [
              'parent_unit_id', 'unit_type', 'is_deleted', 'operation_type',
              'created_at', 'updated_at', 'effective_date', 'end_date',
              'record_id', 'tenant_id', 'parent_code', 'is_current'
            ]
          },
          allowedContexts: {
            type: 'array',
            default: ['test', 'mock', 'fixture', 'migration']
          }
        },
        additionalProperties: false
      }
    ],
    messages: {
      useStandardField: '使用标准字段名 "{{standard}}" 替代 "{{current}}"',
      deprecatedField: '废弃字段 "{{field}}"，请使用 "{{replacement}}"',
      camelCaseRequired: '字段名必须使用camelCase格式: "{{field}}"',
      snakeCaseProhibited: '禁止使用snake_case字段名: "{{field}}"，请使用camelCase',
      nonStandardField: '非标准字段名 "{{field}}"，请检查字段词汇表',
      oauthException: 'OAuth字段 "{{field}}" 使用标准snake_case格式（协议要求）'
    }
  },

  create(context) {
    const options = context.options[0] || {};
    const fieldNamingStyle = options.fieldNamingStyle || 'camelCase';
    const allowedFields = Array.isArray(options.allowedFields) ? options.allowedFields : [];
    const standardFields = options.standardFields || {
      identifiers: ['code', 'parentCode', 'tenantId', 'recordId'],
      timeFields: ['createdAt', 'updatedAt', 'effectiveDate', 'endDate'],
      statusFields: ['status', 'isCurrent', 'isFuture', 'isTemporal'],
      operationFields: ['operationType', 'operatedBy', 'operationReason'],
      hierarchyFields: ['level', 'codePath', 'namePath', 'hierarchyDepth'],
      configFields: ['unitType', 'sortOrder', 'description', 'profile']
    };
    
    const deprecatedFields = options.deprecatedFields || [
      'parent_unit_id', 'unit_type', 'is_deleted', 'operation_type',
      'created_at', 'updated_at', 'effective_date', 'end_date',
      'record_id', 'tenant_id', 'parent_code', 'is_current'
    ];

    const allowedContexts = options.allowedContexts || ['test', 'mock', 'fixture', 'migration'];

    // 构建标准字段列表
    const allStandardFields = Object.values(standardFields).flat();
    
    // 构建废弃字段映射
    const deprecatedFieldMap = {
      'parent_unit_id': 'parentCode',
      'unit_type': 'unitType', 
      'is_deleted': 'status',
      'operation_type': 'operationType',
      'created_at': 'createdAt',
      'updated_at': 'updatedAt',
      'effective_date': 'effectiveDate',
      'end_date': 'endDate',
      'record_id': 'recordId',
      'tenant_id': 'tenantId',
      'parent_code': 'parentCode',
      'is_current': 'isCurrent'
    };

    // OAuth标准字段例外 + 允许字段白名单（用于特定文件夹例外）
    const oauthFields = ['client_id', 'client_secret', 'grant_type', 'refresh_token', 'access_token'];
    const globallyAllowedFields = new Set([...oauthFields, ...allowedFields]);

    // 检查字段命名格式
    function isCamelCase(name) {
      return /^[a-z][a-zA-Z0-9]*$/.test(name);
    }

    function isSnakeCase(name) {
      return /^[a-z][a-z0-9_]*$/.test(name) && name.includes('_');
    }

    // 检查是否在允许的上下文中
    function isInAllowedContext(filename) {
      return allowedContexts.some(context => 
        filename.toLowerCase().includes(context)
      );
    }

    // 转换为camelCase
    function toCamelCase(snakeStr) {
      return snakeStr.replace(/_([a-z])/g, (match, letter) => letter.toUpperCase());
    }

    // 检查字段名
    function checkFieldName(node, fieldName) {
      const filename = context.getFilename();
      
      // 跳过允许的上下文
      if (isInAllowedContext(filename)) {
        return;
      }

      // OAuth/白名单字段例外
      if (globallyAllowedFields.has(fieldName)) {
        // 在OAuth相关文件中允许；其他文件提示协议字段使用不当
        if (filename.includes('auth') || filename.includes('oauth')) {
          return;
        }
        // 对非auth/oauth文件，保留提示但不阻断（降级为提示消息）
        context.report({
          node,
          messageId: 'oauthException',
          data: { field: fieldName }
        });
        return;
      }

      // 检查废弃字段
      if (deprecatedFields.includes(fieldName)) {
        const replacement = deprecatedFieldMap[fieldName] || toCamelCase(fieldName);
        context.report({
          node,
          messageId: 'deprecatedField',
          data: { 
            field: fieldName, 
            replacement 
          },
          fix(fixer) {
            return fixer.replaceText(node, `"${replacement}"`);
          }
        });
        return;
      }

      // 检查命名格式
      if (fieldNamingStyle === 'camelCase') {
        if (isSnakeCase(fieldName) && !isCamelCase(fieldName)) {
          const camelCaseName = toCamelCase(fieldName);
          context.report({
            node,
            messageId: 'snakeCaseProhibited',
            data: { field: fieldName },
            fix(fixer) {
              return fixer.replaceText(node, `"${camelCaseName}"`);
            }
          });
          return;
        }

        if (!isCamelCase(fieldName)) {
          context.report({
            node,
            messageId: 'camelCaseRequired',
            data: { field: fieldName }
          });
          return;
        }
      }

      // 检查是否为标准字段
      if (!allStandardFields.includes(fieldName) && fieldName.length > 2) {
        // 查找最相似的标准字段
        const similarField = allStandardFields.find(std => 
          std.toLowerCase().includes(fieldName.toLowerCase()) ||
          fieldName.toLowerCase().includes(std.toLowerCase())
        );

        if (similarField) {
          context.report({
            node,
            messageId: 'useStandardField',
            data: { 
              current: fieldName, 
              standard: similarField 
            }
          });
        } else {
          context.report({
            node,
            messageId: 'nonStandardField',
            data: { field: fieldName }
          });
        }
      }
    }

    return {
      // 检查对象属性键
      Property(node) {
        if (node.key.type === 'Literal' && typeof node.key.value === 'string') {
          checkFieldName(node.key, node.key.value);
        }
        
        if (node.key.type === 'Identifier') {
          checkFieldName(node.key, node.key.name);
        }
      },

      // 检查GraphQL查询字段
      TemplateElement(node) {
        const value = node.value.raw;
        // 查找GraphQL查询中的字段名
        const graphqlFieldPattern = /(\w+)\s*[{\s]/g;
        let match;
        
        while ((match = graphqlFieldPattern.exec(value)) !== null) {
          const fieldName = match[1];
          if (fieldName && fieldName !== 'query' && fieldName !== 'mutation') {
            // 创建虚拟节点用于报告
            const virtualNode = {
              type: 'Identifier',
              name: fieldName,
              range: node.range,
              loc: node.loc
            };
            checkFieldName(virtualNode, fieldName);
          }
        }
      },

      // 检查字符串中的字段名（用于动态字段访问）
      Literal(node) {
        if (typeof node.value === 'string') {
          // 检查是否为可能的字段名
          if (/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(node.value) && 
              node.value.length > 2 && 
              node.value.length < 50) {
            
            // 检查上下文，确保这确实是字段名
            const parent = node.parent;
            const isFieldAccess = 
              (parent.type === 'MemberExpression' && parent.property === node) ||
              (parent.type === 'Property' && parent.key === node) ||
              (parent.type === 'CallExpression' && 
               parent.callee.name && 
               /^(get|set|has|delete|pick|omit)/.test(parent.callee.name));

            if (isFieldAccess) {
              checkFieldName(node, node.value);
            }
          }
        }
      },

      // 检查TypeScript接口定义
      TSPropertySignature(node) {
        if (node.key && node.key.type === 'Identifier') {
          checkFieldName(node.key, node.key.name);
        }
      },

      // 检查类型别名中的字段
      TSTypeLiteral(node) {
        node.members.forEach(member => {
          if (member.type === 'TSPropertySignature' && 
              member.key && 
              member.key.type === 'Identifier') {
            checkFieldName(member.key, member.key.name);
          }
        });
      }
    };
  }
};
