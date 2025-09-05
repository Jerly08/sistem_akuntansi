import api from './api';

export interface PurchaseItemRequest {
  product_id: number;
  quantity: number;
  unit_price: number;
  discount?: number;
  tax?: number;
  expense_account_id?: number;
  description?: string;
}

export interface PurchaseCreateRequest {
  vendor_id: number;
  date: string; // ISO string
  due_date?: string; // ISO string
  discount?: number;
  tax?: number;
  notes?: string;
  items: PurchaseItemRequest[]; // backend expects `items`
  request_priority?: 'LOW' | 'NORMAL' | 'HIGH' | 'URGENT';
}

export interface PurchaseFilterParams {
  status?: string;
  vendor_id?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
  approval_status?: string;
  requires_approval?: boolean;
  page?: number;
  limit?: number;
}

export interface PurchaseItem {
  id: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  discount: number;
  tax: number;
  total_price: number;
  product?: {
    id: number;
    name: string;
    code: string;
  };
}

export interface Vendor {
  id: number;
  name: string;
  code: string;
  email?: string;
  phone?: string;
}

export interface Purchase {
  id: number;
  code: string;
  vendor_id: number;
  user_id: number;
  date: string;
  due_date?: string;
  subtotal_before_discount: number;
  item_discount_amount: number;
  discount: number; // order-level percent
  order_discount_amount: number;
  net_before_tax: number;
  tax_amount: number;
  total_amount: number; // grand total
  status: string;
  notes?: string;
  approval_status: string;
  approval_amount_basis?: 'SUBTOTAL_BEFORE_DISCOUNT' | 'NET_AFTER_DISCOUNT_BEFORE_TAX' | 'GRAND_TOTAL_AFTER_TAX';
  approval_base_amount?: number;
  requires_approval: boolean;
  approval_request_id?: number;
  approved_at?: string;
  approved_by?: number;
  vendor?: Vendor;
  purchase_items?: PurchaseItem[];
  approval_request?: {
    id: number;
    status: string;
    approval_steps: Array<{
      id: number;
      step_id: number;
      status: string;
      is_active: boolean;
      step: {
        id: number;
        step_order: number;
        step_name: string;
        approver_role: string;
      };
    }>;
  };
  created_at: string;
  updated_at: string;
}

export interface PurchaseListResponse {
  data: Purchase[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface PurchaseSummary {
  total_purchases: number;
  total_amount: number;
  total_approved_amount: number;
  total_paid: number;
  total_outstanding: number;
  avg_order_value: number;
  status_counts: { [key: string]: number };
  approval_status_counts: { [key: string]: number };
}

class PurchaseService {
  async list(params: PurchaseFilterParams): Promise<PurchaseListResponse> {
    const toUpper = (v?: string) => (v ? v.toUpperCase() : undefined);
    const response = await api.get('/purchases', { params: {
      status: toUpper(params.status),
      vendor_id: params.vendor_id,
      start_date: params.start_date,
      end_date: params.end_date,
      search: params.search,
      approval_status: toUpper(params.approval_status),
      requires_approval: params.requires_approval,
      page: params.page ?? 1,
      limit: params.limit ?? 10,
    }});
    return response.data;
  }

  async create(payload: PurchaseCreateRequest): Promise<Purchase> {
    const response = await api.post('/purchases', payload);
    return response.data;
  }

  async submitForApproval(id: number): Promise<{ message: string }> {
    const response = await api.post(`/purchases/${id}/submit-approval`);
    return response.data;
  }

  async getById(id: number): Promise<Purchase> {
    const response = await api.get(`/purchases/${id}`);
    return response.data;
  }

  async update(id: number, payload: Partial<PurchaseCreateRequest>): Promise<Purchase> {
    const response = await api.put(`/purchases/${id}`, payload);
    return response.data;
  }

  async delete(id: number): Promise<{ message: string }> {
    const response = await api.delete(`/purchases/${id}`);
    return response.data;
  }

  async getPendingApproval(page = 1, limit = 10): Promise<PurchaseListResponse> {
    return this.list({
      approval_status: 'PENDING',
      page,
      limit,
    });
  }

  async getSummary(startDate?: string, endDate?: string): Promise<PurchaseSummary> {
    const response = await api.get('/purchases/summary', {
      params: {
        start_date: startDate,
        end_date: endDate,
      },
    });
    return response.data;
  }
}

const purchaseService = new PurchaseService();
export default purchaseService;
