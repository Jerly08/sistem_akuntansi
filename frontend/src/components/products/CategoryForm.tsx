import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  Grid,
  GridItem,
  Switch,
  useToast,
  VStack,
  HStack,
  Text,
} from '@chakra-ui/react';
import { FiSave, FiX } from 'react-icons/fi';
import ProductService, { Category } from '@/services/productService';
import { useTranslation } from '@/hooks/useTranslation';

interface CategoryFormProps {
  category?: Category;
  onSave: (category: Category) => void;
  onCancel: () => void;
}

const CategoryForm: React.FC<CategoryFormProps> = ({ category, onSave, onCancel }) => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<Partial<Category>>({
    code: '',
    name: '',
    description: '',
    parent_id: undefined,
    is_active: true,
  });

  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const toast = useToast();

  useEffect(() => {
    fetchCategories();
    if (category) {
      setFormData(category);
    }
  }, [category]);

  const fetchCategories = async () => {
    try {
      const data = await ProductService.getCategories();
      setCategories(data.data);
    } catch (error) {
      toast({
        title: t('products.category.messages.fetchFailed'),
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleInputChange = (field: keyof Category, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      let result;
      if (category?.id) {
        result = await ProductService.updateCategory(category.id, formData);
      } else {
        result = await ProductService.createCategory(formData as Category);
      }
      
      toast({
        title: category?.id ? t('products.category.messages.updateSuccess') : t('products.category.messages.createSuccess'),
        status: 'success',
        isClosable: true,
      });
      
      onSave(result.data);
    } catch (error: any) {
      toast({
        title: category?.id ? t('products.category.messages.updateFailed') : t('products.category.messages.createFailed'),
        description: error.response?.data?.error || t('messages.toast.unknownErrorDesc'),
        status: 'error',
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Filter out current category and its descendants from parent options
  const availableParentCategories = categories.filter(cat => {
    if (!category?.id) return true;
    return cat.id !== category.id;
  });

  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={6} align="stretch">
        {/* Basic Information */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>{t('products.form.basicInfo')}</Text>
          <Grid templateColumns="repeat(2, 1fr)" gap={4}>
            <GridItem>
              <FormControl isRequired>
                <FormLabel>{t('products.category.categoryCode')}</FormLabel>
                <Input
                  value={formData.code}
                  onChange={(e) => handleInputChange('code', e.target.value)}
                  placeholder={t('products.category.enterCategoryCode')}
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl isRequired>
                <FormLabel>{t('products.category.categoryName')}</FormLabel>
                <Input
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder={t('products.category.enterCategoryName')}
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>{t('products.description')}</FormLabel>
                <Textarea
                  value={formData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  placeholder={t('products.category.enterDescription')}
                  rows={3}
                />
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        {/* Hierarchy */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>{t('products.category.hierarchy')}</Text>
          <FormControl>
            <FormLabel>{t('products.category.parentCategory')}</FormLabel>
            <Select
              value={formData.parent_id || ''}
              onChange={(e) => handleInputChange('parent_id', e.target.value ? Number(e.target.value) : undefined)}
              placeholder={t('products.category.selectParentCategory')}
            >
              {availableParentCategories.map(cat => (
                <option key={cat.id} value={cat.id}>
                  {cat.name} ({cat.code})
                </option>
              ))}
            </Select>
          </FormControl>
        </Box>

        {/* Settings */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>{t('products.form.settings')}</Text>
          <FormControl display="flex" alignItems="center">
            <FormLabel htmlFor="is_active" mb="0">
              {t('products.category.activeStatus')}
            </FormLabel>
            <Switch
              id="is_active"
              isChecked={formData.is_active}
              onChange={(e) => handleInputChange('is_active', e.target.checked)}
            />
          </FormControl>
        </Box>

        {/* Form Actions */}
        <HStack justify="flex-end" spacing={4}>
          <Button
            leftIcon={<FiX />}
            onClick={onCancel}
            variant="outline"
          >
            {t('common.cancel')}
          </Button>
          <Button
            leftIcon={<FiSave />}
            type="submit"
            colorScheme="blue"
            isLoading={isLoading}
            loadingText={category?.id ? t('common.updating') : t('common.creating')}
          >
            {category?.id ? t('products.category.updateCategory') : t('products.category.createCategory')}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
};

export default CategoryForm;
