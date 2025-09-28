import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Button,
  Badge,
  Alert,
  AlertCircle,
  CheckCircle,
  Clock,
  DollarSign,
  FileText,
  Users,
} from '../components/ui';
import { FastPaymentForm } from '../components/FastPaymentForm';
import { fastPaymentService } from '../services/fastPaymentService';
import SummaryStatCard from '../components/SummaryStatCard';

interface Sale {
  id: number;
  code: string;
  invoice_number: string;
  customer: {
    id: number;
    name: string;
    type: string;
  };
  date: string;
  due_date: string;
  total_amount: number;
  paid_amount: number;
  outstanding_amount: number;
  status: 'DRAFT' | 'CONFIRMED' | 'INVOICED' | 'PAID' | 'OVERDUE' | 'CANCELLED';
}

interface SalesStats {
  total_sales: number;
  total_revenue: number;
  total_outstanding: number;
  avg_order_value: number;
}

const SalesManagementFast: React.FC = () => {
  const [sales, setSales] = useState<Sale[]>([]);
  const [stats, setStats] = useState<SalesStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [paymentFormOpen, setPaymentFormOpen] = useState(false);
  const [selectedSale, setSelectedSale] = useState<Sale | null>(null);
  const [processingPayments, setProcessingPayments] = useState<Set<number>>(new Set());
  const [recentSuccessPayments, setRecentSuccessPayments] = useState<Set<number>>(new Set());

  // Load sales data
  const loadSalesData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Mock API call - replace with actual API
      const response = await fetch('/api/v1/sales', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to load sales data');
      }
      
      const data = await response.json();
      setSales(data.data || []);
      
      // Calculate stats
      const totalSales = data.data?.length || 0;
      const totalRevenue = data.data?.reduce((sum: number, sale: Sale) => sum + sale.total_amount, 0) || 0;
      const totalOutstanding = data.data?.reduce((sum: number, sale: Sale) => sum + sale.outstanding_amount, 0) || 0;
      
      setStats({
        total_sales: totalSales,
        total_revenue: totalRevenue,
        total_outstanding: totalOutstanding,
        avg_order_value: totalSales > 0 ? totalRevenue / totalSales : 0,
      });
      
    } catch (err: any) {
      console.error('Failed to load sales data:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSalesData();
  }, [loadSalesData]);

  // Handle fast payment recording
  const handleFastPayment = async (saleId: number, amount: number) => {
    const sale = sales.find(s => s.id === saleId);
    if (!sale) return;

    setProcessingPayments(prev => new Set(prev).add(saleId));

    try {
      const paymentData = {
        amount,
        payment_date: new Date().toISOString().split('T')[0],
        method: 'BANK_TRANSFER',
        cash_bank_id: 1, // Default bank account
        reference: `Fast-${sale.invoice_number}`,
        notes: `Quick payment for ${sale.invoice_number}`,
      };

      const response = await fastPaymentService.recordSalesPayment(saleId, paymentData);
      
      // Update local state optimistically
      setSales(prev => prev.map(s => 
        s.id === saleId 
          ? {
              ...s,
              paid_amount: s.paid_amount + amount,
              outstanding_amount: s.outstanding_amount - amount,
              status: s.outstanding_amount - amount <= 0.01 ? 'PAID' as const : s.status,
            }
          : s
      ));

      // Show success indicator
      setRecentSuccessPayments(prev => new Set(prev).add(saleId));
      setTimeout(() => {
        setRecentSuccessPayments(prev => {
          const newSet = new Set(prev);
          newSet.delete(saleId);
          return newSet;
        });
      }, 3000);

      console.log('Fast payment successful:', response);
      
    } catch (err: any) {
      console.error('Fast payment failed:', err);
      setError(`Payment failed for ${sale.invoice_number}: ${err.message}`);
    } finally {
      setProcessingPayments(prev => {
        const newSet = new Set(prev);
        newSet.delete(saleId);
        return newSet;
      });
    }
  };

  // Handle detailed payment form
  const handleDetailedPayment = (sale: Sale) => {
    setSelectedSale(sale);
    setPaymentFormOpen(true);
  };

  // Handle payment success from form
  const handlePaymentSuccess = (response: any) => {
    console.log('Payment success:', response);
    
    // Update local state
    if (selectedSale && response.success) {
      setSales(prev => prev.map(s => 
        s.id === selectedSale.id 
          ? {
              ...s,
              paid_amount: s.paid_amount + response.amount,
              outstanding_amount: response.outstanding_amount || Math.max(0, s.outstanding_amount - response.amount),
              status: response.new_status || (s.outstanding_amount - response.amount <= 0.01 ? 'PAID' as const : s.status),
            }
          : s
      ));

      // Show success indicator
      setRecentSuccessPayments(prev => new Set(prev).add(selectedSale.id));
      setTimeout(() => {
        setRecentSuccessPayments(prev => {
          const newSet = new Set(prev);
          newSet.delete(selectedSale.id);
          return newSet;
        });
      }, 3000);
    }

    setPaymentFormOpen(false);
    setSelectedSale(null);
  };

  // Get status color
  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'PAID': return 'bg-green-100 text-green-800';
      case 'INVOICED': return 'bg-blue-100 text-blue-800';
      case 'OVERDUE': return 'bg-red-100 text-red-800';
      case 'CONFIRMED': return 'bg-yellow-100 text-yellow-800';
      case 'DRAFT': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  // Format currency
  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading sales data...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          🚀 Sales Management (Fast Payment)
        </h1>
        <p className="text-gray-600">
          Manage your sales transactions with lightning-fast payment processing
        </p>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert className="mb-6">
          <AlertCircle className="h-4 w-4" />
          <div>
            <strong>Error:</strong> {error}
            <Button 
              variant="outline" 
              size="sm" 
              className="ml-4"
              onClick={() => setError(null)}
            >
              Dismiss
            </Button>
          </div>
        </Alert>
      )}

      {/* Statistics Cards - colored like the Sales Management screenshot */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
          <SummaryStatCard
            title="Total Sales"
            value={stats.total_sales}
            subtitle="Transactions this period"
            icon={<FileText className="h-8 w-8" />}
            color="blue"
          />
          <SummaryStatCard
            title="Total Revenue"
            value={formatCurrency(stats.total_revenue)}
            subtitle="Gross revenue"
            icon={<DollarSign className="h-8 w-8" />}
            color="green"
          />
          <SummaryStatCard
            title="Outstanding"
            value={formatCurrency(stats.total_outstanding)}
            subtitle="Unpaid invoices"
            icon={<Clock className="h-8 w-8" />}
            color="orange"
          />
          <SummaryStatCard
            title="Avg Order Value"
            value={formatCurrency(stats.avg_order_value)}
            subtitle="Per transaction"
            icon={<Users className="h-8 w-8" />}
            color="purple"
          />
        </div>
      )}

      {/* Sales Transactions Table */}
      <Card>
        <CardHeader>
          <CardTitle>Sales Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full table-auto">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-4 font-medium">Invoice</th>
                  <th className="text-left p-4 font-medium">Customer</th>
                  <th className="text-right p-4 font-medium">Total</th>
                  <th className="text-right p-4 font-medium">Paid</th>
                  <th className="text-right p-4 font-medium">Outstanding</th>
                  <th className="text-center p-4 font-medium">Status</th>
                  <th className="text-center p-4 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sales.map((sale) => (
                  <tr key={sale.id} className="border-b hover:bg-gray-50">
                    <td className="p-4">
                      <div>
                        <p className="font-medium">{sale.invoice_number}</p>
                        <p className="text-sm text-gray-600">{sale.code}</p>
                      </div>
                    </td>
                    <td className="p-4">
                      <p className="font-medium">{sale.customer.name}</p>
                    </td>
                    <td className="p-4 text-right">
                      <p className="font-medium">{formatCurrency(sale.total_amount)}</p>
                    </td>
                    <td className="p-4 text-right">
                      <p className="text-green-600 font-medium">{formatCurrency(sale.paid_amount)}</p>
                    </td>
                    <td className="p-4 text-right">
                      <p className={`font-medium ${sale.outstanding_amount > 0 ? 'text-orange-600' : 'text-gray-400'}`}>
                        {formatCurrency(sale.outstanding_amount)}
                      </p>
                    </td>
                    <td className="p-4 text-center">
                      <div className="flex items-center justify-center gap-2">
                        <Badge className={getStatusColor(sale.status)}>
                          {sale.status}
                        </Badge>
                        {recentSuccessPayments.has(sale.id) && (
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        )}
                      </div>
                    </td>
                    <td className="p-4">
                      {sale.outstanding_amount > 0 && (sale.status === 'INVOICED' || sale.status === 'OVERDUE') && (
                        <div className="flex items-center gap-2">
                          {/* Fast Payment Button */}
                          <Button
                            size="sm"
                            variant="outline"
                            disabled={processingPayments.has(sale.id)}
                            onClick={() => handleFastPayment(sale.id, sale.outstanding_amount)}
                            className="text-xs"
                          >
                            {processingPayments.has(sale.id) ? (
                              <>
                                <div className="animate-spin rounded-full h-3 w-3 border-b border-current mr-1"></div>
                                Processing...
                              </>
                            ) : (
                              <>⚡ Full Payment</>
                            )}
                          </Button>

                          {/* Detailed Payment Button */}
                          <Button
                            size="sm"
                            variant="default"
                            onClick={() => handleDetailedPayment(sale)}
                            className="text-xs"
                          >
                            💰 Record Payment
                          </Button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {sales.length === 0 && (
              <div className="text-center py-12">
                <p className="text-gray-600">No sales transactions found.</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Fast Payment Form Modal */}
      {paymentFormOpen && selectedSale && (
        <FastPaymentForm
          open={paymentFormOpen}
          onClose={() => {
            setPaymentFormOpen(false);
            setSelectedSale(null);
          }}
          onSuccess={handlePaymentSuccess}
          saleData={{
            sale_id: selectedSale.id,
            invoice_number: selectedSale.invoice_number,
            customer: {
              name: selectedSale.customer.name,
            },
            total_amount: selectedSale.total_amount,
            outstanding_amount: selectedSale.outstanding_amount,
          }}
        />
      )}
    </div>
  );
};

export default SalesManagementFast;