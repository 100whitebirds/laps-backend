-- Add description column to work_experience table
ALTER TABLE work_experience ADD COLUMN IF NOT EXISTS description TEXT;

