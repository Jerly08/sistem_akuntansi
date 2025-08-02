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
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2 } from 'react-icons/fi';

// Define the Asset type
interface Asset {
  id: string;
  name: string;
  category: string;
  purchaseDate: string;
  purchasePrice: number;
  currentValue: number;
  depreciationRate: number;
  depreciationMethod: string;
  usefulLife: number;
  location: string;
  serialNumber: string;
  condition: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

const AssetsPage = () => {
  const { token } = useAuth();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState<Partial<Asset> | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState<Partial<Asset>>({
    name: '',
    category: '',
    purchaseDate: '',
    purchasePrice: 0,
    currentValue: 0,
    depreciationRate: 0,
    depreciationMethod: 'Straight Line',
    usefulLife: 1,
    location: '',
    serialNumber: '',
    condition: 'Good',
    active: true
  });

  // Fetch assets from API
  const fetchAssets = async () => {
    // Data dummy untuk testing di frontend
    setAssets([
      {
        id: '1',
        name: 'Office Building',
        category: 'Real Estate',
        purchaseDate: '2020-01-15',
        purchasePrice: 5000000,
        currentValue: 4800000,
        depreciationRate: 2,
        depreciationMethod: 'Straight Line',
        usefulLife: 50,
        location: 'Jl. Sudirman No. 123',
        serialNumber: 'N/A',
        condition: 'Good',
        active: true,
        createdAt: '2020-01-15',
        updatedAt: '2025-07-01',
      },
      {
        id: '2',
        name: 'Company Car - Toyota Avanza',
        category: 'Vehicle',
        purchaseDate: '2021-06-22',
        purchasePrice: 250000000,
        currentValue: 200000000,
        depreciationRate: 10,
        depreciationMethod: 'Declining Balance',
        usefulLife: 8,
        location: 'Parking Area',
        serialNumber: 'B1234ABC',
        condition: 'Excellent',
        active: true,
        createdAt: '2021-06-22',
        updatedAt: '2025-07-15',
      },
      {
        id: '3',
        name: 'Dell Workstation Computer',
        category: 'Computer Equipment',
        purchasePrice: 25000000,
        currentValue: 18000000,
        depreciationRate: 25,
        purchaseDate: '2022-03-10',
        depreciationMethod: 'Double Declining Balance',
        usefulLife: 4,
        location: 'IT Department',
        serialNumber: 'DL-WS-2022-001',
        condition: 'Good',
        active: true,
        createdAt: '2022-03-10',
        updatedAt: '2025-08-01',
      },
      {
        id: '4',
        name: 'Xerox Printer Machine',
        category: 'Office Equipment',
        purchaseDate: '2021-11-05',
        purchasePrice: 15000000,
        currentValue: 12000000,
        depreciationRate: 15,
        depreciationMethod: 'Straight Line',
        usefulLife: 7,
        location: 'Administrative Office',
        serialNumber: 'XRX-2021-PM-505',
        condition: 'Good',
        active: true,
        createdAt: '2021-11-05',
        updatedAt: '2025-07-20',
      },
      {
        id: '5',
        name: 'Office Desk Set',
        category: 'Furniture',
        purchaseDate: '2020-08-15',
        purchasePrice: 12000000,
        currentValue: 10000000,
        depreciationRate: 10,
        depreciationMethod: 'Straight Line',
        usefulLife: 10,
        location: 'Main Office Floor 2',
        serialNumber: 'FUR-DESK-2020-15',
        condition: 'Fair',
        active: true,
        createdAt: '2020-08-15',
        updatedAt: '2025-06-10',
      },
      {
        id: '6',
        name: 'Network Server HPE',
        category: 'IT Infrastructure',
        purchaseDate: '2021-09-12',
        purchasePrice: 80000000,
        currentValue: 64000000,
        depreciationRate: 20,
        depreciationMethod: 'Straight Line',
        usefulLife: 5,
        location: 'Server Room',
        serialNumber: 'HPE-SRV-2021-G10',
        condition: 'Excellent',
        active: true,
        createdAt: '2021-09-12',
        updatedAt: '2025-08-02',
      },
      {
        id: '7',
        name: 'Industrial Packaging Machine',
        category: 'Machinery',
        purchaseDate: '2019-04-20',
        purchasePrice: 150000000,
        currentValue: 100000000,
        depreciationRate: 12,
        depreciationMethod: 'Declining Balance',
        usefulLife: 12,
        location: 'Production Floor',
        serialNumber: 'IPM-2019-PKG-420',
        condition: 'Good',
        active: true,
        createdAt: '2019-04-20',
        updatedAt: '2025-05-18',
      },
      {
        id: '8',
        name: 'Air Conditioning Central System',
        category: 'Office Equipment',
        purchaseDate: '2020-07-08',
        purchasePrice: 45000000,
        currentValue: 38000000,
        depreciationRate: 8,
        depreciationMethod: 'Straight Line',
        usefulLife: 15,
        location: 'Building Rooftop',
        serialNumber: 'AC-CENT-2020-708',
        condition: 'Good',
        active: true,
        createdAt: '2020-07-08',
        updatedAt: '2025-07-25',
      },
    ]);
    setIsLoading(false);
  };

  // Load assets on component mount
  useEffect(() => {
    if (token) {
      fetchAssets();
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    
    try {
      const url = formData.id
        ? `/api/assets/${formData.id}`
        : '/api/assets';
        
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
        throw new Error(`Failed to ${formData.id ? 'update' : 'create'} asset`);
      }
      
      // Refresh assets list
      fetchAssets();
      
      // Close modal and reset form
      setIsModalOpen(false);
      setSelectedAsset(null);
      setFormData({
        name: '',
        category: '',
        purchaseDate: '',
        purchasePrice: 0,
        currentValue: 0,
        depreciationRate: 0,
        depreciationMethod: 'Straight Line',
        usefulLife: 1,
        location: '',
        serialNumber: '',
        condition: 'Good',
        active: true
      });
    } catch (err) {
      setError(`Error ${formData.id ? 'updating' : 'creating'} asset. Please try again.`);
      console.error('Error submitting asset:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle asset deletion
  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this asset?')) {
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/assets/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete asset');
      }
      
      // Refresh assets list
      fetchAssets();
    } catch (err) {
      setError('Error deleting asset. Please try again.');
      console.error('Error deleting asset:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Open modal for creating a new asset
  const handleCreate = () => {
    setSelectedAsset(null);
    setFormData({
      name: '',
      category: '',
      purchaseDate: '',
      purchasePrice: 0,
      currentValue: 0,
      depreciationRate: 0,
      depreciationMethod: 'Straight Line',
      usefulLife: 1,
      location: '',
      serialNumber: '',
      condition: 'Good',
      active: true
    });
    setIsModalOpen(true);
  };

  // Open modal for editing an existing asset
  const handleEdit = (asset: Asset) => {
    setSelectedAsset(asset);
    setFormData(asset);
    setIsModalOpen(true);
  };

  // Handle form input changes
  const handleInputChange = (field: keyof Asset, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  // Table columns definition
  const columns = [
    { header: 'Name', accessor: 'name' },
    { header: 'Category', accessor: 'category' },
    { header: 'Purchase Price', accessor: (asset: Asset) => `$${asset.purchasePrice.toLocaleString()}` },
    { header: 'Current Value', accessor: (asset: Asset) => `$${asset.currentValue.toLocaleString()}` },
    { header: 'Condition', accessor: 'condition' },
    { header: 'Status', accessor: (asset: Asset) => (asset.active ? 'Active' : 'Inactive') },
  ];

  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // Action buttons for each row
  const renderActions = (asset: Asset) => (
    <>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => handleEdit(asset)}
      >
        Edit
      </Button>
      <Button
        size="sm"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => handleDelete(asset.id)}
      >
        Delete
      </Button>
    </>
  );

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Asset Master</Heading>
          <Button
            colorScheme="brand"
            leftIcon={<FiPlus />}
            onClick={handleCreate}
          >
            Add Asset
          </Button>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Table<Asset>
          columns={columns}
          data={assets}
          keyField="id"
          title="Assets"
          actions={renderActions}
          isLoading={isLoading}
        />
        
        <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="xl">
          <ModalOverlay />
          <ModalContent>
            <form onSubmit={handleSubmit}>
              <ModalHeader>
                {selectedAsset?.id ? 'Edit Asset' : 'Create Asset'}
              </ModalHeader>
              <ModalCloseButton />
              <ModalBody>
                <VStack spacing={4}>
                  <HStack width="100%" spacing={4}>
                    <FormControl isRequired flex={2}>
                      <FormLabel>Asset Name</FormLabel>
                      <Input
                        value={formData.name || ''}
                        onChange={(e) => handleInputChange('name', e.target.value)}
                        placeholder="Enter asset name"
                      />
                    </FormControl>
                    
                    <FormControl isRequired flex={1}>
                      <FormLabel>Category</FormLabel>
                      <Select
                        value={formData.category || ''}
                        onChange={(e) => handleInputChange('category', e.target.value)}
                        placeholder="Select category"
                      >
                        <option value="Real Estate">Real Estate</option>
                        <option value="Computer Equipment">Computer Equipment</option>
                        <option value="Vehicle">Vehicle</option>
                        <option value="Office Equipment">Office Equipment</option>
                        <option value="Furniture">Furniture</option>
                        <option value="IT Infrastructure">IT Infrastructure</option>
                        <option value="Machinery">Machinery</option>
                      </Select>
                    </FormControl>
                  </HStack>
                  
                  <HStack width="100%" spacing={4}>
                    <FormControl isRequired>
                      <FormLabel>Purchase Date</FormLabel>
                      <Input
                        type="date"
                        value={formData.purchaseDate || ''}
                        onChange={(e) => handleInputChange('purchaseDate', e.target.value)}
                      />
                    </FormControl>
                    
                    <FormControl isRequired>
                      <FormLabel>Purchase Price</FormLabel>
                      <NumberInput
                        value={formData.purchasePrice || 0}
                        onChange={(valueString) => handleInputChange('purchasePrice', parseFloat(valueString) || 0)}
                        min={0}
                        precision={2}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                    
                    <FormControl isRequired>
                      <FormLabel>Current Value</FormLabel>
                      <NumberInput
                        value={formData.currentValue || 0}
                        onChange={(valueString) => handleInputChange('currentValue', parseFloat(valueString) || 0)}
                        min={0}
                        precision={2}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                  </HStack>
                  
                  <HStack width="100%" spacing={4}>
                    <FormControl>
                      <FormLabel>Depreciation Rate (%)</FormLabel>
                      <NumberInput
                        value={formData.depreciationRate || 0}
                        onChange={(valueString) => handleInputChange('depreciationRate', parseFloat(valueString) || 0)}
                        min={0}
                        max={100}
                        precision={2}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Depreciation Method</FormLabel>
                      <Select
                        value={formData.depreciationMethod || 'Straight Line'}
                        onChange={(e) => handleInputChange('depreciationMethod', e.target.value)}
                      >
                        <option value="Straight Line">Straight Line</option>
                        <option value="Declining Balance">Declining Balance</option>
                        <option value="Double Declining Balance">Double Declining Balance</option>
                      </Select>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Useful Life (Years)</FormLabel>
                      <NumberInput
                        value={formData.usefulLife || 1}
                        onChange={(valueString) => handleInputChange('usefulLife', parseInt(valueString) || 1)}
                        min={1}
                        max={100}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                  </HStack>
                  
                  <HStack width="100%" spacing={4}>
                    <FormControl flex={2}>
                      <FormLabel>Location</FormLabel>
                      <Input
                        value={formData.location || ''}
                        onChange={(e) => handleInputChange('location', e.target.value)}
                        placeholder="Enter asset location"
                      />
                    </FormControl>
                    
                    <FormControl flex={1}>
                      <FormLabel>Serial Number</FormLabel>
                      <Input
                        value={formData.serialNumber || ''}
                        onChange={(e) => handleInputChange('serialNumber', e.target.value)}
                        placeholder="Enter serial number"
                      />
                    </FormControl>
                  </HStack>
                  
                  <HStack width="100%" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>Condition</FormLabel>
                      <Select
                        value={formData.condition || 'Good'}
                        onChange={(e) => handleInputChange('condition', e.target.value)}
                      >
                        <option value="Excellent">Excellent</option>
                        <option value="Good">Good</option>
                        <option value="Fair">Fair</option>
                        <option value="Poor">Poor</option>
                      </Select>
                    </FormControl>
                    
                    <FormControl flex={1}>
                      <HStack>
                        <FormLabel mb={0}>Active</FormLabel>
                        <Switch
                          isChecked={formData.active !== false}
                          onChange={(e) => handleInputChange('active', e.target.checked)}
                        />
                      </HStack>
                    </FormControl>
                  </HStack>
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
                  loadingText={selectedAsset?.id ? 'Updating...' : 'Creating...'}
                >
                  {selectedAsset?.id ? 'Update' : 'Create'}
                </Button>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
      </Box>
    </Layout>
  );
};

export default AssetsPage;
