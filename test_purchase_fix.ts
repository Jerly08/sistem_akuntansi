// Simple test to verify the PURCHASES endpoint is now available
import { API_ENDPOINTS } from './frontend/src/config/api';

console.log('Testing PURCHASES endpoint:');
console.log('PURCHASES:', API_ENDPOINTS.PURCHASES);
console.log('PURCHASES_BY_ID(123):', API_ENDPOINTS.PURCHASES_BY_ID(123));

// Verify all purchase endpoints are defined
const purchaseEndpoints = [
  'PURCHASES',
  'PURCHASES_BY_ID',
  'PURCHASES_PENDING_APPROVAL',
  'PURCHASES_APPROVE',
  'PURCHASES_REJECT',
  'PURCHASES_APPROVAL_HISTORY',
  'PURCHASES_APPROVAL_STATS',
  'PURCHASES_SUBMIT_APPROVAL',
  'PURCHASES_SUMMARY',
  'PURCHASES_FOR_PAYMENT',
  'PURCHASES_INTEGRATED_PAYMENT',
  'PURCHASES_PAYMENTS',
  'PURCHASES_EXPORT_PDF',
  'PURCHASES_EXPORT_CSV'
];

console.log('\nVerifying all PURCHASES endpoints are defined:');
purchaseEndpoints.forEach(endpoint => {
  if ((API_ENDPOINTS as any)[endpoint] !== undefined) {
    console.log(`✓ ${endpoint}:`, typeof (API_ENDPOINTS as any)[endpoint] === 'function' 
      ? (API_ENDPOINTS as any)[endpoint](123) // Test function endpoints with ID 123
      : (API_ENDPOINTS as any)[endpoint]);
  } else {
    console.log(`✗ ${endpoint}: MISSING`);
  }
});

console.log('\n✅ PURCHASES endpoint fix verification complete!');