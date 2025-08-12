'use client';

import React, { useState, useEffect } from 'react';
import FormField from '../common/FormField';
import { 
  Button, 
  Text, 
  Badge, 
  Tooltip, 
  Icon, 
  HStack,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Box
} from '@chakra-ui/react';
import { FiInfo, FiLock, FiUnlock } from 'react-icons/fi';

import { Account, AccountCreateRequest, AccountUpdateRequest } from '@/types/account';

interface AccountFormProps {
  account?: Account;
  parentAccounts?: Account[];
  onSubmit: (data: AccountCreateRequest | AccountUpdateRequest) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

// Helper function to determine smart category based on account type, account code, and parent
const getSmartCategory = (type: string, parentAccounts: Account[], parentId: number | null, accountCode?: string, accountName?: string): string => {
  // If no parent, use smart defaults based on account type
  if (!parentId) {
    const smartDefaults = {
      ASSET: 'FIXED_ASSET',        // Assets without parent are typically fixed assets
      LIABILITY: 'CURRENT_LIABILITY', 
      EQUITY: 'EQUITY',
      REVENUE: 'OPERATING_REVENUE',
      EXPENSE: 'OPERATING_EXPENSE',
    };
    return smartDefaults[type as keyof typeof smartDefaults] || '';
  }

  // Find parent account
  const parent = parentAccounts.find(p => p.id === parentId);
  if (!parent) return '';

  // Smart categorization based on parent account code/name and account code pattern
  if (type === 'ASSET') {
    // Use account code pattern for better accuracy
    if (accountCode) {
      const codeNum = parseInt(accountCode);
      if (codeNum >= 1100 && codeNum < 1500) {
        return 'CURRENT_ASSET';
      }
      if (codeNum >= 1500) {
        return 'FIXED_ASSET';
      }
    }
    
    // Fallback to parent-based logic
    if (parent.code === '1100' || parent.name.includes('CURRENT')) {
      return 'CURRENT_ASSET';
    }
    if (parent.code === '1500' || parent.name.includes('FIXED')) {
      return 'FIXED_ASSET';
    }
    
    // Account name semantic analysis for better categorization
    if (accountName) {
      const nameLower = accountName.toLowerCase();
      // Current asset indicators
      if (nameLower.includes('kas') || nameLower.includes('bank') || nameLower.includes('piutang') || 
          nameLower.includes('persediaan') || nameLower.includes('inventory')) {
        return 'CURRENT_ASSET';
      }
      // Fixed asset indicators
      if (nameLower.includes('tanah') || nameLower.includes('bangunan') || nameLower.includes('peralatan') || 
          nameLower.includes('kendaraan') || nameLower.includes('mesin') || nameLower.includes('gedung')) {
        return 'FIXED_ASSET';
      }
    }
    
    // If parent is main ASSETS (1000), default to CURRENT_ASSET for codes < 1500
    // This is more logical as most sub-accounts under main assets are current assets
    if (parent.code === '1000') {
      return accountCode && parseInt(accountCode) >= 1500 ? 'FIXED_ASSET' : 'CURRENT_ASSET';
    }
    
    return 'CURRENT_ASSET'; // Default to current asset
  }

  if (type === 'LIABILITY') {
    if (parent.code === '2100' || parent.name.includes('CURRENT')) {
      return 'CURRENT_LIABILITY';
    }
    if (parent.name.includes('LONG') || parent.name.includes('TERM')) {
      return 'LONG_TERM_LIABILITY';
    }
    return 'CURRENT_LIABILITY';
  }

  if (type === 'REVENUE') {
    if (parent.name.includes('OTHER') || parent.name.includes('NON')) {
      return 'OTHER_REVENUE';
    }
    return 'OPERATING_REVENUE';
  }

  if (type === 'EXPENSE') {
    if (parent.name.includes('OTHER') || parent.name.includes('NON')) {
      return 'OTHER_EXPENSE';
    }
    return 'OPERATING_EXPENSE';
  }

  return 'EQUITY';
};

const AccountForm: React.FC<AccountFormProps> = ({
  account,
  parentAccounts = [],
  onSubmit,
  onCancel,
  isSubmitting = false,
}) => {
  const [formData, setFormData] = useState<any>({
    code: '',
    name: '',
    description: '',
    type: 'ASSET',
    category: '',
    parent_id: null,
    opening_balance: 0,
    is_active: true,
    // Convert from account if provided
    ...(account && {
      code: account.code,
      name: account.name,
      description: account.description,
      type: account.type,
      category: account.category,
      parent_id: account.parent_id,
      is_active: account.is_active,
    }),
  });

  // Debug log for form data changes
  useEffect(() => {
    console.log('Form data updated:', formData);
  }, [formData]);

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Reset form when account prop changes
  useEffect(() => {
    setFormData({
      code: '',
      name: '',
      description: '',
      type: 'ASSET',
      category: '',
      parent_id: null,
      opening_balance: 0,
      is_active: true,
      // Convert from account if provided
      ...(account && {
        code: account.code,
        name: account.name,
        description: account.description,
        type: account.type,
        category: account.category,
        parent_id: account.parent_id,
        is_active: account.is_active,
      }),
    });
  }, [account]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value, type } = e.target;
    
    console.log('Form change:', { name, value, type }); // Debug log
    
    // Handle checkbox inputs
    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData((prev) => ({ ...prev, [name]: checked }));
    } else {
      // Handle special field mappings
      let fieldValue: any = value;
      if (name === 'parent_id' && value === '') {
        fieldValue = null;
      } else if (name === 'parent_id' && value !== '') {
        fieldValue = parseInt(value);
      } else if (name === 'opening_balance') {
        fieldValue = parseFloat(value) || 0;
      } else if (name === 'name') {
        // Force account names to uppercase for uniformity
        fieldValue = value.toUpperCase();
      }
      setFormData((prev) => ({ ...prev, [name]: fieldValue }));
    }
    
