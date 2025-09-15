-- Add sample specializations
INSERT INTO specializations (name, description, type, is_active, created_at, updated_at) VALUES
-- Психологи
('Семейная психология', 'Работа с семейными парами и семейными конфликтами', 'psychologist', true, NOW(), NOW()),
('Детская психология', 'Психологическая помощь детям и подросткам', 'psychologist', true, NOW(), NOW()),
('Клиническая психология', 'Диагностика и коррекция психических расстройств', 'psychologist', true, NOW(), NOW()),

-- Юристы
('Семейное право', 'Развод, алименты, опека над детьми', 'lawyer', true, NOW(), NOW()),
('Трудовое право', 'Трудовые споры, увольнения, трудовые договоры', 'lawyer', true, NOW(), NOW()),
('Гражданское право', 'Договоры, недвижимость, наследство', 'lawyer', true, NOW(), NOW());
