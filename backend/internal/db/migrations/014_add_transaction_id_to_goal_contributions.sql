-- Add optional transaction link to goal contributions
ALTER TABLE goal_contributions ADD COLUMN transaction_id TEXT REFERENCES transactions(id);
CREATE INDEX idx_goal_contributions_transaction_id ON goal_contributions(transaction_id);
