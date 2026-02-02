# Tauri IPC Commands Guide (catalogizer-desktop)

This guide documents all IPC commands and the state management architecture used in the `catalogizer-desktop` Tauri application.

## Overview

The Catalogizer desktop app uses Tauri to wrap a React frontend with a Rust backend. Communication between the two layers happens through Tauri's IPC (Inter-Process Communication) system, where the frontend invokes Rust commands and the backend can emit events.

## Architecture

```
React Frontend (TypeScript)
    │  invoke("command_name", { args })
    ▼
Tauri IPC Layer
    │  #[tauri::command]
    ▼
Rust Backend (src-tauri/src/main.rs)
    │  Managed State (ConfigState)
    ▼
System Resources (filesystem, HTTP, OS info)
```

### State Management

The Rust backend manages a shared `AppConfig` state through Tauri's state management:

```rust
#[derive(Debug, Serialize, Deserialize)]
struct AppConfig {
    server_url: Option<String>,   // Catalog API server URL
    auth_token: Option<String>,   // JWT authentication token
    theme: String,                // UI theme ("dark" / "light")
    auto_start: bool,             // Launch on system startup
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            server_url: None,
            auth_token: None,
            theme: "dark".to_string(),
            auto_start: false,
        }
    }
}

type ConfigState = Mutex<AppConfig>;
```

The state is registered in the Tauri builder:

```rust
fn main() {
    let config = ConfigState::default();

    tauri::Builder::default()
        .manage(config)
        .invoke_handler(tauri::generate_handler![
            get_config,
            update_config,
            set_server_url,
            set_auth_token,
            clear_auth_token,
            make_http_request,
            get_app_version,
            get_platform,
            get_arch
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

## IPC Commands Reference

### Configuration Commands

#### `get_config`

Retrieves the current application configuration.

- **Parameters**: None
- **Returns**: `AppConfig` object
- **Errors**: None (infallible in practice)

```rust
#[tauri::command]
async fn get_config(config: State<'_, ConfigState>) -> Result<AppConfig, String> {
    let config = config.lock().await;
    Ok(config.clone())
}
```

**Frontend usage:**
```typescript
import { invoke } from '@tauri-apps/api/tauri'

const config = await invoke<AppConfig>('get_config')
console.log(config.server_url, config.theme)
```

---

#### `update_config`

Replaces the entire application configuration.

- **Parameters**: `new_config: AppConfig` - the complete new configuration
- **Returns**: `()` on success
- **Errors**: String error message on failure

```rust
#[tauri::command]
async fn update_config(
    new_config: AppConfig,
    config: State<'_, ConfigState>,
) -> Result<(), String> {
    let mut config = config.lock().await;
    *config = new_config;
    Ok(())
}
```

**Frontend usage:**
```typescript
await invoke('update_config', {
  newConfig: {
    server_url: 'http://192.168.1.100:8080',
    auth_token: null,
    theme: 'dark',
    auto_start: false,
  }
})
```

---

#### `set_server_url`

Updates only the server URL in the configuration.

- **Parameters**: `url: String` - the catalog API server URL
- **Returns**: `()` on success

```rust
#[tauri::command]
async fn set_server_url(url: String, config: State<'_, ConfigState>) -> Result<(), String> {
    let mut config = config.lock().await;
    config.server_url = Some(url);
    Ok(())
}
```

**Frontend usage:**
```typescript
await invoke('set_server_url', { url: 'http://192.168.1.100:8080' })
```

---

### Authentication Commands

#### `set_auth_token`

Stores the JWT authentication token in the app state.

- **Parameters**: `token: String` - JWT token received after login
- **Returns**: `()` on success

```rust
#[tauri::command]
async fn set_auth_token(token: String, config: State<'_, ConfigState>) -> Result<(), String> {
    let mut config = config.lock().await;
    config.auth_token = Some(token);
    Ok(())
}
```

**Frontend usage:**
```typescript
// After successful login
const loginResponse = await loginToApi(username, password)
await invoke('set_auth_token', { token: loginResponse.token })
```

---

#### `clear_auth_token`

Clears the stored authentication token (logout).

- **Parameters**: None
- **Returns**: `()` on success

```rust
#[tauri::command]
async fn clear_auth_token(config: State<'_, ConfigState>) -> Result<(), String> {
    let mut config = config.lock().await;
    config.auth_token = None;
    Ok(())
}
```

**Frontend usage:**
```typescript
// On logout
await invoke('clear_auth_token')
```

---

### HTTP Proxy Command

#### `make_http_request`

Executes an HTTP request from the Rust backend, bypassing browser CORS restrictions. This is the primary mechanism for the desktop app to communicate with the catalog API.

- **Parameters**:
  - `url: String` - full URL to request
  - `method: String` - HTTP method (GET, POST, PUT, DELETE, PATCH)
  - `headers: HashMap<String, String>` - request headers
  - `body: Option<String>` - optional request body (for POST/PUT/PATCH)
- **Returns**: `String` - response body text
- **Errors**: String error message on network failure or unsupported method

```rust
#[tauri::command]
async fn make_http_request(
    url: String,
    method: String,
    headers: HashMap<String, String>,
    body: Option<String>,
) -> Result<String, String> {
    let client = reqwest::Client::new();

    let mut request = match method.to_uppercase().as_str() {
        "GET" => client.get(&url),
        "POST" => client.post(&url),
        "PUT" => client.put(&url),
        "DELETE" => client.delete(&url),
        "PATCH" => client.patch(&url),
        _ => return Err("Unsupported HTTP method".to_string()),
    };

    for (key, value) in headers {
        request = request.header(&key, &value);
    }

    if let Some(body_content) = body {
        request = request.body(body_content);
    }

    match request.send().await {
        Ok(response) => match response.text().await {
            Ok(text) => Ok(text),
            Err(e) => Err(format!("Failed to read response: {}", e)),
        },
        Err(e) => Err(format!("Request failed: {}", e)),
    }
}
```

**Frontend usage:**
```typescript
// GET request with auth header
const response = await invoke<string>('make_http_request', {
  url: 'http://192.168.1.100:8080/api/v1/media/search?query=star',
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: null,
})
const data = JSON.parse(response)