    // Clear error when field is edited
    if (errors[name]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    
    // Validate required fields
    if (!formData.code) newErrors.code = 'Account code is required';
    if (!formData.name) newErrors.name = 'Account name is required';
    if (!formData.type) newErrors.type = 'Account type is required';
    
    // Validate code format (e.g., numeric only)
    if (formData.code && !/^\d+$/.test(formData.code)) {
      newErrors.code = 'Account code must contain only numbers';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (validateForm()) {
      // Auto-assign smart category before submitting
      const smartCategory = getSmartCategory(formData.type, parentAccounts, formData.parent_id, formData.code, formData.name);
      const dataWithSmartCategory = {
        ...formData,
        category: smartCategory
      };
      
      console.log('Smart category assigned:', smartCategory, 'for type:', formData.type, 'parent:', formData.parent_id, 'code:', formData.code, 'name:', formData.name);
      
      onSubmit(dataWithSmartCategory);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <FormField
          id="code"
          label="Account Code"
          value={formData.code || ''}
          onChange={handleChange}
          placeholder="Enter account code"
          required
          error={errors.code}
          name="code"
        />
        
        <FormField
          id="name"
          label="Account Name"
          value={formData.name || ''}
          onChange={handleChange}
          placeholder="Enter account name"
          required
          error={errors.name}
          name="name"
        />
        
        <FormField
          id="type"
          label="Account Type"
          type="select"
          value={formData.type || ''}
          onChange={handleChange}
          required
          error={errors.type}
          options={[
            { value: 'ASSET', label: 'Asset' },
            { value: 'LIABILITY', label: 'Liability' },
            { value: 'EQUITY', label: 'Equity' },
            { value: 'REVENUE', label: 'Revenue' },
            { value: 'EXPENSE', label: 'Expense' },
          ]}
          name="type"
        />
        
        
        <FormField
          id="parent_id"
          label="Parent Account"
          type="select"
          value={formData.parent_id || ''}
          onChange={handleChange}
          options={[
            { value: '', label: 'No Parent (Top Level)' },
            ...parentAccounts.map((parent) => ({
              value: parent.id.toString(),
              label: `${parent.code} - ${parent.name}`,
            })),
          ]}
          name="parent_id"
        />
        
        <Box>
          <HStack mb={2}>
            <Text fontSize="sm" fontWeight="medium">Opening Balance</Text>
            <Tooltip 
              label="Opening balance is the initial balance when creating the account. It can only be edited if there are no transactions yet."
              hasArrow
            >
              <span>
                <Icon as={FiInfo} color="gray.500" boxSize={3} />
              </span>
            </Tooltip>
            {account && (
              <Badge 
                colorScheme={account.id ? 'orange' : 'green'} 
                size="sm"
                variant="subtle"
              >
                <Icon as={account.id ? FiLock : FiUnlock} mr={1} />
                {account.id ? 'Edit Restricted' : 'Editable'}
              </Badge>
            )}
          </HStack>
          <FormField
            id="opening_balance"
            label=""
            type="number"
            value={formData.opening_balance || ''}
            onChange={handleChange}
            placeholder="Enter opening balance"
            name="opening_balance"
            disabled={account?.id ? true : false}
          />
          {account?.id && (
            <Text fontSize="xs" color="gray.500" mt={1}>
              Note: Opening balance cannot be changed after account creation. Use journal entries to adjust balance.
            </Text>
          )}
        </Box>
        
        <div className="md:col-span-2">
          <FormField
            id="description"
            label="Description"
            type="textarea"
            value={formData.description || ''}
            onChange={handleChange}
            placeholder="Enter account description"
            name="description"
          />
        </div>
        
        <div className="flex items-center">
          <input
            id="is_active"
            type="checkbox"
            checked={formData.is_active}
            onChange={handleChange}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            name="is_active"
          />
          <label htmlFor="is_active" className="ml-2 block text-sm text-gray-700">
            Active
          </label>
        </div>
      </div>
      
      {/* Smart Category Indicator with Tooltip */}
      {formData.type && (
        <Alert status="info" variant="left-accent" mt={4}>
          <AlertIcon />
          <Box>
            <AlertTitle fontSize="sm">Smart Category Assignment</AlertTitle>
            <AlertDescription fontSize="xs">
              <HStack spacing={2} mt={2}>
                <Text>Category will be:</Text>
                <Badge colorScheme="blue" variant="solid">
                  {getSmartCategory(formData.type, parentAccounts, formData.parent_id, formData.code, formData.name).replace(/_/g, ' ')}
                </Badge>
              </HStack>
              <Tooltip 
                label="Categories help organize accounts for financial reporting. They are automatically assigned based on the account type and parent account to ensure consistency."
                hasArrow
                placement="top"
              >
                <HStack spacing={1} mt={2} cursor="help">
                  <Icon as={FiInfo} color="blue.600" boxSize={3} />
                  <Text color="blue.600">What is Category?</Text>
                </HStack>
              </Tooltip>
            </AlertDescription>
          </Box>
        </Alert>
      )}
      
      <div className="mt-6 flex justify-end space-x-3">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          isDisabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          colorScheme="brand"
          isLoading={isSubmitting}
        >
          {account?.id ? 'Update Account' : 'Create Account'}
        </Button>
      </div>
    </form>
  );
};

export default AccountForm; 