import React, { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Button,
  Heading,
  Input,
  InputGroup,
  InputLeftElement,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  Flex,
  Select,
  useToast
} from '@chakra-ui/react';
import { FiSearch, FiEdit, FiTrash2, FiUpload } from 'react-icons/fi';
import ProductService, { Product, Category } from '@/services/productService';
import Modal from '@/components/common/Modal';

const ProductCatalog: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const toast = useToast();

  useEffect(() => {
    fetchProducts();
    fetchCategories();
  }, []);

  const fetchProducts = async () => {
    try {
      const data = await ProductService.getProducts({ search: searchTerm });
      setProducts(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch products',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const fetchCategories = async () => {
    try {
      const data = await ProductService.getCategories();
      setCategories(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch categories',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleSearch = () => {
    fetchProducts();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = async (productId: number) => {
    if (!selectedFile) {
      toast({
        title: 'Please select a file',
        status: 'warning',
        isClosable: true,
      });
      return;
    }
    try {
      await ProductService.uploadProductImage(productId, selectedFile);
      toast({
        title: 'Image uploaded successfully',
        status: 'success',
        isClosable: true,
      });
      setSelectedFile(null); // Reset file input
    } catch (error) {
      toast({
        title: 'Failed to upload image',
        status: 'error',
        isClosable: true,
      });
    }
  };

  return (
    <Layout allowedRoles={['ADMIN', 'INVENTORY_MANAGER']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Box>
            <Heading as="h1" size="xl" mb={2}>Product Catalog</Heading>
          </Box>
          <Button leftIcon={<FiUpload />} colorScheme="brand" size="lg">
            Add Product
          </Button>
        </Flex>

        <InputGroup maxW="400px" mb={4}>
          <InputLeftElement pointerEvents="none">
            <FiSearch color="gray.300" />
          </InputLeftElement>
          <Input
            placeholder="Search products..."
            value={searchTerm}
            onChange={handleSearchChange}
          />
          <Button onClick={handleSearch} ml={2}>
            Search
          </Button>
        </InputGroup>

        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Product ID</Th>
              <Th>Name</Th>
              <Th>Category</Th>
              <Th>Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {products.map((product) => (
              <Tr key={product.id}>
                <Td>{product.code}</Td>
                <Td>{product.name}</Td>
                <Td>{product.category?.name}</Td>
                <Td>
                  <Button size="sm" variant="ghost" leftIcon={<FiEdit />} mr={2}>
                    Edit
                  </Button>
                  <Button size="sm" variant="ghost" colorScheme="red" leftIcon={<FiTrash2 />} mr={2}>
                    Delete
                  </Button>
                  <Input
                    type="file"
                    onChange={handleFileChange}
                    style={{ display: 'none' }}
                    id={`file-upload-${product.id}`}
                  />
                  <Button
                    size="sm"
                    variant="ghost"
                    leftIcon={<FiUpload />}
                    htmlFor={`file-upload-${product.id}`}
                    as="label"
                  >
                    Upload Image
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleUpload(product.id!)}
                    disabled={!selectedFile}
                  >
                    Confirm Upload
                  </Button>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>

        {/* Edit Product Modal */}
        {selectedProduct && (
          <Modal
            isOpen={!!selectedProduct}
            onClose={() => setSelectedProduct(null)}
            title="Edit Product"
            size="lg"
          >
            {/* Form fields for editing will be placed here */}
          </Modal>
        )}
      </Box>
    </Layout>
  );
};

export default ProductCatalog;
