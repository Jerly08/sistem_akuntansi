import api from './api';

export interface ApprovalRequest {
  id: number;
  request_code: string;
  workflow_id: number;
  requester_id: number;
  entity_type: string;
  entity_id: number;
  amount: number;
  status: string;
  priority: string;
  request_title: string;
  request_message?: string;
  reject_reason?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  workflow: {
    id: number;
    name: string;
    module: string;
  };
  requester: {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
    name: string;
  };
  approval_steps: ApprovalAction[];
}

export interface ApprovalAction {
  id: number;
  request_id: number;
  step_id: number;
  approver_id?: number;
  status: string;
  comments?: string;
  action_date?: string;
  is_active: boolean;
  step: {
    id: number;
    step_order: number;
    step_name: string;
    approver_role: string;
    time_limit?: number; // hours until due
  };
  // When this step became active; used for SLA countdown if provided
  activated_at?: string;
  approver?: {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
  };
}

export interface ApprovalHistory {
  id: number;
  request_id: number;
  user_id: number;
  action: string;
  comments: string;
  created_at: string;
  user: {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
  };
}

export interface Purchase {
  id: number;
  code: string;
  vendor_id: number;
  user_id: number;
  date: string;
  due_date?: string;
  total_amount: number;
  discount: number;
  tax: number;
  status: string;
  notes?: string;
  approval_status: string;
  requires_approval: boolean;
  approved_at?: string;
  vendor?: {
    id: number;
    name: string;
  };
  user?: {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
    name: string;
  };
}

export interface ApprovalStats {
  pending_approvals: number;
  approved_this_month: number;
  rejected_this_month: number;
  total_amount_pending: number;
}

export interface PendingApprovalsResponse {
  purchases: Purchase[];
  total: number;
  page: number;
  limit: number;
}

class ApprovalService {
  // Get purchases pending approval for current user
  async getPurchasesForApproval(params: { page?: number; limit?: number } = {}): Promise<PendingApprovalsResponse> {
    const response = await api.get('/purchases/pending-approval', { params });
    return response.data;
  }

  // Approve a purchase
  async approvePurchase(purchaseId: number, data: { comments?: string; escalate_to_director?: boolean }): Promise<{ message: string; purchase_id: number; escalated?: boolean; status?: string; approval_status?: string }> {
    const response = await api.post(`/purchases/${purchaseId}/approve`, data);
    return response.data;
  }

  // Reject a purchase
  async rejectPurchase(purchaseId: number, data: { comments: string }): Promise<{ message: string; purchase_id: number; comments: string }> {
    const response = await api.post(`/purchases/${purchaseId}/reject`, data);
    return response.data;
  }

  // Get approval history for a purchase
  async getApprovalHistory(purchaseId: number): Promise<{ purchase_id: number; approval_history: ApprovalHistory[] }> {
    const response = await api.get(`/purchases/${purchaseId}/approval-history`);
    return response.data;
  }

  // Get approval statistics
  async getApprovalStats(): Promise<ApprovalStats> {
    const response = await api.get('/purchases/approval-stats');
    return response.data;
  }

  // Submit purchase for approval
  async submitPurchaseForApproval(purchaseId: number): Promise<{ message: string; purchase_id: number }> {
    const response = await api.post(`/purchases/${purchaseId}/submit-approval`);
    return response.data;
  }

  // SALES APPROVAL METHODS
  
  // Submit sale for approval
  async submitSaleForApproval(saleId: number, data?: { comments?: string; priority?: string }): Promise<ApprovalRequest> {
    const response = await api.post(`/sales/${saleId}/submit-approval`, data || {});
    return response.data;
  }

