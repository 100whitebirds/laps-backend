-- Add missing columns to chat_sessions table
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS last_message_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS client_name VARCHAR(255);
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS specialist_name VARCHAR(255);

-- Create function to update last_message_at if it doesn't exist
CREATE OR REPLACE FUNCTION update_chat_session_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chat_sessions 
    SET last_message_at = NEW.created_at,
        updated_at = NEW.created_at
    WHERE id = NEW.session_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger if it doesn't exist
DROP TRIGGER IF EXISTS update_chat_session_last_message_trigger ON chat_messages;
CREATE TRIGGER update_chat_session_last_message_trigger
AFTER INSERT ON chat_messages
FOR EACH ROW
EXECUTE FUNCTION update_chat_session_last_message();

