import api from './api';

export interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  role: string;
  status: string;
  department?: string;
  position?: string;
  phone?: string;
  created_at: string;
  updated_at: string;
}

export interface UsersFilter {
  page?: number;
  limit?: number;
  role?: string;
  status?: string;
  search?: string;
}

export interface UsersResult {
  data: User[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface UserCreateRequest {
  username: string;
  email: string;
  password: string;
  full_name: string;
  role: string;
  department?: string;
  position?: string;
  phone?: string;
}

export interface UserUpdateRequest {
  username?: string;
  email?: string;
  password?: string;
  full_name?: string;
  role?: string;
  department?: string;
  position?: string;
  phone?: string;
  status?: string;
}

class UserService {
  async getUsers(filter: UsersFilter = {}): Promise<UsersResult> {
    const params = new URLSearchParams();
    Object.entries(filter).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });
    
    const response = await api.get(`/users?${params}`);
    return response.data;
  }

  async getUser(id: number): Promise<User> {
    const response = await api.get(`/users/${id}`);
    return response.data;
  }

  async createUser(data: UserCreateRequest): Promise<User> {
    const response = await api.post('/users', data);
    return response.data;
  }

  async updateUser(id: number, data: UserUpdateRequest): Promise<User> {
    const response = await api.put(`/users/${id}`, data);
    return response.data;
  }

  async deleteUser(id: number): Promise<void> {
    await api.delete(`/users/${id}`);
  }

  async getUserProfile(): Promise<User> {
    const response = await api.get('/profile');
    return response.data;
  }

  async updateProfile(data: Partial<UserUpdateRequest>): Promise<User> {
    const response = await api.put('/profile', data);
    return response.data;
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await api.post('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword
    });
  }

  // Helper methods
  getRoleLabel(role: string): string {
    const labels: { [key: string]: string } = {
      'admin': 'Administrator',
      'manager': 'Manager',
      'accountant': 'Accountant',
      'finance': 'Finance',
      'sales': 'Sales',
      'inventory': 'Inventory',
      'employee': 'Employee'
    };
    return labels[role] || role;
  }

  getStatusColor(status: string): string {
    const colors: { [key: string]: string } = {
      'active': 'green',
      'inactive': 'gray',
      'suspended': 'red',
      'pending': 'yellow'
    };
    return colors[status] || 'gray';
  }

  getStatusLabel(status: string): string {
    const labels: { [key: string]: string } = {
      'active': 'Active',
      'inactive': 'Inactive',
      'suspended': 'Suspended',
      'pending': 'Pending'
    };
    return labels[status] || status;
  }
}

export default new UserService();
