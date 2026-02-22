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
