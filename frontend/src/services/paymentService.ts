import api from './api';

// Types
export interface Payment {
  id: number;
  code: string;
  contact_id: number;
  contact?: {
    id: number;
    name: string;
    type: 'CUSTOMER' | 'VENDOR';
  };
  user_id: number;
  date: string;
  amount: number;
  method: string;
  reference: string;
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface PaymentFilters {
  page?: number;
  limit?: number;
  contact_id?: number;
  status?: string;
  method?: string;
  start_date?: string;
  end_date?: string;
}

export interface PaymentResult {
  data: Payment[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface PaymentCreateRequest {
  contact_id: number;
  cash_bank_id?: number;
  date: string;
  amount: number;
  method: string;
  reference?: string;
  notes?: string;
  allocations?: PaymentAllocation[];
  bill_allocations?: BillAllocation[];
}

export interface PaymentAllocation {
  invoice_id: number;
  amount: number;
}

export interface BillAllocation {
  bill_id: number;
  amount: number;
}

export interface PaymentSummary {
  total_received: number;
  total_paid: number;
  net_flow: number;
  by_method: Record<string, number>;
  status_counts: Record<string, number>;
}

export interface PaymentAnalytics {
  total_received: number;
  total_paid: number;
  net_flow: number;
  received_growth: number;
  paid_growth: number;
  flow_growth: number;
  total_outstanding: number;
  by_method: Record<string, number>;
  daily_trend: Array<{
    date: string;
    received: number;
    paid: number;
  }>;
  recent_payments: Payment[];
  avg_payment_time: number;
  success_rate: number;
}

class PaymentService {
  private readonly baseUrl = '/payments';

  // Get all payments with filters
  async getPayments(filters: PaymentFilters = {}): Promise<PaymentResult> {
    try {
      const params = new URLSearchParams();
      
      if (filters.page) params.append('page', filters.page.toString());
      if (filters.limit) params.append('limit', filters.limit.toString());
      if (filters.contact_id) params.append('contact_id', filters.contact_id.toString());
      if (filters.status) params.append('status', filters.status);
      if (filters.method) params.append('method', filters.method);
      if (filters.start_date) params.append('start_date', filters.start_date);
      if (filters.end_date) params.append('end_date', filters.end_date);

      const response = await api.get(`${this.baseUrl}?${params}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching payments:', error);
      throw error;
    }
  }

  // Get payment by ID
  async getPaymentById(id: number): Promise<Payment> {
    try {
      const response = await api.get(`${this.baseUrl}/${id}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching payment:', error);
      throw error;
    }
  }

  // Create receivable payment (from customer)
  async createReceivablePayment(data: PaymentCreateRequest): Promise<Payment> {
    try {
      // Convert date to RFC3339 format (ISO 8601 with timezone)
      const formattedData = {
        ...data,
        date: this.formatDateForAPI(data.date)
      };
      
      const response = await api.post(`${this.baseUrl}/receivable`, formattedData);
      return response.data;
    } catch (error: any) {
      console.error('PaymentService - Error creating receivable payment:', error);
      console.log('PaymentService - Error details:', {
        isAuthError: error.isAuthError,
        code: error.code,
        message: error.message,
        responseStatus: error.response?.status,
        responseData: error.response?.data
      });
      
      // Check if it's an authentication error from API interceptor
      if (error.isAuthError || error.code === 'AUTH_SESSION_EXPIRED' || error.message?.includes('Session expired')) {
        console.log('PaymentService - Detected auth error, throwing auth error');
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        throw authError;
      } else if (error.response?.status === 401) {
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        throw authError;
      } else if (error.response?.status === 403) {
        throw new Error('You do not have permission to create payments.');
      } else if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      } else if (error.response?.data?.details) {
        throw new Error(error.response.data.details);
      } else if (error.message) {
        throw new Error(error.message);
      }
      
      throw error;
    }
  }

  // Create payable payment (to vendor)
  async createPayablePayment(data: PaymentCreateRequest): Promise<Payment> {
    try {
      // Convert date to RFC3339 format (ISO 8601 with timezone)
      const formattedData = {
        ...data,
        date: this.formatDateForAPI(data.date)
      };
      
      const response = await api.post(`${this.baseUrl}/payable`, formattedData);
      return response.data;
    } catch (error: any) {
      console.error('PaymentService - Error creating payable payment:', error);
      console.log('PaymentService - Error details:', {
        isAuthError: error.isAuthError,
        code: error.code,
        message: error.message,
        responseStatus: error.response?.status,
        responseData: error.response?.data
      });
      
      // Check if it's an authentication error from API interceptor
      if (error.isAuthError || error.code === 'AUTH_SESSION_EXPIRED' || error.message?.includes('Session expired')) {
        console.log('PaymentService - Detected auth error, throwing auth error');
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        throw authError;
      } else if (error.response?.status === 401) {
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        throw authError;
      } else if (error.response?.status === 403) {
        throw new Error('You do not have permission to create payments.');
      } else if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      } else if (error.response?.data?.details) {
        throw new Error(error.response.data.details);
      } else if (error.message) {
        throw new Error(error.message);
      }
      
      throw error;
    }
  }

  // Cancel payment
  async cancelPayment(id: number, reason: string): Promise<void> {
    try {
      await api.post(`${this.baseUrl}/${id}/cancel`, { reason });
    } catch (error) {
      console.error('Error cancelling payment:', error);
      throw error;
    }
  }

  // Delete payment
  async deletePayment(id: number | string): Promise<void> {
    try {
      await api.delete(`${this.baseUrl}/${id}`);
    } catch (error: any) {
      console.error('Error deleting payment:', error);
      throw new Error(error.response?.data?.error || 'Failed to delete payment');
    }
  }

  // Export payments to Excel/CSV
  async exportPayments(filters?: PaymentFilters): Promise<Blob> {
    try {
      const params = new URLSearchParams();
      
      if (filters?.page) params.append('page', filters.page.toString());
      if (filters?.limit) params.append('limit', filters.limit.toString());
      if (filters?.contact_id) params.append('contact_id', filters.contact_id.toString());
      if (filters?.status) params.append('status', filters.status);
      if (filters?.method) params.append('method', filters.method);
      if (filters?.start_date) params.append('start_date', filters.start_date);
      if (filters?.end_date) params.append('end_date', filters.end_date);

      const response = await api.get(`${this.baseUrl}/export?${params}`, {
        responseType: 'blob',
      });
      
      return response.data;
    } catch (error: any) {
      console.error('Error exporting payments:', error);
      throw new Error(error.response?.data?.error || 'Failed to export payments');
    }
  }

  // Export payment report to PDF
  async exportPaymentReportPDF(startDate?: string, endDate?: string, status?: string, method?: string): Promise<Blob> {
    try {
      const params = new URLSearchParams();
      
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
      if (status) params.append('status', status);
      if (method) params.append('method', method);

      const response = await api.get(`${this.baseUrl}/report/pdf?${params}`, {
        responseType: 'blob',
      });
      
      return response.data;
    } catch (error: any) {
      console.error('Error exporting payment report PDF:', error);
      throw new Error(error.response?.data?.error || 'Failed to export payment report PDF');
    }
  }

  // Export payment report to Excel
  async exportPaymentReportExcel(startDate?: string, endDate?: string, status?: string, method?: string): Promise<Blob> {
    try {
      const params = new URLSearchParams();
      
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
      if (status) params.append('status', status);
      if (method) params.append('method', method);

      const response = await api.get(`${this.baseUrl}/export/excel?${params}`, {
        responseType: 'blob',
      });
      
      return response.data;
    } catch (error: any) {
      console.error('Error exporting payment report Excel:', error);
      throw new Error(error.response?.data?.error || 'Failed to export payment report Excel');
    }
  }

  // Export payment detail to PDF
  async exportPaymentDetailPDF(paymentId: number): Promise<Blob> {
    try {
      const response = await api.get(`${this.baseUrl}/${paymentId}/pdf`, {
        responseType: 'blob',
      });
      
      return response.data;
    } catch (error: any) {
      console.error('Error exporting payment detail PDF:', error);
      throw new Error(error.response?.data?.error || 'Failed to export payment detail PDF');
    }
  }

  // Download payment report PDF
  async downloadPaymentReportPDF(startDate?: string, endDate?: string, status?: string, method?: string): Promise<void> {
    try {
      const blob = await this.exportPaymentReportPDF(startDate, endDate, status, method);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      const filename = `Payment_Report_${startDate || 'all'}_to_${endDate || 'all'}.pdf`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Error downloading payment report PDF:', error);
      throw error;
    }
  }

  // Download payment detail PDF
  async downloadPaymentDetailPDF(paymentId: number, paymentCode: string): Promise<void> {
    try {
      const blob = await this.exportPaymentDetailPDF(paymentId);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `Payment_${paymentCode}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Error downloading payment detail PDF:', error);
      throw error;
    }
  }

  // Download payment report Excel
  async downloadPaymentReportExcel(startDate?: string, endDate?: string, status?: string, method?: string): Promise<void> {
    try {
      const blob = await this.exportPaymentReportExcel(startDate, endDate, status, method);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      const filename = `Payment_Report_${startDate || 'all'}_to_${endDate || 'all'}.xlsx`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Error downloading payment report Excel:', error);
      throw error;
    }
  }

  // Get unpaid invoices for customer
  async getUnpaidInvoices(customerId: number): Promise<any[]> {
    try {
      const response = await api.get(`${this.baseUrl}/unpaid-invoices/${customerId}`);
      return response.data || [];
    } catch (error) {
      console.error('Error fetching unpaid invoices:', error);
      return [];
    }
  }

  // Get unpaid bills for vendor
  async getUnpaidBills(vendorId: number): Promise<any[]> {
    try {
      const response = await api.get(`${this.baseUrl}/unpaid-bills/${vendorId}`);
      return response.data || [];
    } catch (error) {
      console.error('Error fetching unpaid bills:', error);
      return [];
    }
  }

  // Get payment summary
  async getPaymentSummary(startDate: string, endDate: string): Promise<PaymentSummary> {
    try {
      const response = await api.get(`${this.baseUrl}/summary?start_date=${startDate}&end_date=${endDate}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching payment summary:', error);
      throw error;
    }
  }

  // Get payment analytics
  async getPaymentAnalytics(startDate: string, endDate: string): Promise<PaymentAnalytics> {
    try {
      const response = await api.get(`${this.baseUrl}/analytics?start_date=${startDate}&end_date=${endDate}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching payment analytics:', error);
      throw error;
    }
  }

  // Generate payment report
  async generateReport(reportType: 'cash_flow' | 'aging' | 'method_analysis', params: any): Promise<any> {
    try {
      const response = await api.post(`${this.baseUrl}/reports/${reportType}`, params);
      return response.data;
    } catch (error) {
      console.error('Error generating report:', error);
      throw error;
    }
  }

  // Bulk payment processing
  async processBulkPayments(payments: PaymentCreateRequest[]): Promise<any> {
    try {
      const response = await api.post(`${this.baseUrl}/bulk`, { payments });
      return response.data;
    } catch (error) {
      console.error('Error processing bulk payments:', error);
      throw error;
    }
  }

  // Format currency for display
  formatCurrency(amount: number, currency: string = 'IDR'): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(amount);
  }

  // Format date for display
  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    });
  }

  // Format date time for display
  formatDateTime(dateString: string): string {
    return new Date(dateString).toLocaleString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  // Get status color scheme for badges
  getStatusColorScheme(status: string): string {
    switch (status.toUpperCase()) {
      case 'COMPLETED':
        return 'green';
      case 'PENDING':
        return 'yellow';
      case 'FAILED':
        return 'red';
      default:
        return 'gray';
    }
  }

  // Get method display name
  getMethodDisplayName(method: string): string {
    const methodMap: Record<string, string> = {
      'CASH': 'Tunai',
      'BANK_TRANSFER': 'Transfer Bank',
      'CHECK': 'Cek',
      'CREDIT_CARD': 'Kartu Kredit',
      'DEBIT_CARD': 'Kartu Debit',
      'OTHER': 'Lainnya'
    };
    
    return methodMap[method] || method;
  }

  // Format date for API (convert to RFC3339/ISO 8601 with timezone)
  formatDateForAPI(dateString: string): string {
    // If date is already in ISO format, return as is
    if (dateString.includes('T')) {
      return dateString;
    }
    
    // Convert YYYY-MM-DD to YYYY-MM-DDTHH:mm:ssZ (assume local timezone)
    const date = new Date(dateString + 'T00:00:00');
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      throw new Error('Invalid date format');
    }
    
    // Return in ISO format with local timezone
    return date.toISOString();
  }

  // Validate payment data
  validatePaymentData(data: PaymentCreateRequest): string[] {
    const errors: string[] = [];

    if (!data.contact_id) {
      errors.push('Contact is required');
    }

    if (!data.amount || data.amount <= 0) {
      errors.push('Amount must be greater than zero');
    }

    if (!data.date) {
      errors.push('Payment date is required');
    }

    if (!data.method) {
      errors.push('Payment method is required');
    }

    // Check if date is not in the future
    if (data.date && new Date(data.date) > new Date()) {
      errors.push('Payment date cannot be in the future');
    }

    return errors;
  }

  // Calculate total allocation amount
  calculateAllocationTotal(allocations: PaymentAllocation[]): number {
    return allocations.reduce((total, allocation) => total + allocation.amount, 0);
  }
}

export default new PaymentService();
