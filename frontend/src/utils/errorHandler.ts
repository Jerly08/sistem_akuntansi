import { UseToastOptions } from '@chakra-ui/react';

export interface APIError {
  response?: {
    data?: {
      error?: string;
      message?: string;
      details?: any;
    };
    status?: number;
    statusText?: string;
  };
  message?: string;
  code?: string;
}

export interface ErrorHandlerOptions {
  operation?: string;
  showToast?: boolean;
  logToConsole?: boolean;
  fallbackMessage?: string;
  duration?: number;
}

export class ErrorHandler {
  /**
   * Extract meaningful error message from API error response
   */
  static extractErrorMessage(error: APIError, fallback: string = 'An unexpected error occurred'): string {
    // Try to get error message from various possible locations
    if (error?.response?.data?.error) {
      return error.response.data.error;
    }
    
    if (error?.response?.data?.message) {
      return error.response.data.message;
    }
    
    if (error?.message) {
      return error.message;
    }
    
    if (error?.response?.statusText) {
      return `${error.response.status}: ${error.response.statusText}`;
    }
    
    return fallback;
  }

  /**
   * Get error severity based on status code or error type
   */
  static getErrorSeverity(error: APIError): 'error' | 'warning' | 'info' {
    const status = error?.response?.status;
    
    if (status >= 500) return 'error';     // Server errors
    if (status >= 400) return 'warning';   // Client errors
    if (status >= 300) return 'info';      // Redirects
    
    return 'error'; // Default to error for unknown cases
  }

  /**
   * Handle API error with consistent formatting and logging
   */
  static handleAPIError(
    error: APIError, 
    toast: (options: UseToastOptions) => void,
    options: ErrorHandlerOptions = {}
  ): string {
    const {
      operation = 'operation',
      showToast = true,
      logToConsole = true,
      fallbackMessage,
      duration = 5000
    } = options;

    const errorMessage = this.extractErrorMessage(
      error, 
      fallbackMessage || `Failed to ${operation}`
    );
    
    const severity = this.getErrorSeverity(error);
    const status = error?.response?.status;

    // Log to console if enabled
    if (logToConsole) {
      console.group(`🚨 Error in ${operation}`);
      console.error('Error message:', errorMessage);
      console.error('Status:', status);
      console.error('Full error:', error);
      console.groupEnd();
    }

    // Show toast notification if enabled
    if (showToast) {
      const title = this.getErrorTitle(operation, severity);
      
      toast({
        title,
        description: errorMessage,
        status: severity,
        duration,
        isClosable: true,
        position: 'top-right'
      });
    }

    return errorMessage;
  }

  /**
   * Get appropriate error title based on operation and severity
   */
  static getErrorTitle(operation: string, severity: 'error' | 'warning' | 'info'): string {
    const operationName = operation.charAt(0).toUpperCase() + operation.slice(1);
    
    switch (severity) {
      case 'error':
        return `${operationName} Failed`;
      case 'warning':
        return `${operationName} Warning`;
      case 'info':
        return `${operationName} Info`;
      default:
        return `${operationName} Error`;
    }
  }

  /**
   * Handle validation errors specifically
   */
  static handleValidationError(
    errors: string[],
    toast: (options: UseToastOptions) => void,
    operation: string = 'validation'
  ): void {
    const errorMessage = errors.length === 1 
      ? errors[0] 
      : `Multiple validation errors:\n• ${errors.join('\n• ')}`;

    toast({
      title: 'Validation Error',
      description: errorMessage,
      status: 'warning',
      duration: 6000,
      isClosable: true,
      position: 'top-right'
    });

    console.warn(`Validation errors in ${operation}:`, errors);
  }

  /**
   * Handle success messages consistently
   */
  static handleSuccess(
    message: string,
    toast: (options: UseToastOptions) => void,
    operation: string = 'operation'
  ): void {
    toast({
      title: 'Success',
      description: message,
      status: 'success',
      duration: 3000,
      isClosable: true,
      position: 'top-right'
    });

    console.log(`✅ ${operation} success:`, message);
  }

  /**
   * Handle network/connection errors
   */
  static handleNetworkError(
    error: APIError,
    toast: (options: UseToastOptions) => void
  ): string {
    const isNetworkError = !error?.response || error?.code === 'NETWORK_ERROR';
    
    if (isNetworkError) {
      const message = 'Network connection failed. Please check your internet connection.';
      
      toast({
        title: 'Connection Error',
        description: message,
        status: 'error',
        duration: 8000,
        isClosable: true,
        position: 'top-right'
      });
      
      return message;
    }
    
    return this.handleAPIError(error, toast, {
      operation: 'network request',
      fallbackMessage: 'Network request failed'
    });
  }

  /**
   * Handle service unavailable errors
   */
  static handleServiceUnavailable(
    serviceName: string,
    toast: (options: UseToastOptions) => void,
    fallbackAction?: string
  ): void {
    const message = fallbackAction 
      ? `${serviceName} is currently unavailable. ${fallbackAction}`
      : `${serviceName} service is temporarily unavailable. Please try again later.`;

    toast({
      title: 'Service Unavailable',
      description: message,
      status: 'warning',
      duration: 6000,
      isClosable: true,
      position: 'top-right'
    });

    console.warn(`Service unavailable: ${serviceName}`);
  }

  /**
   * Handle loading state errors
   */
  static handleLoadingError(
    resourceName: string,
    error: APIError,
    toast: (options: UseToastOptions) => void
  ): string {
    return this.handleAPIError(error, toast, {
      operation: `load ${resourceName}`,
      fallbackMessage: `Failed to load ${resourceName}. Please refresh and try again.`
    });
  }

  /**
   * Handle save/update errors
   */
  static handleSaveError(
    resourceName: string,
    error: APIError,
    toast: (options: UseToastOptions) => void,
    isUpdate: boolean = false
  ): string {
    const operation = isUpdate ? `update ${resourceName}` : `create ${resourceName}`;
    
    return this.handleAPIError(error, toast, {
      operation,
      fallbackMessage: `Failed to ${operation}. Please check your input and try again.`
    });
  }

  /**
   * Handle delete errors
   */
  static handleDeleteError(
    resourceName: string,
    error: APIError,
    toast: (options: UseToastOptions) => void
  ): string {
    return this.handleAPIError(error, toast, {
      operation: `delete ${resourceName}`,
      fallbackMessage: `Failed to delete ${resourceName}. It may be in use by other records.`
    });
  }
}

// Export convenience functions
export const handleAPIError = ErrorHandler.handleAPIError.bind(ErrorHandler);
export const handleValidationError = ErrorHandler.handleValidationError.bind(ErrorHandler);
export const handleSuccess = ErrorHandler.handleSuccess.bind(ErrorHandler);
export const handleNetworkError = ErrorHandler.handleNetworkError.bind(ErrorHandler);
export const handleServiceUnavailable = ErrorHandler.handleServiceUnavailable.bind(ErrorHandler);
export const handleLoadingError = ErrorHandler.handleLoadingError.bind(ErrorHandler);
export const handleSaveError = ErrorHandler.handleSaveError.bind(ErrorHandler);
export const handleDeleteError = ErrorHandler.handleDeleteError.bind(ErrorHandler);

export default ErrorHandler;
