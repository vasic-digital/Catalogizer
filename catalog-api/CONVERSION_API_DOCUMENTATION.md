# Conversion API Documentation

## Overview
The Conversion API provides endpoints for media file conversion including video, audio, image, and document formats. The API supports asynchronous job processing with status tracking and error handling.

## Base URL
```
https://your-domain.com/api/v1/conversion
```

## Authentication
All endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <JWT_TOKEN>
```

## Required Permissions
- `conversion:create` - Create new conversion jobs
- `conversion:view` - View conversion job details and lists  
- `conversion:manage` - Cancel and manage conversion jobs

## Endpoints

### 1. Create Conversion Job
Create a new conversion job for media file processing.

**Endpoint**: `POST /jobs`

**Headers**:
```
Content-Type: application/json
Authorization: Bearer <JWT_TOKEN>
```

**Request Body**:
```json
{
  "source_path": "/path/to/source/file.ext",
  "target_path": "/path/to/target/file.ext",
  "source_format": "avi",
  "target_format": "mp4", 
  "conversion_type": "video",
  "quality": "high",
  "priority": 1,
  "settings": "{\"crf\": 23, \"preset\": \"medium\"}"
}
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|-------|----------|-------------|
| source_path | string | Yes | Path to source file |
| target_path | string | Yes | Path for converted output file |
| source_format | string | Yes | Source file format (avi, mp4, mp3, jpg, etc.) |
| target_format | string | Yes | Target file format (mp4, mp3, jpg, etc.) |
| conversion_type | string | Yes | Type of conversion: `video`, `audio`, `image`, `document` |
| quality | string | No | Quality level: `low`, `medium`, `high` (default: `medium`) |
| priority | integer | No | Job priority 1-10, lower numbers = higher priority (default: 5) |
| settings | string | No | JSON string with conversion-specific settings |

**Response**:
```json
{
  "id": 123,
  "user_id": 1,
  "source_path": "/path/to/source/file.ext",
  "target_path": "/path/to/target/file.ext",
  "source_format": "avi",
  "target_format": "mp4",
  "conversion_type": "video",
  "quality": "high",
  "settings": "{\"crf\": 23, \"preset\": \"medium\"}",
  "priority": 1,
  "status": "pending",
  "created_at": "2025-01-01T10:00:00Z",
  "started_at": null,
  "completed_at": null,
  "scheduled_for": null,
  "duration": null,
  "error_message": null
}
```

**Status Codes**:
- `201 Created` - Job created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Insufficient permissions
- `422 Unprocessable Entity` - Validation errors
- `500 Internal Server Error` - Server error

### 2. List Conversion Jobs
Retrieve a paginated list of user's conversion jobs.

**Endpoint**: `GET /jobs`

**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|-------|----------|-------------|
| status | string | No | Filter by status: `pending`, `running`, `completed`, `failed`, `cancelled` |
| limit | integer | No | Number of results per page (default: 20, max: 100) |
| offset | integer | No | Number of results to skip (default: 0) |

**Response**:
```json
{
  "jobs": [
    {
      "id": 123,
      "user_id": 1,
      "source_path": "/path/to/source/video.avi",
      "target_path": "/path/to/target/video.mp4",
      "source_format": "avi",
      "target_format": "mp4",
      "conversion_type": "video",
      "quality": "high",
      "status": "completed",
      "created_at": "2025-01-01T10:00:00Z",
      "started_at": "2025-01-01T10:00:05Z",
      "completed_at": "2025-01-01T10:05:30Z",
      "duration": 325000000000,
      "error_message": null
    },
    {
      "id": 124,
      "user_id": 1,
      "source_path": "/path/to/source/image.jpg",
      "target_path": "/path/to/target/image.png",
      "source_format": "jpg",
      "target_format": "png",
      "conversion_type": "image",
      "quality": "medium",
      "status": "running",
      "created_at": "2025-01-01T10:10:00Z",
      "started_at": "2025-01-01T10:10:02Z",
      "completed_at": null,
      "duration": null,
      "error_message": null
    }
  ],
  "pagination": {
    "total": 2,
    "limit": 20,
    "offset": 0,
    "has_more": false
  }
}
```

**Status Codes**:
- `200 OK` - Jobs retrieved successfully
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Insufficient permissions
- `500 Internal Server Error` - Server error

### 3. Get Conversion Job
Retrieve details of a specific conversion job.

