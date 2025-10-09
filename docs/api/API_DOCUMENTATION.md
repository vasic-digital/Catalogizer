# Catalogizer v3.0 - Complete API Documentation

![API Overview](../screenshots/api/api-overview.png)
*Complete API endpoint overview in the admin dashboard*

## 📚 Table of Contents

1. [Authentication APIs](#authentication-apis)
2. [Media Management APIs](#media-management-apis)
3. [Analytics APIs](#analytics-apis)
4. [Collections & Favorites APIs](#collections--favorites-apis)
5. [Advanced Features APIs](#advanced-features-apis)
6. [Administration APIs](#administration-apis)
7. [WebSocket APIs](#websocket-apis)
8. [Error Handling](#error-handling)
9. [Rate Limiting](#rate-limiting)
10. [SDK Examples](#sdk-examples)

---

## 🔐 Authentication APIs

### Login
**POST** `/api/auth/login`

![Login API](../screenshots/api/auth-login.png)
*Login API response in developer tools*

**Request Body:**
```json
{
  "username": "user@example.com",
  "password": "securepassword",
  "remember_me": true
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "username": "user@example.com",
    "role": "user",
    "permissions": ["media:read", "media:write"]
  }
}
```

### Register
**POST** `/api/auth/register`

![Registration API](../screenshots/api/auth-register.png)
*User registration form with API integration*

**Request Body:**
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "securepassword",
  "full_name": "New User",
  "terms_accepted": true
}
```

### Logout
**POST** `/api/auth/logout`

**Headers:**
```
Authorization: Bearer <token>
```

### Refresh Token
**POST** `/api/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

---

## 📁 Media Management APIs

### List Media Items
**GET** `/api/media`

![Media List API](../screenshots/api/media-list.png)
*Media items displayed from API response*

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `type` (string): Filter by media type (video, audio, image, document)
- `search` (string): Search in title, description, tags
- `sort` (string): Sort by (created_at, updated_at, size, title)
- `order` (string): asc or desc

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Sample Video",
      "type": "video",
      "format": "mp4",
      "size": 1048576,
      "duration": 120.5,
      "thumbnail_url": "/api/media/1/thumbnail",
      "download_url": "/api/media/1/download",
      "metadata": {
        "width": 1920,
        "height": 1080,
        "codec": "h264"
      },
      "created_at": "2023-10-01T12:00:00Z",
      "updated_at": "2023-10-01T12:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 95,
    "per_page": 20
  }
}
```

### Upload Media
**POST** `/api/media/upload`

![Media Upload API](../screenshots/api/media-upload.png)
*File upload interface with progress tracking*

**Request:** Multipart form data
- `file`: The media file
- `title` (optional): Custom title
- `description` (optional): Description
- `tags` (optional): Comma-separated tags
- `collection_id` (optional): Add to collection

**Response:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "id": 123,
    "title": "Uploaded Video",
    "processing_status": "queued",
    "upload_progress": 100
  }
}
```

### Get Media Details
**GET** `/api/media/{id}`

![Media Details API](../screenshots/api/media-details.png)
*Detailed media information from API*

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Sample Video",
    "description": "A sample video file",
    "type": "video",
    "format": "mp4",
    "size": 1048576,
    "duration": 120.5,
    "metadata": {
      "width": 1920,
      "height": 1080,
      "fps": 30,
      "codec": "h264",
      "bitrate": 2500
    },
    "thumbnails": [
      "/api/media/1/thumbnail?size=small",
      "/api/media/1/thumbnail?size=medium",
      "/api/media/1/thumbnail?size=large"
    ],
    "tags": ["sample", "video", "demo"],
    "collections": [1, 3],
    "created_at": "2023-10-01T12:00:00Z",
    "updated_at": "2023-10-01T12:00:00Z"
  }
}
```

### Update Media
**PUT** `/api/media/{id}`

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "tags": ["updated", "tag1", "tag2"]
}
```

### Delete Media
**DELETE** `/api/media/{id}`

**Response:**
```json
{
  "success": true,
  "message": "Media item deleted successfully"
}
```

### Download Media
**GET** `/api/media/{id}/download`

Returns the actual media file with appropriate headers.

### Get Thumbnail
**GET** `/api/media/{id}/thumbnail`

**Query Parameters:**
- `size` (string): thumbnail size (small, medium, large)
- `format` (string): image format (jpg, png, webp)

---

## 📊 Analytics APIs

### Track Event
**POST** `/api/analytics/track`

![Analytics Tracking](../screenshots/api/analytics-track.png)
*Event tracking in real-time dashboard*

**Request Body:**
```json
{
  "event_type": "media_view",
  "entity_type": "media_item",
  "entity_id": 123,
  "metadata": {
    "duration_watched": 45.2,
    "quality": "1080p",
    "device_type": "desktop"
  },
  "session_id": "session_123",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.1"
}
```

### Get User Events
**GET** `/api/analytics/events`

![User Events](../screenshots/api/analytics-events.png)
*User activity timeline from analytics*

**Query Parameters:**
- `event_type` (string): Filter by event type
- `start_date` (string): ISO 8601 date
- `end_date` (string): ISO 8601 date
- `limit` (int): Number of events to return

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "event_type": "media_view",
      "entity_type": "media_item",
      "entity_id": 123,
      "metadata": {
        "duration_watched": 45.2
      },
      "timestamp": "2023-10-01T12:00:00Z"
    }
  ]
}
```

### Dashboard Metrics
**GET** `/api/analytics/dashboard`

![Dashboard Metrics](../screenshots/api/analytics-dashboard.png)
*Analytics dashboard powered by API data*

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "total_events": 1250,
    "events_today": 45,
    "events_this_week": 320,
    "events_this_month": 890,
    "top_event_types": {
      "media_view": 650,
      "media_download": 300,
      "collection_create": 25
    },
    "activity_timeline": [
      {
        "date": "2023-10-01",
        "events": 45
      }
    ],
    "device_breakdown": {
      "desktop": 70,
      "mobile": 25,
      "tablet": 5
    }
  }
}
```

### Generate Report
**POST** `/api/analytics/reports`

![Analytics Report](../screenshots/api/analytics-report.png)
*Generated analytics report interface*

**Request Body:**
```json
{
  "report_type": "user_activity",
  "start_date": "2023-09-01T00:00:00Z",
  "end_date": "2023-09-30T23:59:59Z",
  "format": "json",
  "filters": {
    "event_types": ["media_view", "media_download"],
    "include_metadata": true
  }
}
```

---

## ⭐ Collections & Favorites APIs

### List Collections
**GET** `/api/collections`

![Collections List](../screenshots/api/collections-list.png)
*Collections overview from API*

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "My Vacation Videos",
      "description": "Family vacation memories",
      "type": "custom",
      "item_count": 25,
      "created_at": "2023-10-01T12:00:00Z",
      "thumbnail": "/api/collections/1/thumbnail"
    }
  ]
}
```

### Create Collection
**POST** `/api/collections`

![Create Collection](../screenshots/api/collections-create.png)
*Collection creation form*

**Request Body:**
```json
{
  "name": "New Collection",
  "description": "Collection description",
  "type": "custom",
  "tags": ["tag1", "tag2"],
  "privacy": "private"
}
```

### Add Item to Collection
**POST** `/api/collections/{id}/items`

**Request Body:**
```json
{
  "media_id": 123
}
```

### List Favorites
**GET** `/api/favorites`

![Favorites List](../screenshots/api/favorites-list.png)
*User favorites displayed from API*

**Query Parameters:**
- `entity_type` (string): Filter by entity type
- `limit` (int): Number of items to return

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "entity_type": "media_item",
      "entity_id": 123,
      "entity": {
        "id": 123,
        "title": "Favorite Video",
        "thumbnail_url": "/api/media/123/thumbnail"
      },
      "created_at": "2023-10-01T12:00:00Z"
    }
  ]
}
```

