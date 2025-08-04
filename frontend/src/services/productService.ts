import api from './api';

export interface Product {
  id?: number;
  code: string;
  name: string;
  description?: string;
  category_id?: number;
  brand?: string;
  model?: string;
  unit: string;
  purchase_price: number;
  sale_price: number;
  pricing_tier?: string;
  stock: number;
  min_stock: number;
  max_stock: number;
  reorder_level: number;
  barcode?: string;
  sku?: string;
  weight?: number;
  dimensions?: string;
  is_active: boolean;
  is_service: boolean;
  taxable: boolean;
  image_path?: string;
  notes?: string;
  category?: Category;
  variants?: ProductVariant[];
}

export interface ProductVariant {
  id?: number;
  product_id: number;
  name: string;
  sku?: string;
  price: number;
  stock: number;
  is_active: boolean;
}

export interface Category {
  id?: number;
  code: string;
  name: string;
  description?: string;
  parent_id?: number;
  is_active: boolean;
  parent?: Category;
  children?: Category[];
}

export interface InventoryMovement {
  id: number;
  product_id: number;
  reference_type: string;
  reference_id: number;
  type: 'IN' | 'OUT';
  quantity: number;
  unit_cost: number;
  total_cost: number;
  notes?: string;
  transaction_date: string;
  product: Product;
}

export interface StockAdjustment {
  product_id: number;
  quantity: number;
  type: 'IN' | 'OUT';
  notes?: string;
}

export interface StockOpname {
  product_id: number;
  new_stock: number;
  notes?: string;
}

export interface BulkPriceUpdate {
  updates: {
    product_id: number;
    purchase_price?: number;
    sale_price?: number;
  }[];
}

class ProductService {
  // Products
  async getProducts(params?: {
    search?: string;
    category?: string;
    page?: number;
    limit?: number;
  }) {
    const response = await api.get('/products', { params });
    return response.data;
  }

  async getProduct(id: number) {
    const response = await api.get(`/products/${id}`);
    return response.data;
  }

  async createProduct(product: Product) {
    const response = await api.post('/products', product);
    return response.data;
  }

  async updateProduct(id: number, product: Partial<Product>) {
    const response = await api.put(`/products/${id}`, product);
    return response.data;
  }

  async deleteProduct(id: number) {
    const response = await api.delete(`/products/${id}`);
    return response.data;
  }

  // Categories
  async getCategories(params?: {
    include_relations?: boolean;
    parent_id?: string;
  }) {
    const response = await api.get('/categories', { params });
    return response.data;
  }

  async getCategory(id: number) {
    const response = await api.get(`/categories/${id}`);
    return response.data;
  }

  async createCategory(category: Category) {
    const response = await api.post('/categories', category);
    return response.data;
  }

  async updateCategory(id: number, category: Partial<Category>) {
    const response = await api.put(`/categories/${id}`, category);
    return response.data;
  }

  async deleteCategory(id: number) {
    const response = await api.delete(`/categories/${id}`);
    return response.data;
  }

  async getCategoryTree() {
    const response = await api.get('/categories/tree');
    return response.data;
  }

  async getCategoryProducts(id: number, search?: string) {
    const response = await api.get(`/categories/${id}/products`, {
      params: { search }
    });
    return response.data;
  }

  // Inventory
  async getInventoryMovements(params?: {
    product_id?: number;
    start_date?: string;
    end_date?: string;
    type?: 'IN' | 'OUT';
  }) {
    const response = await api.get('/inventory/movements', { params });
    return response.data;
  }

  async getLowStockProducts() {
    const response = await api.get('/inventory/low-stock');
    return response.data;
  }

  async getStockValuation(params?: {
    method?: 'FIFO' | 'LIFO' | 'Average';
    product_id?: number;
  }) {
    const response = await api.get('/inventory/valuation', { params });
    return response.data;
  }

  async getStockReport(params?: {
    category_id?: number;
  }) {
    const response = await api.get('/inventory/report', { params });
    return response.data;
  }

  async bulkPriceUpdate(data: BulkPriceUpdate) {
    const response = await api.post('/inventory/bulk-price-update', data);
    return response.data;
  }

  // Stock Operations
  async adjustStock(data: StockAdjustment) {
    const response = await api.post('/products/adjust-stock', data);
    return response.data;
  }

  async stockOpname(data: StockOpname) {
    const response = await api.post('/products/opname', data);
    return response.data;
  }

  // File Upload
  async uploadProductImage(productId: number, file: File) {
    const formData = new FormData();
    formData.append('image', file);
    formData.append('product_id', productId.toString());
    
    const response = await api.post('/products/upload-image', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }
}

export default new ProductService();