**Endpoint**: `GET /jobs/{id}`

**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Path Parameters**:
| Parameter | Type | Description |
|-----------|-------|-------------|
| id | integer | ID of the conversion job |

**Response**:
```json
{
  "id": 123,
  "user_id": 1,
  "source_path": "/path/to/source/video.avi",
  "target_path": "/path/to/target/video.mp4",
  "source_format": "avi",
  "target_format": "mp4",
  "conversion_type": "video",
  "quality": "high",
  "settings": "{\"crf\": 23, \"preset\": \"medium\"}",
  "priority": 1,
  "status": "completed",
  "created_at": "2025-01-01T10:00:00Z",
  "started_at": "2025-01-01T10:00:05Z",
  "completed_at": "2025-01-01T10:05:30Z",
  "scheduled_for": null,
  "duration": 325000000000,
  "error_message": null
}
```

**Status Codes**:
- `200 OK` - Job retrieved successfully
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Insufficient permissions or not job owner
- `404 Not Found` - Job not found
- `500 Internal Server Error` - Server error

### 4. Cancel Conversion Job
Cancel a running or pending conversion job.

**Endpoint**: `POST /jobs/{id}/cancel`

**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Path Parameters**:
| Parameter | Type | Description |
|-----------|-------|-------------|
| id | integer | ID of the conversion job to cancel |

**Response**:
```json
{
  "success": true,
  "message": "Job cancelled successfully"
}
```

**Status Codes**:
- `200 OK` - Job cancelled successfully
- `400 Bad Request` - Job cannot be cancelled (already completed/failed)
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Insufficient permissions or not job owner
- `404 Not Found` - Job not found
- `500 Internal Server Error` - Server error

### 5. Get Supported Formats
Retrieve list of supported input and output formats for each media type.

**Endpoint**: `GET /formats`

**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response**:
```json
{
  "video": {
    "input": ["avi", "mp4", "mov", "mkv", "wmv", "flv", "webm"],
    "output": ["mp4", "webm", "avi", "mov"]
  },
  "audio": {
    "input": ["mp3", "wav", "flac", "aac", "ogg", "m4a"],
    "output": ["mp3", "wav", "flac", "aac", "ogg"]
  },
  "image": {
    "input": ["jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp"],
    "output": ["jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp"]
  },
  "document": {
    "input": ["pdf", "doc", "docx", "txt", "rtf"],
    "output": ["pdf", "docx", "txt", "html"]
  }
}
```

**Status Codes**:
- `200 OK` - Formats retrieved successfully
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Insufficient permissions
- `500 Internal Server Error` - Server error

## Job Status Lifecycle

### Status Values
| Status | Description | Can Transition To |
|---------|-------------|-------------------|
| `pending` | Job queued but not started | `running`, `cancelled` |
| `running` | Job currently processing | `completed`, `failed`, `cancelled` |
| `completed` | Job finished successfully | - |
| `failed` | Job failed with errors | - |
| `cancelled` | Job was cancelled by user | - |

### Status Flow
```
pending → running → completed
          ↘ failed
          ↘ cancelled
```

## Quality Levels

### Video Quality
| Quality | Resolution | Bitrate | Use Case |
|---------|------------|----------|----------|
| `low` | 480p | ~1 Mbps | Fast processing, small file size |
| `medium` | 720p | ~2.5 Mbps | Balanced quality/size |
| `high` | 1080p | ~5 Mbps | Best quality for most uses |

### Audio Quality  
| Quality | Bitrate | Codec | Use Case |
|---------|----------|-------|----------|
| `low` | 128 kbps | AAC | Small file size |
| `medium` | 192 kbps | AAC | Good balance |
| `high` | 320 kbps | AAC | Best quality |

### Image Quality
| Quality | Compression | Use Case |
|---------|-------------|----------|
| `low` | High compression | Web/ thumbnails |
| `medium` | Balanced | General use |
| `high` | Low compression | Print/ archival |

## Conversion Settings

### Video Settings (JSON string)
```json
{
  "crf": 23,           // Constant rate factor (0-51, lower = better quality)
  "preset": "medium",   // Encoding speed: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
  "tune": "film",      // Optimization: film, animation, stillimage
  "profile": "high",     // H.264 profile: baseline, main, high
  "level": "4.0",        // H.264 level
  "pixel_format": "yuv420p", // Pixel format
  "threads": 4,         // Number of encoding threads
  "gpu_acceleration": false  // Enable GPU acceleration
}
```

