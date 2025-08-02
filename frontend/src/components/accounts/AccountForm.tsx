'use client';

import React, { useState, useEffect } from 'react';
import FormField from '../common/FormField';
import { Button } from '@chakra-ui/react';

// Define the Account type based on the Prisma schema
interface Account {
  id: string;
  code: string;
  name: string;
  description?: string;
  type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  subType?: string;
  parentAccountId?: string;
  active: boolean;
  balance: number;
}

interface AccountFormProps {
  account?: Partial<Account>;
  parentAccounts?: Account[];
  onSubmit: (data: Partial<Account>) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

const AccountForm: React.FC<AccountFormProps> = ({
  account,
  parentAccounts = [],
  onSubmit,
  onCancel,
  isSubmitting = false,
}) => {
  const [formData, setFormData] = useState<Partial<Account>>({
    code: '',
    name: '',
    description: '',
    type: 'ASSET',
    subType: '',
    parentAccountId: '',
    active: true,
    ...account,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Reset form when account prop changes
  useEffect(() => {
    setFormData({
      code: '',
      name: '',
      description: '',
      type: 'ASSET',
      subType: '',
      parentAccountId: '',
      active: true,
      ...account,
    });
  }, [account]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value, type } = e.target;
    
    // Handle checkbox inputs
    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData((prev) => ({ ...prev, [name]: checked }));
    } else {
      setFormData((prev) => ({ ...prev, [name]: value }));
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
      onSubmit(formData);
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
          id="subType"
          label="Sub Type"
          value={formData.subType || ''}
          onChange={handleChange}
          placeholder="Enter sub type"
          name="subType"
        />
        
        <FormField
          id="parentAccountId"
          label="Parent Account"
          type="select"
          value={formData.parentAccountId || ''}
          onChange={handleChange}
          options={[
            { value: '', label: 'No Parent (Top Level)' },
            ...parentAccounts.map((parent) => ({
              value: parent.id,
              label: `${parent.code} - ${parent.name}`,
            })),
          ]}
          name="parentAccountId"
        />
        
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
            id="active"
            type="checkbox"
            checked={formData.active}
            onChange={handleChange}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            name="active"
          />
          <label htmlFor="active" className="ml-2 block text-sm text-gray-700">
            Active
          </label>
        </div>
      </div>
      
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