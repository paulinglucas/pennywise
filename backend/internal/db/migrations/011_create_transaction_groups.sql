-- Transaction groups for split transactions (e.g., paycheck splits)
CREATE TABLE IF NOT EXISTS transaction_groups (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_transaction_groups_user ON transaction_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_transaction_groups_deleted ON transaction_groups(deleted_at);
