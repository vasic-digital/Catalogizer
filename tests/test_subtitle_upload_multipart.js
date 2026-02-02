const http = require('http');
const fs = require('fs');

// Token from previous login
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3R1c2VyMiIsInJvbGVfaWQiOjIsInNlc3Npb25faWQiOiIxIiwiaXNzIjoiY2F0YWxvZ2l6ZXIiLCJzdWIiOiIyIiwiZXhwIjoxNzY1MzcxMzM5LCJpYXQiOjE3NjUyODQ5Mzl9.kPs3zcDYePLbVhXUh9fA4OZmjlK-2ujCk8IrO2hs9Cg';

// Create test subtitle content
const subtitleContent = `1
00:00:01,000 --> 00:00:03,000
Test subtitle line 1

2
00:00:04,000 --> 00:00:06,000
Test subtitle line 2`;

// Create temporary subtitle file
const tempFile = '/tmp/test_subtitle.srt';
fs.writeFileSync(tempFile, subtitleContent);

// Read file as buffer
const fileBuffer = fs.readFileSync(tempFile);

// Create multipart form data
const boundary = '----WebKitFormBoundary' + Math.random().toString(16);

const formData = [
  `--${boundary}`,
  'Content-Disposition: form-data; name="media_item_id"',
  '',
  '1',
  `--${boundary}`,
  'Content-Disposition: form-data; name="language"',
  '',
  'English',
  `--${boundary}`,
  'Content-Disposition: form-data; name="language_code"',
  '',
  'en',
  `--${boundary}`,
  'Content-Disposition: form-data; name="format"',
  '',
  'srt',
  `--${boundary}`,
  `Content-Disposition: form-data; name="file"; filename="test_subtitle.srt"`,
  'Content-Type: application/octet-stream',
  '',
  fileBuffer,
  `--${boundary}--`,
  ''
];

// Calculate content length (for buffers and strings)
let totalLength = 0;
for (const part of formData) {
  if (typeof part === 'string') {
    totalLength += Buffer.byteLength(part, 'utf8');
  } else if (Buffer.isBuffer(part)) {
    totalLength += part.length;
  }
}

// Create the final form data as a single buffer
const formDataBuffer = Buffer.concat(formData.map(part => 
  typeof part === 'string' ? Buffer.from(part, 'utf8') : part
));

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/v1/subtitles/upload',
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': `multipart/form-data; boundary=${boundary}`,
    'Content-Length': formDataBuffer.length
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
        console.log(`✅ Subtitle ID: ${parsed.subtitle_id || parsed.track?.id}`);
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

req.write(formDataBuffer);
req.end();

// Clean up temp file
setTimeout(() => {
  fs.unlinkSync(tempFile);
}, 1000);