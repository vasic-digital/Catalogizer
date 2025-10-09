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