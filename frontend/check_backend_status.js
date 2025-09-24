/**
 * Backend Status Checker
 * 
 * Script untuk mengecek status backend dan mencari endpoint yang tidak memerlukan auth
 */

const API_V1_BASE = 'http://localhost:8080/api/v1';

async function checkBackendStatus() {
  console.log('🔍 Backend Status Checker');
  console.log('='.repeat(50));
  console.log(`🌐 API Base: ${API_V1_BASE}`);
  console.log('='.repeat(50));
  
  // List of endpoints to test
  const endpointsToTest = [
    { path: '/health', name: 'Health Check', requiresAuth: false },
    { path: '/docs', name: 'Documentation', requiresAuth: false },
    { path: '/swagger', name: 'Swagger UI', requiresAuth: false },
    { path: '/auth/login', name: 'Login Endpoint', requiresAuth: false },
    { path: '/journals/summary', name: 'Journal Summary', requiresAuth: true },
    { path: '/journals/account-balances', name: 'Account Balances', requiresAuth: true },
    { path: '/accounts', name: 'Accounts', requiresAuth: true },
    { path: '/journals', name: 'Journals', requiresAuth: true },
  ];

  console.log('\n📡 Testing Public Endpoints (No Auth Required)...');
  console.log('-'.repeat(50));
  
  for (const endpoint of endpointsToTest.filter(e => !e.requiresAuth)) {
    try {
      const url = `${API_V1_BASE}${endpoint.path}`;
      const startTime = Date.now();
      
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      });
      
      const endTime = Date.now();
      const responseTime = endTime - startTime;
      
      if (response.ok) {
        console.log(`✅ ${endpoint.name}: ${response.status} ${response.statusText} (${responseTime}ms)`);
        
        // Try to get response data
        try {
          const contentType = response.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const data = await response.json();
            console.log(`   Response structure: ${Object.keys(data).join(', ')}`);
          } else {
            const text = await response.text();
            console.log(`   Response type: ${contentType || 'unknown'}, length: ${text.length}`);
          }
        } catch (parseError) {
          console.log(`   Could not parse response: ${parseError.message}`);
        }
      } else {
        const errorText = await response.text();
        console.log(`❌ ${endpoint.name}: ${response.status} ${response.statusText}`);
        console.log(`   Error: ${errorText.substring(0, 100)}...`);
      }
    } catch (error) {
      console.log(`❌ ${endpoint.name}: ${error.message}`);
    }
  }

  console.log('\n🔐 Testing Protected Endpoints (Auth Required)...');
  console.log('-'.repeat(50));
  
  for (const endpoint of endpointsToTest.filter(e => e.requiresAuth)) {
    try {
      const url = `${API_V1_BASE}${endpoint.path}`;
      
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        }
      });
      
      if (response.status === 401) {
        console.log(`🔒 ${endpoint.name}: Requires Authentication (${response.status})`);
      } else if (response.ok) {
        console.log(`⚠️  ${endpoint.name}: Unexpectedly accessible without auth (${response.status})`);
      } else {
        console.log(`❌ ${endpoint.name}: ${response.status} ${response.statusText}`);
      }
    } catch (error) {
      console.log(`❌ ${endpoint.name}: ${error.message}`);
    }
  }

  console.log('\n🎯 Login Test...');
  console.log('-'.repeat(50));
  
  // Try to get login endpoint info
  try {
    const loginResponse = await fetch(`${API_V1_BASE}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify({
        email: 'test@example.com',
        password: 'testpassword'
      })
    });
    
    if (loginResponse.status === 422 || loginResponse.status === 400) {
      console.log('✅ Login endpoint is accessible (validation error expected)');
      try {
        const errorData = await loginResponse.json();
        console.log('   Expected validation response structure:', Object.keys(errorData));
      } catch (e) {
        console.log('   Could not parse validation response');
      }
    } else if (loginResponse.status === 401) {
      console.log('✅ Login endpoint is accessible (authentication failed as expected)');
    } else {
      console.log(`⚠️  Login endpoint returned unexpected status: ${loginResponse.status}`);
    }
  } catch (error) {
    console.log(`❌ Login endpoint test failed: ${error.message}`);
  }

  console.log('\n🔍 Backend Discovery...');
  console.log('-'.repeat(50));
  
  // Try some common discovery endpoints
  const discoveryEndpoints = [
    '/api',
    '/api/v1',
    '/api/v1/status',
    '/status',
    '/ping',
    '/info',
    '/version',
  ];

  for (const endpoint of discoveryEndpoints) {
    try {
      const url = `http://localhost:8080${endpoint}`;
      const response = await fetch(url, {
        headers: { 'Accept': 'application/json' }
      });
      
      if (response.ok) {
        console.log(`✅ Found endpoint: ${endpoint} (${response.status})`);
        try {
          const data = await response.json();
          console.log(`   Data: ${JSON.stringify(data).substring(0, 100)}...`);
        } catch (e) {
          const text = await response.text();
          console.log(`   Text: ${text.substring(0, 100)}...`);
        }
      }
    } catch (error) {
      // Silently continue for discovery
    }
  }

  console.log('\n📋 Summary');
  console.log('='.repeat(50));
  console.log('✅ Backend Status Check Completed');
  console.log(`📅 ${new Date().toLocaleString('id-ID')}`);
  console.log(`🌐 Backend URL: http://localhost:8080`);
  console.log(`🔗 API URL: ${API_V1_BASE}`);
  
  console.log('\n💡 Next Steps for Balance Sheet:');
  console.log('1. If login endpoint works, create a test account');
  console.log('2. Use valid credentials to get auth token');
  console.log('3. Add token to Balance Sheet test script');
  console.log('4. Run Balance Sheet test with authentication');
  
  console.log('='.repeat(50));
}

// Run the check
checkBackendStatus()
  .then(() => {
    console.log('\n🎉 Status check completed!');
  })
  .catch(error => {
    console.error('\n💥 Status check failed:', error);
  });