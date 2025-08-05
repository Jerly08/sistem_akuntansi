'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import Table from '@/components/common/Table';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useToast,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  Switch,
  VStack,
  HStack,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2 } from 'react-icons/fi';

// Define the Contact type
interface Contact {
  id: number;
  code?: string;
  name: string;
  type: 'CUSTOMER' | 'VENDOR' | 'EMPLOYEE';
  category?: string;
  email: string;
  phone: string;
  mobile?: string;
  fax?: string;
  website?: string;
  tax_number?: string;
  credit_limit?: number;
  payment_terms?: number;
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
  addresses?: ContactAddress[];
}

interface ContactAddress {
  id: number;
  contact_id: number;
  type: 'BILLING' | 'SHIPPING' | 'MAILING';
  address1: string;
  address2?: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  is_default: boolean;
}

const ContactsPage = () => {
  const { token } = useAuth();
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedContact, setSelectedContact] = useState<Partial<Contact> | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState<Partial<Contact>>({
    name: '',
    type: 'CUSTOMER',
    email: '',
    phone: '',
    mobile: '',
    notes: '',
    is_active: true
  });
  // Fetch contacts from API
  const fetchContacts = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/contacts`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch contacts');
      }

      const data = await response.json();
      setContacts(data);
    } catch (err) {
      setError('Failed to fetch contacts. Please try again.');
      console.error('Error fetching contacts:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Load contacts on component mount
  useEffect(() => {
    if (token) {
      fetchContacts();
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    
    try {
      const url = formData.id
        ? `http://localhost:8080/api/v1/contacts/${formData.id}`
        : 'http://localhost:8080/api/v1/contacts';
        
      const method = formData.id ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(formData),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to ${formData.id ? 'update' : 'create'} contact`);
      }
      
      // Refresh contacts list
      fetchContacts();
      
      // Show success message
      toast({
        title: formData.id ? 'Contact Updated' : 'Contact Created',
        description: `Contact has been ${formData.id ? 'updated' : 'created'} successfully.`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
      // Close modal and reset form
      setIsModalOpen(false);
      setSelectedContact(null);
      setFormData({
        name: '',
        type: 'CUSTOMER',
        email: '',
        phone: '',
        mobile: '',
        notes: '',
        is_active: true
      });
    } catch (err) {
      setError(`Error ${formData.id ? 'updating' : 'creating'} contact. Please try again.`);
      console.error('Error submitting contact:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle contact deletion
  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this contact?')) {
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`http://localhost:8080/api/v1/contacts/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete contact');
      }
      
      // Refresh contacts list
      fetchContacts();
      
      // Show success message
      toast({
        title: 'Contact Deleted',
        description: 'Contact has been deleted successfully.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err) {
      setError('Error deleting contact. Please try again.');
      console.error('Error deleting contact:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Open modal for creating a new contact
  const handleCreate = () => {
    setSelectedContact(null);
    setFormData({
      name: '',
      type: 'CUSTOMER',
      email: '',
      phone: '',
      mobile: '',
      notes: '',
      is_active: true
    });
    setIsModalOpen(true);
  };

  // Open modal for editing an existing contact
  const handleEdit = (contact: Contact) => {
    setSelectedContact(contact);
    setFormData(contact);
    setIsModalOpen(true);
  };

  // Handle form input changes
  const handleInputChange = (field: keyof Contact, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  // Table columns definition
  const columns = [
    { header: 'Name', accessor: 'name' },
    { header: 'Type', accessor: 'type' },
    { header: 'Email', accessor: 'email' },
    { header: 'Phone', accessor: 'phone' },
    { header: 'Status', accessor: (contact: Contact) => (contact.is_active ? 'Active' : 'Inactive') },
  ];

  const toast = useToast();

  // Action buttons for each row
  const renderActions = (contact: Contact) => (
    <>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => handleEdit(contact)}
      >
        Edit
      </Button>
      <Button
        size="sm"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => handleDelete(contact.id)}
      >
        Delete
      </Button>
    </>
  );

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'INVENTORY_MANAGER']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Contact Master</Heading>
          <Button
            colorScheme="brand"
            leftIcon={<FiPlus />}
            onClick={handleCreate}
          >
            Add Contact
          </Button>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Table<Contact>
          columns={columns}
          data={contacts}
          keyField="id"
          title="Contacts"
          actions={renderActions}
          isLoading={isLoading}
        />
        
        <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <form onSubmit={handleSubmit}>
              <ModalHeader>
                {selectedContact?.id ? 'Edit Contact' : 'Create Contact'}
              </ModalHeader>
              <ModalCloseButton />
              <ModalBody>
                <VStack spacing={4}>
                  <FormControl isRequired>
                    <FormLabel>Name</FormLabel>
                    <Input
                      value={formData.name || ''}
                      onChange={(e) => handleInputChange('name', e.target.value)}
                      placeholder="Enter contact name"
                    />
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>Type</FormLabel>
                    <Select
                      value={formData.type || 'CUSTOMER'}
                      onChange={(e) => handleInputChange('type', e.target.value as 'CUSTOMER' | 'VENDOR' | 'EMPLOYEE')}
                    >
                      <option value="CUSTOMER">Customer</option>
                      <option value="VENDOR">Vendor</option>
                      <option value="EMPLOYEE">Employee</option>
                    </Select>
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      value={formData.email || ''}
                      onChange={(e) => handleInputChange('email', e.target.value)}
                      placeholder="Enter email address"
                    />
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>Phone</FormLabel>
                    <Input
                      value={formData.phone || ''}
                      onChange={(e) => handleInputChange('phone', e.target.value)}
                      placeholder="Enter phone number"
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>Mobile</FormLabel>
                    <Input
                      value={formData.mobile || ''}
                      onChange={(e) => handleInputChange('mobile', e.target.value)}
                      placeholder="Enter mobile number"
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>Notes</FormLabel>
                    <Textarea
                      value={formData.notes || ''}
                      onChange={(e) => handleInputChange('notes', e.target.value)}
                      placeholder="Enter notes"
                      rows={3}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <HStack>
                      <FormLabel mb={0}>Active</FormLabel>
                      <Switch
                        isChecked={formData.is_active !== false}
                        onChange={(e) => handleInputChange('is_active', e.target.checked)}
                      />
                    </HStack>
                  </FormControl>
                </VStack>
              </ModalBody>
              <ModalFooter>
                <Button variant="ghost" mr={3} onClick={() => setIsModalOpen(false)}>
                  Cancel
                </Button>
                <Button
                  colorScheme="brand"
                  type="submit"
                  isLoading={isSubmitting}
                  loadingText={selectedContact?.id ? 'Updating...' : 'Creating...'}
                >
                  {selectedContact?.id ? 'Update' : 'Create'}
                </Button>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
      </Box>
    </Layout>
  );
};

export default ContactsPage;
