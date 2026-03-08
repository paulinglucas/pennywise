-- Add original_balance column for tracking debt progress
ALTER TABLE accounts ADD COLUMN original_balance REAL;
