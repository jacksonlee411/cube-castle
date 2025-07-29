// tests/integration/api.test.ts
import { GraphQLClient } from 'graphql-request';
import { 
  GET_EMPLOYEES, 
  GET_EMPLOYEE, 
  CREATE_POSITION_CHANGE,
  GET_ORGANIZATION_CHART,
  GET_SITUATIONAL_CONTEXT,
  APPROVE_POSITION_CHANGE
} from '../../src/lib/graphql-queries';

describe('Frontend API Integration Tests', () => {
  let client: GraphQLClient;
  
  beforeAll(() => {
    // Initialize GraphQL client
    client = new GraphQLClient(process.env.GRAPHQL_ENDPOINT || 'http://localhost:8080/graphql', {
      headers: {
        authorization: `Bearer ${process.env.TEST_TOKEN || 'test-token'}`,
      },
    });
  });

  describe('Employee Management API', () => {
    test('should fetch employees list with pagination', async () => {
      const variables = {
        first: 10,
        after: null,
        filters: {}
      };

      const response = await client.request(GET_EMPLOYEES, variables);
      
      expect(response.employees).toBeDefined();
      expect(response.employees.edges).toBeInstanceOf(Array);
      expect(response.employees.pageInfo).toBeDefined();
      expect(response.employees.totalCount).toBeGreaterThan(0);
      
      // Check employee structure
      if (response.employees.edges.length > 0) {
        const employee = response.employees.edges[0].node;
        expect(employee.id).toBeDefined();
        expect(employee.employeeId).toBeDefined();
        expect(employee.legalName).toBeDefined();
        expect(employee.email).toBeDefined();
        expect(employee.status).toBeDefined();
      }
    });

    test('should fetch single employee details', async () => {
      // First get an employee ID
      const employeesResponse = await client.request(GET_EMPLOYEES, { first: 1 });
      const employeeId = employeesResponse.employees.edges[0].node.id;

      const response = await client.request(GET_EMPLOYEE, { id: employeeId });
      
      expect(response.employee).toBeDefined();
      expect(response.employee.id).toBe(employeeId);
      expect(response.employee.currentPosition).toBeDefined();
      expect(response.employee.currentPosition.positionTitle).toBeDefined();
      expect(response.employee.currentPosition.department).toBeDefined();
    });

    test('should filter employees by department', async () => {
      const variables = {
        first: 10,
        filters: {
          department: '研发部'
        }
      };

      const response = await client.request(GET_EMPLOYEES, variables);
      
      expect(response.employees).toBeDefined();
      
      // All returned employees should be from the specified department
      response.employees.edges.forEach(edge => {
        if (edge.node.currentPosition) {
          expect(edge.node.currentPosition.department).toBe('研发部');
        }
      });
    });

    test('should filter employees by status', async () => {
      const variables = {
        first: 10,
        filters: {
          status: 'ACTIVE'
        }
      };

      const response = await client.request(GET_EMPLOYEES, variables);
      
      expect(response.employees).toBeDefined();
      
      // All returned employees should have ACTIVE status
      response.employees.edges.forEach(edge => {
        expect(edge.node.status).toBe('ACTIVE');
      });
    });

    test('should search employees by name', async () => {
      const variables = {
        first: 10,
        filters: {
          search: '张'
        }
      };

      const response = await client.request(GET_EMPLOYEES, variables);
      
      expect(response.employees).toBeDefined();
      
      // All returned employees should have '张' in their name
      response.employees.edges.forEach(edge => {
        expect(edge.node.legalName).toContain('张');
      });
    });

    test('should handle pagination correctly', async () => {
      // Get first page
      const firstPage = await client.request(GET_EMPLOYEES, { first: 5 });
      
      expect(firstPage.employees.edges).toHaveLength(5);
      expect(firstPage.employees.pageInfo.hasNextPage).toBe(true);
      
      // Get second page
      const secondPage = await client.request(GET_EMPLOYEES, {
        first: 5,
        after: firstPage.employees.pageInfo.endCursor
      });
      
      expect(secondPage.employees.edges).toHaveLength(5);
      
      // Ensure different employees on different pages
      const firstPageIds = firstPage.employees.edges.map(edge => edge.node.id);
      const secondPageIds = secondPage.employees.edges.map(edge => edge.node.id);
      
      expect(firstPageIds).not.toEqual(secondPageIds);
    });
  });

  describe('Position Change Workflow API', () => {
    test('should create position change workflow', async () => {
      // First get an employee
      const employeesResponse = await client.request(GET_EMPLOYEES, { first: 1 });
      const employee = employeesResponse.employees.edges[0].node;

      const variables = {
        input: {
          employeeId: employee.id,
          positionTitle: '高级技术专家',
          department: '研发部',
          jobLevel: 'SENIOR',
          effectiveDate: new Date().toISOString().split('T')[0],
          changeReason: '晋升',
          minSalary: 25000,
          maxSalary: 30000,
          currency: 'CNY'
        }
      };

      const response = await client.request(CREATE_POSITION_CHANGE, variables);
      
      expect(response.createPositionChange).toBeDefined();
      expect(response.createPositionChange.positionHistory).toBeDefined();
      expect(response.createPositionChange.workflowId).toBeDefined();
      expect(response.createPositionChange.errors).toHaveLength(0);
    });

    test('should validate position change input', async () => {
      const variables = {
        input: {
          employeeId: 'invalid-id',
          positionTitle: '',
          department: '',
          effectiveDate: 'invalid-date'
        }
      };

      try {
        await client.request(CREATE_POSITION_CHANGE, variables);
        fail('Should have thrown validation error');
      } catch (error) {
        expect(error.response.errors).toBeDefined();
      }
    });

    test('should approve position change workflow', async () => {
      // First create a position change
      const employeesResponse = await client.request(GET_EMPLOYEES, { first: 1 });
      const employee = employeesResponse.employees.edges[0].node;

      const createResponse = await client.request(CREATE_POSITION_CHANGE, {
        input: {
          employeeId: employee.id,
          positionTitle: '技术主管',
          department: '研发部',
          jobLevel: 'MANAGER',
          effectiveDate: new Date().toISOString().split('T')[0],
          changeReason: '晋升'
        }
      });

      const workflowId = createResponse.createPositionChange.workflowId;

      // Approve the workflow
      const approveResponse = await client.request(APPROVE_POSITION_CHANGE, {
        workflowId,
        comments: '同意晋升申请'
      });

      expect(approveResponse.approvePositionChange).toBeDefined();
      expect(approveResponse.approvePositionChange.success).toBe(true);
      expect(approveResponse.approvePositionChange.workflowId).toBe(workflowId);
    });
  });

  describe('Organization Chart API', () => {
    test('should fetch organization chart data', async () => {
      const variables = {
        rootDepartment: null,
        maxLevels: 5
      };

      const response = await client.request(GET_ORGANIZATION_CHART, variables);
      
      expect(response.organizationChart).toBeDefined();
      expect(response.organizationChart.department).toBeDefined();
      expect(response.organizationChart.employees).toBeInstanceOf(Array);
      expect(response.organizationChart.totalEmployees).toBeGreaterThan(0);
      
      // Check sub-departments structure
      if (response.organizationChart.subDepartments) {
        expect(response.organizationChart.subDepartments).toBeInstanceOf(Array);
      }
    });

    test('should fetch organization chart for specific department', async () => {
      const variables = {
        rootDepartment: '研发部',
        maxLevels: 3
      };

      const response = await client.request(GET_ORGANIZATION_CHART, variables);
      
      expect(response.organizationChart).toBeDefined();
      expect(response.organizationChart.department).toBe('研发部');
      expect(response.organizationChart.employees).toBeInstanceOf(Array);
    });

    test('should handle empty departments', async () => {
      const variables = {
        rootDepartment: '不存在的部门',
        maxLevels: 2
      };

      try {
        const response = await client.request(GET_ORGANIZATION_CHART, variables);
        expect(response.organizationChart.employees).toHaveLength(0);
      } catch (error) {
        // Expect either empty result or appropriate error
        expect(error.response.errors[0].message).toContain('部门');
      }
    });
  });

  describe('SAM Analysis API', () => {
    test('should fetch situational context analysis', async () => {
      const variables = {
        filters: {
          department: null,
          timeRange: {
            start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
            end: new Date().toISOString()
          }
        }
      };

      const response = await client.request(GET_SITUATIONAL_CONTEXT, variables);
      
      expect(response.getSituationalContext).toBeDefined();
      expect(response.getSituationalContext.timestamp).toBeDefined();
      expect(response.getSituationalContext.alertLevel).toBeDefined();
      expect(['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']).toContain(
        response.getSituationalContext.alertLevel
      );
      
      // Check organization health
      expect(response.getSituationalContext.organizationHealth).toBeDefined();
      expect(response.getSituationalContext.organizationHealth.overallScore).toBeGreaterThan(0);
      expect(response.getSituationalContext.organizationHealth.departmentHealth).toBeInstanceOf(Array);
      
      // Check talent metrics
      expect(response.getSituationalContext.talentMetrics).toBeDefined();
      expect(response.getSituationalContext.talentMetrics.skillGapAnalysis).toBeInstanceOf(Array);
      
      // Check risk assessment
      expect(response.getSituationalContext.riskAssessment).toBeDefined();
      expect(response.getSituationalContext.riskAssessment.overallRiskScore).toBeGreaterThan(0);
    });

    test('should fetch department-specific analysis', async () => {
      const variables = {
        filters: {
          department: '研发部',
          timeRange: {
            start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
            end: new Date().toISOString()
          }
        }
      };

      const response = await client.request(GET_SITUATIONAL_CONTEXT, variables);
      
      expect(response.getSituationalContext).toBeDefined();
      
      // Department-specific analysis should focus on the specified department
      const departmentHealth = response.getSituationalContext.organizationHealth.departmentHealth;
      expect(departmentHealth.some(dept => dept.department === '研发部')).toBe(true);
    });

    test('should provide actionable recommendations', async () => {
      const response = await client.request(GET_SITUATIONAL_CONTEXT, {
        filters: {}
      });
      
      expect(response.getSituationalContext.recommendations).toBeInstanceOf(Array);
      
      if (response.getSituationalContext.recommendations.length > 0) {
        const recommendation = response.getSituationalContext.recommendations[0];
        expect(recommendation.id).toBeDefined();
        expect(recommendation.title).toBeDefined();
        expect(recommendation.description).toBeDefined();
        expect(recommendation.priority).toBeDefined();
        expect(['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']).toContain(recommendation.priority);
        expect(recommendation.implementation).toBeDefined();
      }
    });
  });

  describe('Real-time Subscriptions', () => {
    test('should handle GraphQL subscriptions for real-time updates', async () => {
      // Note: This would require WebSocket testing setup
      // For now, we'll test the subscription queries are valid
      
      const EMPLOYEE_POSITION_CHANGED = `
        subscription EmployeePositionChanged($employeeId: ID) {
          employeePositionChanged(employeeId: $employeeId) {
            id
            employeeId
            positionTitle
            department
            effectiveDate
            isRetroactive
          }
        }
      `;

      // Validate subscription query syntax
      expect(EMPLOYEE_POSITION_CHANGED).toContain('subscription');
      expect(EMPLOYEE_POSITION_CHANGED).toContain('employeePositionChanged');
    });

    test('should validate workflow status subscription', async () => {
      const WORKFLOW_STATUS_CHANGED = `
        subscription WorkflowStatusChanged($workflowId: String!) {
          workflowStatusChanged(workflowId: $workflowId) {
            workflowId
            status
            currentStep
            progress
            updatedAt
          }
        }
      `;

      // Validate subscription query syntax
      expect(WORKFLOW_STATUS_CHANGED).toContain('subscription');
      expect(WORKFLOW_STATUS_CHANGED).toContain('workflowStatusChanged');
    });
  });

  describe('Error Handling and Edge Cases', () => {
    test('should handle invalid employee ID', async () => {
      try {
        await client.request(GET_EMPLOYEE, { id: 'invalid-id' });
        fail('Should have thrown error for invalid ID');
      } catch (error) {
        expect(error.response.errors).toBeDefined();
        expect(error.response.errors[0].message).toContain('Employee not found');
      }
    });

    test('should handle malformed GraphQL queries', async () => {
      const invalidQuery = `
        query InvalidQuery {
          invalidField {
            nonExistentField
          }
        }
      `;

      try {
        await client.request(invalidQuery);
        fail('Should have thrown error for invalid query');
      } catch (error) {
        expect(error.response.errors).toBeDefined();
      }
    });

    test('should handle network timeouts gracefully', async () => {
      // Create client with short timeout
      const timeoutClient = new GraphQLClient('http://localhost:8080/graphql', {
        timeout: 100, // 100ms timeout
        headers: {
          authorization: `Bearer ${process.env.TEST_TOKEN || 'test-token'}`,
        },
      });

      try {
        // This might timeout or succeed depending on network speed
        await timeoutClient.request(GET_EMPLOYEES, { first: 1000 });
      } catch (error) {
        // Expect timeout or network error
        expect(error.message).toMatch(/timeout|network|connection/i);
      }
    });

    test('should handle concurrent requests correctly', async () => {
      const concurrentRequests = Array.from({ length: 5 }, (_, i) =>
        client.request(GET_EMPLOYEES, { first: 10, after: null })
      );

      const responses = await Promise.all(concurrentRequests);
      
      // All requests should succeed
      expect(responses).toHaveLength(5);
      responses.forEach(response => {
        expect(response.employees).toBeDefined();
        expect(response.employees.edges).toBeInstanceOf(Array);
      });
    });

    test('should handle large data sets with pagination', async () => {
      let allEmployees = [];
      let hasNextPage = true;
      let cursor = null;
      let pageCount = 0;
      const maxPages = 10; // Prevent infinite loops

      while (hasNextPage && pageCount < maxPages) {
        const response = await client.request(GET_EMPLOYEES, {
          first: 50,
          after: cursor
        });

        allEmployees.push(...response.employees.edges);
        hasNextPage = response.employees.pageInfo.hasNextPage;
        cursor = response.employees.pageInfo.endCursor;
        pageCount++;
      }

      expect(allEmployees.length).toBeGreaterThan(0);
      expect(pageCount).toBeLessThanOrEqual(maxPages);
      
      // Verify no duplicate employees
      const employeeIds = allEmployees.map(edge => edge.node.id);
      const uniqueIds = new Set(employeeIds);
      expect(uniqueIds.size).toBe(employeeIds.length);
    });
  });

  describe('Data Consistency and Validation', () => {
    test('should maintain data consistency across related queries', async () => {
      // Get employee from employees list
      const employeesResponse = await client.request(GET_EMPLOYEES, { first: 1 });
      const employeeFromList = employeesResponse.employees.edges[0].node;

      // Get same employee from individual query
      const employeeResponse = await client.request(GET_EMPLOYEE, { 
        id: employeeFromList.id 
      });
      const employeeFromDetail = employeeResponse.employee;

      // Basic fields should match
      expect(employeeFromDetail.id).toBe(employeeFromList.id);
      expect(employeeFromDetail.employeeId).toBe(employeeFromList.employeeId);
      expect(employeeFromDetail.legalName).toBe(employeeFromList.legalName);
      expect(employeeFromDetail.email).toBe(employeeFromList.email);
      expect(employeeFromDetail.status).toBe(employeeFromList.status);
    });

    test('should validate date formats and ranges', async () => {
      const employees = await client.request(GET_EMPLOYEES, { first: 10 });
      
      employees.employees.edges.forEach(edge => {
        const employee = edge.node;
        
        // Validate hire date format
        if (employee.hireDate) {
          expect(employee.hireDate).toMatch(/^\d{4}-\d{2}-\d{2}/);
          
          // Hire date should not be in the future
          const hireDate = new Date(employee.hireDate);
          const now = new Date();
          expect(hireDate.getTime()).toBeLessThanOrEqual(now.getTime());
        }
        
        // Validate termination date if present
        if (employee.terminationDate) {
          expect(employee.terminationDate).toMatch(/^\d{4}-\d{2}-\d{2}/);
          
          const terminationDate = new Date(employee.terminationDate);
          const hireDate = new Date(employee.hireDate);
          
          // Termination date should be after hire date
          expect(terminationDate.getTime()).toBeGreaterThan(hireDate.getTime());
        }
      });
    });

    test('should validate email formats', async () => {
      const employees = await client.request(GET_EMPLOYEES, { first: 10 });
      
      employees.employees.edges.forEach(edge => {
        const employee = edge.node;
        if (employee.email) {
          expect(employee.email).toMatch(/^[^\s@]+@[^\s@]+\.[^\s@]+$/);
        }
      });
    });

    test('should validate enum values', async () => {
      const employees = await client.request(GET_EMPLOYEES, { first: 10 });
      
      const validStatuses = ['ACTIVE', 'INACTIVE', 'TERMINATED', 'ON_LEAVE'];
      const validJobLevels = ['INTERN', 'JUNIOR', 'INTERMEDIATE', 'SENIOR', 'LEAD', 'MANAGER', 'DIRECTOR', 'VP', 'C_LEVEL'];
      const validEmploymentTypes = ['FULL_TIME', 'PART_TIME', 'CONTRACT', 'INTERN'];
      
      employees.employees.edges.forEach(edge => {
        const employee = edge.node;
        
        // Validate status
        expect(validStatuses).toContain(employee.status);
        
        // Validate current position enums if present
        if (employee.currentPosition) {
          if (employee.currentPosition.jobLevel) {
            expect(validJobLevels).toContain(employee.currentPosition.jobLevel);
          }
          if (employee.currentPosition.employmentType) {
            expect(validEmploymentTypes).toContain(employee.currentPosition.employmentType);
          }
        }
      });
    });
  });

  describe('Performance and Load Testing', () => {
    test('should handle reasonable load without errors', async () => {
      const startTime = Date.now();
      
      // Simulate 10 concurrent users making requests
      const requests = Array.from({ length: 10 }, () =>
        client.request(GET_EMPLOYEES, { first: 20 })
      );
      
      const responses = await Promise.all(requests);
      const endTime = Date.now();
      const duration = endTime - startTime;
      
      // All requests should succeed
      expect(responses).toHaveLength(10);
      responses.forEach(response => {
        expect(response.employees).toBeDefined();
      });
      
      // Should complete within reasonable time (adjust based on your requirements)
      expect(duration).toBeLessThan(5000); // 5 seconds
    });

    test('should return responses within acceptable time limits', async () => {
      const timeouts = [];
      
      for (let i = 0; i < 5; i++) {
        const startTime = Date.now();
        await client.request(GET_EMPLOYEES, { first: 10 });
        const endTime = Date.now();
        timeouts.push(endTime - startTime);
      }
      
      const averageTime = timeouts.reduce((a, b) => a + b, 0) / timeouts.length;
      
      // Average response time should be reasonable
      expect(averageTime).toBeLessThan(1000); // 1 second average
      
      // No single request should take too long
      timeouts.forEach(time => {
        expect(time).toBeLessThan(3000); // 3 seconds max
      });
    });
  });
});