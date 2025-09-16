-- Sync production chat_sessions schema to match local development schema
-- This ensures consistency between environments

-- Step 1: Add missing columns that exist in local but not in production
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS specialization_id bigint;
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS started_at timestamp with time zone;
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS ended_at timestamp with time zone;

-- Step 2: Remove columns that exist in production but not in local
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS client_name;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS specialist_name;
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS last_message_at;

-- Step 3: Update status column constraints to match local
-- First drop the old constraint
ALTER TABLE chat_sessions DROP CONSTRAINT IF EXISTS chat_sessions_status_check;

-- Add the new constraint with local status values
ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_status_check 
    CHECK (status IN ('pending', 'active', 'ended'));

-- Step 4: Update default status value to match local
ALTER TABLE chat_sessions ALTER COLUMN status SET DEFAULT 'pending';

-- Step 5: Populate specialization_id for existing records
-- Get the first specialization for each specialist
UPDATE chat_sessions 
SET specialization_id = (
    SELECT ss.specialization_id 
    FROM specialist_specializations ss 
    WHERE ss.specialist_id = chat_sessions.specialist_id 
    LIMIT 1
)
WHERE specialization_id IS NULL;

-- Step 6: Make specialization_id NOT NULL after populating data
ALTER TABLE chat_sessions ALTER COLUMN specialization_id SET NOT NULL;

-- Step 7: Add foreign key constraint for specialization_id
ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_specialization_id_fkey 
    FOREIGN KEY (specialization_id) REFERENCES specializations(id) ON DELETE CASCADE;

-- Step 8: Add missing indexes to match local schema
CREATE INDEX IF NOT EXISTS idx_chat_sessions_specialization ON chat_sessions(specialization_id);
CREATE UNIQUE INDEX IF NOT EXISTS unique_appointment_chat ON chat_sessions(appointment_id);

-- Step 9: Add date constraint to match local schema (if it doesn't exist)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_dates') THEN
        ALTER TABLE chat_sessions ADD CONSTRAINT check_dates 
            CHECK (ended_at IS NULL OR ended_at >= started_at);
    END IF;
END $$;

-- Step 10: Update existing 'archived' status to 'ended' to match local schema
UPDATE chat_sessions SET status = 'ended' WHERE status = 'archived';
