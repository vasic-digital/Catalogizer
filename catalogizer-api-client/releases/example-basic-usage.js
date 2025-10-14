// Basic Catalogizer API Client Usage Example

const CatalogizerClient = require('@catalogizer/api-client').default;

async function main() {
  // Create client instance
  const client = new CatalogizerClient({
    baseURL: 'http://localhost:8080',
    enableWebSocket: true,
    webSocketURL: 'ws://localhost:8080/ws'
  });

  try {
    // Connect and authenticate
    console.log('Connecting to Catalogizer...');
    const loginResponse = await client.connect({
      username: 'demo',
      password: 'demo'
    });

    console.log('Logged in as:', loginResponse.user.username);

    // Search for media
    console.log('Searching for media...');
    const searchResults = await client.media.search({
      query: 'action',
      limit: 10
    });

    console.log(`Found ${searchResults.total} items:`);
    searchResults.items.forEach(item => {
      console.log(`- ${item.title} (${item.year || 'Unknown year'})`);
    });

    // Get media statistics
    const stats = await client.media.getStats();
    console.log('Library stats:', stats);

    // Listen for real-time updates
    client.on('download:progress', (progress) => {
      console.log(`Download ${progress.job_id}: ${progress.progress}%`);
    });

    // Disconnect
    await client.disconnect();
    console.log('Disconnected');

  } catch (error) {
    console.error('Error:', error.message);
  }
}

main();
