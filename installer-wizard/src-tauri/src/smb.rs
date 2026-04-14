use crate::SMBShare;
use anyhow::{anyhow, Result};
use reqwest;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Serialize, Deserialize)]
pub struct FileEntry {
    pub name: String,
    pub path: String,
    pub is_directory: bool,
    pub size: Option<u64>,
    pub modified: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct SMBShareApiResponse {
    pub host: String,
    pub share_name: String,
    pub path: String,
    pub writable: bool,
    pub description: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct FileEntryApiResponse {
    pub name: String,
    pub path: String,
    pub is_directory: bool,
    pub size: Option<i64>,
    pub modified: Option<String>,
}

/// Scan SMB shares on a specific host using catalog-api
pub async fn scan_shares(host: &str) -> Result<Vec<SMBShare>> {
    scan_shares_with_credentials(host, "guest", "", None).await
}

/// Scan SMB shares with specific credentials
pub async fn scan_shares_with_credentials(
    host: &str,
    username: &str,
    password: &str,
    domain: Option<&str>,
) -> Result<Vec<SMBShare>> {
    // Use the catalog-api SMB discovery endpoint
    let client = reqwest::Client::new();
    let api_url = get_api_base_url();

    let mut request_body = HashMap::new();
    request_body.insert("host", host);
    request_body.insert("username", username);
    request_body.insert("password", password);

    if let Some(d) = domain {
        request_body.insert("domain", d);
    }

    let response = client
        .post(&format!("{}/api/v1/smb/discover", api_url))
        .json(&request_body)
        .send()
        .await;

    match response {
        Ok(resp) if resp.status().is_success() => {
            let shares: Vec<SMBShareApiResponse> = resp
                .json()
                .await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(shares
                .into_iter()
                .map(|s| SMBShare {
                    host: s.host,
                    share_name: s.share_name,
                    path: s.path,
                    writable: s.writable,
                    description: s.description,
                })
                .collect())
        }
        Ok(resp) => {
            // API call failed, return error
            log::warn!("SMB discovery API failed with status: {}", resp.status());
            Err(anyhow!(
                "SMB discovery API failed with status: {}",
                resp.status()
            ))
        }
        Err(e) => {
            // Network error, return error
            log::warn!("SMB discovery API network error: {}", e);
            Err(anyhow!("SMB discovery API network error: {}", e))
        }
    }
}

/// Browse files and directories in an SMB share
pub async fn browse_share(host: &str, share: &str, path: Option<&str>) -> Result<Vec<FileEntry>> {
    browse_share_with_credentials(host, share, path, "guest", "", None).await
}

/// Browse files and directories in an SMB share with credentials
pub async fn browse_share_with_credentials(
    host: &str,
    share: &str,
    path: Option<&str>,
    username: &str,
    password: &str,
    domain: Option<&str>,
) -> Result<Vec<FileEntry>> {
    // Use the catalog-api SMB browse endpoint
    let client = reqwest::Client::new();
    let api_url = get_api_base_url();

    let mut request_body = HashMap::new();
    request_body.insert("host", host);
    request_body.insert("share", share);
    request_body.insert("username", username);
    request_body.insert("password", password);
    request_body.insert("port", "445");

    if let Some(p) = path {
        request_body.insert("path", p);
    } else {
        request_body.insert("path", ".");
    }

    if let Some(d) = domain {
        request_body.insert("domain", d);
    }

    let response = client
        .post(&format!("{}/api/v1/smb/browse", api_url))
        .json(&request_body)
        .send()
        .await;

    match response {
        Ok(resp) if resp.status().is_success() => {
            let entries: Vec<FileEntryApiResponse> = resp
                .json()
                .await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(entries
                .into_iter()
                .map(|e| FileEntry {
                    name: e.name,
                    path: e.path,
                    is_directory: e.is_directory,
                    size: e.size.map(|s| s as u64),
                    modified: e.modified,
                })
                .collect())
        }
        Ok(resp) => {
            log::warn!("SMB browse API failed with status: {}", resp.status());
            Err(anyhow!(
                "SMB browse API failed with status: {}",
                resp.status()
            ))
        }
        Err(e) => {
            log::warn!("SMB browse API network error: {}", e);
            Err(anyhow!("SMB browse API network error: {}", e))
        }
    }
}

/// Test SMB connection with credentials
pub async fn test_connection(
    host: &str,
    share: &str,
    username: &str,
    password: &str,
    domain: Option<&str>,
) -> Result<bool> {
    // Use the catalog-api SMB test endpoint
    let client = reqwest::Client::new();
    let api_url = get_api_base_url();

    let mut request_body = HashMap::new();
    request_body.insert("host", host);
    request_body.insert("share", share);
    request_body.insert("username", username);
    request_body.insert("password", password);
    request_body.insert("port", "445");

    if let Some(d) = domain {
        request_body.insert("domain", d);
    }

    let response = client
        .post(&format!("{}/api/v1/smb/test", api_url))
        .json(&request_body)
        .send()
        .await;

    match response {
        Ok(resp) if resp.status().is_success() => {
            let result: serde_json::Value = resp
                .json()
                .await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(result
                .get("success")
                .and_then(|v| v.as_bool())
                .unwrap_or(false))
        }
        Ok(resp) => {
            log::warn!("SMB test API failed with status: {}", resp.status());
            Ok(false)
        }
        Err(e) => {
            log::warn!("SMB test API network error: {}", e);
            Ok(false)
        }
    }
}

/// Get the API base URL - assumes catalog-api is running on localhost:8080
fn get_api_base_url() -> String {
    std::env::var("CATALOG_API_URL").unwrap_or_else(|_| "http://localhost:8080".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Tests for FileEntry struct
    mod file_entry_tests {
        use super::*;

        #[test]
        fn test_file_entry_serialization() {
            let entry = FileEntry {
                name: "movie.mkv".to_string(),
                path: "/media/movies/movie.mkv".to_string(),
                is_directory: false,
                size: Some(4294967296),
                modified: Some("2024-06-15 10:30:00".to_string()),
            };

            let json = serde_json::to_string(&entry).unwrap();
            let deserialized: FileEntry = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.name, "movie.mkv");
            assert_eq!(deserialized.path, "/media/movies/movie.mkv");
            assert!(!deserialized.is_directory);
            assert_eq!(deserialized.size, Some(4294967296));
            assert!(deserialized.modified.is_some());
        }

        #[test]
        fn test_file_entry_directory() {
            let entry = FileEntry {
                name: "Documents".to_string(),
                path: "Documents".to_string(),
                is_directory: true,
                size: None,
                modified: None,
            };

            assert!(entry.is_directory);
            assert!(entry.size.is_none());
        }

        #[test]
        fn test_file_entry_debug_trait() {
            let entry = FileEntry {
                name: "test.txt".to_string(),
                path: "test.txt".to_string(),
                is_directory: false,
                size: Some(1024),
                modified: None,
            };

            let debug = format!("{:?}", entry);
            assert!(debug.contains("FileEntry"));
            assert!(debug.contains("test.txt"));
        }
    }

    /// Tests for API base URL
    mod api_url_tests {
        use super::*;

        #[test]
        fn test_default_api_base_url() {
            // Clear any existing env var to test default
            std::env::remove_var("CATALOG_API_URL");
            let url = get_api_base_url();
            assert_eq!(url, "http://localhost:8080");
        }

        #[test]
        fn test_custom_api_base_url() {
            std::env::set_var("CATALOG_API_URL", "http://custom:9090");
            let url = get_api_base_url();
            assert_eq!(url, "http://custom:9090");
            std::env::remove_var("CATALOG_API_URL");
        }
    }

    /// Tests for SMBShareApiResponse struct
    mod smb_share_api_response_tests {
        use super::*;

        #[test]
        fn test_smb_share_api_response_serialization() {
            let response = SMBShareApiResponse {
                host: "192.168.1.50".to_string(),
                share_name: "media".to_string(),
                path: "\\\\192.168.1.50\\media".to_string(),
                writable: true,
                description: Some("Media share".to_string()),
            };

            let json = serde_json::to_string(&response).unwrap();
            let deserialized: SMBShareApiResponse = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.host, "192.168.1.50");
            assert_eq!(deserialized.share_name, "media");
            assert!(deserialized.writable);
            assert_eq!(deserialized.description, Some("Media share".to_string()));
        }

        #[test]
        fn test_smb_share_api_response_without_description() {
            let response = SMBShareApiResponse {
                host: "nas.local".to_string(),
                share_name: "public".to_string(),
                path: "/public".to_string(),
                writable: false,
                description: None,
            };

            let json = serde_json::to_string(&response).unwrap();
            assert!(json.contains("\"description\":null"));
        }

        #[test]
        fn test_smb_share_api_response_debug_trait() {
            let response = SMBShareApiResponse {
                host: "host".to_string(),
                share_name: "share".to_string(),
                path: "/share".to_string(),
                writable: false,
                description: None,
            };
            let debug = format!("{:?}", response);
            assert!(debug.contains("SMBShareApiResponse"));
        }
    }

    /// Tests for FileEntryApiResponse struct
    mod file_entry_api_response_tests {
        use super::*;

        #[test]
        fn test_file_entry_api_response_serialization() {
            let entry = FileEntryApiResponse {
                name: "video.mp4".to_string(),
                path: "/media/video.mp4".to_string(),
                is_directory: false,
                size: Some(1073741824),
                modified: Some("2024-01-15 12:00:00".to_string()),
            };

            let json = serde_json::to_string(&entry).unwrap();
            let deserialized: FileEntryApiResponse = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.name, "video.mp4");
            assert!(!deserialized.is_directory);
            assert_eq!(deserialized.size, Some(1073741824));
        }

        #[test]
        fn test_file_entry_api_response_directory() {
            let entry = FileEntryApiResponse {
                name: "Movies".to_string(),
                path: "/media/Movies".to_string(),
                is_directory: true,
                size: None,
                modified: None,
            };

            assert!(entry.is_directory);
            assert!(entry.size.is_none());
            assert!(entry.modified.is_none());
        }

        #[test]
        fn test_file_entry_api_response_negative_size() {
            // API may return negative size (i64), conversion to u64 handles it
            let entry = FileEntryApiResponse {
                name: "file.dat".to_string(),
                path: "file.dat".to_string(),
                is_directory: false,
                size: Some(-1),
                modified: None,
            };

            let file_entry = FileEntry {
                name: entry.name.clone(),
                path: entry.path.clone(),
                is_directory: entry.is_directory,
                size: entry.size.map(|s| s as u64),
                modified: entry.modified.clone(),
            };

            // Negative i64 wraps around when cast to u64
            assert!(file_entry.size.is_some());
        }

        #[test]
        fn test_file_entry_api_response_zero_size() {
            let entry = FileEntryApiResponse {
                name: "empty.txt".to_string(),
                path: "empty.txt".to_string(),
                is_directory: false,
                size: Some(0),
                modified: None,
            };

            let file_entry = FileEntry {
                name: entry.name,
                path: entry.path,
                is_directory: entry.is_directory,
                size: entry.size.map(|s| s as u64),
                modified: entry.modified,
            };

            assert_eq!(file_entry.size, Some(0));
        }
    }

