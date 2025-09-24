// Test CSV conversion functionality
const testData = {
  trial_balance: {
    accounts: [
      {
        account_code: "1101",
        account_name: "Cash Account",
        account_type: "ASSET",
        debit_balance: 5000000,
        credit_balance: 0
      },
      {
        account_code: "1102", 
        account_name: "Bank BCA",
        account_type: "ASSET",
        debit_balance: 25000000,
        credit_balance: 0
      }
    ]
  },
  balance_sheet: {
    as_of_date: "2025-01-22",
    total_assets: 30000000,
    total_liabilities: 10000000, 
    total_equity: 20000000,
    assets: {
      current_assets: {
        items: [
          { account_name: "Cash", amount: 5000000 },
          { account_name: "Bank", amount: 25000000 }
        ]
      }
    }
  }
};

// Helper function to flatten complex values
const flattenValue = (value) => {
  if (value === null || value === undefined) return '';
  if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
    return String(value);
  }
  if (typeof value === 'object') {
    if (Array.isArray(value)) {
      return value.map(item => flattenValue(item)).join('; ');
    } else {
      return Object.entries(value)
        .map(([k, v]) => `${k}: ${flattenValue(v)}`)
        .join('; ');
    }
  }
  return String(value);
};

// Convert JSON to CSV
const convertJSONToCSV = (data, reportType) => {
  try {
    if (!data) return 'No data available';
    
    console.log('Converting to CSV. Report type:', reportType);
    console.log('Data structure:', data);
    
    let records = [];
    
    if (reportType === 'trial-balance' && data.accounts) {
      records = data.accounts.map((account) => ({
        'Account Code': account.account_code || '',
        'Account Name': account.account_name || account.name || '',
        'Account Type': account.account_type || '',
        'Debit Balance': account.debit_balance || 0,
        'Credit Balance': account.credit_balance || 0
      }));
    } else if (reportType === 'balance-sheet') {
      const reportDate = data.as_of_date || new Date().toISOString().split('T')[0];
      records = [
        { 'Report': 'Balance Sheet', 'As Of': reportDate, 'Value': '' },
        { 'Report': 'Total Assets', 'As Of': reportDate, 'Value': data.total_assets || 0 },
        { 'Report': 'Total Liabilities', 'As Of': reportDate, 'Value': data.total_liabilities || 0 },
        { 'Report': 'Total Equity', 'As Of': reportDate, 'Value': data.total_equity || 0 }
      ];
      
      if (data.assets && data.assets.current_assets && data.assets.current_assets.items) {
        data.assets.current_assets.items.forEach((item) => {
          records.push({
            'Report': 'Asset - ' + (item.account_name || item.name || 'Unknown'),
            'As Of': reportDate,
            'Value': item.amount || 0
          });
        });
      }
    }
    
    if (records.length === 0) {
      return 'No data available for export';
    }
    
    const headers = Object.keys(records[0]);
    
    const csvContent = [
      headers.join(','),
      ...records.map(record => 
        headers.map(header => {
          const value = record[header];
          const stringValue = flattenValue(value);
          if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
            return `"${stringValue.replace(/"/g, '""')}"`;
          }
          return stringValue;
        }).join(',')
      )
    ].join('\n');
    
    return csvContent;
    
  } catch (error) {
    console.error('Error converting to CSV:', error);
    return `Error converting data to CSV format: ${error.message}`;
  }
};

// Test the conversion
console.log('=== Trial Balance CSV Test ===');
console.log(convertJSONToCSV(testData.trial_balance, 'trial-balance'));

console.log('\n=== Balance Sheet CSV Test ===');
console.log(convertJSONToCSV(testData.balance_sheet, 'balance-sheet'));