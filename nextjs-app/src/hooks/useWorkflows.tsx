// src/hooks/useWorkflows.tsx
import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useSubscription } from '@apollo/client';
import { 
  GET_WORKFLOW_STATUS,
  APPROVE_POSITION_CHANGE,
  REJECT_POSITION_CHANGE,
  WORKFLOW_STATUS_CHANGED,
  POSITION_CHANGE_APPROVAL_REQUIRED
} from '@/lib/graphql-queries';
import { toast } from 'react-hot-toast';
import { Button } from '@/components/ui/button';
import { createElement } from 'react';

export interface WorkflowStatus {
  workflowId: string;
  status: 'PENDING' | 'IN_PROGRESS' | 'APPROVED' | 'REJECTED' | 'COMPLETED' | 'FAILED';
  currentStep?: string;
  progress: number;
  startedAt: string;
  updatedAt: string;
  completedAt?: string;
  error?: string;
}

export interface ApprovalRequest {
  workflowId: string;
  employeeId: string;
  currentPosition?: {
    positionTitle: string;
    department: string;
  };
  newPosition: {
    positionTitle: string;
    department: string;
    minSalary?: number;
    maxSalary?: number;
    currency?: string;
  };
  requestedBy: string;
  requestedAt: string;
  dueDate: string;
  priority: 'LOW' | 'NORMAL' | 'HIGH' | 'URGENT';
}

// Hook for fetching workflow status with real-time updates
export const useWorkflowStatus = (workflowId: string) => {
  const [workflowStatus, setWorkflowStatus] = useState<WorkflowStatus | null>(null);

  const { data, loading, error, refetch } = useQuery(GET_WORKFLOW_STATUS, {
    variables: { workflowId },
    skip: !workflowId,
    errorPolicy: 'all',
    pollInterval: 5000, // Poll every 5 seconds as fallback
  });

  // Real-time subscription for workflow status changes
  useSubscription(WORKFLOW_STATUS_CHANGED, {
    variables: { workflowId },
    skip: !workflowId,
    onData: ({ data: subscriptionData }) => {
      if (subscriptionData?.data?.workflowStatusChanged) {
        const newStatus = subscriptionData.data.workflowStatusChanged;
        setWorkflowStatus(newStatus);
        
        // Show notification for status changes
        if (newStatus.status === 'COMPLETED') {
          toast.success('工作流已完成');
        } else if (newStatus.status === 'FAILED') {
          toast.error(newStatus.error || '工作流执行过程中发生错误');
        } else if (newStatus.status === 'APPROVED') {
          toast.success('审批通过');
        } else if (newStatus.status === 'REJECTED') {
          toast.error('审批被拒');
        }
      }
    },
    onError: (error) => {
      // Workflow subscription error - falling back to polling
    },
  });

  useEffect(() => {
    if (data?.workflowStatus) {
      setWorkflowStatus(data.workflowStatus);
    }
  }, [data]);

  return {
    workflowStatus,
    loading,
    error,
    refetch,
  };
};

// Hook for handling position change approvals
export const usePositionChangeApproval = () => {
  const [approvePositionChange, { loading: approveLoading }] = useMutation(APPROVE_POSITION_CHANGE);
  const [rejectPositionChange, { loading: rejectLoading }] = useMutation(REJECT_POSITION_CHANGE);

  const approve = async (workflowId: string, comments?: string) => {
    try {
      const result = await approvePositionChange({
        variables: { workflowId, comments },
      });

      const data = result.data?.approvePositionChange;
      
      if (data?.errors && data.errors.length > 0) {
        const errorMessages = data.errors.map((err: any) => err.message).join(', ');
        throw new Error(errorMessages);
      }

      toast.success('审批成功');

      return {
        success: true,
        workflowId: data?.workflowId,
      };
    } catch (err: any) {
      toast.error(`审批失败: ${err.message || '处理审批时发生错误'}`);
      
      return {
        success: false,
        error: err.message,
      };
    }
  };

  const reject = async (workflowId: string, reason: string) => {
    try {
      const result = await rejectPositionChange({
        variables: { workflowId, reason },
      });

      const data = result.data?.rejectPositionChange;
      
      if (data?.errors && data.errors.length > 0) {
        const errorMessages = data.errors.map((err: any) => err.message).join(', ');
        throw new Error(errorMessages);
      }

      toast.success('已拒绝审批');

      return {
        success: true,
        workflowId: data?.workflowId,
      };
    } catch (err: any) {
      toast.error(`操作失败: ${err.message || '处理拒绝时发生错误'}`);
      
      return {
        success: false,
        error: err.message,
      };
    }
  };

  return {
    approve,
    reject,
    loading: approveLoading || rejectLoading,
  };
};

