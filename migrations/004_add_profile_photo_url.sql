-- Add missing profile_photo_url column to specialists table
ALTER TABLE specialists ADD COLUMN IF NOT EXISTS profile_photo_url VARCHAR(500);

-- Add index for profile photo URL (optional, for performance)
CREATE INDEX IF NOT EXISTS idx_specialists_profile_photo_url ON specialists(profile_photo_url) WHERE profile_photo_url IS NOT NULL;
