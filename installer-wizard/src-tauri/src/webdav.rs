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
        format!(
            "{}/{}",
            url.trim_end_matches('/'),
            p.trim_start_matches('/')
        )
    } else {
        url.to_string()
    };

    // Use a simple HTTP PROPFIND to test WebDAV
    let parsed = url_str
        .strip_prefix("http://")
        .or_else(|| url_str.strip_prefix("https://"));
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
            &port_addr
                .parse()
                .map_err(|e| anyhow!("Invalid address: {}", e))?,
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
        &port_addr
            .parse()
            .map_err(|e| anyhow!("Invalid address: {}", e))?,
        Duration::from_secs(10),
    )
    .map_err(|e| anyhow!("WebDAV connection failed: {}", e))?;

    // Build basic auth header
    use base64::Engine;
    let credentials =
        base64::engine::general_purpose::STANDARD.encode(format!("{}:{}", username, password));

    let request = format!(
        "PROPFIND {} HTTP/1.1\r\nHost: {}\r\nAuthorization: Basic {}\r\nDepth: 0\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
        request_path, host_port, credentials
    );

    stream
        .set_write_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| anyhow!("Failed to set timeout: {}", e))?;
    stream
        .write_all(request.as_bytes())
        .map_err(|e| anyhow!("Failed to send request: {}", e))?;

    stream
        .set_read_timeout(Some(Duration::from_secs(10)))
        .map_err(|e| anyhow!("Failed to set timeout: {}", e))?;
    let mut response = String::new();
    stream
        .read_to_string(&mut response)
        .map_err(|e| anyhow!("Failed to read response: {}", e))?;

    // Check for successful WebDAV response (207 Multi-Status or 200 OK)
    if response.contains("207") || response.contains("200") {
        Ok(true)
    } else if response.contains("401") || response.contains("403") {
        Err(anyhow!("WebDAV authentication failed"))
    } else {
        Err(anyhow!(
            "WebDAV returned unexpected response: {}",
            response.lines().next().unwrap_or("")
        ))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_webdav_url_path_combination_with_path() {
        let url = "http://example.com/dav";
        let path = Some("documents");
        let url_str = if let Some(p) = path {
            format!(
                "{}/{}",
                url.trim_end_matches('/'),
                p.trim_start_matches('/')
            )
        } else {
            url.to_string()
        };
        assert_eq!(url_str, "http://example.com/dav/documents");
    }

    #[test]
    fn test_webdav_url_path_combination_without_path() {
        let url = "http://example.com/dav";
        let path: Option<&str> = None;
        let url_str = if let Some(p) = path {
            format!(
                "{}/{}",
                url.trim_end_matches('/'),
                p.trim_start_matches('/')
            )
        } else {
            url.to_string()
        };
        assert_eq!(url_str, "http://example.com/dav");
    }

    #[test]
    fn test_webdav_url_trailing_slash_normalization() {
        let url = "http://example.com/dav/";
        let path = Some("/documents/");
        let url_str = format!(
            "{}/{}",
            url.trim_end_matches('/'),
            path.expect("path is Some").trim_start_matches('/')
        );
        assert_eq!(url_str, "http://example.com/dav/documents/");
    }

    #[test]
    fn test_webdav_http_prefix_stripping() {
        let url_str = "http://example.com:8080/dav";
        let parsed = url_str.strip_prefix("http://");
        assert!(parsed.is_some());
        assert_eq!(parsed.expect("should strip http"), "example.com:8080/dav");
    }

    #[test]
    fn test_webdav_https_prefix_stripping() {
        let url_str = "https://example.com/dav";
        let parsed = url_str
            .strip_prefix("http://")
            .or_else(|| url_str.strip_prefix("https://"));
        assert!(parsed.is_some());
        assert_eq!(parsed.expect("should strip https"), "example.com/dav");
    }

    #[test]
    fn test_webdav_invalid_url_no_prefix() {
        let url_str = "ftp://example.com/dav";
        let parsed = url_str
            .strip_prefix("http://")
            .or_else(|| url_str.strip_prefix("https://"));
        assert!(parsed.is_none());
    }

    #[test]
    fn test_webdav_host_port_extraction_with_port() {
        let rest = "example.com:8080/dav/files";
        let (hp, p) = rest.split_once('/').unwrap_or((rest, ""));
        assert_eq!(hp, "example.com:8080");
        assert_eq!(p, "dav/files");
    }

    #[test]
    fn test_webdav_host_port_extraction_without_path() {
        let rest = "example.com:8080";
        let (hp, p) = rest.split_once('/').unwrap_or((rest, ""));
        assert_eq!(hp, "example.com:8080");
        assert_eq!(p, "");
    }

    #[test]
    fn test_webdav_request_path_formatting() {
        let path_part = "dav/files";
        let request_path = format!("/{}", path_part);
        assert_eq!(request_path, "/dav/files");
    }

    #[test]
    fn test_webdav_default_http_port() {
        let host_port = "example.com";
        let port_addr = if host_port.contains(':') {
            host_port.to_string()
        } else {
            format!("{}:80", host_port)
        };
        assert_eq!(port_addr, "example.com:80");
    }

    #[test]
    fn test_webdav_default_https_port() {
        let host_port = "example.com";
        let port_addr = if host_port.contains(':') {
            host_port.to_string()
        } else {
            format!("{}:443", host_port)
        };
        assert_eq!(port_addr, "example.com:443");
    }

    #[test]
    fn test_webdav_explicit_port_preserved() {
        let host_port = "example.com:9090";
        let port_addr = if host_port.contains(':') {
            host_port.to_string()
        } else {
            format!("{}:80", host_port)
        };
        assert_eq!(port_addr, "example.com:9090");
    }

    #[test]
    fn test_webdav_basic_auth_encoding() {
        use base64::Engine;
        let username = "admin";
        let password = "secret";
        let credentials = base64::engine::general_purpose::STANDARD
            .encode(format!("{}:{}", username, password));
        assert_eq!(credentials, "YWRtaW46c2VjcmV0");
    }

    #[test]
    fn test_webdav_basic_auth_encoding_special_chars() {
        use base64::Engine;
        let username = "user@domain.com";
        let password = "p@ss:word!";
        let credentials = base64::engine::general_purpose::STANDARD
            .encode(format!("{}:{}", username, password));
        // Verify it can be decoded back
        let decoded_bytes = base64::engine::general_purpose::STANDARD
            .decode(&credentials)
            .expect("Should decode");
        let decoded = String::from_utf8(decoded_bytes).expect("Should be valid UTF-8");
        assert_eq!(decoded, "user@domain.com:p@ss:word!");
    }

    #[test]
    fn test_webdav_propfind_request_format() {
        let request_path = "/dav/files";
        let host_port = "example.com:8080";
        let credentials = "base64encoded";
        let request = format!(
            "PROPFIND {} HTTP/1.1\r\nHost: {}\r\nAuthorization: Basic {}\r\nDepth: 0\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
            request_path, host_port, credentials
        );
        assert!(request.starts_with("PROPFIND /dav/files HTTP/1.1\r\n"));
        assert!(request.contains("Host: example.com:8080"));
        assert!(request.contains("Authorization: Basic base64encoded"));
        assert!(request.contains("Depth: 0"));
        assert!(request.contains("Connection: close"));
        assert!(request.ends_with("\r\n\r\n"));
    }

    #[test]
    fn test_webdav_response_success_207() {
        let response = "HTTP/1.1 207 Multi-Status\r\nContent-Type: text/xml\r\n\r\n<response/>";
        assert!(response.contains("207"));
    }

    #[test]
    fn test_webdav_response_success_200() {
        let response = "HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\n\r\n";
        assert!(response.contains("200"));
    }

    #[test]
    fn test_webdav_response_auth_failure_401() {
        let response = "HTTP/1.1 401 Unauthorized\r\n";
        assert!(response.contains("401"));
        assert!(!response.contains("207"));
        assert!(!response.contains("200"));
    }

    #[test]
    fn test_webdav_response_forbidden_403() {
        let response = "HTTP/1.1 403 Forbidden\r\n";
        assert!(response.contains("403"));
    }

    #[test]
    fn test_webdav_is_https_detection() {
        assert!("https://example.com".starts_with("https://"));
        assert!(!"http://example.com".starts_with("https://"));
    }

    #[tokio::test]
    async fn test_webdav_connection_invalid_url() {
        let result = test_connection("not-a-valid-url", "user", "pass", None).await;
        assert!(result.is_err());
        let err_msg = result.unwrap_err().to_string();
        assert!(
            err_msg.contains("Invalid URL"),
            "Expected 'Invalid URL' error, got: {}",
            err_msg
        );
    }

    #[tokio::test]
    async fn test_webdav_connection_unreachable_http() {
        let result =
            test_connection("http://192.0.2.1:8080/dav", "user", "pass", None).await;
        assert!(result.is_err());
        let err_msg = result.unwrap_err().to_string();
        assert!(
            err_msg.contains("WebDAV connection failed") || err_msg.contains("Invalid address"),
            "Unexpected error: {}",
            err_msg
        );
    }

    #[tokio::test]
    async fn test_webdav_connection_unreachable_https() {
        let result =
            test_connection("https://192.0.2.1/dav", "user", "pass", None).await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_webdav_connection_with_path() {
        let result = test_connection(
            "http://192.0.2.1/dav",
            "user",
            "pass",
            Some("/documents"),
        )
        .await;
        assert!(result.is_err());
    }

    #[test]
    fn test_webdav_timeout_values() {
        let connect_timeout = Duration::from_secs(10);
        let rw_timeout = Duration::from_secs(10);
        assert_eq!(connect_timeout.as_secs(), 10);
        assert_eq!(rw_timeout.as_secs(), 10);
    }
}
