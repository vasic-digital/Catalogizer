const http = require('http');

const postData = JSON.stringify({
  username: 'testuser2',
  email: 'test2@example.com',
  password: 'password123',
  first_name: 'Test',
  last_name: 'User2',
  device_info: {
    platform: 'web',
    device_id: 'test-device-456',
    device_name: 'Test Browser',
    os_version: 'macOS'
  }
});

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/v1/auth/register',
  method: 'POST',
  headers: {
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
      console.log('Parsed response:', JSON.stringify(parsed, null, 2));
      if (parsed.token) {
        console.log('✅ Registration successful! Token:', parsed.token.substring(0, 50) + '...');
      }
    } catch (e) {
      console.log('Response is not valid JSON');
    }
  });
});

req.on('error', (e) => {
  console.error(`Problem with request: ${e.message}`);
});

req.write(postData);
req.end();