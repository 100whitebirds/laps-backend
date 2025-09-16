-- Minimal schema sync - only add missing pieces safely
-- Most of the schema sync was already done manually

-- Ensure specialization_id exists and is populated
DO $$ 
BEGIN
    -- Add specialization_id column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'chat_sessions' AND column_name = 'specialization_id') THEN
        ALTER TABLE chat_sessions ADD COLUMN specialization_id bigint;
        
        -- Populate specialization_id for existing records
        UPDATE chat_sessions 
        SET specialization_id = (
            SELECT ss.specialization_id 
            FROM specialist_specializations ss 
            WHERE ss.specialist_id = chat_sessions.specialist_id 
            LIMIT 1
        )
        WHERE specialization_id IS NULL;
        
        -- Make it NOT NULL
        ALTER TABLE chat_sessions ALTER COLUMN specialization_id SET NOT NULL;
        
        -- Add foreign key constraint
        ALTER TABLE chat_sessions ADD CONSTRAINT chat_sessions_specialization_id_fkey 
            FOREIGN KEY (specialization_id) REFERENCES specializations(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Ensure indexes exist
CREATE INDEX IF NOT EXISTS idx_chat_sessions_specialization ON chat_sessions(specialization_id);
CREATE UNIQUE INDEX IF NOT EXISTS unique_appointment_chat ON chat_sessions(appointment_id);