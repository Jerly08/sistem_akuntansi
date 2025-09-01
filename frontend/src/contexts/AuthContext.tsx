'use client';

import React, { createContext, useState, useContext, useEffect } from 'react';

// Define user type
export type UserRole = 'ADMIN' | 'FINANCE' | 'INVENTORY_MANAGER' | 'DIRECTOR' | 'EMPLOYEE';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
}

// Define auth context type
interface AuthContextType {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<boolean>;
  handleUnauthorized: () => void;
}

// Create context
export const AuthContext = createContext<AuthContextType | undefined>(undefined);

// API URL - ensure this is correctly defined
const API_URL = process.env.NEXT_PUBLIC_API_URL;

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isClient, setIsClient] = useState<boolean>(false);

  // Initialize auth state from localStorage on component mount
  useEffect(() => {
    const initializeAuth = () => {
      setIsClient(true);
      
      // Only access localStorage on client-side
      if (typeof window !== 'undefined') {
        try {
          const storedToken = localStorage.getItem('token');
          const storedRefreshToken = localStorage.getItem('refreshToken');
          const storedUser = localStorage.getItem('user');

          if (storedToken && storedUser) {
            setToken(storedToken);
            setRefreshToken(storedRefreshToken);
            setUser(JSON.parse(storedUser));
          }
        } catch (error) {
          console.error('Error parsing stored user:', error);
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      }
      
      setIsLoading(false);
    };

    initializeAuth();
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    try {
      setIsLoading(true);
      
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Login failed');
      }

      const data = await response.json();
      
      // Save auth data
      setToken(data.token);
      setRefreshToken(data.refreshToken);
      setUser(data.user);
      
      // Store in localStorage (client-side only)
      if (typeof window !== 'undefined') {
        localStorage.setItem('token', data.token);
        localStorage.setItem('refreshToken', data.refreshToken);
        localStorage.setItem('user', JSON.stringify(data.user));
      }
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  // Register function
  const register = async (name: string, email: string, password: string) => {
    try {
      setIsLoading(true);
      
      const response = await fetch(`${API_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, email, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Registration failed');
      }

      const data = await response.json();
      
      // Save auth data
      setToken(data.token);
      setRefreshToken(data.refreshToken);
      setUser(data.user);
      
      // Store in localStorage (client-side only)
      if (typeof window !== 'undefined') {
        localStorage.setItem('token', data.token);
        localStorage.setItem('refreshToken', data.refreshToken);
        localStorage.setItem('user', JSON.stringify(data.user));
      }
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  // Clear all auth data function
  const clearAuthData = () => {
    // Clear state
    setToken(null);
    setRefreshToken(null);
    setUser(null);
    
    // Clear localStorage (client-side only)
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
      // Also clear any other auth-related items
      localStorage.removeItem('authData');
      localStorage.removeItem('userData');
    }
  };

// Handle unauthorized access (e.g., invalid token)
  const handleUnauthorized = () => {
    console.warn('Unauthorized access detected');
    // Don't auto-logout immediately, let components handle 401s gracefully
    // Only logout if the user is actually not authenticated
    if (!token || !user) {
      console.error('No valid auth tokens found, logging out...');
      logout();
    } else {
      console.warn('Auth tokens present but 401 received - might be a permission issue');
    }
  };

  // Logout function
  const logout = () => {
    clearAuthData();
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
  };

  // Check if user is authenticated
  const checkAuth = async (): Promise<boolean> => {
    if (!token) return false;
    
    try {
      // Try to refresh the token if it's about to expire
      if (refreshToken) {
        const response = await fetch(`${API_URL}/auth/refresh`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ refreshToken }),
        });

        if (response.ok) {
          const data = await response.json();
          
          // Update auth data
          setToken(data.token);
          setRefreshToken(data.refreshToken);
          setUser(data.user);
          
          // Update localStorage (client-side only)
          if (typeof window !== 'undefined') {
            localStorage.setItem('token', data.token);
            localStorage.setItem('refreshToken', data.refreshToken);
            localStorage.setItem('user', JSON.stringify(data.user));
          }
          
          return true;
        } else {
          // If refresh fails, clear auth data
          clearAuthData();
          return false;
        }
      }
      
      return !!token;
    } catch (error) {
      console.error('Token refresh error:', error);
      clearAuthData();
      return false;
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        refreshToken,
        isAuthenticated: !!user,
        isLoading,
        login,
        register,
        logout,
        checkAuth,
        handleUnauthorized
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use the auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}; 