# Service Discovery Implementation

## Overview

Catalogizer now includes a service discovery mechanism that allows containers to bind to dynamic ports and clients to automatically discover service endpoints. This ensures seamless operation in environments where default ports may be occupied and enables both local network and cloud-hosted deployments.

## Key Features

1. **Dynamic Port Binding**: Services attempt to bind to their default port, and if unavailable, automatically try subsequent ports (up to 10 attempts).
2. **Port Registration**: Bound ports are written to a `.service-port` file for local discovery.
3. **API Discovery Endpoint**: Services expose a `/api/v1/discovery` endpoint that returns the bound host and port.
4. **Frontend Auto-Configuration**: The React frontend automatically reads the port file and configures its API proxy accordingly.
5. **Container Runtime Integration**: Uses the `digital.vasic.containers` module's TCP discovery for port availability checks.

## Implementation Details

### Backend (catalog-api)

**Dynamic Port Selection** (`main.go`):
- Added `findAvailablePort()` function that uses `digital.vasic.containers/pkg/discovery.TCPDiscoverer` to check port availability.
- The server tries ports starting from `cfg.Server.Port` (default 8080) and increments until finding an available port.
- Selected port is written to `catalog-api/.service-port` and logged.
- The `cfg.Server.Port` is updated to reflect the bound port, ensuring all internal references use the correct port.

**Discovery API Endpoint**:
- Added `GET /api/v1/discovery` endpoint that returns JSON: `{"host": "...", "port": ...}`.
- This endpoint is authenticated (requires JWT) and can be used by clients to discover the service location.

**Port File**:
- Created `.service-port` in the catalog-api directory with the numeric port.
- Used by local clients (including the frontend dev server) to determine the API location.

### Frontend (catalog-web)

**Vite Proxy Configuration** (`vite.config.ts`):
- Added `getApiPort()` function that reads `../catalog-api/.service-port`.
- Updated proxy targets to use the discovered port: `http://localhost:${getApiPort()}`.
- Removed hardcoded port 8083 from configuration.

**Environment Configuration**:
- Updated `.env.local` to set `VITE_API_BASE_URL=` (empty), causing the frontend to use relative URLs.
- API requests are now proxied through the dev server, eliminating hardcoded backend addresses.

### Container Integration

**Containers Submodule**:
- Added `digital.vasic.containers` as a submodule and dependency.
- Used the `discovery` package for TCP-based port availability checking.
- The module provides a foundation for future service discovery enhancements (DNS, mDNS, etc.).

## Usage

### Local Development

1. Start the catalog-api server:
   ```bash
   cd catalog-api
   go run main.go
   ```
   The server will log the selected port (e.g., "Selected HTTP port 8083").

2. Start the frontend:
   ```bash
   cd catalog-web
   npm run dev
   ```
   The dev server will read `.service-port` and proxy API requests to the correct port.

3. Verify discovery endpoint:
   ```bash
   curl -H "Authorization: Bearer <token>" http://localhost:<port>/api/v1/discovery
   ```

### Multi-Instance Deployment

When running multiple instances (e.g., for testing):
- Each instance will automatically find an available port.
- The `.service-port` file will be overwritten by the last started instance in the same directory.
- For isolated instances, set different working directories or use environment variable `PORT` to specify the starting port.

### Cloud Deployment

For cloud-hosted deployments where services have constant addresses:
- Set the `PORT` environment variable to the desired port.
- Clients should use the provided constant address (bypassing discovery).
- The discovery endpoint still provides the bound port for verification.

## Configuration

### Environment Variables

- `PORT`: Overrides the starting port for dynamic binding.
- `CATALOG_API_URL`: Used by challenges and tests to specify the API URL (defaults to `http://localhost:8080`).

### File Locations

- `catalog-api/.service-port`: Contains the bound port number (plain text).
- `catalog-api/config.json`: Server configuration (default port 8080).

## Future Enhancements

1. **Network Discovery**: Implement mDNS/DNS-SD for automatic service discovery across local networks.
2. **Service Registry**: Central registry for multiple services (using etcd, Consul, or the Containers module's registry capabilities).
3. **Health Checks**: Integrate with the Containers health checking system for automated service validation.
4. **Client Libraries**: Provide discovery-enabled clients for all platform components (Android, Tauri, etc.).

## Testing

The service discovery mechanism is validated through:
- Manual verification of dynamic port binding.
- Frontend proxy configuration reading the port file.
- Discovery API endpoint returning correct host/port.
- Integration with the Containers module's TCP discovery.

Run comprehensive tests with:
```bash
cd catalog-api
go test ./... -short
```

## Constraints

- **Host Resource Limits**: All tests and container workloads must stay within 30-40% of total host resources.
- **Containerized Builds**: All builds and services must use containers (Podman/Docker).
- **HTTP/3 Requirement**: All network communication must use HTTP/3 (QUIC) with Brotli compression where possible.

## References

- `digital.vasic.containers` module documentation
- `catalog-api/main.go` - dynamic port binding and discovery endpoint
- `catalog-web/vite.config.ts` - frontend proxy configuration
- `catalog-api/config/config.go` - server configuration structure