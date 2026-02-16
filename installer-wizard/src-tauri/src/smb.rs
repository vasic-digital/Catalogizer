use crate::SMBShare;
use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use reqwest;
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
            let shares: Vec<SMBShareApiResponse> = resp.json().await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(shares.into_iter().map(|s| SMBShare {
                host: s.host,
                share_name: s.share_name,
                path: s.path,
                writable: s.writable,
                description: s.description,
            }).collect())
        }
        Ok(resp) => {
            // API call failed, fallback to common shares
            log::warn!("SMB discovery API failed with status: {}", resp.status());
            Ok(get_common_shares(host))
        }
        Err(e) => {
            // Network error, fallback to common shares
            log::warn!("SMB discovery API network error: {}", e);
            Ok(get_common_shares(host))
        }
    }
}


/// Get common SMB share names to try
fn get_common_shares(host: &str) -> Vec<SMBShare> {
    let common_shares = vec![
        ("shared", "Shared folder"),
        ("public", "Public folder"),
        ("media", "Media files"),
        ("downloads", "Downloads"),
        ("documents", "Documents"),
        ("music", "Music files"),
        ("videos", "Video files"),
        ("pictures", "Pictures"),
        ("backup", "Backup files"),
    ];

    common_shares
        .into_iter()
        .map(|(name, desc)| SMBShare {
            host: host.to_string(),
            share_name: name.to_string(),
            path: format!("\\\\{}\\{}", host, name),
            writable: false,
            description: Some(desc.to_string()),
        })
        .collect()
}

/// Browse files and directories in an SMB share
pub async fn browse_share(
    host: &str,
    share: &str,
    path: Option<&str>,
) -> Result<Vec<FileEntry>> {
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
            let entries: Vec<FileEntryApiResponse> = resp.json().await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(entries.into_iter().map(|e| FileEntry {
                name: e.name,
                path: e.path,
                is_directory: e.is_directory,
                size: e.size.map(|s| s as u64),
                modified: e.modified,
            }).collect())
        }
        Ok(resp) => {
            log::warn!("SMB browse API failed with status: {}", resp.status());
            Ok(get_mock_entries())
        }
        Err(e) => {
            log::warn!("SMB browse API network error: {}", e);
            Ok(get_mock_entries())
        }
    }
}

/// Get mock entries for fallback
fn get_mock_entries() -> Vec<FileEntry> {
    vec![
        FileEntry {
            name: "..".to_string(),
            path: "..".to_string(),
            is_directory: true,
            size: None,
            modified: None,
        },
        FileEntry {
            name: "Example Folder".to_string(),
            path: "Example Folder".to_string(),
            is_directory: true,
            size: None,
            modified: Some("2024-01-01 12:00:00".to_string()),
        },
        FileEntry {
            name: "example.txt".to_string(),
            path: "example.txt".to_string(),
            is_directory: false,
            size: Some(1024),
            modified: Some("2024-01-01 12:00:00".to_string()),
        },
    ]
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
            let result: serde_json::Value = resp.json().await
                .map_err(|e| anyhow!("Failed to parse API response: {}", e))?;

            Ok(result.get("success").and_then(|v| v.as_bool()).unwrap_or(false))
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

    /// Tests for get_common_shares
    mod common_shares_tests {
        use super::*;

        #[test]
        fn test_get_common_shares_returns_expected_count() {
            let shares = get_common_shares("192.168.1.10");
            assert_eq!(shares.len(), 9);
        }

        #[test]
        fn test_get_common_shares_has_expected_names() {
            let shares = get_common_shares("testhost");
            let names: Vec<&str> = shares.iter().map(|s| s.share_name.as_str()).collect();

            assert!(names.contains(&"shared"));
            assert!(names.contains(&"public"));
            assert!(names.contains(&"media"));
            assert!(names.contains(&"downloads"));
            assert!(names.contains(&"documents"));
            assert!(names.contains(&"music"));
            assert!(names.contains(&"videos"));
            assert!(names.contains(&"pictures"));
            assert!(names.contains(&"backup"));
        }

        #[test]
        fn test_get_common_shares_host_is_set() {
            let host = "192.168.1.100";
            let shares = get_common_shares(host);

            for share in &shares {
                assert_eq!(share.host, host);
            }
        }

        #[test]
        fn test_get_common_shares_path_format() {
            let shares = get_common_shares("myhost");

            for share in &shares {
                assert!(share.path.starts_with("\\\\myhost\\"));
                assert!(share.path.contains(&share.share_name));
            }
        }

        #[test]
        fn test_get_common_shares_all_read_only() {
            let shares = get_common_shares("host");

            for share in &shares {
                assert!(!share.writable, "Share {} should not be writable", share.share_name);
            }
        }

        #[test]
        fn test_get_common_shares_all_have_descriptions() {
            let shares = get_common_shares("host");

            for share in &shares {
                assert!(share.description.is_some(), "Share {} should have a description", share.share_name);
            }
        }
    }

    /// Tests for get_mock_entries
    mod mock_entries_tests {
        use super::*;

        #[test]
        fn test_get_mock_entries_returns_three_entries() {
            let entries = get_mock_entries();
            assert_eq!(entries.len(), 3);
        }

        #[test]
        fn test_get_mock_entries_has_parent_directory() {
            let entries = get_mock_entries();
            let parent = &entries[0];
            assert_eq!(parent.name, "..");
            assert!(parent.is_directory);
        }

        #[test]
        fn test_get_mock_entries_has_folder() {
            let entries = get_mock_entries();
            let folder = &entries[1];
            assert_eq!(folder.name, "Example Folder");
            assert!(folder.is_directory);
        }

        #[test]
        fn test_get_mock_entries_has_file() {
            let entries = get_mock_entries();
            let file = &entries[2];
            assert_eq!(file.name, "example.txt");
            assert!(!file.is_directory);
            assert_eq!(file.size, Some(1024));
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
}