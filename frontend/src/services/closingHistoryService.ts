import api from './api';
import { API_ENDPOINTS } from '@/config/api';

export interface ClosingHistoryItem {
  id: number;
  code: string;
  description: string;
  entry_date: string;
  created_at: string;
  total_debit: number;
  total_credit?: number;
  net_income?: number;
  fiscal_year?: string;
  status?: 'completed' | 'pending';
}

export interface ClosingHistoryResponse {
  success: boolean;
  data: ClosingHistoryItem[];
  error?: string;
}

class ClosingHistoryService {
  /**
   * Get fiscal year closing history
   */
  async getFiscalClosingHistory(): Promise<ClosingHistoryItem[]> {
    try {
      const response = await api.get<ClosingHistoryResponse>(
        API_ENDPOINTS.FISCAL_CLOSING.HISTORY
      );
      
      if (response.data.success) {
        return response.data.data || [];
      }
      
      throw new Error(response.data.error || 'Failed to fetch closing history');
    } catch (error: any) {
      console.error('Error fetching fiscal closing history:', error);
      throw error;
    }
  }

  /**
   * Get period closing history (if available)
   */
  async getPeriodClosingHistory(): Promise<ClosingHistoryItem[]> {
    try {
      const response = await api.get<ClosingHistoryResponse>(
        API_ENDPOINTS.PERIOD_CLOSING.HISTORY
      );
      
      if (response.data.success) {
        return response.data.data || [];
      }
      
      // If not implemented, return empty array
      return [];
    } catch (error: any) {
      // Handle 501 Not Implemented gracefully - period closing history is not yet available
      if (error.response?.status === 501) {
        console.log('Period closing history not yet implemented, using fiscal closing only');
        return [];
      }
      console.warn('Period closing history not available:', error.message);
      return [];
    }
  }

  /**
   * Get combined closing history (both fiscal and period)
   */
  async getAllClosingHistory(): Promise<ClosingHistoryItem[]> {
    try {
      const [fiscalHistory, periodHistory] = await Promise.allSettled([
        this.getFiscalClosingHistory(),
        this.getPeriodClosingHistory()
      ]);

      const allHistory: ClosingHistoryItem[] = [];

      if (fiscalHistory.status === 'fulfilled') {
        allHistory.push(...fiscalHistory.value);
      }

      if (periodHistory.status === 'fulfilled') {
        allHistory.push(...periodHistory.value);
      }

      // Sort by date (newest first)
      return allHistory.sort((a, b) => {
        const dateA = new Date(a.entry_date).getTime();
        const dateB = new Date(b.entry_date).getTime();
        return dateB - dateA;
      });
    } catch (error: any) {
      console.error('Error fetching closing history:', error);
      return [];
    }
  }

  /**
   * Check if a specific date falls within a closed period
   */
  async isDateInClosedPeriod(date: string): Promise<boolean> {
    try {
      const response = await api.get<{success: boolean; is_closed: boolean}>(
        API_ENDPOINTS.PERIOD_CLOSING.CHECK_DATE,
        { params: { date } }
      );
      
      return response.data.is_closed || false;
    } catch (error: any) {
      console.error('Error checking closed period:', error);
      return false;
    }
  }

  /**
   * Filter closing history by date range
   */
  filterHistoryByDateRange(
    history: ClosingHistoryItem[], 
    startDate?: string, 
    endDate?: string
  ): ClosingHistoryItem[] {
    return history.filter(item => {
      const itemDate = new Date(item.entry_date);
      
      if (startDate && itemDate < new Date(startDate)) {
        return false;
      }
      
      if (endDate && itemDate > new Date(endDate)) {
        return false;
      }
      
      return true;
    });
  }

  /**
   * Format closing history for display
   */
  formatClosingHistory(item: ClosingHistoryItem): {
    date: string;
    period: string;
    netIncome: string;
    status: string;
    description: string;
  } {
    const date = new Date(item.entry_date);
    const formattedDate = date.toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });

    // Extract period from description or use entry date
    let period = item.fiscal_year || '';
    if (!period && item.description) {
      // Try to extract period from description
      const periodMatch = item.description.match(/(\d{4}-\d{2}-\d{2})\s+to\s+(\d{4}-\d{2}-\d{2})/);
      if (periodMatch) {
        period = `${periodMatch[1]} to ${periodMatch[2]}`;
      } else {
        // Use fiscal year if available in description
        const yearMatch = item.description.match(/\d{4}/);
        period = yearMatch ? `FY ${yearMatch[0]}` : formattedDate;
      }
    }

    // Calculate net income if not provided
    const netIncome = item.net_income !== undefined 
      ? item.net_income 
      : item.total_debit || 0;

    return {
      date: formattedDate,
      period,
      netIncome: this.formatCurrency(netIncome),
      status: item.status || 'completed',
      description: item.description || 'Fiscal Year Closing'
    };
  }

  /**
   * Format currency for display
   */
  private formatCurrency(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  }
}

export default new ClosingHistoryService();