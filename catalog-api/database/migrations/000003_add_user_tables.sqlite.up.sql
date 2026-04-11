-- Add missing user authentication and role tables for SQLite
-- This migration fixes the missing users table issue

-- Users table with all required columns
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL,
    role_id INTEGER NOT NULL,
    first_name TEXT,
    last_name TEXT,
    display_name TEXT,
    avatar_url TEXT,
    time_zone TEXT,
    language TEXT,
    settings TEXT DEFAULT '{}',
    is_active INTEGER DEFAULT 1,
    is_locked INTEGER DEFAULT 0,
    locked_until DATETIME,
    failed_login_attempts INTEGER DEFAULT 0,
    last_login_at DATETIME,
    last_login_ip TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT DEFAULT '[]',
    is_system INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT,
    device_info TEXT,
    ip_address TEXT,
    user_agent TEXT,
    is_active INTEGER DEFAULT 1,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert default admin role
INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
VALUES (1, 'Admin', 'Administrator role with all permissions', '["*"]', 1);

-- Insert default user role
INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
VALUES (2, 'User', 'Standard user role', '["media.view", "media.download"]', 1);

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Indexes for sessions table
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);