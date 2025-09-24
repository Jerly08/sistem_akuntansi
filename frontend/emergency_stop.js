/**
 * EMERGENCY STOP FUNCTION
 * Paste this in browser console to immediately stop all toast spam
 */

console.log('🚨 EMERGENCY STOP - Clearing all toasts and WebSocket connections');

// 1. Force close all toasts immediately
function clearAllToasts() {
  try {
    // Chakra UI toast removal
    const toastContainer = document.querySelector('[data-toast-manager]');
    if (toastContainer) {
      const allToasts = toastContainer.querySelectorAll('[role="alert"], [role="status"]');
      console.log(`Removing ${allToasts.length} active toasts`);
      allToasts.forEach(toast => {
        toast.style.display = 'none';
        toast.remove();
      });
    }
    
    // Alternative method - remove all toast-like elements
    const toastElements = document.querySelectorAll('.chakra-toast, .chakra-alert, .Toastify__toast');
    toastElements.forEach(el => {
      el.style.display = 'none';
      el.remove();
    });
    
    console.log('✅ All toasts cleared');
  } catch (error) {
    console.error('Error clearing toasts:', error);
  }
}

// 2. Block new toast creation temporarily
let originalToast = null;
function blockToasts() {
  try {
    // Find toast function and override it
    if (window.toast) {
      if (!originalToast) originalToast = window.toast;
      window.toast = function() {
        console.log('Toast blocked:', arguments);
        return null;
      };
    }
    
    // Override console methods that might trigger toasts
    const originalConsoleError = console.error;
    console.error = function(...args) {
      if (args[0] && args[0].toString().includes('toast')) {
        console.log('Toast error blocked:', args);
        return;
      }
      originalConsoleError.apply(console, args);
    };
    
    console.log('✅ Toast creation blocked');
  } catch (error) {
    console.error('Error blocking toasts:', error);
  }
}

// 3. Stop all WebSocket connections
function stopWebSockets() {
  try {
    // Close any existing WebSockets
    if (window.WebSocket) {
      let wsCount = 0;
      const OriginalWebSocket = window.WebSocket;
      
      // Override WebSocket to prevent new connections temporarily
      window.WebSocket = function(url) {
        console.log('WebSocket creation blocked:', url);
        return {
          close: () => {},
          send: () => {},
          addEventListener: () => {},
          removeEventListener: () => {},
          readyState: 3 // CLOSED
        };
      };
      
      // Restore after 10 seconds
      setTimeout(() => {
        window.WebSocket = OriginalWebSocket;
        console.log('✅ WebSocket restored');
      }, 10000);
      
      console.log('✅ WebSocket connections blocked');
    }
  } catch (error) {
    console.error('Error stopping WebSockets:', error);
  }
}

// 4. Clear React state if accessible
function clearReactState() {
  try {
    // Try to access React DevTools if available
    if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
      console.log('React DevTools detected - attempting state reset');
      
      // This is a hack - in production you'd need proper state management
      const reactFiberNodes = document.querySelectorAll('[data-reactroot] *');
      reactFiberNodes.forEach(node => {
        if (node._reactInternalFiber || node._reactInternalInstance) {
          console.log('Found React fiber node');
        }
      });
    }
    
    console.log('⚠️ React state clearing attempted');
  } catch (error) {
    console.log('Could not access React state:', error.message);
  }
}

// Execute emergency stop
console.log('🛑 Executing emergency stop sequence...');
clearAllToasts();
blockToasts();
stopWebSockets();
clearReactState();

console.log('✅ Emergency stop completed!');
console.log('💡 Refresh the page to restore normal functionality');
console.log('🔧 Or run: location.reload() to refresh');

// Auto-refresh after 5 seconds option
console.log('⏰ Auto-refresh in 5 seconds (cancel with clearTimeout if needed)');
const refreshTimer = setTimeout(() => {
  console.log('🔄 Auto-refreshing page...');
  location.reload();
}, 5000);

console.log(`Timer ID: ${refreshTimer} (use clearTimeout(${refreshTimer}) to cancel)`);