    /// Tests for API endpoint URL construction
    mod api_endpoint_tests {
        use super::*;

        #[test]
        fn test_discover_endpoint_url() {
            std::env::remove_var("CATALOG_API_URL");
            let api_url = get_api_base_url();
            let endpoint = format!("{}/api/v1/smb/discover", api_url);
            assert_eq!(endpoint, "http://localhost:8080/api/v1/smb/discover");
        }

        #[test]
        fn test_browse_endpoint_url() {
            std::env::remove_var("CATALOG_API_URL");
            let api_url = get_api_base_url();
            let endpoint = format!("{}/api/v1/smb/browse", api_url);
            assert_eq!(endpoint, "http://localhost:8080/api/v1/smb/browse");
        }

        #[test]
        fn test_test_endpoint_url() {
            std::env::remove_var("CATALOG_API_URL");
            let api_url = get_api_base_url();
            let endpoint = format!("{}/api/v1/smb/test", api_url);
            assert_eq!(endpoint, "http://localhost:8080/api/v1/smb/test");
        }
    }

    /// Tests for request body construction
    mod request_body_tests {
        use super::*;

        #[test]
        fn test_scan_request_body_without_domain() {
            let mut body = HashMap::new();
            body.insert("host", "192.168.1.100");
            body.insert("username", "guest");
            body.insert("password", "");

            assert_eq!(body.len(), 3);
            assert_eq!(body.get("host"), Some(&"192.168.1.100"));
            assert!(!body.contains_key("domain"));
        }

