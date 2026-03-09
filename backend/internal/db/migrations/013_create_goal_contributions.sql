-- Goal contributions: tracks individual payments toward goals
CREATE TABLE IF NOT EXISTS goal_contributions (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL REFERENCES goals(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    amount REAL NOT NULL,
    notes TEXT,
    contributed_at DATE NOT NULL DEFAULT (date('now')),
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_goal_contributions_goal_id ON goal_contributions(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_contributions_user_id ON goal_contributions(user_id);
CREATE INDEX IF NOT EXISTS idx_goal_contributions_contributed_at ON goal_contributions(contributed_at);
