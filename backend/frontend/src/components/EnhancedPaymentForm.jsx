import React, { useState, useEffect, useCallback } from 'react';
import { 
  Card, 
  Form, 
  Input, 
  Select, 
  DatePicker, 
  Button, 
  Alert, 
  Spin, 
  Row, 
  Col, 
  Divider, 
  Tag, 
  Tooltip, 
  Badge,
  Modal,
  Table,
  Typography,
  Space,
  message
} from 'antd';
import {
  UserOutlined,
  BankOutlined,
  DollarOutlined,
  CalendarOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
  LoadingOutlined
} from '@ant-design/icons';
import moment from 'moment';
import { debounce } from 'lodash';

const { Option } = Select;
const { TextArea } = Input;
const { Title, Text } = Typography;

const EnhancedPaymentForm = ({ onSubmit, loading = false, contacts = [], cashBanks = [] }) => {
  const [form] = Form.useForm();
  
  // State management
  const [selectedContact, setSelectedContact] = useState(null);
  const [selectedCashBank, setSelectedCashBank] = useState(null);
  const [paymentMethod, setPaymentMethod] = useState('');
  const [validationResults, setValidationResults] = useState(null);
  const [availableAllocations, setAvailableAllocations] = useState([]);
  const [selectedAllocations, setSelectedAllocations] = useState([]);
  const [isValidating, setIsValidating] = useState(false);
  const [showAllocationModal, setShowAllocationModal] = useState(false);
  const [formData, setFormData] = useState({});
  
  // Real-time validation states
  const [warnings, setWarnings] = useState([]);
  const [errors, setErrors] = useState([]);
  const [validationStatus, setValidationStatus] = useState('');

  // 🎯 AUTO-DETECTION: When contact is selected
  const handleContactChange = useCallback(async (contactId) => {
    const contact = contacts.find(c => c.id === contactId);
    setSelectedContact(contact);
    
    if (contact) {
      // Auto-detect payment method based on contact type
      let detectedMethod = '';
      if (contact.type === 'CUSTOMER') {
        detectedMethod = 'RECEIVABLE';
      } else if (contact.type === 'VENDOR') {
        detectedMethod = 'PAYABLE';
      }
      
      setPaymentMethod(detectedMethod);
      form.setFieldsValue({ method: detectedMethod });
      
      // Show contact info and auto-detection notification
      message.success(`Auto-detected payment method: ${detectedMethod} for ${contact.type.toLowerCase()}`);
      
      // Load available allocations (invoices/bills)
      await loadAvailableAllocations(contact);
      
      // Auto-select appropriate cash/bank account
      autoSelectCashBank(detectedMethod, form.getFieldValue('amount') || 0);
    } else {
      resetFormState();
    }
  }, [contacts, form]);

  // 🏦 AUTO-SELECT CASH/BANK based on method and amount
  const autoSelectCashBank = useCallback((method, amount) => {
    if (cashBanks.length === 0) return;
    
    let suitable = cashBanks.filter(cb => cb.is_active);
    
    if (method === 'PAYABLE' && amount > 0) {
      // For outgoing payments, prefer accounts with sufficient balance
      suitable = suitable.filter(cb => cb.balance >= amount);
      if (suitable.length === 0) {
        // Fallback to account with highest balance
        suitable = cashBanks.filter(cb => cb.is_active)
          .sort((a, b) => b.balance - a.balance);
      }
      
      // Prefer BANK for large amounts
      if (amount > 1000000) {
        const banks = suitable.filter(cb => cb.type === 'BANK');
        if (banks.length > 0) suitable = banks;
      }
    }
    
    if (suitable.length > 0) {
      const selected = suitable[0];
      setSelectedCashBank(selected);
      form.setFieldsValue({ cash_bank_id: selected.id });
      
      if (method === 'PAYABLE' && selected.balance < amount) {
        setWarnings(prev => [...prev, `Selected account ${selected.name} has insufficient balance (${selected.balance.toLocaleString()}). Payment will result in negative balance.`]);
      }
    }
  }, [cashBanks, form]);

  // 📋 Load available invoices/bills for allocation
  const loadAvailableAllocations = useCallback(async (contact) => {
    if (!contact) return;
    
    try {
      const endpoint = contact.type === 'CUSTOMER' 
        ? `/api/sales/outstanding/${contact.id}`
        : `/api/purchases/outstanding/${contact.id}`;
      
      const response = await fetch(endpoint);
      const data = await response.json();
      
      if (response.ok) {
        setAvailableAllocations(data.data || []);
      }
    } catch (error) {
      console.error('Error loading available allocations:', error);
    }
  }, []);

  // 🔍 Real-time validation with debouncing
  const performRealTimeValidation = useCallback(
    debounce(async (formValues) => {
      if (!formValues.contact_id || !formValues.amount) return;
      
      setIsValidating(true);
      setValidationResults(null);
      setWarnings([]);
      setErrors([]);
      
      try {
        const response = await fetch('/api/payments/validate', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(formValues),
        });
        
        const result = await response.json();
        
        if (result.validation) {
          setValidationResults(result.validation);
          setWarnings(result.validation.warnings || []);
          setErrors(result.validation.errors || []);
          setValidationStatus(result.validation.passed ? 'success' : 'error');
        }
      } catch (error) {
        console.error('Validation error:', error);
        setValidationStatus('error');
        setErrors(['Failed to validate payment']);
      } finally {
        setIsValidating(false);
      }
    }, 500),
    []
  );

  // 📝 Form value change handler
  const handleFormValuesChange = (changedValues, allValues) => {
    setFormData(allValues);
    
    // Auto-select cash/bank when amount changes
    if (changedValues.amount && paymentMethod) {
      autoSelectCashBank(paymentMethod, changedValues.amount);
    }
    
    // Perform real-time validation
    if (allValues.contact_id && allValues.amount && allValues.date) {
      performRealTimeValidation({
        ...allValues,
        date: allValues.date?.format('YYYY-MM-DD'),
        method: paymentMethod,
      });
    }
  };

  // 🎯 Show allocation selection modal
  const showAllocationSelection = () => {
    if (availableAllocations.length === 0) {
      message.info(`No outstanding ${selectedContact?.type === 'CUSTOMER' ? 'invoices' : 'bills'} found for ${selectedContact?.name}`);
      return;
    }
    setShowAllocationModal(true);
  };

  // 📊 Allocation selection columns
  const allocationColumns = [
    {
      title: selectedContact?.type === 'CUSTOMER' ? 'Invoice' : 'Bill',
      dataIndex: 'code',
      key: 'code',
      render: (code, record) => (
        <Space>
          <Text strong>{code}</Text>
          <Tag color={record.status === 'PAID' ? 'green' : record.status === 'PARTIAL' ? 'orange' : 'blue'}>
            {record.status}
          </Tag>
        </Space>
      ),
    },
    {
      title: 'Date',
      dataIndex: 'date',
      key: 'date',
      render: (date) => moment(date).format('DD/MM/YYYY'),
    },
    {
      title: 'Total Amount',
      dataIndex: 'total_amount',
      key: 'total_amount',
      render: (amount) => amount?.toLocaleString(),
      align: 'right',
    },
    {
      title: 'Outstanding',
      dataIndex: 'outstanding_amount',
      key: 'outstanding_amount',
      render: (amount) => (
        <Text strong style={{ color: amount > 0 ? '#1890ff' : '#52c41a' }}>
          {amount?.toLocaleString()}
        </Text>
      ),
      align: 'right',
    },
  ];

  // 🏷️ Validation status indicator
  const ValidationIndicator = () => {
    if (isValidating) {
      return (
        <div className="validation-indicator">
          <LoadingOutlined spin /> Validating...
        </div>
      );
    }
    
    if (validationStatus === 'success') {
      return (
        <div className="validation-indicator success">
          <CheckCircleOutlined style={{ color: '#52c41a' }} /> Validation Passed
          {warnings.length > 0 && (
            <Tooltip title={warnings.join('; ')}>
              <WarningOutlined style={{ color: '#faad14', marginLeft: 8 }} />
            </Tooltip>
          )}
        </div>
      );
    }
    
    if (validationStatus === 'error') {
      return (
        <div className="validation-indicator error">
          <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} /> Validation Failed
        </div>
      );
    }
    
    return null;
  };

  // 🧹 Reset form state
  const resetFormState = () => {
    setSelectedContact(null);
    setSelectedCashBank(null);
    setPaymentMethod('');
    setValidationResults(null);
    setAvailableAllocations([]);
    setSelectedAllocations([]);
    setWarnings([]);
    setErrors([]);
    setValidationStatus('');
  };

  // 📝 Form submission
  const handleSubmit = async (values) => {
    try {
      const submissionData = {
        ...values,
        date: values.date.format('YYYY-MM-DD'),
        method: paymentMethod,
        target_invoice_id: selectedAllocations.find(a => a.type === 'invoice')?.id || null,
        target_bill_id: selectedAllocations.find(a => a.type === 'bill')?.id || null,
        auto_allocate: selectedAllocations.length === 0,
      };
      
      await onSubmit(submissionData);
      
      // Reset form after successful submission
      form.resetFields();
      resetFormState();
      message.success('Payment recorded successfully!');
    } catch (error) {
      message.error('Failed to submit payment: ' + error.message);
    }
  };

  // Set default date to today
  useEffect(() => {
    form.setFieldsValue({
      date: moment(),
    });
  }, [form]);

  return (
    <div className="enhanced-payment-form">
      <Card
        title={
          <Space>
            <DollarOutlined />
            <Title level={4} style={{ margin: 0 }}>Record Payment</Title>
            <ValidationIndicator />
          </Space>
        }
        extra={
          selectedContact && (
            <Tag color={selectedContact.type === 'CUSTOMER' ? 'green' : 'blue'}>
              {selectedContact.type}: {selectedContact.name}
            </Tag>
          )
        }
      >
        {/* Validation Alerts */}
        {errors.length > 0 && (
          <Alert
            message="Validation Errors"
            description={
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                {errors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            }
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}
        
        {warnings.length > 0 && errors.length === 0 && (
          <Alert
            message="Validation Warnings"
            description={
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                {warnings.map((warning, index) => (
                  <li key={index}>{warning}</li>
                ))}
              </ul>
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          onValuesChange={handleFormValuesChange}
          size="large"
        >
          <Row gutter={24}>
            {/* Contact Selection */}
            <Col span={12}>
              <Form.Item
                name="contact_id"
                label={
                  <Space>
                    <UserOutlined />
                    Contact
                  </Space>
                }
                rules={[{ required: true, message: 'Please select a contact' }]}
              >
                <Select
                  placeholder="Select customer or vendor"
                  showSearch
                  filterOption={(input, option) =>
                    option?.children?.toLowerCase().includes(input.toLowerCase())
                  }
                  onChange={handleContactChange}
                  size="large"
                >
                  {contacts.map(contact => (
                    <Option key={contact.id} value={contact.id}>
                      <Space>
                        <Tag 
                          color={contact.type === 'CUSTOMER' ? 'green' : 'blue'}
                          size="small"
                        >
                          {contact.type}
                        </Tag>
                        {contact.name}
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>

            {/* Cash/Bank Account */}
            <Col span={12}>
              <Form.Item
                name="cash_bank_id"
                label={
                  <Space>
                    <BankOutlined />
                    Cash/Bank Account
                    <Tooltip title="Will be auto-selected if not specified">
                      <InfoCircleOutlined style={{ color: '#1890ff' }} />
                    </Tooltip>
                  </Space>
                }
              >
                <Select
                  placeholder="Auto-select or choose manually"
                  allowClear
                  size="large"
                >
                  {cashBanks.map(cashBank => (
                    <Option key={cashBank.id} value={cashBank.id}>
                      <Space>
                        <Tag color={cashBank.type === 'BANK' ? 'blue' : 'green'}>
                          {cashBank.type}
                        </Tag>
                        {cashBank.name}
                        <Text type="secondary">
                          ({cashBank.balance?.toLocaleString()})
                        </Text>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={24}>
            {/* Payment Date */}
            <Col span={12}>
              <Form.Item
                name="date"
                label={
                  <Space>
                    <CalendarOutlined />
                    Payment Date
                  </Space>
                }
                rules={[{ required: true, message: 'Please select payment date' }]}
              >
                <DatePicker 
                  style={{ width: '100%' }} 
                  format="DD/MM/YYYY"
                  size="large"
                />
              </Form.Item>
            </Col>

            {/* Amount */}
            <Col span={12}>
              <Form.Item
                name="amount"
                label={
                  <Space>
                    <DollarOutlined />
                    Amount
                  </Space>
                }
                rules={[
                  { required: true, message: 'Please enter amount' },
                  { type: 'number', min: 0.01, message: 'Amount must be greater than 0' }
                ]}
              >
                <Input
                  type="number"
                  placeholder="Enter payment amount"
                  prefix="Rp"
                  size="large"
                />
              </Form.Item>
            </Col>
          </Row>

          {/* Payment Method (Read-only, auto-detected) */}
          {paymentMethod && (
            <Form.Item
              name="method"
              label="Payment Method (Auto-detected)"
            >
              <Input 
                value={paymentMethod}
                disabled
                addonAfter={
                  <Tag color={paymentMethod === 'RECEIVABLE' ? 'green' : 'red'}>
                    {paymentMethod === 'RECEIVABLE' ? 'Incoming' : 'Outgoing'}
                  </Tag>
                }
                size="large"
              />
            </Form.Item>
          )}

          {/* Allocation Selection */}
          {selectedContact && availableAllocations.length > 0 && (
            <Form.Item label="Target Allocation (Optional)">
              <Button
                type="dashed"
                onClick={showAllocationSelection}
                size="large"
                style={{ width: '100%' }}
              >
                Select Specific {selectedContact.type === 'CUSTOMER' ? 'Invoice' : 'Bill'} to Pay
                <Badge count={availableAllocations.length} showZero />
              </Button>
            </Form.Item>
          )}

          <Row gutter={24}>
            {/* Reference */}
            <Col span={12}>
              <Form.Item
                name="reference"
                label="Reference"
              >
                <Input
                  placeholder="Enter payment reference (optional)"
                  size="large"
                />
              </Form.Item>
            </Col>

            {/* Notes */}
            <Col span={12}>
              <Form.Item
                name="notes"
                label="Notes"
              >
                <TextArea
                  placeholder="Add any notes about this payment"
                  rows={2}
                  size="large"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          {/* Submit Button */}
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading || isValidating}
              disabled={validationStatus === 'error' || errors.length > 0}
              size="large"
              style={{ width: '100%' }}
            >
              {loading ? 'Recording Payment...' : 'Record Payment'}
            </Button>
          </Form.Item>
        </Form>

        {/* Validation Details */}
        {validationResults && (
          <Card 
            title="Validation Results" 
            size="small" 
            style={{ marginTop: 16 }}
            type="inner"
          >
            <div style={{ maxHeight: 200, overflowY: 'auto' }}>
              {validationResults.checks?.map((check, index) => (
                <div key={index} style={{ marginBottom: 8 }}>
                  <Space>
                    {check.status === 'PASS' && <CheckCircleOutlined style={{ color: '#52c41a' }} />}
                    {check.status === 'WARN' && <WarningOutlined style={{ color: '#faad14' }} />}
                    {check.status === 'FAIL' && <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />}
                    <Text strong>{check.name}</Text>
                    <Text type="secondary">{check.message}</Text>
                  </Space>
                </div>
              ))}
            </div>
          </Card>
        )}
      </Card>

      {/* Allocation Selection Modal */}
      <Modal
        title={`Select ${selectedContact?.type === 'CUSTOMER' ? 'Invoice' : 'Bill'} to Pay`}
        open={showAllocationModal}
        onCancel={() => setShowAllocationModal(false)}
        onOk={() => {
          // Handle allocation selection
          setShowAllocationModal(false);
        }}
        width={800}
      >
        <Table
          dataSource={availableAllocations}
          columns={allocationColumns}
          rowKey="id"
          size="small"
          pagination={false}
          rowSelection={{
            type: 'radio',
            onChange: (selectedKeys, selectedRows) => {
              setSelectedAllocations(selectedRows.map(row => ({
                id: row.id,
                type: selectedContact?.type === 'CUSTOMER' ? 'invoice' : 'bill',
                code: row.code,
                amount: row.outstanding_amount,
              })));
            },
          }}
        />
      </Modal>

      <style jsx>{`
        .enhanced-payment-form .validation-indicator {
          display: flex;
          align-items: center;
          font-size: 12px;
        }
        
        .enhanced-payment-form .validation-indicator.success {
          color: #52c41a;
        }
        
        .enhanced-payment-form .validation-indicator.error {
          color: #ff4d4f;
        }
        
        .enhanced-payment-form .ant-form-item-label > label {
          font-weight: 600;
        }
        
        .enhanced-payment-form .ant-tag {
          border-radius: 4px;
        }
      `}</style>
    </div>
  );
};

export default EnhancedPaymentForm;