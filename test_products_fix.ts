// Simple test to verify the PRODUCTS endpoint is now available
import { API_ENDPOINTS } from './frontend/src/config/api';

console.log('Testing PRODUCTS endpoint:');
console.log('PRODUCTS:', API_ENDPOINTS.PRODUCTS);
console.log('PRODUCTS_BY_ID(123):', API_ENDPOINTS.PRODUCTS_BY_ID(123));

// Verify the PRODUCTS endpoint is defined
if (API_ENDPOINTS.PRODUCTS !== undefined) {
  console.log('✓ PRODUCTS endpoint is defined:', API_ENDPOINTS.PRODUCTS);
} else {
  console.log('✗ PRODUCTS endpoint is missing');
}

if (typeof API_ENDPOINTS.PRODUCTS_BY_ID === 'function') {
  console.log('✓ PRODUCTS_BY_ID function is defined:', API_ENDPOINTS.PRODUCTS_BY_ID(123));
} else {
  console.log('✗ PRODUCTS_BY_ID function is missing');
}

console.log('\n✅ PRODUCTS endpoint fix verification complete!');