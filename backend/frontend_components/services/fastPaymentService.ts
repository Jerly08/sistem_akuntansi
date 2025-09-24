import { API_V1_BASE } from '@/config/api';

export interface LightweightPaymentRequest {
  sale_id: number;
  amount: number;
  payment_date: string;
  method: string;
  cash_bank_id: number;
  reference: string;
  notes: string;
}

export interface LightweightPaymentResponse {
  success: boolean;
  payment_id: number;
  payment_code: string;
  amount: number;
  new_status: string;
  outstanding_amount: number;
  processing_time: string;
  message: string;
}

class FastPaymentService {
  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  // Fast payment recording with async journal processing (recommended)
  async recordPaymentAsync(data: LightweightPaymentRequest): Promise<LightweightPaymentResponse> {
    console.log('🚀 Recording payment asynchronously:', data);
    
    const response = await fetch(`${API_V1_BASE}/payments/fast/record-async`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Network error' }));
      throw new Error(error.message || error.details || 'Failed to record payment');
    }

    const result = await response.json();
    console.log('✅ Payment recorded successfully:', result);
    return result;
  }

  // Fast payment recording with synchronous processing
  async recordPaymentFast(data: LightweightPaymentRequest): Promise<LightweightPaymentResponse> {
    console.log('⚡ Recording payment synchronously:', data);
    
    const response = await fetch(`${API_V1_BASE}/payments/fast/record`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Network error' }));
      throw new Error(error.message || error.details || 'Failed to record payment');
    }

    const result = await response.json();
    console.log('✅ Payment recorded successfully:', result);
    return result;
  }

  // Validate payment data without recording
  async validatePayment(data: LightweightPaymentRequest): Promise<{ valid: boolean; message: string }> {
    console.log('🔍 Validating payment data:', data);
    
    try {
      const response = await fetch(`${API_V1_BASE}/payments/fast/validate`, {
        method: 'POST',
        headers: this.getAuthHeaders(),
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json().catch(() => ({ message: 'Validation failed' }));
        throw new Error(error.message || error.details || 'Validation failed');
      }

      const result = await response.json();
      console.log('✅ Validation result:', result);
      return result;
    } catch (err: any) {
      console.log('⚠️ Validation warning:', err.message);
      return { valid: false, message: err.message };
    }
  }

  // Get payment status
  async getPaymentStatus(paymentId: number): Promise<any> {
    const response = await fetch(`${API_V1_BASE}/payments/fast/status/${paymentId}`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to get payment status' }));
      throw new Error(error.message || 'Failed to get payment status');
    }

    return await response.json();
  }

  // Service health check
  async checkServiceHealth(): Promise<any> {
    try {
      const response = await fetch(`${API_V1_BASE}/payments/fast/health`);
      
      if (!response.ok) {
        return { status: 'unhealthy', message: 'Service unavailable' };
      }

      return await response.json();
    } catch (err) {
      return { status: 'unhealthy', message: 'Network error' };
    }
  }

  // Sales-specific fast payment recording
  async recordSalesPayment(
    saleId: number, 
    paymentData: Omit<LightweightPaymentRequest, 'sale_id'>
  ): Promise<LightweightPaymentResponse> {
    console.log(`🚀 Recording payment for sale ${saleId}:`, paymentData);
    
    const response = await fetch(`${API_V1_BASE}/sales/fast-payment/${saleId}/record`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(paymentData),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Network error' }));
      throw new Error(error.message || error.details || 'Failed to record payment');
    }

    const result = await response.json();
    console.log('✅ Sales payment recorded successfully:', result);
    return result;
  }

  // Bulk payment validation (for multiple payments)
  async validateBulkPayments(payments: LightweightPaymentRequest[]): Promise<any[]> {
    const validationPromises = payments.map(payment => 
      this.validatePayment(payment).catch(err => ({ 
        valid: false, 
        message: err.message,
        sale_id: payment.sale_id 
      }))
    );
    
    return await Promise.all(validationPromises);
  }

  // Get optimal payment method based on amount and account
  getOptimalPaymentMethod(amount: number): string {
    if (amount >= 10000000) { // 10 million IDR
      return 'BANK_TRANSFER'; // Large amounts should use bank transfer
    } else if (amount >= 1000000) { // 1 million IDR
      return 'BANK_TRANSFER'; // Medium amounts prefer bank transfer
    } else {
      return 'CASH'; // Small amounts can use cash
    }
  }

  // Format payment amount for display
  formatPaymentAmount(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  // Calculate processing fee (if applicable)
  calculateProcessingFee(amount: number, method: string): number {
    switch (method) {
      case 'CREDIT_CARD':
        return amount * 0.025; // 2.5% for credit card
      case 'BANK_TRANSFER':
        return amount > 1000000 ? 6500 : 2500; // Fixed fee based on amount
      case 'CHECK':
        return 1000; // Fixed fee for check
      case 'CASH':
      default:
        return 0; // No fee for cash
    }
  }
}

export const fastPaymentService = new FastPaymentService();
export default fastPaymentService;