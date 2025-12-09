const http = require('http');

// Token from previous login
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

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

// Test upload endpoint with simple JSON
function uploadSubtitleJSON() {
  return new Promise((resolve) => {
    const data = JSON.stringify({
      media_item_id: 1,
      language: "Spanish",
      language_code: "es",
      format: "srt",
      content: "1\n00:00:01,000 --> 00:00:03,000\nLínea de subtítulo 1\n\n2\n00:00:04,000 --> 00:00:06,000\nLínea de subtítulo 2",
      is_default: false,
      is_forced: false,
      encoding: "utf-8",
      sync_offset: 0.0
    });

    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/subtitles/upload',
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

async function testAdvancedFeatures() {
  console.log('=== Testing Advanced Subtitle Features ===\n');

  // Test 1: Upload subtitle via JSON
  console.log('1. Testing POST /upload (JSON payload)');
  const uploadResult = await uploadSubtitleJSON();
  console.log(`   Status: ${uploadResult.statusCode}`);
  if (uploadResult.statusCode === 200) {
    try {
      const data = JSON.parse(uploadResult.body);
      console.log(`   ✅ Success - Subtitle ID: ${data.subtitle_id}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${uploadResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${uploadResult.body}`);
  }

  // Test 2: Verify sync
  console.log('\n2. Testing GET /:subtitle_id/verify-sync/:media_id');
  const syncResult = await verifySubtitleSync();
  console.log(`   Status: ${syncResult.statusCode}`);
  if (syncResult.statusCode === 200) {
    try {
      const data = JSON.parse(syncResult.body);
      console.log(`   ✅ Success - Sync verified: ${data.success}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${syncResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${syncResult.body}`);
  }

  // Test 3: Translation
  console.log('\n3. Testing POST /translate');
  const translateResult = await translateSubtitle();
  console.log(`   Status: ${translateResult.statusCode}`);
  if (translateResult.statusCode === 200) {
    try {
      const data = JSON.parse(translateResult.body);
      console.log(`   ✅ Success - Translation completed: ${data.success}`);
    } catch (e) {
      console.log(`   ❌ Invalid JSON: ${translateResult.body}`);
    }
  } else {
    console.log(`   ❌ Failed: ${translateResult.body}`);
  }

  console.log('\n=== Advanced Features Test Complete ===');
}

testAdvancedFeatures().catch(console.error);