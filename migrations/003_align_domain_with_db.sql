-- Удаление полей field_of_study и from_year из структуры Education в коде
ALTER TABLE work_experience
ADD COLUMN IF NOT EXISTS from_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS to_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS description TEXT;

-- Добавление колонок в таблицу education для соответствия доменной модели
ALTER TABLE education 
ADD COLUMN IF NOT EXISTS field_of_study VARCHAR(255),
ADD COLUMN IF NOT EXISTS from_year INT,
ADD COLUMN IF NOT EXISTS to_year INT; 