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
  id: string;
  name: string;
  type: 'Customer' | 'Vendor' | 'Employee';
  email: string;
  phone: string;
  address: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
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
    type: 'Customer',
    email: '',
    phone: '',
    address: '',
    active: true
  });

  // Fetch contacts from API
  const fetchContacts = async () => {
    // Data dummy untuk testing di frontend
    setContacts([
      {
        id: '1',
        name: 'PT Maju Jaya',
        type: 'Customer',
        email: 'info@majujaya.com',
        phone: '+62-21-5551234',
        address: 'Jl. Sudirman No. 123, Jakarta Pusat',
        active: true,
        createdAt: '2025-01-01',
        updatedAt: '2025-08-01',
      },
      {
        id: '2',
        name: 'CV Sumber Rejeki',
        type: 'Vendor',
        email: 'sales@sumberrejeki.co.id',
        phone: '+62-21-5555678',
        address: 'Jl. Gatot Subroto No. 456, Jakarta Selatan',
        active: true,
        createdAt: '2025-01-15',
        updatedAt: '2025-08-01',
      },
      {
        id: '3',
        name: 'Ahmad Subandi',
        type: 'Employee',
        email: 'ahmad.subandi@company.com',
        phone: '+62-812-3456789',
        address: 'Jl. Kebon Jeruk No. 789, Jakarta Barat',
        active: true,
        createdAt: '2025-02-01',
        updatedAt: '2025-08-01',
      },
      {
        id: '4',
        name: 'PT Global Tech',
        type: 'Customer',
        email: 'contact@globaltech.id',
        phone: '+62-21-7771234',
        address: 'Jl. HR Rasuna Said No. 321, Jakarta Selatan',
        active: true,
        createdAt: '2025-02-15',
        updatedAt: '2025-08-01',
      },
      {
        id: '5',
        name: 'Toko Elektronik Sejati',
        type: 'Vendor',
        email: 'admin@elektroniksejati.com',
        phone: '+62-21-6661111',
        address: 'Jl. Mangga Besar No. 88, Jakarta Barat',
        active: true,
        createdAt: '2025-03-01',
        updatedAt: '2025-08-01',
      },
      {
        id: '6',
        name: 'Siti Nurhaliza',
        type: 'Employee',
        email: 'siti.nurhaliza@company.com',
        phone: '+62-813-9876543',
        address: 'Jl. Cempaka Putih No. 55, Jakarta Pusat',
        active: true,
        createdAt: '2025-03-15',
        updatedAt: '2025-08-01',
      },
      {
        id: '7',
        name: 'PT Indah Karya',
        type: 'Customer',
        email: 'info@indahkarya.co.id',
        phone: '+62-21-4441234',
        address: 'Jl. Thamrin No. 567, Jakarta Pusat',
        active: true,
        createdAt: '2025-04-01',
        updatedAt: '2025-08-01',
      },
      {
        id: '8',
        name: 'CV Berkah Jaya',
        type: 'Vendor',
        email: 'order@berkahjaya.net',
        phone: '+62-21-3332222',
        address: 'Jl. Hayam Wuruk No. 99, Jakarta Barat',
        active: false,
        createdAt: '2025-04-15',
        updatedAt: '2025-08-01',
      },
      {
        id: '9',
        name: 'Budi Santoso',
        type: 'Employee',
        email: 'budi.santoso@company.com',
        phone: '+62-814-5551234',
        address: 'Jl. Kemayoran No. 77, Jakarta Pusat',
        active: true,
        createdAt: '2025-05-01',
        updatedAt: '2025-08-01',
      },
      {
        id: '10',
        name: 'PT Solusi Digital',
        type: 'Customer',
        email: 'hello@solusidigital.com',
        phone: '+62-21-8881111',
        address: 'Jl. Kuningan No. 234, Jakarta Selatan',
        active: true,
        createdAt: '2025-05-15',
        updatedAt: '2025-08-01',
      },
    ]);
    setIsLoading(false);
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
        ? `/api/contacts/${formData.id}`
        : '/api/contacts';
        
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
      
      // Close modal and reset form
      setIsModalOpen(false);
      setSelectedContact(null);
      setFormData({
        name: '',
        type: 'Customer',
        email: '',
        phone: '',
        address: '',
        active: true
      });
    } catch (err) {
      setError(`Error ${formData.id ? 'updating' : 'creating'} contact. Please try again.`);
      console.error('Error submitting contact:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle contact deletion
  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this contact?')) {
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/contacts/${id}`, {
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
      type: 'Customer',
      email: '',
      phone: '',
      address: '',
      active: true
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
    { header: 'Status', accessor: (contact: Contact) => (contact.active ? 'Active' : 'Inactive') },
  ];

  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

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
                      value={formData.type || 'Customer'}
                      onChange={(e) => handleInputChange('type', e.target.value as 'Customer' | 'Vendor' | 'Employee')}
                    >
                      <option value="Customer">Customer</option>
                      <option value="Vendor">Vendor</option>
                      <option value="Employee">Employee</option>
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
                    <FormLabel>Address</FormLabel>
                    <Textarea
                      value={formData.address || ''}
                      onChange={(e) => handleInputChange('address', e.target.value)}
                      placeholder="Enter address"
                      rows={3}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <HStack>
                      <FormLabel mb={0}>Active</FormLabel>
                      <Switch
                        isChecked={formData.active !== false}
                        onChange={(e) => handleInputChange('active', e.target.checked)}
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
