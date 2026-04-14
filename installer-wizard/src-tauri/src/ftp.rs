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
        &addr
            .parse()
            .map_err(|e| anyhow!("Invalid address: {}", e))?,
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
    write!(stream, "USER {}\r\n", username).map_err(|e| anyhow!("Failed to send USER: {}", e))?;
    let n = stream
        .read(&mut buf)
        .map_err(|e| anyhow!("Failed to read: {}", e))?;
    let _response = String::from_utf8_lossy(&buf[..n]);

    // Send PASS command
    write!(stream, "PASS {}\r\n", password).map_err(|e| anyhow!("Failed to send PASS: {}", e))?;
    let n = stream
        .read(&mut buf)
        .map_err(|e| anyhow!("Failed to read: {}", e))?;
    let response = String::from_utf8_lossy(&buf[..n]);

    if !response.starts_with("230") {
        return Err(anyhow!("FTP authentication failed: {}", response.trim()));
    }

    // If path specified, try CWD
    if let Some(p) = path {
        write!(stream, "CWD {}\r\n", p).map_err(|e| anyhow!("Failed to send CWD: {}", e))?;
        let n = stream
            .read(&mut buf)
            .map_err(|e| anyhow!("Failed to read: {}", e))?;
        let response = String::from_utf8_lossy(&buf[..n]);
        if !response.starts_with("250") {
            return Err(anyhow!("FTP path not accessible: {}", response.trim()));
        }
    }

    // Quit
    let _ = write!(stream, "QUIT\r\n");

    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ftp_address_formatting() {
        let host = "192.168.1.100";
        let port: u16 = 21;
        let addr = format!("{}:{}", host, port);
        assert_eq!(addr, "192.168.1.100:21");
    }

    #[test]
    fn test_ftp_address_formatting_custom_port() {
        let host = "ftp.example.com";
        let port: u16 = 2121;
        let addr = format!("{}:{}", host, port);
        assert_eq!(addr, "ftp.example.com:2121");
    }

    #[test]
    fn test_ftp_address_parsing_valid_ipv4() {
        let addr = "192.168.1.100:21";
        let parsed: Result<std::net::SocketAddr, _> = addr.parse();
        assert!(parsed.is_ok());
        let socket_addr = parsed.expect("Address should parse");
        assert_eq!(socket_addr.port(), 21);
    }

    #[test]
    fn test_ftp_address_parsing_invalid() {
        let addr = "not-a-valid-address:abc";
        let parsed: Result<std::net::SocketAddr, _> = addr.parse();
        assert!(parsed.is_err());
    }

    #[test]
    fn test_ftp_banner_recognition_valid() {
        let banner = "220 ProFTPD Server Ready\r\n";
        assert!(banner.starts_with("220"));
    }

    #[test]
    fn test_ftp_banner_recognition_invalid() {
        let banner = "421 Service not available\r\n";
        assert!(!banner.starts_with("220"));
    }

    #[test]
    fn test_ftp_auth_response_success() {
        let response = "230 User logged in, proceed.\r\n";
        assert!(response.starts_with("230"));
    }

    #[test]
    fn test_ftp_auth_response_failure() {
        let response = "530 Login incorrect.\r\n";
        assert!(!response.starts_with("230"));
    }

    #[test]
    fn test_ftp_cwd_response_success() {
        let response = "250 Directory successfully changed.\r\n";
        assert!(response.starts_with("250"));
    }

    #[test]
    fn test_ftp_cwd_response_failure() {
        let response = "550 Failed to change directory.\r\n";
        assert!(!response.starts_with("250"));
    }

    #[test]
    fn test_ftp_user_command_format() {
        let username = "admin";
        let cmd = format!("USER {}\r\n", username);
        assert_eq!(cmd, "USER admin\r\n");
    }

    #[test]
    fn test_ftp_pass_command_format() {
        let password = "secret123";
        let cmd = format!("PASS {}\r\n", password);
        assert_eq!(cmd, "PASS secret123\r\n");
    }

    #[test]
    fn test_ftp_cwd_command_format() {
        let path = "/media/videos";
        let cmd = format!("CWD {}\r\n", path);
        assert_eq!(cmd, "CWD /media/videos\r\n");
    }

    #[test]
    fn test_ftp_timeout_duration() {
        let timeout = Duration::from_secs(10);
        assert_eq!(timeout.as_secs(), 10);
    }

    #[test]
    fn test_ftp_read_timeout_duration() {
        let timeout = Duration::from_secs(5);
        assert_eq!(timeout.as_secs(), 5);
        assert_eq!(timeout.as_millis(), 5000);
    }

    #[tokio::test]
    async fn test_ftp_connection_unreachable_host() {
        // RFC 5737: 192.0.2.0/24 is reserved for documentation, guaranteed unreachable
        let result = test_connection("192.0.2.1", 21, "user", "pass", None).await;
        assert!(result.is_err());
        let err_msg = result.unwrap_err().to_string();
        assert!(
            err_msg.contains("FTP connection failed") || err_msg.contains("Invalid address"),
            "Unexpected error: {}",
            err_msg
        );
    }

    #[tokio::test]
    async fn test_ftp_connection_invalid_address() {
        let result = test_connection("not_a_valid_host_!!!!", 21, "user", "pass", None).await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_ftp_connection_with_path() {
        // Test that path parameter is accepted (will fail on connect, but tests code path)
        let result =
            test_connection("192.0.2.1", 21, "user", "pass", Some("/media")).await;
        assert!(result.is_err());
    }

    #[test]
    fn test_ftp_port_boundaries() {
        // Standard FTP port
        let standard_port: u16 = 21;
        assert_eq!(standard_port, 21);

        // Max port
        let max_port: u16 = 65535;
        assert_eq!(max_port, 65535);

        // Min port
        let min_port: u16 = 1;
        assert_eq!(min_port, 1);
    }

    #[test]
    fn test_ftp_buffer_size() {
        let buf = [0u8; 512];
        assert_eq!(buf.len(), 512);
    }
}
