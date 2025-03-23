-- Добавление недостающих колонок в таблицу specialists
ALTER TABLE specialists ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE specialists ADD COLUMN IF NOT EXISTS experience_years INT NOT NULL DEFAULT 0; 