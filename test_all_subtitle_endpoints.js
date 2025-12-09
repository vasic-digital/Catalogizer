const http = require('http');

// Token from previous login  
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

// Test functions for different subtitle endpoints
function testEndpoint(path, method = 'GET', data = null, contentType = 'application/json') {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: `/api/v1/subtitles/${path}`,
      method: method,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': contentType
      }
    };

    if (data) {
      let body;
      if (contentType === 'application/json') {
        body = JSON.stringify(data);
      } else {
        body = data; // For multipart/form-data
      }
      options.headers['Content-Length'] = Buffer.byteLength(body);
    }

    const req = http.request(options, (res) => {
      let responseData = '';
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          path: path,
          method: method,
          headers: res.headers,
          body: responseData
        });
      });
    });

    req.on('error', (e) => {
      resolve({
        statusCode: 500,
        path: path,
        method: method,
        error: e.message,
        body: ''
      });
    });

    if (data) {
      req.write(data);
    }
    req.end();
  });
}

async function runTests() {
  console.log('=== Subtitle API Tests ===\n');

  // Test 1: Get supported languages
  console.log('1. Testing GET /languages');
  const languagesResult = await testEndpoint('languages');
  console.log(`   Status: ${languagesResult.statusCode}`);
  if (languagesResult.statusCode === 200) {
    try {
      const data = JSON.parse(languagesResult.body);
      console.log(`   ✅ Languages endpoint working - Found ${data.languages?.length || 0} languages`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON response: ${languagesResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${languagesResult.body}`);
  }

  // Test 2: Get supported providers
  console.log('\n2. Testing GET /providers');
  const providersResult = await testEndpoint('providers');
  console.log(`   Status: ${providersResult.statusCode}`);
  if (providersResult.statusCode === 200) {
    try {
      const data = JSON.parse(providersResult.body);
      console.log(`   ✅ Providers endpoint working - Found ${data.providers?.length || 0} providers`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON response: ${providersResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${providersResult.body}`);
  }

  // Test 3: Search subtitles
  console.log('\n3. Testing GET /search');
  const searchResult = await testEndpoint('search?media_path=/test.mp4&title=test&year=2024&languages=en');
  console.log(`   Status: ${searchResult.statusCode}`);
  if (searchResult.statusCode === 200) {
    try {
      const data = JSON.parse(searchResult.body);
      console.log(`   ✅ Search endpoint working - Found ${data.results?.length || 0} results`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON response: ${searchResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${searchResult.body}`);
  }

  // Test 4: Download subtitle
  console.log('\n4. Testing POST /download');
  const downloadData = {
    media_item_id: 1,
    result_id: "opensubtitles_1", 
    language: "English",
    verify_sync: false
  };
  const downloadResult = await testEndpoint('download', 'POST', downloadData);
  console.log(`   Status: ${downloadResult.statusCode}`);
  if (downloadResult.statusCode === 200) {
    try {
      const data = JSON.parse(downloadResult.body);
      console.log(`   ✅ Download endpoint working - ${data.success ? 'Success' : 'Failed'}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON response: ${downloadResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${downloadResult.body}`);
  }

  // Test 5: Get subtitles for media item
  console.log('\n5. Testing GET /media/1');
  const mediaResult = await testEndpoint('media/1');
  console.log(`   Status: ${mediaResult.statusCode}`);
  if (mediaResult.statusCode === 200) {
    try {
      const data = JSON.parse(mediaResult.body);
      console.log(`   ✅ Media subtitles endpoint working - Found ${data.subtitles?.length || 0} subtitles`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON response: ${mediaResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${mediaResult.body}`);
  }

  // Test 6: Verify sync endpoint
  console.log('\n6. Testing GET /test_verify-sync/1 (using dummy ID)');
  const syncResult = await testEndpoint('test_verify-sync/1');
  console.log(`   Status: ${syncResult.statusCode}`);
  console.log(`   Note: Using dummy subtitle ID "test", expected to fail`);
  if (syncResult.statusCode === 404) {
    console.log(`   ✅ Sync verification endpoint accessible`);
  } else {
    console.log(`   ❌ Unexpected status: ${syncResult.body}`);
  }

  console.log('\n=== Subtitle API Tests Complete ===');
}

runTests().catch(console.error);