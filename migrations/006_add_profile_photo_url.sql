ALTER TABLE specialists ADD COLUMN IF NOT EXISTS profile_photo_url VARCHAR(255);

COMMENT ON COLUMN specialists.profile_photo_url IS 'URL фото профиля специалиста';