        #[test]
        fn test_scan_request_body_with_domain() {
            let mut body = HashMap::new();
            body.insert("host", "192.168.1.100");
            body.insert("username", "admin");
            body.insert("password", "pass");
            body.insert("domain", "WORKGROUP");

            assert_eq!(body.len(), 4);
            assert_eq!(body.get("domain"), Some(&"WORKGROUP"));
        }

        #[test]
        fn test_browse_request_body_with_path() {
            let mut body = HashMap::new();
            body.insert("host", "192.168.1.100");
            body.insert("share", "media");
            body.insert("username", "guest");
            body.insert("password", "");
            body.insert("port", "445");
            body.insert("path", "/movies");

            assert_eq!(body.len(), 6);
            assert_eq!(body.get("path"), Some(&"/movies"));
            assert_eq!(body.get("port"), Some(&"445"));
        }

        #[test]
        fn test_browse_request_body_default_path() {
            let mut body = HashMap::new();
            body.insert("host", "192.168.1.100");
            body.insert("share", "media");
            body.insert("path", ".");

            assert_eq!(body.get("path"), Some(&"."));
        }
    }

    /// Tests for SMBShare conversion from API response
    mod conversion_tests {
        use super::*;

        #[test]
        fn test_api_response_to_smb_share_conversion() {
            let api_response = SMBShareApiResponse {
                host: "nas.local".to_string(),
                share_name: "videos".to_string(),
                path: "\\\\nas.local\\videos".to_string(),
                writable: true,
                description: Some("Video library".to_string()),
            };

            let share = SMBShare {
                host: api_response.host,
                share_name: api_response.share_name,
                path: api_response.path,
                writable: api_response.writable,
                description: api_response.description,
            };

            assert_eq!(share.host, "nas.local");
            assert_eq!(share.share_name, "videos");
            assert!(share.writable);
        }

        #[test]
        fn test_file_entry_api_to_file_entry_conversion() {
            let api_entry = FileEntryApiResponse {
                name: "photo.jpg".to_string(),
                path: "/photos/photo.jpg".to_string(),
                is_directory: false,
                size: Some(2048000),
                modified: Some("2024-06-01 09:00:00".to_string()),
            };

            let entry = FileEntry {
                name: api_entry.name,
                path: api_entry.path,
                is_directory: api_entry.is_directory,
                size: api_entry.size.map(|s| s as u64),
                modified: api_entry.modified,
            };

            assert_eq!(entry.name, "photo.jpg");
            assert_eq!(entry.size, Some(2048000u64));
        }
    }
}
