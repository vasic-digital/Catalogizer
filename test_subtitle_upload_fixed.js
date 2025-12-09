const http = require('http');

// Token from previous login
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

// Create test subtitle content
const subtitleContent = `1
00:00:01,000 --> 00:00:03,000
Test subtitle line 1

2
00:00:04,000 --> 00:00:06,000
Test subtitle line 2`;

// Create JSON payload for upload
const uploadData = {
  media_item_id: 1,
  language: "English",
  language_code: "en",
  format: "srt",
  content: subtitleContent,
  is_default: false,
  is_forced: false,
  encoding: "utf-8",
  sync_offset: 0.0
};

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/v1/subtitles/upload',
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(JSON.stringify(uploadData))
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
        console.log('✅ Subtitle upload test successful!');
        console.log(`✅ Subtitle ID: ${parsed.subtitle_id}`);
      } else {
        console.log('❌ Subtitle upload failed:', parsed.error);
      }
    } catch (e) {
      console.log('❌ Response is not valid JSON');
    }
  });
});

req.on('error', (e) => {
  console.error(`Problem with request: ${e.message}`);
});

req.write(JSON.stringify(uploadData));
req.end();