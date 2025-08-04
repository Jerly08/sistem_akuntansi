'use client';

import React, { useState, useEffect } from 'react';
import FormField from '../common/FormField';
import { Button } from '@chakra-ui/react';

import { Account, AccountCreateRequest, AccountUpdateRequest } from '@/types/account';

interface AccountFormProps {
  account?: Account;
  parentAccounts?: Account[];
  onSubmit: (data: AccountCreateRequest | AccountUpdateRequest) => void;
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
          id="category"
          label="Category"
          value={formData.category || ''}
          onChange={handleChange}
          placeholder="Enter category"
          name="category"
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
        
        <FormField
          id="opening_balance"
          label="Opening Balance"
          type="number"
          value={formData.opening_balance || ''}
          onChange={handleChange}
          placeholder="Enter opening balance"
          name="opening_balance"
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