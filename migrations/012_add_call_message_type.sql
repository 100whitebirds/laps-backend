-- Add 'call' message type to chat_messages constraint if needed
DO $$
BEGIN
    -- Check if the constraint exists and drop it to recreate with 'call'
    IF EXISTS (
        SELECT 1 
        FROM information_schema.constraint_column_usage ccu
        JOIN information_schema.table_constraints tc
          ON tc.constraint_name = ccu.constraint_name
        WHERE ccu.table_name = 'chat_messages'
          AND ccu.column_name = 'message_type'
          AND tc.constraint_type = 'CHECK'
    ) THEN
        -- Try to drop any existing check constraint by name pattern
        -- Note: constraint name may vary across environments; attempt common names
        BEGIN
            ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chat_messages_message_type_check;
            ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS chk_chat_messages_message_type;
        EXCEPTION WHEN others THEN
            -- Ignore if cannot drop by unknown name
            NULL;
        END;
    END IF;

    -- Recreate the check constraint allowing 'call'
    ALTER TABLE chat_messages
    ADD CONSTRAINT chat_messages_message_type_check
    CHECK (message_type IN ('text','image','file','system','call'));
END$$;