### Audio Settings
```json
{
  "bitrate": "192k",    // Audio bitrate
  "sample_rate": 44100,  // Sample rate in Hz
  "channels": 2,         // Number of audio channels
  "codec": "aac"         // Audio codec
}
```

### Image Settings
```json
{
  "quality": 85,         // Image quality (0-100)
  "compression": "jpeg",   // Compression type
  "progressive": true,    // Progressive loading
  "optimize": true,       // Optimize for web
  "resize": {            // Optional resize
    "width": 1920,
    "height": 1080,
    "maintain_aspect": true
  }
}
```

## Error Responses

### Standard Error Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid source format",
    "details": {
      "field": "source_format",
      "value": "xyz",
      "valid_options": ["avi", "mp4", "mov", "mkv"]
    }
  }
}
```

### Common Error Codes
| Code | Description | HTTP Status |
|-------|-------------|-------------|
| `VALIDATION_ERROR` | Request validation failed | 400 |
| `AUTHENTICATION_REQUIRED` | JWT token missing or invalid | 401 |
| `PERMISSION_DENIED` | Insufficient permissions | 403 |
| `RESOURCE_NOT_FOUND` | Job not found | 404 |
| `CONVERSION_LIMIT_EXCEEDED` | Too many concurrent jobs | 429 |
| `FILE_NOT_FOUND` | Source file not found | 400 |
| `UNSUPPORTED_FORMAT` | Format not supported | 400 |
| `INTERNAL_ERROR` | Server internal error | 500 |

## Rate Limiting

### Limits
- **Job Creation**: 10 jobs per minute per user
- **Job Queries**: 100 requests per minute per user  
- **Format Queries**: 60 requests per minute per user

### Rate Limit Headers
```http
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1640995200
```

## Examples

### Create Video Conversion
```bash
curl -X POST https://your-domain.com/api/v1/conversion/jobs \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "source_path": "/videos/vacation.avi",
    "target_path": "/videos/vacation.mp4",
    "source_format": "avi", 
    "target_format": "mp4",
    "conversion_type": "video",
    "quality": "high",
    "settings": "{\"crf\": 20, \"preset\": \"slow\"}"
  }'
```

### List Jobs with Status Filter
```bash
curl -X GET "https://your-domain.com/api/v1/conversion/jobs?status=running&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Get Job Details
```bash
curl -X GET https://your-domain.com/api/v1/conversion/jobs/123 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Cancel Job
```bash
curl -X POST https://your-domain.com/api/v1/conversion/jobs/123/cancel \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Get Supported Formats
```bash
curl -X GET https://your-domain.com/api/v1/conversion/formats \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## SDK Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'https://your-domain.com/api/v1/conversion',
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
    'Content-Type': 'application/json'
  }
});

// Create job
const job = await api.post('/jobs', {
  source_path: '/input/video.avi',
  target_path: '/output/video.mp4',
  source_format: 'avi',
  target_format: 'mp4',
  conversion_type: 'video',
  quality: 'high'
});

// List jobs
const jobs = await api.get('/jobs', {
  params: { status: 'completed', limit: 20 }
});

// Get job status
const status = await api.get(`/jobs/${job.data.id}`);

// Cancel job
await api.post(`/jobs/${job.data.id}/cancel`);
```

### Python
```python
import requests

headers = {
    'Authorization': f'Bearer {jwt_token}',
    'Content-Type': 'application/json'
}

# Create job
response = requests.post(
    'https://your-domain.com/api/v1/conversion/jobs',
    json={
        'source_path': '/input/video.avi',
        'target_path': '/output/video.mp4',
        'source_format': 'avi',
        'target_format': 'mp4',
        'conversion_type': 'video',
        'quality': 'high'
    },
    headers=headers
)
job = response.json()

# List jobs
response = requests.get(
    'https://your-domain.com/api/v1/conversion/jobs',
    params={'status': 'completed', 'limit': 20},
    headers=headers
)
jobs = response.json()

# Get job status
response = requests.get(
    f'https://your-domain.com/api/v1/conversion/jobs/{job["id"]}',
    headers=headers
)
status = response.json()
```

---

**API Version**: v1  
**Last Updated**: November 27, 2025  
**Documentation Version**: 1.0