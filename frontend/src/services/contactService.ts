import { 
  Contact, 
  ContactAddress,
  ApiResponse,
  ApiError
} from '@/types/contact';

// Base API URL - should be moved to environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL;

class ContactService {

  private getHeaders(token?: string): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    
    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let errorData: ApiError;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'Network error',
          code: 'NETWORK_ERROR',
        };
      }

      // Handle specific error cases
      if (response.status === 401) {
        throw new Error('Unauthorized: Authentication token is invalid or expired');
      }
      if (response.status === 403) {
        throw new Error('Forbidden: Insufficient permissions to perform this action');
      }
      if (response.status === 404) {
        throw new Error('Not found: The requested resource does not exist');
      }
      if (response.status === 400) {
        throw new Error(errorData.error || 'Bad request: Invalid data provided');
      }
      if (response.status >= 500) {
        throw new Error('Server error: Please try again later');
      }

      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    // Check if response has content
    const contentLength = response.headers.get('content-length');
    if (contentLength === '0') {
      console.warn('ContactService: Empty response received');
      return {} as T;
    }
    
    try {
      const result = await response.json();
      if (result === null || result === undefined) {
        console.warn('ContactService: Null or undefined response received');
        return {} as T;
      }
      return result;
    } catch (error) {
      console.error('ContactService: Error parsing JSON response:', error);
      throw new Error('Invalid JSON response from server');
    }
  }

  // Get all contacts
  async getContacts(token: string, type?: string): Promise<Contact[]> {
    const url = new URL(`${API_BASE_URL}/contacts`);
    if (type) {
      url.searchParams.append('type', type);
    }
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result = await this.handleResponse(response);
    return Array.isArray(result) ? result : result.data || [];
  }

  // Get single contact by ID
  async getContact(token: string, id: string): Promise<Contact> {
    const response = await fetch(`${API_BASE_URL}/contacts/${id}`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Contact> = await this.handleResponse(response);
    return result.data;
  }

  // Create new contact
  async createContact(token: string, contactData: Partial<Contact>): Promise<Contact> {
    console.log('ContactService: Creating contact with data:', contactData);
    
    const response = await fetch(`${API_BASE_URL}/contacts`, {
      method: 'POST',
      headers: this.getHeaders(token),
      body: JSON.stringify(contactData),
    });
    
    console.log('ContactService: Response status:', response.status, response.statusText);
    
    const result = await this.handleResponse(response);
    console.log('ContactService: Parsed result:', result);
    
    // Handle different response structures
    if (result && typeof result === 'object') {
      // Check if result has data property (wrapped response)
      if ('data' in result && result.data) {
        console.log('ContactService: Returning result.data:', result.data);
        return result.data as Contact;
      }
      
      // Check if result is the contact object directly
      if ('id' in result || 'name' in result) {
        console.log('ContactService: Returning result directly:', result);
        return result as Contact;
      }
    }
    
    console.error('ContactService: Invalid response structure:', result);
    throw new Error('Invalid response format from server');
  }

  // Update existing contact
  async updateContact(token: string, id: string, contactData: Partial<Contact>): Promise<Contact> {
    const response = await fetch(`${API_BASE_URL}/contacts/${id}`, {
      method: 'PUT',
      headers: this.getHeaders(token),
      body: JSON.stringify(contactData),
    });
    
    const result: ApiResponse<Contact> = await this.handleResponse(response);
    return result.data;
  }

  // Delete contact
  async deleteContact(token: string, id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/contacts/${id}`, {
      method: 'DELETE',
      headers: this.getHeaders(token),
    });
    
    await this.handleResponse(response);
  }

  // Helper: Get contact type label
  getContactTypeLabel(type: string): string {
    switch (type) {
      case 'CUSTOMER':
        return 'Pelanggan';
      case 'VENDOR':
        return 'Supplier';
      case 'EMPLOYEE':
        return 'Karyawan';
      default:
        return type;
    }
  }

  // Helper: Get contact type color
  getContactTypeColor(type: string): string {
    switch (type) {
      case 'CUSTOMER':
        return 'blue';
      case 'VENDOR':
        return 'green';
      case 'EMPLOYEE':
        return 'purple';
      default:
        return 'gray';
    }
  }
}

export const contactService = new ContactService();
export default contactService;
