ALTER TABLE education ADD COLUMN IF NOT EXISTS field_of_study VARCHAR(255) DEFAULT '';
ALTER TABLE education ADD COLUMN IF NOT EXISTS from_year INT DEFAULT 0;
ALTER TABLE education ADD COLUMN IF NOT EXISTS to_year INT DEFAULT 0;


COMMENT ON COLUMN education.field_of_study IS 'Область изучения';
COMMENT ON COLUMN education.from_year IS 'Год начала обучения';
COMMENT ON COLUMN education.to_year IS 'Год окончания обучения'; 