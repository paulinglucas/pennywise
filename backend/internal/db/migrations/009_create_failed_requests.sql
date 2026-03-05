-- Create failed_requests table
CREATE TABLE IF NOT EXISTS failed_requests (
    id TEXT PRIMARY KEY,
    request_id TEXT,
    user_id TEXT,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    request_body TEXT,
    request_headers TEXT,
    error_code TEXT,
    error_message TEXT,
    stack_trace TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_failed_requests_created ON failed_requests(created_at);