### Add Favorite
**POST** `/api/favorites`

**Request Body:**
```json
{
  "entity_type": "media_item",
  "entity_id": 123
}
```

### Remove Favorite
**DELETE** `/api/favorites/{id}`

---

## 🛠️ Advanced Features APIs

### Format Conversion

#### Create Conversion Job
**POST** `/api/conversion/jobs`

![Conversion Job](../screenshots/api/conversion-create.png)
*Format conversion interface*

**Request Body:**
```json
{
  "source_path": "/media/video.mp4",
  "target_path": "/media/video.mp3",
  "source_format": "mp4",
  "target_format": "mp3",
  "quality_settings": {
    "bitrate": "320k",
    "sample_rate": "44100"
  }
}
```

#### Get Conversion Status
**GET** `/api/conversion/jobs/{id}`

![Conversion Status](../screenshots/api/conversion-status.png)
*Conversion progress tracking*

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "processing",
    "progress": 45,
    "estimated_completion": "2023-10-01T12:15:00Z",
    "error_message": null
  }
}
```

### Sync & Backup

#### List Sync Endpoints
**GET** `/api/sync/endpoints`

![Sync Endpoints](../screenshots/api/sync-endpoints.png)
*WebDAV sync configuration*

#### Create Sync Session
**POST** `/api/sync/sessions`

**Request Body:**
```json
{
  "endpoint_id": 1,
  "direction": "bidirectional",
  "local_path": "/media",
  "remote_path": "/backup"
}
```

### Error Reporting

#### Report Error
**POST** `/api/errors/report`

![Error Reporting](../screenshots/api/error-report.png)
*Error reporting interface*

**Request Body:**
```json
{
  "level": "error",
  "message": "Failed to process media file",
  "error_code": "MEDIA_PROCESS_ERROR",
  "component": "media_processor",
  "stack_trace": "Stack trace here...",
  "context": {
    "file_id": 123,
    "operation": "thumbnail_generation"
  }
}
```

#### Get Error Statistics
**GET** `/api/errors/statistics`

![Error Statistics](../screenshots/api/error-statistics.png)
*Error statistics dashboard*

### Log Management

#### Create Log Collection
**POST** `/api/logs/collections`

**Request Body:**
```json
{
  "name": "Debug Session",
  "components": ["api", "media_processor"],
  "log_level": "debug",
  "start_time": "2023-10-01T00:00:00Z",
  "end_time": "2023-10-01T23:59:59Z"
}
```

#### Stream Logs
**GET** `/api/logs/stream` (WebSocket)

![Log Streaming](../screenshots/api/log-stream.png)
*Real-time log streaming interface*

---

## 👨‍💼 Administration APIs

### User Management

#### List Users
**GET** `/api/admin/users`

![User Management](../screenshots/api/admin-users.png)
*User management interface*

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "status": "active",
      "last_login": "2023-10-01T12:00:00Z",
      "created_at": "2023-09-01T12:00:00Z"
    }
  ]
}
```

