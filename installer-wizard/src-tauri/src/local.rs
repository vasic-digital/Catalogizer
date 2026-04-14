use anyhow::{anyhow, Result};

/// Test local filesystem connection
pub async fn test_connection(base_path: &str) -> Result<bool> {
    let path = std::path::Path::new(base_path);

    if !path.exists() {
        return Err(anyhow!("Path does not exist: {}", base_path));
    }

    if !path.is_dir() {
        return Err(anyhow!("Path is not a directory: {}", base_path));
    }

    // Check read permission by trying to read the directory
    std::fs::read_dir(path).map_err(|e| anyhow!("Cannot read directory '{}': {}", base_path, e))?;

    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_local_connection_valid_directory() {
        let result = test_connection("/tmp").await;
        assert!(result.is_ok());
        assert!(result.expect("Should succeed for /tmp"));
    }

    #[tokio::test]
    async fn test_local_connection_nonexistent_path() {
        let result = test_connection("/nonexistent/path/that/does/not/exist/12345").await;
        assert!(result.is_err());
        let err_msg = result.unwrap_err().to_string();
        assert!(
            err_msg.contains("does not exist"),
            "Error should mention path does not exist, got: {}",
            err_msg
        );
    }

    #[tokio::test]
    async fn test_local_connection_file_not_directory() {
        // /etc/hostname is a file, not a directory (exists on most Linux systems)
        let test_path = "/etc/hostname";
        if std::path::Path::new(test_path).exists() {
            let result = test_connection(test_path).await;
            assert!(result.is_err());
            let err_msg = result.unwrap_err().to_string();
            assert!(
                err_msg.contains("not a directory"),
                "Error should mention path is not a directory, got: {}",
                err_msg
            );
        }
    }

    #[tokio::test]
    async fn test_local_connection_root_directory() {
        let result = test_connection("/").await;
        assert!(result.is_ok());
        assert!(result.expect("Root directory should be readable"));
    }

    #[tokio::test]
    async fn test_local_connection_home_directory() {
        if let Ok(home) = std::env::var("HOME") {
            let result = test_connection(&home).await;
            assert!(result.is_ok());
            assert!(result.expect("Home directory should be readable"));
        }
    }

    #[tokio::test]
    async fn test_local_connection_empty_path() {
        let result = test_connection("").await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_local_connection_relative_path_nonexistent() {
        let result = test_connection("relative/nonexistent/path").await;
        assert!(result.is_err());
    }

    #[test]
    fn test_path_exists_check() {
        let path = std::path::Path::new("/tmp");
        assert!(path.exists());
        assert!(path.is_dir());
    }

    #[test]
    fn test_path_not_exists_check() {
        let path = std::path::Path::new("/nonexistent_path_abc123");
        assert!(!path.exists());
    }

    #[test]
    fn test_path_is_file_not_dir() {
        let test_path = "/etc/hostname";
        let path = std::path::Path::new(test_path);
        if path.exists() {
            assert!(!path.is_dir());
        }
    }
}
