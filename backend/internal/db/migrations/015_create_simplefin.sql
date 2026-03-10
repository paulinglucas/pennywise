-- SimpleFIN integration tables
CREATE TABLE IF NOT EXISTS simplefin_connections (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id),
    access_url TEXT NOT NULL,
    last_sync_at DATETIME,
    sync_error TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

ALTER TABLE accounts ADD COLUMN simplefin_id TEXT;