  // Get approval status for a sale
  async getSaleApprovalStatus(saleId: number): Promise<ApprovalRequest | null> {
    try {
      const response = await api.get(`/sales/${saleId}/approval`);
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 404) {
        return null;
      }
      throw error;
    }
  }

  // Get sales pending approval
  async getSalesForApproval(params: { page?: number; limit?: number; status?: string; priority?: string } = {}): Promise<any> {
    const response = await api.get('/sales/pending-approval', { params });
    return response.data;
  }

  // Approve a sale step
  async approveSaleStep(approvalId: number, stepId: number, data: { comments?: string }): Promise<void> {
    await api.post(`/approval-requests/${approvalId}/steps/${stepId}/approve`, data);
  }

  // Reject a sale step
  async rejectSaleStep(approvalId: number, stepId: number, data: { comments?: string }): Promise<void> {
    await api.post(`/approval-requests/${approvalId}/steps/${stepId}/reject`, data);
  }

  // Get sale approval history
  async getSaleApprovalHistory(saleId: number): Promise<{ sale_id: number; approval_history: ApprovalHistory[] }> {
    const response = await api.get(`/sales/${saleId}/approval-history`);
    return response.data;
  }

  // Get all approval requests (generic)
  async getApprovalRequests(params: { 
    page?: number; 
    limit?: number; 
    status?: string; 
    entity_type?: string;
    priority?: string;
    requester_id?: number;
    my_approvals?: boolean;
    date_from?: string;
    date_to?: string;
  } = {}): Promise<any> {
    const response = await api.get('/approval-requests', { params });
    return response.data;
  }

  // Get approval request by ID
  async getApprovalRequest(requestId: number): Promise<ApprovalRequest> {
    const response = await api.get(`/approval-requests/${requestId}`);
    return response.data;
  }

  // Cancel approval request
  async cancelApprovalRequest(requestId: number, data: { reason?: string } = {}): Promise<void> {
    await api.post(`/approval-requests/${requestId}/cancel`, data);
  }

  // Escalate approval step
  async escalateApprovalStep(approvalId: number, stepId: number, data: { comments?: string; escalate_to?: number }): Promise<void> {
    await api.post(`/approval-requests/${approvalId}/steps/${stepId}/escalate`, data);
  }

  // Get pending approvals for current user (any entity type)
  async getMyPendingApprovals(): Promise<ApprovalRequest[]> {
    const response = await api.get('/approval-requests/my-pending');
    return response.data.requests || [];
  }

  // Get approval summary/dashboard data
  async getApprovalSummary(): Promise<{
    total_pending: number;
    total_approved: number;
    total_rejected: number;
    my_pending: number;
    urgent_count: number;
    overdue_count: number;
    by_entity_type: Record<string, number>;
    by_status: Record<string, number>;
  }> {
    const response = await api.get('/approval-requests/summary');
    return response.data;
  }

  // Get approval workflows
  async getApprovalWorkflows(module?: string): Promise<any[]> {
    const response = await api.get('/approval-workflows', { params: { module } });
    return response.data.workflows || [];
  }

  // Get notifications
  async getNotifications(params: { page?: number; limit?: number; type?: string } = {}): Promise<any> {
    const response = await api.get('/notifications', { params });
    return response.data;
  }

  // Mark notification as read
  async markNotificationAsRead(notificationId: number): Promise<void> {
    await api.put(`/notifications/${notificationId}/read`);
  }

  // Get unread notification count
  async getUnreadNotificationCount(): Promise<{ count: number }> {
    const response = await api.get('/notifications/unread-count');
    return response.data;
  }

  // Simplified sales approval methods (for compatibility)
  async approveSale(saleId: number, data: { comments?: string }): Promise<{ message: string; sale_id: number }> {
    const response = await api.post(`/sales/${saleId}/approve`, data);
    return response.data;
  }

  async rejectSale(saleId: number, data: { comments: string }): Promise<{ message: string; sale_id: number }> {
    const response = await api.post(`/sales/${saleId}/reject`, data);
    return response.data;
  }
}

const approvalService = new ApprovalService();
export default approvalService;
