-- Add video_call to communication_method constraint
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_communication_method_check;
ALTER TABLE appointments ADD CONSTRAINT appointments_communication_method_check 
    CHECK (communication_method IN ('phone', 'whatsapp', 'video_call'));
