import { useEnhancedToast, ToastNotificationOptions } from '@/components/common/ToastNotification';
import { useTranslation } from './useTranslation';

/**
 * Custom hook that wraps useEnhancedToast with translation support.
 * All toast messages are automatically translated based on the current language.
 */
export const useTranslatedToast = () => {
  const toast = useEnhancedToast();
  const { t } = useTranslation();

  /**
   * Helper function to interpolate values into translated strings
   * Supports {{key}} syntax for interpolation
   */
  const interpolate = (text: string, values?: Record<string, string | number>): string => {
    if (!values) return text;
    return Object.entries(values).reduce((result, [key, value]) => {
      return result.replace(new RegExp(`{{${key}}}`, 'g'), String(value));
    }, text);
  };

  /**
   * Show a success toast with translated messages
   */
  const success = (
    titleKey: string,
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const title = interpolate(t(titleKey), interpolations);
    const description = interpolate(t(descriptionKey), interpolations);
    toast.success(title, description, options);
  };

  /**
   * Show an error toast with translated messages
   */
  const error = (
    titleKey: string,
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const title = interpolate(t(titleKey), interpolations);
    const description = interpolate(t(descriptionKey), interpolations);
    toast.error(title, description, options);
  };

  /**
   * Show a warning toast with translated messages
   */
  const warning = (
    titleKey: string,
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const title = interpolate(t(titleKey), interpolations);
    const description = interpolate(t(descriptionKey), interpolations);
    toast.warning(title, description, options);
  };

  /**
   * Show an info toast with translated messages
   */
  const info = (
    titleKey: string,
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const title = interpolate(t(titleKey), interpolations);
    const description = interpolate(t(descriptionKey), interpolations);
    toast.info(title, description, options);
  };

  /**
   * Show a validation error toast with translated message
   */
  const validationError = (
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const description = interpolate(t(descriptionKey), interpolations);
    toast.showToast({
      type: 'warning',
      title: t('messages.toast.validationError'),
      description,
      duration: 6000,
      ...options,
    });
  };

  /**
   * Show a network error toast with translated messages
   */
  const networkError = (options?: Partial<ToastNotificationOptions>) => {
    toast.showToast({
      type: 'error',
      title: t('messages.toast.connectionError'),
      description: t('messages.toast.connectionErrorDesc'),
      duration: 8000,
      ...options,
    });
  };

  /**
   * Show a server error toast with translated messages
   */
  const serverError = (options?: Partial<ToastNotificationOptions>) => {
    toast.showToast({
      type: 'error',
      title: t('messages.toast.serverError'),
      description: t('messages.toast.serverErrorDesc'),
      duration: 6000,
      ...options,
    });
  };

  /**
   * Show a permission error toast with translated messages
   */
  const permissionError = (
    actionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const action = interpolate(t(actionKey), interpolations);
    toast.showToast({
      type: 'warning',
      title: t('messages.toast.accessDenied'),
      description: `${t('messages.toast.accessDeniedDesc')} (${action})`,
      duration: 5000,
      ...options,
    });
  };

  /**
   * Show a save success toast with translated messages
   */
  const saveSuccess = (
    resourceNameKey: string,
    isUpdate = false,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const resourceName = t(resourceNameKey);
    const titleKey = isUpdate ? 'messages.toast.updateSuccess' : 'messages.toast.saveSuccess';
    const descKey = isUpdate ? 'messages.toast.updateSuccessDesc' : 'messages.toast.saveSuccessDesc';
    
    toast.showToast({
      type: 'success',
      title: t(titleKey),
      description: `${resourceName} - ${t(descKey)}`,
      duration: 3000,
      ...options,
    });
  };

  /**
   * Show a create success toast with translated messages
   */
  const createSuccess = (
    resourceNameKey: string,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const resourceName = t(resourceNameKey);
    toast.showToast({
      type: 'success',
      title: t('messages.toast.createSuccess'),
      description: `${resourceName} - ${t('messages.toast.createSuccessDesc')}`,
      duration: 3000,
      ...options,
    });
  };

  /**
   * Show a delete success toast with translated messages
   */
  const deleteSuccess = (
    resourceNameKey: string,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const resourceName = t(resourceNameKey);
    toast.showToast({
      type: 'success',
      title: t('messages.toast.deleteSuccess'),
      description: `${resourceName} - ${t('messages.toast.deleteSuccessDesc')}`,
      duration: 3000,
      ...options,
    });
  };

  /**
   * Show an insufficient stock warning toast with translated messages
   */
  const insufficientStock = (
    productName: string,
    available: number,
    options?: Partial<ToastNotificationOptions>
  ) => {
    toast.showToast({
      type: 'warning',
      title: t('messages.toast.insufficientStock'),
      description: `${t('messages.toast.insufficientStockDesc')} (${productName}: ${available})`,
      duration: 6000,
      ...options,
    });
  };

  /**
   * Show a credit limit exceeded warning toast with translated messages
   */
  const creditLimitExceeded = (
    availableCredit: number,
    options?: Partial<ToastNotificationOptions>
  ) => {
    toast.showToast({
      type: 'warning',
      title: t('messages.toast.creditLimitExceeded'),
      description: `${t('messages.toast.creditLimitExceededDesc')} (${availableCredit})`,
      duration: 6000,
      ...options,
    });
  };

  /**
   * Show a session expired error toast with translated messages
   */
  const sessionExpired = (options?: Partial<ToastNotificationOptions>) => {
    toast.showToast({
      type: 'error',
      title: t('messages.toast.sessionExpired'),
      description: t('messages.toast.sessionExpiredDesc'),
      duration: 6000,
      ...options,
    });
  };

  /**
   * Show an export success toast with translated messages
   */
  const exportSuccess = (options?: Partial<ToastNotificationOptions>) => {
    toast.showToast({
      type: 'success',
      title: t('messages.toast.exportSuccess'),
      description: t('messages.toast.exportSuccessDesc'),
      duration: 3000,
      ...options,
    });
  };

  /**
   * Show an import success toast with translated messages
   */
  const importSuccess = (options?: Partial<ToastNotificationOptions>) => {
    toast.showToast({
      type: 'success',
      title: t('messages.toast.importSuccess'),
      description: t('messages.toast.importSuccessDesc'),
      duration: 3000,
      ...options,
    });
  };

  /**
   * Show a generic toast with custom translated keys
   */
  const showTranslated = (
    type: 'success' | 'error' | 'warning' | 'info',
    titleKey: string,
    descriptionKey: string,
    interpolations?: Record<string, string | number>,
    options?: Partial<ToastNotificationOptions>
  ) => {
    const title = interpolate(t(titleKey), interpolations);
    const description = interpolate(t(descriptionKey), interpolations);
    toast.showToast({
      type,
      title,
      description,
      ...options,
    });
  };

  return {
    // Basic toast types with translation
    success,
    error,
    warning,
    info,
    
    // Business logic specific toasts
    validationError,
    networkError,
    serverError,
    permissionError,
    saveSuccess,
    createSuccess,
    deleteSuccess,
    insufficientStock,
    creditLimitExceeded,
    sessionExpired,
    exportSuccess,
    importSuccess,
    
    // Generic translated toast
    showTranslated,
    
    // Access to underlying toast for custom usage
    showToast: toast.showToast,
  };
};

export default useTranslatedToast;
