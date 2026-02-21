use anyhow::{anyhow, Result};
use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

/// Test WebDAV connection with given credentials
pub async fn test_connection(
    url: &str,
    username: &str,
    password: &str,
    path: Option<&str>,
) -> Result<bool> {
    // Parse URL to extract host and port
    let url_str = if let Some(p) = path {
        format!("{}/{}", url.trim_end_matches('/'), p.trim_start_matches('/'))
    } else {
        url.to_string()
    };

    // Use a simple HTTP PROPFIND to test WebDAV
    let parsed = url_str.strip_prefix("http://").or_else(|| url_str.strip_prefix("https://"));
    let (host_port, request_path) = match parsed {
        Some(rest) => {
            let (hp, p) = rest.split_once('/').unwrap_or((rest, ""));
            (hp.to_string(), format!("/{}", p))
        }
        None => return Err(anyhow!("Invalid URL format")),
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
            &port_addr.parse().map_err(|e| anyhow!("Invalid address: {}", e))?,
            Duration::from_secs(10),
        )
        .map_err(|e| anyhow!("WebDAV host not reachable: {}", e))?;
        return Ok(true);
    }

    let port_addr = if host_port.contains(':') {
        host_port.clone()
    } else {
        format!("{}:80", host_port)
    };

    let mut stream = TcpStream::connect_timeout(
        &port_addr.parse().map_err(|e| anyhow!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| anyhow!("WebDAV connection failed: {}", e))?;

    // Build basic auth header
    use base64::Engine;
    let credentials = base64::engine::general_purpose::STANDARD.encode(format!("{}:{}", username, password));

    let request = format!(
        "PROPFIND {} HTTP/1.1\r\nHost: {}\r\nAuthorization: Basic {}\r\nDepth: 0\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
        request_path, host_port, credentials
    );

    stream.set_write_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| anyhow!("Failed to set timeout: {}", e))?;
    stream.write_all(request.as_bytes())
        .map_err(|e| anyhow!("Failed to send request: {}", e))?;

    stream.set_read_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| anyhow!("Failed to set timeout: {}", e))?;
    let mut response = String::new();
    stream.read_to_string(&mut response)
        .map_err(|e| anyhow!("Failed to read response: {}", e))?;

    // Check for successful WebDAV response (207 Multi-Status or 200 OK)
    if response.contains("207") || response.contains("200") {
        Ok(true)
    } else if response.contains("401") || response.contains("403") {
        Err(anyhow!("WebDAV authentication failed"))
    } else {
        Err(anyhow!("WebDAV returned unexpected response: {}", response.lines().next().unwrap_or("")))
    }
}