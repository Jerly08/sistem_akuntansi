'use client';

import React, { createContext, useState, useContext, useEffect } from 'react';

// Define user type - backend sends lowercase roles
export type UserRole = 'admin' | 'finance' | 'inventory_manager' | 'director' | 'employee' | 'auditor' | 'operational_user';

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

// API URL - use environment variable with fallback and correct API path
const API_URL = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') + '/api/v1';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isClient, setIsClient] = useState<boolean>(false);

  // Initialize auth state from localStorage on component mount
  useEffect(() => {
    const initializeAuth = async () => {
      setIsClient(true);
      
      // Only access localStorage on client-side
      if (typeof window !== 'undefined') {
        try {
          const storedToken = localStorage.getItem('token');
          const storedRefreshToken = localStorage.getItem('refreshToken');
          const storedUser = localStorage.getItem('user');
          const lastLogoutTime = localStorage.getItem('lastLogoutTime');
          
          // Check if user explicitly logged out recently (within last 5 minutes)
          const isRecentLogout = lastLogoutTime && 
            (Date.now() - parseInt(lastLogoutTime)) < 5 * 60 * 1000; // 5 minutes
          
          if (isRecentLogout) {
            console.log('Recent logout detected, clearing any remaining auth data');
            // Clear any remaining auth data from recent logout
            localStorage.removeItem('token');
            localStorage.removeItem('refreshToken');
            localStorage.removeItem('user');
            localStorage.removeItem('lastLogoutTime');
          } else if (storedToken && storedUser) {
            // Validate token before auto-login
            const userData = JSON.parse(storedUser);
            
            // Basic validation of stored data
            if (userData && userData.id && userData.email && storedToken.length > 20) {
              // Try to validate the token by making a quick API call
              try {
                const response = await fetch(`${API_URL}/auth/validate-token`, {
                  method: 'GET',
                  headers: {
                    'Authorization': `Bearer ${storedToken}`,
                    'Content-Type': 'application/json',
                  },
                });
                
                if (response.ok) {
                  setToken(storedToken);
                  setRefreshToken(storedRefreshToken);
                  setUser(userData);
                  console.log('Auth initialized from localStorage - token validated:', { user: userData, hasToken: !!storedToken });
                } else if (response.status === 401 || response.status === 403) {
                  console.log('Stored token is invalid (401/403), clearing auth data');
                  localStorage.removeItem('token');
                  localStorage.removeItem('refreshToken');
                  localStorage.removeItem('user');
                } else {
                  console.log('Token validation failed with status:', response.status, 'proceeding with stored data');
                  // If validation fails due to server error, proceed with stored data
                  setToken(storedToken);
                  setRefreshToken(storedRefreshToken);
                  setUser(userData);
                }
              } catch (error) {
                console.log('Token validation failed (network/server error), proceeding with stored data:', error);
                // If validation fails due to network issues, proceed with stored data
                // The API interceptors will handle auth errors when actual requests are made
                setToken(storedToken);
                setRefreshToken(storedRefreshToken);
                setUser(userData);
              }
            } else {
              console.log('Invalid stored auth data, clearing');
              localStorage.removeItem('token');
              localStorage.removeItem('refreshToken');
              localStorage.removeItem('user');
            }
          } else {
            console.log('No stored auth data found');
          }
        } catch (error) {
          console.error('Error parsing stored user:', error);
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
          localStorage.removeItem('lastLogoutTime');
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
      
      // Save auth data - ensure role is lowercase
      const userData = {
        ...data.user,
        role: data.user.role.toLowerCase() // Ensure role is lowercase
      };
      
      setToken(data.access_token || data.token);
      setRefreshToken(data.refresh_token || data.refreshToken);
      setUser(userData);
      
      // Store in localStorage (client-side only)
      if (typeof window !== 'undefined') {
        localStorage.setItem('token', data.access_token || data.token);
        localStorage.setItem('refreshToken', data.refresh_token || data.refreshToken);
        localStorage.setItem('user', JSON.stringify(userData));
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
  const clearAuthData = (setLogoutTimestamp = false) => {
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
      
      // Set logout timestamp to prevent auto-login for a few minutes
      if (setLogoutTimestamp) {
        localStorage.setItem('lastLogoutTime', Date.now().toString());
        console.log('Logout timestamp set to prevent immediate auto-login');
      }
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
    console.log('Logout initiated by user');
    clearAuthData(true); // Set logout timestamp
    if (typeof window !== 'undefined') {
      // Use setTimeout to ensure localStorage is updated before redirect
      setTimeout(() => {
        window.location.href = '/login';
      }, 100);
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
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (response.ok) {
          const data = await response.json();
          
          // Handle different possible response field names
          const accessToken = data.access_token || data.token;
          const newRefreshToken = data.refresh_token || data.refreshToken;
          const userData = {
            ...data.user,
            role: data.user.role.toLowerCase() // Ensure role is lowercase
          };
          
          // Update auth data
          setToken(accessToken);
          setRefreshToken(newRefreshToken);
          setUser(userData);
          
          // Update localStorage (client-side only)
          if (typeof window !== 'undefined') {
            localStorage.setItem('token', accessToken);
            localStorage.setItem('refreshToken', newRefreshToken);
            localStorage.setItem('user', JSON.stringify(userData));
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