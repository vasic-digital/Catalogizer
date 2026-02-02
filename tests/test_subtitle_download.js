const http = require('http');

// Token from previous login
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

// Test subtitle download with proper parameters
const postData = JSON.stringify({
  media_item_id: 1,
  result_id: 'opensubtitles_1',
  language: 'en',
  verify_sync: false
});

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/v1/subtitles/download',
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(postData)
  }
};

const req = http.request(options, (res) => {
  console.log(`Status: ${res.statusCode}`);
  console.log(`Headers: ${JSON.stringify(res.headers, null, 2)}`);
  
  let data = '';
  res.on('data', (chunk) => {
    data += chunk;
  });
  
  res.on('end', () => {
    console.log('Response body:', data);
    try {
      const parsed = JSON.parse(data);
      if (parsed.success) {
        console.log('✅ Subtitle download test successful!');
        console.log(`✅ Download URL: ${parsed.download_url}`);
        console.log(`✅ File path: ${parsed.file_path}`);
        console.log(`✅ File size: ${parsed.file_size}`);
      } else {
        console.log('❌ Subtitle download failed:', parsed.error);
      }
    } catch (e) {
      console.log('❌ Response is not valid JSON');
    }
  });
});

req.on('error', (e) => {
  console.error(`Problem with request: ${e.message}`);
});

req.write(postData);
req.end();