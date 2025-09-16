-- Add dietologist specializations

-- Update the specialists table check constraint to include dietologist
ALTER TABLE specialists DROP CONSTRAINT IF EXISTS specialists_type_check;
ALTER TABLE specialists ADD CONSTRAINT specialists_type_check 
    CHECK (type::text = ANY (ARRAY['lawyer'::character varying, 'psychologist'::character varying, 'dietologist'::character varying]::text[]));

-- Insert dietologist specializations
INSERT INTO specializations (name, description, type, is_active, created_at, updated_at) VALUES
('Клиническая диетология', 'Лечебное питание при различных заболеваниях', 'dietologist', true, NOW(), NOW()),
('Спортивная диетология', 'Питание для спортсменов и активного образа жизни', 'dietologist', true, NOW(), NOW()),
('Детская диетология', 'Питание детей и подростков', 'dietologist', true, NOW(), NOW()),
('Диетология снижения веса', 'Программы здорового похудения', 'dietologist', true, NOW(), NOW());
