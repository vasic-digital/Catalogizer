const http = require('http');

// Test direct comparison of search vs upload endpoints

function testEndpoint(path, method = 'GET', data = null) {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: `/api/v1/subtitles/${path}`,
      method: method,
      headers: {
        'Content-Type': 'application/json'
      }
    };

    if (data) {
      const body = JSON.stringify(data);
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
          headers: res.headers,
          body: responseData
        });
      });
    });

    req.on('error', (e) => {
      resolve({
        statusCode: 500,
        path: path,
        error: e.message,
        body: ''
      });
    });

    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

async function testRoutes() {
  console.log('Testing subtitle routes...\n');
  
  // Test search endpoint (should work)
  const searchResult = await testEndpoint('search?media_path=/test.mp4&title=test&year=2024&languages=en');
  console.log('SEARCH ENDPOINT:');
  console.log(`Status: ${searchResult.statusCode}`);
  console.log(`Body: ${searchResult.body}\n`);

  // Test upload endpoint (should return 404)
  const uploadResult = await testEndpoint('upload', 'POST');
  console.log('UPLOAD ENDPOINT:');
  console.log(`Status: ${uploadResult.statusCode}`);
  console.log(`Body: ${uploadResult.body}\n`);

  // Test languages endpoint (should work)
  const languagesResult = await testEndpoint('languages');
  console.log('LANGUAGES ENDPOINT:');
  console.log(`Status: ${languagesResult.statusCode}`);
  console.log(`Body: ${languagesResult.body}\n`);
}

testRoutes();