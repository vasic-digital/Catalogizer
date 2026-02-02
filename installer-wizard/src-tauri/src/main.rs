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