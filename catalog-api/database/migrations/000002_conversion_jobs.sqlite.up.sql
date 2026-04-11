-- Conversion jobs table migration (SQLite variant)
-- SERIAL -> INTEGER PRIMARY KEY AUTOINCREMENT
-- No other changes needed; TEXT/INTEGER/TIMESTAMP are native to SQLite.

CREATE TABLE IF NOT EXISTS conversion_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    source_file_path TEXT NOT NULL,
    target_file_path TEXT NOT NULL,
    source_format TEXT NOT NULL,
    target_format TEXT NOT NULL,
    quality_level TEXT DEFAULT 'medium',
    status TEXT DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_conversion_jobs_user_id ON conversion_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_status ON conversion_jobs(status);
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_created_at ON conversion_jobs(created_at);
