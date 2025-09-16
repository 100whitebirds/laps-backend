-- Remove communication_method column since only video calls are supported
ALTER TABLE appointments DROP COLUMN IF EXISTS communication_method;
