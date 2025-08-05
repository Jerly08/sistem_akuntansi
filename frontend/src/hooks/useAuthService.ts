import { useContext, useEffect } from 'react';
import { AuthContext } from '@/contexts/AuthContext';
import { accountService } from '@/services/accountService';
import { contactService } from '@/services/contactService';

/**
 * Custom hook to setup unauthorized error handling for API services
 * This should be used in the main app or layout components
 */
export const useAuthService = () => {
  const { handleUnauthorized } = useContext(AuthContext);

  useEffect(() => {
    // Set up the unauthorized handler for all services
    accountService.setUnauthorizedHandler(handleUnauthorized);
    contactService.setUnauthorizedHandler(handleUnauthorized);

    // Add other services here as they are created
    // e.g., transactionService.setUnauthorizedHandler(handleUnauthorized);

    // Cleanup function (optional - services are singletons, so this might not be necessary)
    return () => {
      // Could clear handlers here if needed
    };
  }, [handleUnauthorized]);
};
