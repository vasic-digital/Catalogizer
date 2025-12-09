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

// Create multipart form data properly
const boundary = '----WebKitFormBoundary7MA4YWxkTrZu0gW';

let formData = '';

// Add media_item_id field
formData += `--${boundary}\r\n`;
formData += `Content-Disposition: form-data; name="media_item_id"\r\n\r\n`;
formData += `1\r\n`;

// Add language field
formData += `--${boundary}\r\n`;
formData += `Content-Disposition: form-data; name="language"\r\n\r\n`;
formData += `English\r\n`;

// Add language_code field
formData += `--${boundary}\r\n`;
formData += `Content-Disposition: form-data; name="language_code"\r\n\r\n`;
formData += `en\r\n`;

// Add format field
formData += `--${boundary}\r\n`;
formData += `Content-Disposition: form-data; name="format"\r\n\r\n`;
formData += `srt\r\n`;

// Add file field
formData += `--${boundary}\r\n`;
formData += `Content-Disposition: form-data; name="file"; filename="test_subtitle.srt"\r\n`;
formData += `Content-Type: application/octet-stream\r\n\r\n`;

// Create the complete form data buffer
const formDataHeader = Buffer.from(formData, 'utf8');
const formDataFooter = Buffer.from(`\r\n--${boundary}--\r\n`, 'utf8');

const totalLength = formDataHeader.length + fileBuffer.length + formDataFooter.length;

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/v1/subtitles/upload',
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': `multipart/form-data; boundary=${boundary}`,
    'Content-Length': totalLength.toString()
  }
};

const req = http.request(options, (res) => {
  console.log(`Status: ${res.statusCode}`);
  console.log('Headers:', res.headers);
  
  let data = '';
  res.on('data', (chunk) => {
    data += chunk;
  });
  
  res.on('end', () => {
    console.log('Response body:', data);
    try {
      const jsonResponse = JSON.parse(data);
      if (jsonResponse.success) {
        console.log('✅ Subtitle upload successful!');
        console.log('Response:', JSON.stringify(jsonResponse, null, 2));
      } else {
        console.log('❌ Subtitle upload failed:', jsonResponse.error);
      }
    } catch (e) {
      console.log('❌ Failed to parse JSON response:', e.message);
      console.log('Raw response:', data);
    }
  });
});

req.on('error', (e) => {
  console.error('❌ Request error:', e);
});

// Write the form data in the correct order
req.write(formDataHeader);
req.write(fileBuffer);
req.write(formDataFooter);
req.end();

console.log('📤 Sending subtitle upload request...');