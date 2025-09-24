/**
 * EMERGENCY KILL SWITCH
 * Run this in browser console to stop all loops immediately
 */

console.log('🚨 EMERGENCY KILL SWITCH ACTIVATED');

// 1. Close all existing toasts
try {
  document.querySelectorAll('[role="alert"], [role="status"], .chakra-toast').forEach(el => {
    el.style.display = 'none';
    el.remove();
  });
  console.log('✅ All toasts removed');
} catch (e) {
  console.log('⚠️ Could not remove toasts:', e.message);
}

// 2. Block all WebSocket connections
const originalWebSocket = window.WebSocket;
let blockedConnections = 0;
window.WebSocket = function(url) {
  blockedConnections++;
  console.log(`🚫 WebSocket connection blocked #${blockedConnections}: ${url}`);
  return {
    close: () => {},
    send: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    readyState: 3, // CLOSED
    onopen: null,
    onclose: null,
    onerror: null,
    onmessage: null
  };
};

// 3. Block all API calls to ssot-reports
const originalFetch = window.fetch;
let blockedAPIs = 0;
window.fetch = function(url, options) {
  if (url.includes('/ssot-reports/')) {
    blockedAPIs++;
    console.log(`🚫 SSOT API call blocked #${blockedAPIs}: ${url}`);
    return Promise.reject(new Error('API call blocked by kill switch'));
  }
  return originalFetch.apply(this, arguments);
};

// 4. Override toast functions
if (window.toast) {
  window.toast = () => console.log('🚫 Toast blocked by kill switch');
}

// 5. Block setInterval/setTimeout for auto-refresh
const originalSetInterval = window.setInterval;
let blockedIntervals = 0;
window.setInterval = function(callback, delay) {
  // Block intervals shorter than 10 seconds (likely polling/refresh)
  if (delay < 10000) {
    blockedIntervals++;
    console.log(`🚫 Interval blocked #${blockedIntervals}: ${delay}ms`);
    return -1; // Invalid interval ID
  }
  return originalSetInterval.apply(this, arguments);
};

console.log('🛡️ Kill switch active:');
console.log(`   - WebSocket connections blocked: ${blockedConnections}`);
console.log(`   - API calls blocked: ${blockedAPIs}`);
console.log(`   - Intervals blocked: ${blockedIntervals}`);
console.log('');
console.log('🔄 To restore normal functionality, refresh the page:');
console.log('   location.reload()');
console.log('');
console.log('⏰ Auto-refresh in 10 seconds...');

setTimeout(() => {
  console.log('🔄 Auto-refreshing to restore functionality...');
  location.reload();
}, 10000);