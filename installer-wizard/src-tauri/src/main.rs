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
            load_configuration,
            save_configuration,
            get_default_config_path
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}