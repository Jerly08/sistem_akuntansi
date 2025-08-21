// Token Refresh Test Utility
// Use in browser console to test token refresh flow

async function testTokenRefresh() {
  console.log("=== TOKEN REFRESH TEST ===");
  
  // Get current tokens
  const accessToken = localStorage.getItem('accessToken');
  const refreshToken = localStorage.getItem('refreshToken');
  
  if (!accessToken || !refreshToken) {
    console.error("❌ No tokens found in localStorage. Please login first.");
    return;
  }
  
  console.log("Current Access Token:", accessToken);
  console.log("Current Refresh Token:", refreshToken);
  
  // Display decoded token data using jwt-decode (if available in your app)
  try {
    // This requires jwt-decode to be loaded
    const decoded = window.jwt_decode(accessToken);
    console.log("Decoded Access Token:", decoded);
    
    // Check token expiration
    const expiryDate = new Date(decoded.exp * 1000);
    const now = new Date();
    console.log("Token expires:", expiryDate);
    console.log("Current time:", now);
    console.log("Seconds until expiry:", Math.floor((expiryDate - now) / 1000));
    
    // Check role claim
    if (decoded.role) {
      console.log("👤 User Role:", decoded.role);
    } else {
      console.error("❌ No 'role' claim in token!");
    }
  } catch (e) {
    console.log("Could not decode token:", e);
  }
  
  // Test token refresh
  console.log("\n🔄 Testing token refresh...");
  try {
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      console.error(`❌ Token refresh failed: ${response.status} ${response.statusText}`);
      console.error("Error details:", errorText);
      return;
    }
    
    const data = await response.json();
    console.log("✅ Token refresh successful!");
    console.log("New Access Token:", data.access_token);
    console.log("Refresh Token (should be same):", data.refresh_token);
    
    // Try to decode the new token
    try {
      const newDecoded = window.jwt_decode(data.access_token);
      console.log("New Decoded Token:", newDecoded);
      
      // Check role claim in new token
      if (newDecoded.role) {
        console.log("👤 User Role in new token:", newDecoded.role);
      } else {
        console.error("❌ No 'role' claim in new token!");
      }
    } catch (e) {
      console.log("Could not decode new token:", e);
    }
    
    // Test a protected endpoint that requires role
    console.log("\n🔒 Testing protected endpoint with new token...");
    const testResponse = await fetch('/api/payments', {
      headers: {
        'Authorization': `Bearer ${data.access_token}`
      }
    });
    
    if (testResponse.ok) {
      console.log("✅ Protected endpoint access successful!");
    } else {
      console.error(`❌ Protected endpoint access failed: ${testResponse.status} ${testResponse.statusText}`);
      const errorDetails = await testResponse.text();
      console.error("Error details:", errorDetails);
    }
    
  } catch (error) {
    console.error("❌ Error during token refresh test:", error);
  }
  
  console.log("=== TEST COMPLETE ===");
}

// Execute the test
testTokenRefresh();
