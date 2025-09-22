-- Add trainer specializations

-- First, update the specializations table check constraint to include trainer
ALTER TABLE specializations DROP CONSTRAINT IF EXISTS specializations_type_check;
ALTER TABLE specializations ADD CONSTRAINT specializations_type_check 
    CHECK (type IN ('lawyer', 'psychologist', 'dietologist', 'trainer'));

-- Then, update the specialists table check constraint to include trainer
ALTER TABLE specialists DROP CONSTRAINT IF EXISTS specialists_type_check;
ALTER TABLE specialists ADD CONSTRAINT specialists_type_check 
    CHECK (type::text = ANY (ARRAY['lawyer'::character varying, 'psychologist'::character varying, 'dietologist'::character varying, 'trainer'::character varying]::text[]));

-- Insert trainer specializations
INSERT INTO specializations (name, description, type, is_active, created_at, updated_at) VALUES
('Персональный тренер', 'Индивидуальные тренировки и программы фитнеса', 'trainer', true, NOW(), NOW()),
('Тренер по йоге', 'Занятия йогой для всех уровней подготовки', 'trainer', true, NOW(), NOW()),
('Тренер по пилатесу', 'Занятия пилатесом для укрепления мышц', 'trainer', true, NOW(), NOW()),
('Тренер по функциональному тренингу', 'Функциональные тренировки для повседневной жизни', 'trainer', true, NOW(), NOW()),
('Тренер по единоборствам', 'Обучение различным видам единоборств', 'trainer', true, NOW(), NOW());
