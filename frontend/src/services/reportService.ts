import { API_BASE_URL } from '@/config/api';

export interface Report {
  id: string;
  name: string;
  description: string;
  type: 'Financial' | 'Operational' | 'Analytical';
  category: string;
  parameters: string[];
}

export interface ReportData {
  id: string;
  title: string;
  type: string;
  period: string;
  generated_at: string;
  data: any;
  summary: { [key: string]: number };
  parameters: { [key: string]: any };
}

export interface ReportParameters {
  start_date?: string;
  end_date?: string;
  as_of_date?: string;
  group_by?: 'month' | 'quarter' | 'year';
  customer_id?: string;
  vendor_id?: string;
  account_code?: string;
  include_valuation?: boolean;
  period?: 'current' | 'ytd' | 'comparative';
  format?: 'json' | 'pdf' | 'excel';
}

class ReportService {
  private getAuthHeaders() {
    const token = localStorage.getItem('token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  private buildQueryString(params: ReportParameters): string {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        searchParams.append(key, value.toString());
      }
    });

    return searchParams.toString();
  }

  // Get list of available reports
  async getAvailableReports(): Promise<Report[]> {
    const response = await fetch(`${API_BASE_URL}/reports`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch available reports');
    }

    const result = await response.json();
    return result.data || [];
  }

  // Generate Balance Sheet
  async generateBalanceSheet(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/balance-sheet${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate balance sheet');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Profit & Loss Statement
  async generateProfitLoss(params: ReportParameters): Promise<ReportData | Blob> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for profit & loss statement');
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/profit-loss${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate profit & loss statement');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Cash Flow Statement
  async generateCashFlow(params: ReportParameters): Promise<ReportData | Blob> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for cash flow statement');
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/cash-flow${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate cash flow statement');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Trial Balance
  async generateTrialBalance(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/trial-balance${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate trial balance');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate General Ledger
  async generateGeneralLedger(params: ReportParameters): Promise<ReportData | Blob> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for general ledger');
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/general-ledger${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate general ledger');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Accounts Receivable Report
  async generateAccountsReceivable(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/accounts-receivable${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate accounts receivable report');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Accounts Payable Report
  async generateAccountsPayable(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/accounts-payable${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate accounts payable report');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Sales Summary Report
  async generateSalesSummary(params: ReportParameters): Promise<ReportData | Blob> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for sales summary');
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/sales-summary${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate sales summary report');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Purchase Summary Report
  async generatePurchaseSummary(params: ReportParameters): Promise<ReportData | Blob> {
    if (!params.start_date || !params.end_date) {
      throw new Error('Start date and end date are required for purchase summary');
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/purchase-summary${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate purchase summary report');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Inventory Report
  async generateInventoryReport(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/inventory-report${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate inventory report');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generate Financial Ratios Analysis
  async generateFinancialRatios(params: ReportParameters): Promise<ReportData | Blob> {
    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/financial-ratios${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to generate financial ratios analysis');
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Generic report generator
  async generateReport(reportId: string, params: ReportParameters): Promise<ReportData | Blob> {
    const endpointMap: { [key: string]: string } = {
      'balance-sheet': 'balance-sheet',
      'profit-loss': 'profit-loss',
      'cash-flow': 'cash-flow',
      'trial-balance': 'trial-balance',
      'general-ledger': 'general-ledger',
      'accounts-receivable': 'accounts-receivable',
      'accounts-payable': 'accounts-payable',
      'sales-summary': 'sales-summary',
      'purchase-summary': 'purchase-summary',
      'inventory-report': 'inventory-report',
      'financial-ratios': 'financial-ratios'
    };

    const endpoint = endpointMap[reportId];
    if (!endpoint) {
      throw new Error(`Unknown report type: ${reportId}`);
    }

    const queryString = this.buildQueryString(params);
    const url = `${API_BASE_URL}/reports/${endpoint}${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to generate report: ${errorText}`);
    }

    if (params.format === 'pdf') {
      return await response.blob();
    }

    const result = await response.json();
    return result.data;
  }

  // Download report as file
  async downloadReport(reportData: Blob, fileName: string): Promise<void> {
    const url = window.URL.createObjectURL(reportData);
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  }

  // Get report templates (if implemented)
  async getReportTemplates(): Promise<any[]> {
    const response = await fetch(`${API_BASE_URL}/reports/templates`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch report templates');
    }

    const result = await response.json();
    return result.data || [];
  }

  // Save report template (if implemented)
  async saveReportTemplate(template: {
    name: string;
    type: string;
    description: string;
    template: string;
    is_default: boolean;
  }): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/reports/templates`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(template),
    });

    if (!response.ok) {
      throw new Error('Failed to save report template');
    }

    const result = await response.json();
    return result.data;
  }
}

export const reportService = new ReportService();