// POST request with body
const createResponse = await invoke<string>('make_http_request', {
  url: 'http://192.168.1.100:8080/api/v1/auth/login',
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'admin', password: 'secret' }),
})
```

---

### System Information Commands

#### `get_app_version`

Returns the application version from `Cargo.toml`.

- **Parameters**: None
- **Returns**: `String` - version string (e.g., "1.0.0")

```rust
#[tauri::command]
fn get_app_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}
```

---

#### `get_platform`

Returns the operating system name.

- **Parameters**: None
- **Returns**: `String` - OS identifier (e.g., "linux", "windows", "macos")

```rust
#[tauri::command]
fn get_platform() -> String {
    std::env::consts::OS.to_string()
}
```

---

#### `get_arch`

Returns the CPU architecture.

- **Parameters**: None
- **Returns**: `String` - architecture identifier (e.g., "x86_64", "aarch64")

```rust
#[tauri::command]
fn get_arch() -> String {
    std::env::consts::ARCH.to_string()
}
```

**Frontend usage for system info:**
```typescript
const version = await invoke<string>('get_app_version')
const platform = await invoke<string>('get_platform')
const arch = await invoke<string>('get_arch')
console.log(`Catalogizer Desktop v${version} on ${platform} (${arch})`)
```

## Command Summary Table

| Command | Type | Parameters | Returns | Purpose |
|---------|------|-----------|---------|---------|
| `get_config` | async | None | `AppConfig` | Read current configuration |
| `update_config` | async | `new_config: AppConfig` | `()` | Replace entire configuration |
| `set_server_url` | async | `url: String` | `()` | Set catalog API server URL |
| `set_auth_token` | async | `token: String` | `()` | Store JWT auth token |
| `clear_auth_token` | async | None | `()` | Clear auth token (logout) |
| `make_http_request` | async | `url, method, headers, body` | `String` | HTTP proxy (bypasses CORS) |
| `get_app_version` | sync | None | `String` | App version from Cargo.toml |
| `get_platform` | sync | None | `String` | OS identifier |
| `get_arch` | sync | None | `String` | CPU architecture |

## Typical Frontend Integration Pattern

A typical API call from the desktop frontend goes through this flow:

```typescript
// 1. Get config to know server URL
const config = await invoke<AppConfig>('get_config')
const baseUrl = config.server_url

// 2. Make authenticated API request through Rust backend
const response = await invoke<string>('make_http_request', {
  url: `${baseUrl}/api/v1/media/search?query=${query}`,
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${config.auth_token}`,
    'Content-Type': 'application/json',
  },
  body: null,
})

// 3. Parse response
const data = JSON.parse(response)
```

## Adding a New IPC Command

### Step 1: Define the command in Rust

```rust
#[tauri::command]
async fn my_new_command(
    param1: String,
    param2: i32,
    config: State<'_, ConfigState>,
) -> Result<String, String> {
    // Implementation
    Ok("result".to_string())
}
```

### Step 2: Register in the handler list

```rust
tauri::Builder::default()
    .invoke_handler(tauri::generate_handler![
        // ... existing commands ...
        my_new_command,
    ])
```

### Step 3: Call from the frontend

```typescript
const result = await invoke<string>('my_new_command', {
  param1: 'hello',
  param2: 42,
})
```

## Build and Development

```bash
# Development (hot-reload)
cd catalogizer-desktop
npm run tauri:dev

# Production build
npm run tauri:build
```

The Rust backend is compiled by Cargo as part of the Tauri build process. The `build.rs` script handles any build-time code generation.