#### Create User
**POST** `/api/admin/users`

**Request Body:**
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "securepassword",
  "role": "user",
  "permissions": ["media:read", "media:write"]
}
```

### System Configuration

#### Get Configuration
**GET** `/api/admin/config`

![System Config](../screenshots/api/admin-config.png)
*System configuration interface*

#### Update Configuration
**PUT** `/api/admin/config`

**Request Body:**
```json
{
  "database": {
    "type": "sqlite",
    "path": "/data/catalogizer.db"
  },
  "storage": {
    "media_directory": "/media",
    "max_file_size": 1073741824
  },
  "features": {
    "media_conversion": true,
    "webdav_sync": false
  }
}
```

### System Health

#### Health Check
**GET** `/api/health`

![Health Check](../screenshots/api/health-check.png)
*System health monitoring*

**Response:**
```json
{
  "success": true,
  "status": "healthy",
  "timestamp": "2023-10-01T12:00:00Z",
  "version": "3.0.0",
  "components": {
    "database": "healthy",
    "storage": "healthy",
    "external_services": "degraded"
  },
  "metrics": {
    "uptime": "72h 30m",
    "memory_usage": "45%",
    "disk_usage": "30%",
    "active_users": 25
  }
}
```

---

## 🔌 WebSocket APIs

### Real-time Notifications
**WebSocket** `/ws/notifications`

![WebSocket Notifications](../screenshots/api/websocket-notifications.png)
*Real-time notification system*

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/notifications?token=jwt_token');

ws.onmessage = function(event) {
  const notification = JSON.parse(event.data);
  console.log('Notification:', notification);
};
```

