const http = require('http');

// Token from previous login  
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

// Test media subtitles endpoint
function getMediaSubtitles() {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/subtitles/media/1',
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    const req = http.request(options, (res) => {
      let responseData = '';
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          body: responseData
        });
      });
    });

    req.on('error', (e) => {
      resolve({
        statusCode: 500,
        error: e.message,
        body: ''
      });
    });

    req.end();
  });
}

// Test sync verification endpoint
function verifySubtitleSync() {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/subtitles/sub_1765287471139464000/verify-sync/1',
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    const req = http.request(options, (res) => {
      let responseData = '';
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          body: responseData
        });
      });
    });

    req.on('error', (e) => {
      resolve({
        statusCode: 500,
        error: e.message,
        body: ''
      });
    });

    req.end();
  });
}

// Test translation endpoint
function translateSubtitle() {
  return new Promise((resolve) => {
    const data = JSON.stringify({
      subtitle_id: "sub_1765287471139464000",
      source_language: "en",
      target_language: "es",
      use_cache: true
    });

    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/subtitles/translate',
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      }
    };

    const req = http.request(options, (res) => {
      let responseData = '';
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          body: responseData
        });
      });
    });

    req.on('error', (e) => {
      resolve({
        statusCode: 500,
        error: e.message,
        body: ''
      });
    });

    req.write(data);
    req.end();
  });
}

async function testRemainingEndpoints() {
  console.log('=== Testing Remaining Subtitle Endpoints ===\n');

  // Test 1: Get subtitles for media item
  console.log('1. Testing GET /media/1 (subtitles for media item)');
  const mediaResult = await getMediaSubtitles();
  console.log(`   Status: ${mediaResult.statusCode}`);
  if (mediaResult.statusCode === 200) {
    try {
      const data = JSON.parse(mediaResult.body);
      console.log(`   ✅ Success - Found ${data.subtitles?.length || 0} subtitles`);
      if (data.subtitles && data.subtitles.length > 0) {
        console.log(`   📝 First subtitle: ${data.subtitles[0].language} (${data.subtitles[0].language_code})`);
      }
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${mediaResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${mediaResult.body}`);
  }

  // Test 2: Verify sync endpoint
  console.log('\n2. Testing GET /:subtitle_id/verify-sync/:media_id');
  const syncResult = await verifySubtitleSync();
  console.log(`   Status: ${syncResult.statusCode}`);
  if (syncResult.statusCode === 200) {
    try {
      const data = JSON.parse(syncResult.body);
      console.log(`   ✅ Success - Sync verification: ${data.success ? 'Verified' : 'Failed'}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${syncResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${syncResult.body}`);
  }

  // Test 3: Translation endpoint
  console.log('\n3. Testing POST /translate');
  const translateResult = await translateSubtitle();
  console.log(`   Status: ${translateResult.statusCode}`);
  if (translateResult.statusCode === 200) {
    try {
      const data = JSON.parse(translateResult.body);
      console.log(`   ✅ Success - Translation: ${data.success ? 'Completed' : 'Failed'}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${translateResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${translateResult.body}`);
  }

  console.log('\n=== Remaining Endpoints Test Complete ===');
}

testRemainingEndpoints().catch(console.error);