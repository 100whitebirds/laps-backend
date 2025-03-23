-- Таблица расписания специалистов
CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time VARCHAR(5) NOT NULL, -- Формат "HH:MM"
    end_time VARCHAR(5) NOT NULL, -- Формат "HH:MM"
    slot_time INT NOT NULL CHECK (slot_time >= 10 AND slot_time <= 120), -- Длительность слота в минутах
    exclude_times VARCHAR[] DEFAULT '{}', -- Исключенные временные слоты
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (specialist_id, date) -- Уникальное расписание на день для специалиста
);

CREATE INDEX IF NOT EXISTS idx_schedules_specialist_id ON schedules(specialist_id);
CREATE INDEX IF NOT EXISTS idx_schedules_date ON schedules(date);

COMMENT ON TABLE schedules IS 'Таблица расписания специалистов';
COMMENT ON COLUMN schedules.specialist_id IS 'ID специалиста';
COMMENT ON COLUMN schedules.date IS 'Дата расписания';
COMMENT ON COLUMN schedules.start_time IS 'Время начала работы в формате HH:MM';
COMMENT ON COLUMN schedules.end_time IS 'Время окончания работы в формате HH:MM';
COMMENT ON COLUMN schedules.slot_time IS 'Длительность временного слота в минутах';
COMMENT ON COLUMN schedules.exclude_times IS 'Массив исключенных временных слотов в формате HH:MM';
COMMENT ON COLUMN schedules.created_at IS 'Дата и время создания записи';
COMMENT ON COLUMN schedules.updated_at IS 'Дата и время последнего обновления записи'; 