// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::{Deserialize, Serialize};

mod network;
mod smb;

#[derive(Debug, Serialize, Deserialize)]
pub struct NetworkHost {
    pub ip: String,
    pub hostname: Option<String>,
    pub mac_address: Option<String>,
    pub vendor: Option<String>,
    pub open_ports: Vec<u16>,
    pub smb_shares: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SMBShare {
    pub host: String,
    pub share_name: String,
    pub path: String,
    pub writable: bool,
    pub description: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConfigurationSource {
    pub r#type: String,
    pub url: String,
    pub access: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConfigurationAccess {
    pub name: String,
    pub r#type: String,
    pub account: String,
    pub secret: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Configuration {
    pub accesses: Vec<ConfigurationAccess>,
    pub sources: Vec<ConfigurationSource>,
}

// Learn more about Tauri commands at https://tauri.app/v1/guides/features/command
#[tauri::command]
async fn scan_network() -> Result<Vec<NetworkHost>, String> {
    network::scan_network().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn scan_smb_shares(host: String) -> Result<Vec<SMBShare>, String> {
    smb::scan_shares(&host).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn browse_smb_share(host: String, share: String, path: Option<String>) -> Result<Vec<smb::FileEntry>, String> {
    smb::browse_share(&host, &share, path.as_deref()).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn test_smb_connection(
    host: String,
    share: String,
    username: String,
    password: String,
    domain: Option<String>,
) -> Result<bool, String> {
    smb::test_connection(&host, &share, &username, &password, domain.as_deref())
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn load_configuration(file_path: String) -> Result<Configuration, String> {
    use std::fs;

    let content = fs::read_to_string(&file_path)
        .map_err(|e| format!("Failed to read file: {}", e))?;

    let config: Configuration = serde_json::from_str(&content)
        .map_err(|e| format!("Failed to parse JSON: {}", e))?;

    Ok(config)
}

#[tauri::command]
async fn save_configuration(file_path: String, config: Configuration) -> Result<(), String> {
    use std::fs;

    let content = serde_json::to_string_pretty(&config)
        .map_err(|e| format!("Failed to serialize configuration: {}", e))?;

    fs::write(&file_path, content)
        .map_err(|e| format!("Failed to write file: {}", e))?;

    Ok(())
}

#[tauri::command]
async fn test_ftp_connection(
    host: String,
    port: u16,
    username: String,
    password: String,
    path: Option<String>,
) -> Result<bool, String> {
    use std::net::TcpStream;
    use std::time::Duration;

    let addr = format!("{}:{}", host, port);
    let stream = TcpStream::connect_timeout(
        &addr.parse().map_err(|e| format!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| format!("FTP connection failed: {}", e))?;

    // Read the FTP banner
    use std::io::Read;
    let mut stream = stream;
    stream
        .set_read_timeout(Some(Duration::from_secs(5)))
        .map_err(|e| format!("Failed to set timeout: {}", e))?;
    let mut buf = [0u8; 512];
    let n = stream
        .read(&mut buf)
        .map_err(|e| format!("Failed to read FTP banner: {}", e))?;
    let banner = String::from_utf8_lossy(&buf[..n]);

    if !banner.starts_with("220") {
        return Err(format!("Unexpected FTP response: {}", banner.trim()));
    }

    // Send USER command
    use std::io::Write;
    write!(stream, "USER {}\r\n", username)
        .map_err(|e| format!("Failed to send USER: {}", e))?;
    let n = stream.read(&mut buf).map_err(|e| format!("Failed to read: {}", e))?;
    let _response = String::from_utf8_lossy(&buf[..n]);

    // Send PASS command
    write!(stream, "PASS {}\r\n", password)
        .map_err(|e| format!("Failed to send PASS: {}", e))?;
    let n = stream.read(&mut buf).map_err(|e| format!("Failed to read: {}", e))?;
    let response = String::from_utf8_lossy(&buf[..n]);

    if !response.starts_with("230") {
        return Err(format!("FTP authentication failed: {}", response.trim()));
    }

    // If path specified, try CWD
    if let Some(ref p) = path {
        write!(stream, "CWD {}\r\n", p)
            .map_err(|e| format!("Failed to send CWD: {}", e))?;
        let n = stream.read(&mut buf).map_err(|e| format!("Failed to read: {}", e))?;
        let response = String::from_utf8_lossy(&buf[..n]);
        if !response.starts_with("250") {
            return Err(format!("FTP path not accessible: {}", response.trim()));
        }
    }

    // Quit
    let _ = write!(stream, "QUIT\r\n");

    Ok(true)
}

#[tauri::command]
async fn test_nfs_connection(
    host: String,
    path: String,
    mount_point: String,
    options: Option<String>,
) -> Result<bool, String> {
    // Test NFS by checking if the host is reachable on port 2049 (standard NFS port)
    use std::net::TcpStream;
    use std::time::Duration;

    let addr = format!("{}:2049", host);
    TcpStream::connect_timeout(
        &addr.parse().map_err(|e| format!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| format!("NFS host not reachable on port 2049: {}", e))?;

    // Verify the mount point directory exists or can be created
    let mount_path = std::path::Path::new(&mount_point);
    if !mount_path.exists() {
        std::fs::create_dir_all(mount_path)
            .map_err(|e| format!("Cannot create mount point '{}': {}", mount_point, e))?;
    }

    let _ = (path, options); // path and options would be used during actual mount

    Ok(true)
}

#[tauri::command]
async fn test_webdav_connection(
    url: String,
    username: String,
    password: String,
    path: Option<String>,
) -> Result<bool, String> {
    use std::io::{Read, Write};
    use std::net::TcpStream;
    use std::time::Duration;

    // Parse URL to extract host and port
    let url_str = if let Some(ref p) = path {
        format!("{}/{}", url.trim_end_matches('/'), p.trim_start_matches('/'))
    } else {
        url.clone()
    };

    // Use a simple HTTP PROPFIND to test WebDAV
    let parsed = url_str.strip_prefix("http://").or_else(|| url_str.strip_prefix("https://"));
    let (host_port, request_path) = match parsed {
        Some(rest) => {
            let (hp, p) = rest.split_once('/').unwrap_or((rest, ""));
            (hp.to_string(), format!("/{}", p))
        }
        None => return Err("Invalid URL format".to_string()),
    };

    let is_https = url_str.starts_with("https://");
    if is_https {
        // For HTTPS, just verify the host is reachable
        let port_addr = if host_port.contains(':') {
            host_port.clone()
        } else {
            format!("{}:443", host_port)
        };
        TcpStream::connect_timeout(
            &port_addr.parse().map_err(|e| format!("Invalid address: {}", e))?,
            Duration::from_secs(10),
        )
        .map_err(|e| format!("WebDAV host not reachable: {}", e))?;
        return Ok(true);
    }

    let port_addr = if host_port.contains(':') {
        host_port.clone()
    } else {
        format!("{}:80", host_port)
    };

    let mut stream = TcpStream::connect_timeout(
        &port_addr.parse().map_err(|e| format!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| format!("WebDAV connection failed: {}", e))?;

    // Build basic auth header
    use base64::Engine;
    let credentials = base64::engine::general_purpose::STANDARD.encode(format!("{}:{}", username, password));

    let request = format!(
        "PROPFIND {} HTTP/1.1\r\nHost: {}\r\nAuthorization: Basic {}\r\nDepth: 0\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
        request_path, host_port, credentials
    );

    stream.set_write_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| format!("Failed to set timeout: {}", e))?;
    stream.write_all(request.as_bytes())
        .map_err(|e| format!("Failed to send request: {}", e))?;

    stream.set_read_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| format!("Failed to set timeout: {}", e))?;
    let mut response = String::new();
    stream.read_to_string(&mut response)
        .map_err(|e| format!("Failed to read response: {}", e))?;

    // Check for successful WebDAV response (207 Multi-Status or 200 OK)
    if response.contains("207") || response.contains("200") {
        Ok(true)
    } else if response.contains("401") || response.contains("403") {
        Err("WebDAV authentication failed".to_string())
    } else {
        Err(format!("WebDAV returned unexpected response: {}", response.lines().next().unwrap_or("")))
    }
}

#[tauri::command]
async fn test_local_connection(base_path: String) -> Result<bool, String> {
    let path = std::path::Path::new(&base_path);

    if !path.exists() {
        return Err(format!("Path does not exist: {}", base_path));
    }

    if !path.is_dir() {
        return Err(format!("Path is not a directory: {}", base_path));
    }

    // Check read permission by trying to read the directory
    std::fs::read_dir(path)
        .map_err(|e| format!("Cannot read directory '{}': {}", base_path, e))?;

    Ok(true)
}

#[tauri::command]
async fn get_default_config_path() -> Result<String, String> {
    use std::env;
    use std::path::PathBuf;

    let home_dir = env::var("HOME")
        .or_else(|_| env::var("USERPROFILE"))
        .map_err(|_| "Unable to determine home directory")?;

    let mut path = PathBuf::from(home_dir);
    path.push(".catalogizer");
    path.push("config.json");

    Ok(path.to_string_lossy().to_string())
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .invoke_handler(tauri::generate_handler![
            scan_network,
            scan_smb_shares,
            browse_smb_share,
            test_smb_connection,
            test_ftp_connection,
            test_nfs_connection,
            test_webdav_connection,
            test_local_connection,
            load_configuration,
            save_configuration,
            get_default_config_path
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Tests for NetworkHost struct
    mod network_host_tests {
        use super::*;

        #[test]
        fn test_network_host_serialization() {
            let host = NetworkHost {
                ip: "192.168.1.100".to_string(),
                hostname: Some("fileserver".to_string()),
                mac_address: Some("AA:BB:CC:DD:EE:FF".to_string()),
                vendor: None,
                open_ports: vec![22, 80, 445],
                smb_shares: vec!["shared".to_string(), "media".to_string()],
            };

            let json = serde_json::to_string(&host).unwrap();
            let deserialized: NetworkHost = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.ip, "192.168.1.100");
            assert_eq!(deserialized.hostname, Some("fileserver".to_string()));
            assert_eq!(deserialized.mac_address, Some("AA:BB:CC:DD:EE:FF".to_string()));
            assert!(deserialized.vendor.is_none());
            assert_eq!(deserialized.open_ports, vec![22, 80, 445]);
            assert_eq!(deserialized.smb_shares.len(), 2);
        }

        #[test]
        fn test_network_host_with_no_optional_fields() {
            let host = NetworkHost {
                ip: "10.0.0.1".to_string(),
                hostname: None,
                mac_address: None,
                vendor: None,
                open_ports: vec![],
                smb_shares: vec![],
            };

            let json = serde_json::to_string(&host).unwrap();
            assert!(json.contains("\"ip\":\"10.0.0.1\""));
            assert!(json.contains("\"hostname\":null"));
        }

        #[test]
        fn test_network_host_deserialization_from_json() {
            let json = r#"{
                "ip": "172.16.0.1",
                "hostname": "nas",
                "mac_address": null,
                "vendor": "Synology",
                "open_ports": [80, 443, 5000],
                "smb_shares": ["homes", "video"]
            }"#;

            let host: NetworkHost = serde_json::from_str(json).unwrap();
            assert_eq!(host.ip, "172.16.0.1");
            assert_eq!(host.hostname, Some("nas".to_string()));
            assert!(host.mac_address.is_none());
            assert_eq!(host.vendor, Some("Synology".to_string()));
            assert_eq!(host.open_ports.len(), 3);
        }

        #[test]
        fn test_network_host_debug_trait() {
            let host = NetworkHost {
                ip: "1.2.3.4".to_string(),
                hostname: None,
                mac_address: None,
                vendor: None,
                open_ports: vec![],
                smb_shares: vec![],
            };
            let debug = format!("{:?}", host);
            assert!(debug.contains("NetworkHost"));
            assert!(debug.contains("1.2.3.4"));
        }
    }

    /// Tests for SMBShare struct
    mod smb_share_tests {
        use super::*;

        #[test]
        fn test_smb_share_serialization() {
            let share = SMBShare {
                host: "192.168.1.50".to_string(),
                share_name: "media".to_string(),
                path: "\\\\192.168.1.50\\media".to_string(),
                writable: true,
                description: Some("Media files".to_string()),
            };

            let json = serde_json::to_string(&share).unwrap();
            let deserialized: SMBShare = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.host, "192.168.1.50");
            assert_eq!(deserialized.share_name, "media");
            assert!(deserialized.writable);
            assert_eq!(deserialized.description, Some("Media files".to_string()));
        }

        #[test]
        fn test_smb_share_without_description() {
            let share = SMBShare {
                host: "10.0.0.5".to_string(),
                share_name: "data".to_string(),
                path: "\\\\10.0.0.5\\data".to_string(),
                writable: false,
                description: None,
            };

            let json = serde_json::to_string(&share).unwrap();
            assert!(json.contains("\"description\":null"));
        }
    }

    /// Tests for ConfigurationSource struct
    mod configuration_source_tests {
        use super::*;

        #[test]
        fn test_configuration_source_serialization() {
            let source = ConfigurationSource {
                r#type: "smb".to_string(),
                url: "smb://192.168.1.100/media".to_string(),
                access: "my-access".to_string(),
            };

            let json = serde_json::to_string(&source).unwrap();
            let deserialized: ConfigurationSource = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.r#type, "smb");
            assert_eq!(deserialized.url, "smb://192.168.1.100/media");
            assert_eq!(deserialized.access, "my-access");
        }

        #[test]
        fn test_configuration_source_type_field() {
            // Test that the `type` reserved keyword is handled correctly via r#type
            let json = r#"{"type": "ftp", "url": "ftp://example.com", "access": "ftp-creds"}"#;
            let source: ConfigurationSource = serde_json::from_str(json).unwrap();
            assert_eq!(source.r#type, "ftp");
        }
    }

    /// Tests for ConfigurationAccess struct
    mod configuration_access_tests {
        use super::*;

        #[test]
        fn test_configuration_access_serialization() {
            let access = ConfigurationAccess {
                name: "NAS Credentials".to_string(),
                r#type: "smb".to_string(),
                account: "admin".to_string(),
                secret: "password123".to_string(),
            };

            let json = serde_json::to_string(&access).unwrap();
            let deserialized: ConfigurationAccess = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.name, "NAS Credentials");
            assert_eq!(deserialized.r#type, "smb");
            assert_eq!(deserialized.account, "admin");
            assert_eq!(deserialized.secret, "password123");
        }

        #[test]
        fn test_configuration_access_type_field() {
            let json = r#"{"name": "test", "type": "ftp", "account": "user", "secret": "pass"}"#;
            let access: ConfigurationAccess = serde_json::from_str(json).unwrap();
            assert_eq!(access.r#type, "ftp");
        }
    }

    /// Tests for Configuration struct
    mod configuration_tests {
        use super::*;

        #[test]
        fn test_configuration_serialization_roundtrip() {
            let config = Configuration {
                accesses: vec![
                    ConfigurationAccess {
                        name: "NAS".to_string(),
                        r#type: "smb".to_string(),
                        account: "admin".to_string(),
                        secret: "pass".to_string(),
                    },
                ],
                sources: vec![
                    ConfigurationSource {
                        r#type: "smb".to_string(),
                        url: "smb://nas/media".to_string(),
                        access: "NAS".to_string(),
                    },
                ],
            };

            let json = serde_json::to_string_pretty(&config).unwrap();
            let deserialized: Configuration = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.accesses.len(), 1);
            assert_eq!(deserialized.sources.len(), 1);
            assert_eq!(deserialized.accesses[0].name, "NAS");
            assert_eq!(deserialized.sources[0].url, "smb://nas/media");
        }

        #[test]
        fn test_empty_configuration() {
            let config = Configuration {
                accesses: vec![],
                sources: vec![],
            };

            let json = serde_json::to_string(&config).unwrap();
            let deserialized: Configuration = serde_json::from_str(&json).unwrap();

            assert!(deserialized.accesses.is_empty());
            assert!(deserialized.sources.is_empty());
        }

        #[test]
        fn test_configuration_with_multiple_entries() {
            let config = Configuration {
                accesses: vec![
                    ConfigurationAccess {
                        name: "NAS1".to_string(),
                        r#type: "smb".to_string(),
                        account: "admin".to_string(),
                        secret: "pass1".to_string(),
                    },
                    ConfigurationAccess {
                        name: "FTP Server".to_string(),
                        r#type: "ftp".to_string(),
                        account: "ftpuser".to_string(),
                        secret: "ftppass".to_string(),
                    },
                ],
                sources: vec![
                    ConfigurationSource {
                        r#type: "smb".to_string(),
                        url: "smb://nas1/media".to_string(),
                        access: "NAS1".to_string(),
                    },
                    ConfigurationSource {
                        r#type: "ftp".to_string(),
                        url: "ftp://ftp.example.com/files".to_string(),
                        access: "FTP Server".to_string(),
                    },
                ],
            };

            let json = serde_json::to_string(&config).unwrap();
            let deserialized: Configuration = serde_json::from_str(&json).unwrap();

            assert_eq!(deserialized.accesses.len(), 2);
            assert_eq!(deserialized.sources.len(), 2);
        }

        #[test]
        fn test_configuration_json_deserialization() {
            let json = r#"{
                "accesses": [
                    {"name": "test", "type": "smb", "account": "user", "secret": "pass"}
                ],
                "sources": [
                    {"type": "smb", "url": "smb://host/share", "access": "test"}
                ]
            }"#;

            let config: Configuration = serde_json::from_str(json).unwrap();
            assert_eq!(config.accesses.len(), 1);
            assert_eq!(config.sources.len(), 1);
        }
    }

    /// Tests for local connection validation
    mod local_connection_tests {
        #[tokio::test]
        async fn test_local_connection_with_nonexistent_path() {
            let result = super::test_local_connection("/nonexistent/path/12345".to_string()).await;
            assert!(result.is_err());
            assert!(result.unwrap_err().contains("does not exist"));
        }

        #[tokio::test]
        async fn test_local_connection_with_valid_directory() {
            let result = super::test_local_connection("/tmp".to_string()).await;
            assert!(result.is_ok());
            assert!(result.unwrap());
        }
    }

    /// Tests for default config path
    mod config_path_tests {
        #[tokio::test]
        async fn test_get_default_config_path_returns_path() {
            let result = super::get_default_config_path().await;
            assert!(result.is_ok());
            let path = result.unwrap();
            assert!(path.contains(".catalogizer"));
            assert!(path.contains("config.json"));
        }

        #[tokio::test]
        async fn test_get_default_config_path_format() {
            let result = super::get_default_config_path().await.unwrap();
            // Should end with the expected filename
            assert!(result.ends_with("config.json"));
        }
    }
}