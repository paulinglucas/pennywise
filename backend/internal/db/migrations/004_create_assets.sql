-- Create assets and asset_history tables
CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    account_id TEXT REFERENCES accounts(id),
    name TEXT NOT NULL,
    asset_type TEXT NOT NULL,
    current_value REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_assets_user_type ON assets(user_id, asset_type);

CREATE TABLE IF NOT EXISTS asset_history (
    id TEXT PRIMARY KEY,
    asset_id TEXT NOT NULL REFERENCES assets(id),
    value REAL NOT NULL,
    recorded_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_asset_history_asset_recorded ON asset_history(asset_id, recorded_at);