**Message Format:**
```json
{
  "type": "conversion_complete",
  "data": {
    "job_id": 123,
    "status": "completed",
    "output_path": "/media/converted/video.mp3"
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

### Live Log Streaming
**WebSocket** `/ws/logs`

### Progress Updates
**WebSocket** `/ws/progress/{job_id}`

---

## ⚠️ Error Handling

### Standard Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  },
  "request_id": "req_123456789"
}
```

### HTTP Status Codes

![Error Codes](../screenshots/api/error-codes.png)
*Error handling in the API*

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Validation Error |
| 429 | Rate Limited |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

### Error Types

- `AUTHENTICATION_ERROR`: Authentication failed
- `AUTHORIZATION_ERROR`: Insufficient permissions
- `VALIDATION_ERROR`: Input validation failed
- `NOT_FOUND_ERROR`: Resource not found
- `RATE_LIMIT_ERROR`: Rate limit exceeded
- `INTERNAL_ERROR`: Server internal error
- `SERVICE_UNAVAILABLE`: Service temporarily unavailable

---

## 🚦 Rate Limiting

![Rate Limiting](../screenshots/api/rate-limiting.png)
*Rate limiting configuration interface*

### Rate Limit Headers
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1609459200
X-RateLimit-Window: 3600
```

### Default Limits
- **Authentication**: 10 requests/minute
- **Media Upload**: 50 requests/hour
- **API Calls**: 1000 requests/hour
- **WebSocket**: 100 connections/user

---

## 📚 SDK Examples

### JavaScript/Node.js
```javascript
// Install: npm install catalogizer-sdk

import { CatalogizerClient } from 'catalogizer-sdk';

const client = new CatalogizerClient({
  baseURL: 'https://api.catalogizer.com',
  apiKey: 'your-api-key'
});

// Upload media
const uploadResult = await client.media.upload({
  file: fileBuffer,
  title: 'My Video',
  tags: ['vacation', 'family']
});

// Track analytics
await client.analytics.track({
  event_type: 'media_view',
  entity_id: uploadResult.id
});

// Create collection
const collection = await client.collections.create({
  name: 'Vacation Videos',
  description: 'Family vacation memories'
});
```

### Python
```python
# Install: pip install catalogizer-sdk

from catalogizer import CatalogizerClient

client = CatalogizerClient(
    base_url="https://api.catalogizer.com",
    api_key="your-api-key"
)

# Upload media
with open("video.mp4", "rb") as f:
    result = client.media.upload(
        file=f,
        title="My Video",
        tags=["vacation", "family"]
    )

# Get analytics
analytics = client.analytics.get_dashboard_metrics()
print(f"Total events: {analytics['total_events']}")
```

### cURL Examples

#### Upload File
```bash
curl -X POST "https://api.catalogizer.com/api/media/upload" \
  -H "Authorization: Bearer your-jwt-token" \
  -F "file=@video.mp4" \
  -F "title=My Video" \
  -F "tags=vacation,family"
```

#### Get Media List
```bash
curl -X GET "https://api.catalogizer.com/api/media?limit=10&type=video" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Accept: application/json"
```

#### Track Event
```bash
curl -X POST "https://api.catalogizer.com/api/analytics/track" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "media_view",
    "entity_type": "media_item",
    "entity_id": 123
  }'
```

---

## 🔧 Development Tools

### API Testing
![API Testing](../screenshots/api/api-testing.png)
*Built-in API testing interface*

- **Postman Collection**: Available for download
- **OpenAPI/Swagger**: Interactive documentation
- **Test Environment**: Sandbox API for testing

### Webhooks
Configure webhooks to receive real-time notifications:

```json
{
  "url": "https://your-app.com/webhooks/catalogizer",
  "events": [
    "media.uploaded",
    "conversion.completed",
    "user.registered"
  ],
  "secret": "webhook-secret-key"
}
```

### API Versioning
- Current Version: `v3.0`
- Version Header: `API-Version: 3.0`
- Backward Compatibility: 2 major versions

---

This comprehensive API documentation provides complete coverage of all Catalogizer v3.0 endpoints with visual examples and practical usage scenarios. All screenshots are automatically captured through our QA automation system to ensure accuracy and up-to-date documentation.