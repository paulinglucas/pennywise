-- Create goals table
CREATE TABLE IF NOT EXISTS goals (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    goal_type TEXT NOT NULL,
    target_amount REAL NOT NULL,
    current_amount REAL NOT NULL DEFAULT 0,
    deadline DATE,
    linked_account_id TEXT REFERENCES accounts(id),
    priority_rank INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_goals_user_priority ON goals(user_id, priority_rank);
