import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

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
const subscribers: ((token: string) => void)[] = [];

function onRefreshed(token: string) {
  subscribers.forEach((cb) => cb(token));
  subscribers.length = 0;
}

function subscribeTokenRefresh(cb: (token: string) => void) {
  subscribers.push(cb);
}

api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config || {};
    const status = error.response?.status;

    if (status === 401 && typeof window !== 'undefined') {
      try {
        const storedRefresh = window.localStorage.getItem('refreshToken');
        if (!storedRefresh) {
          // No refresh token, propagate error
          return Promise.reject(error);
        }

        if (!isRefreshing) {
          isRefreshing = true;
          refreshPromise = (async () => {
            try {
              const resp = await fetch(`${API_BASE_URL}/auth/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
body: JSON.stringify({ refresh_token: storedRefresh }),
              });
              if (!resp.ok) throw new Error('Failed to refresh token');
              const data = await resp.json();
              // Persist new tokens
              window.localStorage.setItem('token', data.token);
              if (data.refreshToken) {
                window.localStorage.setItem('refreshToken', data.refreshToken);
              }
              if (data.user) {
                window.localStorage.setItem('user', JSON.stringify(data.user));
              }
              onRefreshed(data.token);
              return data.token as string;
            } finally {
              isRefreshing = false;
              refreshPromise = null;
            }
          })();
        }

        const newToken = await (refreshPromise as Promise<string>);

        // Retry original request with new token
        return new Promise((resolve) => {
          subscribeTokenRefresh((token: string) => {
            if (!originalRequest.headers) originalRequest.headers = {};
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(api(originalRequest));
          });
        });
      } catch (e) {
        // If refresh fails, propagate the original error
        return Promise.reject(error);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
