use anyhow::{anyhow, Result};
use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

/// Test FTP connection with given credentials
pub async fn test_connection(
    host: &str,
    port: u16,
    username: &str,
    password: &str,
    path: Option<&str>,
) -> Result<bool> {
    let addr = format!("{}:{}", host, port);
    let stream = TcpStream::connect_timeout(
        &addr.parse().map_err(|e| anyhow!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| anyhow!("FTP connection failed: {}", e))?;

    // Read the FTP banner
    let mut stream = stream;
    stream
        .set_read_timeout(Some(Duration::from_secs(5)))
        .map_err(|e| anyhow!("Failed to set timeout: {}", e))?;
    let mut buf = [0u8; 512];
    let n = stream
        .read(&mut buf)
        .map_err(|e| anyhow!("Failed to read FTP banner: {}", e))?;
    let banner = String::from_utf8_lossy(&buf[..n]);

    if !banner.starts_with("220") {
        return Err(anyhow!("Unexpected FTP response: {}", banner.trim()));
    }

    // Send USER command
    write!(stream, "USER {}\r\n", username)
        .map_err(|e| anyhow!("Failed to send USER: {}", e))?;
    let n = stream.read(&mut buf).map_err(|e| anyhow!("Failed to read: {}", e))?;
    let _response = String::from_utf8_lossy(&buf[..n]);

    // Send PASS command
    write!(stream, "PASS {}\r\n", password)
        .map_err(|e| anyhow!("Failed to send PASS: {}", e))?;
    let n = stream.read(&mut buf).map_err(|e| anyhow!("Failed to read: {}", e))?;
    let response = String::from_utf8_lossy(&buf[..n]);

    if !response.starts_with("230") {
        return Err(anyhow!("FTP authentication failed: {}", response.trim()));
    }

    // If path specified, try CWD
    if let Some(p) = path {
        write!(stream, "CWD {}\r\n", p)
            .map_err(|e| anyhow!("Failed to send CWD: {}", e))?;
        let n = stream.read(&mut buf).map_err(|e| anyhow!("Failed to read: {}", e))?;
        let response = String::from_utf8_lossy(&buf[..n]);
        if !response.starts_with("250") {
            return Err(anyhow!("FTP path not accessible: {}", response.trim()));
        }
    }

    // Quit
    let _ = write!(stream, "QUIT\r\n");

    Ok(true)
}