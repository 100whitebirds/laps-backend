-- Drop the problematic check_file_fields constraint if it exists
-- This constraint was preventing call-type messages from being inserted
ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS check_file_fields;

-- Recreate a more permissive constraint that allows:
-- - All file fields to be NULL (for call/system messages)
-- - All file fields to be NOT NULL together (for file/image messages)
ALTER TABLE chat_messages
ADD CONSTRAINT check_file_fields CHECK (
    -- Either all file fields are NULL (for call/system/text messages)
    (file_url IS NULL AND file_name IS NULL AND file_size IS NULL)
    OR
    -- Or all file fields are NOT NULL (for file/image messages with actual files)
    (file_url IS NOT NULL AND file_name IS NOT NULL AND file_size IS NOT NULL)
);

-- Verify the message_type constraint includes 'call'
DO $$
BEGIN
    -- Drop old constraint if exists
    ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chat_messages_message_type_check;
    ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chat_messages_message_type_check1;
    ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chk_chat_messages_message_type;
EXCEPTION WHEN OTHERS THEN
    NULL;
END$$;

-- Add the correct constraint
ALTER TABLE chat_messages
ADD CONSTRAINT chat_messages_message_type_check
CHECK (message_type IN ('text','image','file','system','call'));

