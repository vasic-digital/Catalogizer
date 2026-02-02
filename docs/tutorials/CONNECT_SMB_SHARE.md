# Connect an SMB Share

This tutorial walks through configuring an SMB/CIFS network share as a storage source in Catalogizer, testing the connection, scanning media, and verifying results in the catalog.

## Prerequisites

- Catalogizer API running (see [Quick Start](QUICK_START.md))
- An SMB/CIFS share accessible from the server
- SMB credentials (username, password, optional domain)
- Network access to the SMB server on ports 139 and 445

## Step 1: Verify Network Connectivity

Before configuring Catalogizer, confirm the SMB share is reachable.

```bash
# Test port connectivity
nc -zv <smb-server-ip> 445

# List available shares (from a Linux host)
smbclient -L //<smb-server-ip> -U <username>
```

**Expected result:** Port 445 is open. `smbclient` lists available shares on the server.

If running inside Docker, test from the API container:

```bash
docker compose exec api sh -c "nc -zv <smb-server-ip> 445"
```

## Step 2: Configure the SMB Source via API

Send a POST request to add the SMB source:

```bash
curl -X POST http://localhost:8080/api/v1/sources \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "name": "NAS Movies",
    "protocol": "smb",
    "settings": {
      "host": "192.168.1.100",
      "port": 445,
      "share": "Movies",
      "username": "media_user",
      "password": "your_password",
      "domain": "WORKGROUP"
    },
    "max_depth": 5,
    "enabled": true
  }'
```

Configuration fields:
- **host**: IP address or hostname of the SMB server
- **port**: SMB port (default 445)
- **share**: Name of the shared folder
- **username** / **password**: Credentials with at least read access
- **domain**: Windows domain or workgroup name (default: WORKGROUP)
- **max_depth**: How many directory levels to scan (default: 5)

**Expected result:** JSON response with the created source, including a unique `id` field.

## Step 3: Test the Connection

Verify the connection before scanning:

```bash
curl -X POST http://localhost:8080/api/v1/sources/<source-id>/test \
  -H "Authorization: Bearer <your-token>"
```

**Expected result:** A success response confirming the connection is valid:

```json
{
  "success": true,
  "message": "Connection successful",
  "latency_ms": 12
}
```

If the connection fails, the response includes error details to help diagnose the issue.

## Step 4: Trigger a Media Scan

Once the connection is verified, initiate a scan of the share:

```bash
curl -X POST http://localhost:8080/api/v1/sources/<source-id>/scan \
  -H "Authorization: Bearer <your-token>"
```

**Expected result:** The API returns a confirmation that the scan has started. Depending on the size of the share, scanning may take from seconds to several minutes.

Monitor scan progress:

```bash
curl -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/v1/sources/<source-id>/status
```

## Step 5: Verify Media in the Catalog

After the scan completes, query the catalog to see detected media:

```bash
# List all media from this source
curl -H "Authorization: Bearer <your-token>" \
  "http://localhost:8080/api/v1/media?source_id=<source-id>"

# Search for a specific title
curl -H "Authorization: Bearer <your-token>" \
  "http://localhost:8080/api/v1/media/search?q=inception"
```

**Expected result:** A JSON array of media items detected on the SMB share. Each item includes:
- Title and detected media type (movie, TV show, music, etc.)
- File path on the share
- Quality information (resolution, codec)
- External metadata from providers like TMDB, IMDB (if configured)

## Step 6: Enable Real-Time Monitoring

Catalogizer can continuously monitor the SMB share for changes. The watcher interval is controlled by the `WATCH_INTERVAL_SECONDS` environment variable (default: 30 seconds).

Verify the watcher is running:

```bash
curl -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/v1/sources/<source-id>/status
```

When new files are added to the share, they are automatically detected, analyzed, and added to the catalog. If you have WebSocket connections open (via the web UI or API client), you receive real-time notifications.

## SMB Resilience Features

Catalogizer includes built-in resilience for SMB connections:

- **Circuit breaker**: Prevents repeated connection attempts to a failing server
- **Exponential backoff retry**: Automatically retries failed connections with increasing delays
- **Offline caching**: Cached metadata remains available when the SMB source is temporarily unreachable
- **Health checks**: Periodic connectivity checks with configurable intervals

These features are configured via environment variables:

```env
SMB_RETRY_ATTEMPTS=5
SMB_RETRY_DELAY_SECONDS=30
SMB_HEALTH_CHECK_INTERVAL=60
SMB_CONNECTION_TIMEOUT=30
SMB_OFFLINE_CACHE_SIZE=10000
```

## Troubleshooting

### "Connection refused" or timeout errors

- Verify the SMB server is running and accepting connections on port 445
- Check firewall rules between the Catalogizer host/container and the SMB server
- If using Docker, ensure the container network can reach the SMB server (host network or proper routing)

### "Access denied" errors

- Verify the username and password are correct
- Confirm the user has at least read permission on the share
- Check if the domain/workgroup is specified correctly
- Some servers require `domain\username` format

### Scan finds no media files

- Verify the share path points to a directory containing media files
- Increase `max_depth` if media is nested deeply in subdirectories
- Check that Catalogizer's media detection engine recognizes the file types present (50+ media types are supported)

### Intermittent disconnections

- This is normal for network shares. Catalogizer's resilience layer handles this automatically
- Check `SMB_RETRY_ATTEMPTS` and `SMB_HEALTH_CHECK_INTERVAL` settings
- Review API logs for detailed connection status: `docker compose logs -f api`

### Slow scanning performance

- Reduce `MAX_CONCURRENT_ANALYSIS` if the SMB server is under heavy load
- Increase `SMB_CONNECTION_TIMEOUT` for high-latency connections
- Consider limiting `max_depth` to avoid scanning unnecessary subdirectories
