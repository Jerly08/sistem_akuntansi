import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL;

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    try {
      if (typeof window !== 'undefined') {
        const token = window.localStorage.getItem('token');
        if (token) {
          (config.headers as any).Authorization = `Bearer ${token}`;
          // Debug: log token usage for payment requests only
          if (config.url?.includes('payments')) {
            console.log('Request interceptor - Using token (length):', token.length);
          }
        }
      }
    } catch (e) {
      // ignore localStorage errors (SSR or privacy mode)
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle common errors with silent refresh on 401
let isRefreshing = false;
let refreshPromise: Promise<any> | null = null;
let refreshAttempts = 0;
const MAX_REFRESH_ATTEMPTS = 3;
const failedQueue: Array<{resolve: (token: string) => void, reject: (error: any) => void}> = [];

function processQueue(error: any, token: string | null = null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
    } else {
      resolve(token!);
    }
  });
  
  failedQueue.length = 0;
}

api.interceptors.response.use(
  (response) => {
    // Reset refresh attempts on successful response
    refreshAttempts = 0;
    return response;
  },
  async (error) => {
    const originalRequest = error.config || {};
    const status = error.response?.status;

    // Log only for 401 errors that might need refresh
    if (status === 401) {
      console.log('API Interceptor - 401 Error on:', originalRequest.url);
    }

    if (status === 401 && typeof window !== 'undefined' && !originalRequest._retry) {
      originalRequest._retry = true;
      
      // Circuit breaker: prevent infinite refresh loops
      if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
        console.error('API Interceptor - Max refresh attempts reached, forcing logout');
        
        // Clear auth data
        window.localStorage.removeItem('token');
        window.localStorage.removeItem('refreshToken');
        window.localStorage.removeItem('user');
        
        // Reset counter
        refreshAttempts = 0;
        
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        return Promise.reject(authError);
      }
      
      refreshAttempts++;
      console.log(`API Interceptor - Attempting token refresh for 401 (attempt ${refreshAttempts}/${MAX_REFRESH_ATTEMPTS})`);

      try {
        const storedRefresh = window.localStorage.getItem('refreshToken');
        const storedToken = window.localStorage.getItem('token');

        if (!storedRefresh) {
          // No refresh token, clear auth and return specific error
          console.warn('API Interceptor - No refresh token available');
          window.localStorage.removeItem('token');
          window.localStorage.removeItem('refreshToken');
          window.localStorage.removeItem('user');
          
          // Return a more specific error for the frontend
          const authError = new Error('Session expired. Please login again.');
          (authError as any).isAuthError = true;
          (authError as any).code = 'AUTH_SESSION_EXPIRED';
          return Promise.reject(authError);
        }

        if (!isRefreshing) {
          isRefreshing = true;
          refreshPromise = (async () => {
            try {
              // Simple timeout wrapper
              const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => reject(new Error('Request timeout')), 10000);
              });
              
              const fetchPromise = fetch(`${API_BASE_URL}/auth/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: storedRefresh }),
              });
              
              const resp = await Promise.race([fetchPromise, timeoutPromise]) as Response;
              
              if (!resp.ok) {
                const errorText = await resp.text();
                console.error('Refresh failed with:', errorText);
                
                if (resp.status === 401 || resp.status === 403 || resp.status === 400) {
                  throw new Error('Refresh token expired or invalid');
                }
                throw new Error(`Failed to refresh token: ${resp.status}`);
              }
              
              const data = await resp.json();
              
              console.log('Token refresh response data:', data);
              
              // Handle different possible response structures
              const newAccessToken = data.token || data.accessToken || data.access_token;
              const newRefreshToken = data.refreshToken || data.refresh_token;
              
              if (!newAccessToken) {
                console.error('No access token received from refresh:', data);
                throw new Error('No access token received from refresh');
              }
              
              console.log('Token refresh successful, new token length:', newAccessToken.length);
              
              // Persist new tokens
              window.localStorage.setItem('token', newAccessToken);
              if (newRefreshToken) {
                window.localStorage.setItem('refreshToken', newRefreshToken);
              }
              if (data.user) {
                window.localStorage.setItem('user', JSON.stringify(data.user));
              }
              
              processQueue(null, newAccessToken);
              return newAccessToken as string;
            } catch (refreshError) {
              // If refresh fails, clear auth data
              console.error('Token refresh failed:', refreshError);
              window.localStorage.removeItem('token');
              window.localStorage.removeItem('refreshToken');
              window.localStorage.removeItem('user');
              
              processQueue(refreshError, null);
              throw refreshError;
            } finally {
              isRefreshing = false;
              refreshPromise = null;
            }
          })();
        }

        const newToken = await (refreshPromise as Promise<string>);
        console.log('API Interceptor - Retrying request with new token length:', newToken.length);
        console.log('API Interceptor - Original request URL:', originalRequest.url);
        console.log('API Interceptor - Original request headers before update:', originalRequest.headers);

        // Create a fresh config instead of modifying the original
        const retryConfig = {
          ...originalRequest,
          headers: {
            ...originalRequest.headers,
            Authorization: `Bearer ${newToken}`
          }
        };
        delete retryConfig._retry;
        
        console.log('API Interceptor - Retry config headers:', retryConfig.headers);
        
        return api(retryConfig);
      } catch (refreshError) {
        // If refresh fails, handle gracefully
        console.error('API Interceptor - Token refresh completely failed:', refreshError);
        
        // Clear all auth data
        window.localStorage.removeItem('token');
        window.localStorage.removeItem('refreshToken');
        window.localStorage.removeItem('user');
        
        // Return a more specific error for the frontend
        const authError = new Error('Session expired. Please login again.');
        (authError as any).isAuthError = true;
        (authError as any).code = 'AUTH_SESSION_EXPIRED';
        (authError as any).originalError = refreshError;
        return Promise.reject(authError);
      }
    }

    // For other errors, pass them through
    return Promise.reject(error);
  }
);

export default api;
