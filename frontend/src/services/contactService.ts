import { 
  Contact, 
  ContactAddress,
  ApiResponse,
  ApiError
} from '@/types/contact';

// Base API URL - should be moved to environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Type for the unauthorized handler callback
type UnauthorizedHandler = () => void;

class ContactService {
  private unauthorizedHandler?: UnauthorizedHandler;

  // Set the unauthorized handler (to be called from components)
  setUnauthorizedHandler(handler: UnauthorizedHandler) {
    this.unauthorizedHandler = handler;
  }

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
      
      if (response.status === 401 && this.unauthorizedHandler) {
        this.unauthorizedHandler();
      }

      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    return response.json();
  }

  // Get all contacts
  async getContacts(token: string, type?: string): Promise<Contact[]> {
    const url = new URL(`${API_BASE_URL}/api/v1/contacts`);
    if (type) {
      url.searchParams.append('type', type);
    }
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Contact[]> = await this.handleResponse(response);
    return result.data;
  }

  // Get single contact by ID
  async getContact(token: string, id: string): Promise<Contact> {
    const response = await fetch(`${API_BASE_URL}/api/v1/contacts/${id}`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Contact> = await this.handleResponse(response);
    return result.data;
  }

  // Create new contact
  async createContact(token: string, contactData: Partial<Contact>): Promise<Contact> {
    const response = await fetch(`${API_BASE_URL}/api/v1/contacts`, {
      method: 'POST',
      headers: this.getHeaders(token),
      body: JSON.stringify(contactData),
    });
    
    const result: ApiResponse<Contact> = await this.handleResponse(response);
    return result.data;
  }

  // Update existing contact
  async updateContact(token: string, id: string, contactData: Partial<Contact>): Promise<Contact> {
    const response = await fetch(`${API_BASE_URL}/api/v1/contacts/${id}`, {
      method: 'PUT',
      headers: this.getHeaders(token),
      body: JSON.stringify(contactData),
    });
    
    const result: ApiResponse<Contact> = await this.handleResponse(response);
    return result.data;
  }

  // Delete contact
  async deleteContact(token: string, id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/v1/contacts/${id}`, {
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
