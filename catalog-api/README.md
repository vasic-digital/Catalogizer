# Catalog API

A REST API for browsing and searching multi-protocol file catalogs, built with Go and Gin framework.

## Features

- **Catalog Browsing**: Browse files and directories from cataloged storage sources (SMB, FTP, NFS, WebDAV, Local)
- **Advanced Search**: Search files by name, size, type, modification date, and more
- **Duplicate Detection**: Find and analyze duplicate files across all storage sources
- **File Downloads**: Download individual files or create archives (ZIP, TAR, TAR.GZ)
- **Multi-Protocol Operations**: Copy files between different storage protocols
- **Statistics**: Directory size analysis and duplicate file statistics
- **RESTful API**: Clean REST API with JSON responses
- **CORS Support**: Enable cross-origin requests for web frontends

## API Endpoints

### Catalog Browsing
- `GET /api/v1/catalog` - List available storage root directories
- `GET /api/v1/catalog/{path}` - List files and directories in path
- `GET /api/v1/catalog-info/{path}?id={id}` - Get detailed file information

### Search
- `GET /api/v1/search` - Search files with various filters
- `GET /api/v1/search/duplicates` - Find duplicate file groups

### Downloads
- `GET /api/v1/download/file/{id}` - Download a single file
- `GET /api/v1/download/directory/{path}` - Download directory as archive
- `POST /api/v1/download/archive` - Create custom archive from file list

### Storage Operations
- `POST /api/v1/copy/storage` - Copy between storage sources
- `POST /api/v1/copy/local` - Copy from storage to local filesystem
- `POST /api/v1/copy/upload` - Upload file to storage source
- `GET /api/v1/storage/list/{path}?root={root_id}` - List storage directory contents
- `GET /api/v1/storage/roots` - Get available storage roots

### Statistics
- `GET /api/v1/stats/directories/by-size` - Get directories sorted by size
- `GET /api/v1/stats/duplicates/count` - Get duplicate file statistics

## Configuration

Create a `config.json` file in the project root:

```json
{
  "server": {
    "host": "localhost",
    "port": "8080",
    "enable_cors": true
  },
  "smb": {
    "hosts": [
      {
        "name": "nas1",
        "host": "192.168.1.100",
        "port": 445,
        "share": "shared",
        "username": "user",
        "password": "password",
        "domain": "WORKGROUP"
      }
    ]
  },
  "catalog": {
    "temp_dir": "/tmp",
    "max_archive_size": 1073741824,
    "download_chunk_size": 1048576
  }
}
```

## Environment Variables

- `CATALOG_CONFIG_PATH` - Path to configuration file (default: `config.json`)

## Running the API

1. Install dependencies:
```bash
go mod tidy
```

2. Run the server:
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## Health Check

Check if the API is running:
```bash
curl http://localhost:8080/health
```

## Example Requests

### Search for files
```bash
curl "http://localhost:8080/api/v1/search?query=document&extension=pdf&min_size=1000"
```

### Get directories by size
```bash
curl "http://localhost:8080/api/v1/stats/directories/by-size?smb_root=nas1&limit=10"
```

### Find duplicates
```bash
curl "http://localhost:8080/api/v1/search/duplicates?smb_root=nas1&min_count=2"
```

### Copy file between SMB shares
```bash
curl -X POST http://localhost:8080/api/v1/copy/smb \
  -H "Content-Type: application/json" \
  -d '{
    "source_path": "nas1:/documents/file.pdf",
    "destination_path": "nas2:/backup/file.pdf",
    "overwrite": false
  }'
```

## Architecture

The API is structured with the following components:

- **Models**: Data structures for files, search requests, etc.
- **Services**: Business logic for catalog operations and SMB connectivity
- **Handlers**: HTTP request handlers for each endpoint group
- **Middleware**: Cross-cutting concerns (CORS, logging, error handling)
- **Config**: Configuration management

## Dependencies

- **Gin**: HTTP web framework
- **go-smb2**: SMB/CIFS client library
- **Zap**: Structured logging
- **UUID**: Request ID generation
- **SQLite**: Database driver (for catalog storage)

## Development

The project follows Go best practices with:

- Clean architecture with separated concerns
- Dependency injection
- Structured logging
- Error handling middleware
- Configuration management
- Graceful shutdown

## Integration with Catalogizer

This API is designed to work with the existing Catalogizer Kotlin application by:

1. Reading from the same file catalog database
2. Using compatible SMB connection configurations
3. Providing REST endpoints for catalog operations
4. Supporting the same file metadata structure

The API can be deployed separately and accessed by web frontends, mobile apps, or other services that need programmatic access to the file catalog.