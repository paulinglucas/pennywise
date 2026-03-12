-- Track SimpleFIN accounts the user has chosen not to link
CREATE TABLE IF NOT EXISTS dismissed_simplefin_accounts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    simplefin_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, simplefin_id)
);