// Hook for monitoring approval requests
export const useApprovalRequests = (approverId: string) => {
  const [approvalRequests, setApprovalRequests] = useState<ApprovalRequest[]>([]);

  // Subscribe to new approval requests
  useSubscription(POSITION_CHANGE_APPROVAL_REQUIRED, {
    variables: { approverId },
    skip: !approverId,
    onData: ({ data: subscriptionData }) => {
      if (subscriptionData?.data?.positionChangeApprovalRequired) {
        const newRequest = subscriptionData.data.positionChangeApprovalRequired;
        
        setApprovalRequests(prev => {
          // Check if request already exists
          const exists = prev.find(req => req.workflowId === newRequest.workflowId);
          if (exists) {
            return prev;
          }
          
          // Add new request
          const updated = [newRequest, ...prev];
          
          // Show notification with toast and custom button
          const toastId = toast.custom(
            (t) => (
              <div className="bg-blue-500 text-white p-4 rounded shadow-lg max-w-md">
                <div className="flex items-start">
                  <div className="flex-1">
                    <h4 className="font-medium">新的审批请求</h4>
                    <p className="text-sm mt-1">员工 {newRequest.employeeId} 的职位变更需要您的审批</p>
                    <div className="mt-3 flex gap-2">
                      <Button 
                        size="sm"
                        onClick={() => {
                          window.location.href = `/workflows/${newRequest.workflowId}`;
                          toast.dismiss(toastId);
                        }}
                        className="bg-white text-blue-500 hover:bg-gray-100"
                      >
                        立即处理
                      </Button>
                      <Button 
                        size="sm" 
                        variant="outline"
                        onClick={() => toast.dismiss(toastId)}
                        className="bg-transparent border-white text-white hover:bg-white hover:text-blue-500"
                      >
                        关闭
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            ),
            {
              duration: Infinity, // Don't auto-close
              id: newRequest.workflowId,
            }
          );
          
          return updated;
        });
      }
    },
    onError: (error) => {
      // Approval subscription error - falling back to polling
    },
  });

  const markAsHandled = (workflowId: string) => {
    setApprovalRequests(prev => 
      prev.filter(req => req.workflowId !== workflowId)
    );
    
    // Close the notification
    toast.dismiss(workflowId);
  };

  return {
    approvalRequests,
    markAsHandled,
    pendingCount: approvalRequests.length,
  };
};

// Hook for workflow statistics
export const useWorkflowStats = () => {
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    inProgress: 0,
    completed: 0,
    failed: 0,
    averageProcessingTime: 0,
    successRate: 0,
  });

  // This would typically fetch from a GraphQL query
  // For now, we'll simulate with mock data
  useEffect(() => {
    // Simulate API call
    const fetchStats = async () => {
      // Mock data - in real app would come from GraphQL
      setStats({
        total: 156,
        pending: 12,
        inProgress: 8,
        completed: 132,
        failed: 4,
        averageProcessingTime: 2.5, // hours
        successRate: 97.4, // percentage
      });
    };

    fetchStats();
  }, []);

  return { stats };
};

// Hook for workflow history
export const useWorkflowHistory = (employeeId?: string, limit: number = 20) => {
  const [workflows, setWorkflows] = useState<WorkflowStatus[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchWorkflowHistory = async () => {
      setLoading(true);
      
      try {
        // This would typically be a GraphQL query
        // For now, we'll simulate with mock data
        const mockWorkflows: WorkflowStatus[] = [
          {
            workflowId: 'wf-001',
            status: 'COMPLETED',
            currentStep: 'completed',
            progress: 100,
            startedAt: '2024-07-25T10:00:00Z',
            updatedAt: '2024-07-25T15:30:00Z',
            completedAt: '2024-07-25T15:30:00Z',
          },
          {
            workflowId: 'wf-002',
            status: 'IN_PROGRESS',
            currentStep: 'hr-approval',
            progress: 60,
            startedAt: '2024-07-28T09:00:00Z',
            updatedAt: '2024-07-28T11:15:00Z',
          },
          {
            workflowId: 'wf-003',
            status: 'REJECTED',
            currentStep: 'manager-approval',
            progress: 30,
            startedAt: '2024-07-20T14:00:00Z',
            updatedAt: '2024-07-20T16:45:00Z',
          },
        ];

        // Filter by employee if specified
        const filteredWorkflows = employeeId 
          ? mockWorkflows.filter(wf => wf.workflowId.includes(employeeId.slice(-3)))
          : mockWorkflows;

        setWorkflows(filteredWorkflows.slice(0, limit));
      } catch (error) {
        // Failed to fetch workflow history - error handled by UI
      } finally {
        setLoading(false);
      }
    };

    fetchWorkflowHistory();
  }, [employeeId, limit]);

  return {
    workflows,
    loading,
  };
};