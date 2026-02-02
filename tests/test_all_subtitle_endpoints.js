const http = require('http');

const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

const testEndpoint = (name, path, method = 'GET', postData = null) => {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: path,
      method: method,
      headers: {
        'Authorization': 'Bearer ' + token,
        'Content-Type': 'application/json'
      }
    };

    if (postData) {
      options.headers['Content-Length'] = Buffer.byteLength(postData);
    }

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          const response = JSON.parse(data);
          console.log(`\n📡 ${name}`);
          console.log(`   Status: ${res.statusCode}`);
          console.log(`   Success: ${response.success}`);
          if (response.success) {
            console.log(`   ✅ WORKING`);
          } else {
            console.log(`   ❌ FAILED: ${response.error}`);
          }
        } catch (e) {
          console.log(`   ❌ PARSING ERROR: ${e.message}`);
        }
        resolve();
      });
    });

    req.on('error', (e) => {
      console.log(`\n📡 ${name}`);
      console.log(`   ❌ REQUEST ERROR: ${e.message}`);
      resolve();
    });

    if (postData) {
      req.write(postData);
    }
    req.end();
  });
};

async function runAllTests() {
  console.log('🚀 Running comprehensive subtitle endpoint tests...\n');

  // Test all subtitle endpoints
  await testEndpoint('GET /api/v1/subtitles/search', '/api/v1/subtitles/search?query=test&media_path=/test/path');
  await testEndpoint('POST /api/v1/subtitles/download', '/api/v1/subtitles/download', 'POST', JSON.stringify({
    media_item_id: 1,
    result_id: 'opensub_123',
    language: 'English',
    language_code: 'en',
    format: 'srt'
  }));
  await testEndpoint('GET /api/v1/subtitles/media/1', '/api/v1/subtitles/media/1');
  await testEndpoint('GET /api/v1/subtitles/1/verify-sync/1', '/api/v1/subtitles/1/verify-sync/1');
  await testEndpoint('POST /api/v1/subtitles/translate', '/api/v1/subtitles/translate', 'POST', JSON.stringify({
    subtitle_id: '1',
    source_language: 'en',
    target_language: 'es'
  }));
  await testEndpoint('GET /api/v1/subtitles/languages', '/api/v1/subtitles/languages');
  await testEndpoint('GET /api/v1/subtitles/providers', '/api/v1/subtitles/providers');

  console.log('\n✅ Comprehensive subtitle testing complete!');
}

runAllTests();