CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('client', 'specialist', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);


CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) NOT NULL UNIQUE,
    user_agent TEXT,
    ip VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);


CREATE TABLE IF NOT EXISTS specializations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('lawyer', 'psychologist')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_specializations_type ON specializations(type);
CREATE INDEX IF NOT EXISTS idx_specializations_is_active ON specializations(is_active);


CREATE TABLE IF NOT EXISTS specialists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('lawyer', 'psychologist')),
    specialization_id BIGINT NOT NULL REFERENCES specializations(id) ON DELETE SET NULL,
    experience INT NOT NULL,
    association_member BOOLEAN NOT NULL DEFAULT false,
    rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    reviews_count INT NOT NULL DEFAULT 0,
    recommendation_rate INT NOT NULL DEFAULT 0,
    primary_consult_price DECIMAL(10,2) NOT NULL,
    secondary_consult_price DECIMAL(10,2) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_specialists_user_id ON specialists(user_id);
CREATE INDEX IF NOT EXISTS idx_specialists_type ON specialists(type);
CREATE INDEX IF NOT EXISTS idx_specialists_rating ON specialists(rating);


CREATE TABLE IF NOT EXISTS specialist_specializations (
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    specialization_id BIGINT NOT NULL REFERENCES specializations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (specialist_id, specialization_id)
);


CREATE TABLE IF NOT EXISTS education (
    id BIGSERIAL PRIMARY KEY,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    institution VARCHAR(255) NOT NULL,
    specialization VARCHAR(255) NOT NULL,
    degree VARCHAR(100) NOT NULL,
    graduation_year INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_education_specialist_id ON education(specialist_id);


CREATE TABLE IF NOT EXISTS work_experience (
    id BIGSERIAL PRIMARY KEY,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    company VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    start_year INT NOT NULL,
    end_year INT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_work_experience_specialist_id ON work_experience(specialist_id);


CREATE TABLE IF NOT EXISTS appointments (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    consultation_type VARCHAR(20) NOT NULL CHECK (consultation_type IN ('primary', 'secondary')),
    specialization_id BIGINT REFERENCES specializations(id) ON DELETE SET NULL,
    price DECIMAL(10,2) NOT NULL,
    appointment_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'paid', 'completed', 'cancelled')),
    payment_id VARCHAR(255),
    communication_method VARCHAR(20) NOT NULL CHECK (communication_method IN ('phone', 'whatsapp')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appointments_client_id ON appointments(client_id);
CREATE INDEX IF NOT EXISTS idx_appointments_specialist_id ON appointments(specialist_id);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_appointment_date ON appointments(appointment_date);


CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE CASCADE,
    appointment_id BIGINT NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text TEXT NOT NULL,
    is_recommended BOOLEAN NOT NULL DEFAULT false,
    service_rating INT CHECK (service_rating IS NULL OR (service_rating >= 1 AND service_rating <= 5)),
    meeting_efficiency INT CHECK (meeting_efficiency IS NULL OR (meeting_efficiency >= 1 AND meeting_efficiency <= 5)),
    professionalism INT CHECK (professionalism IS NULL OR (professionalism >= 1 AND professionalism <= 5)),
    price_quality INT CHECK (price_quality IS NULL OR (price_quality >= 1 AND price_quality <= 5)),
    cleanliness INT CHECK (cleanliness IS NULL OR (cleanliness >= 1 AND cleanliness <= 5)),
    attentiveness INT CHECK (attentiveness IS NULL OR (attentiveness >= 1 AND attentiveness <= 5)),
    specialist_experience INT CHECK (specialist_experience IS NULL OR (specialist_experience >= 1 AND specialist_experience <= 5)),
    grammar INT CHECK (grammar IS NULL OR (grammar >= 1 AND grammar <= 5)),
    reply_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reviews_specialist_id ON reviews(specialist_id);
CREATE INDEX IF NOT EXISTS idx_reviews_client_id ON reviews(client_id);
CREATE INDEX IF NOT EXISTS idx_reviews_appointment_id ON reviews(appointment_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);


CREATE TABLE IF NOT EXISTS review_replies (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_replies_review_id ON review_replies(review_id);


ALTER TABLE reviews ADD CONSTRAINT fk_reviews_reply_id FOREIGN KEY (reply_id) REFERENCES review_replies(id) ON DELETE SET NULL;


CREATE OR REPLACE FUNCTION update_specialist_rating()
RETURNS TRIGGER AS $$
DECLARE
    avg_rating DECIMAL(3,2);
    rec_count INT;
BEGIN
    SELECT AVG(rating), COUNT(*) INTO avg_rating, rec_count
    FROM reviews
    WHERE specialist_id = NEW.specialist_id;
    
    UPDATE specialists
    SET rating = COALESCE(avg_rating, 0),
        reviews_count = rec_count,
        recommendation_rate = (
            SELECT COALESCE(ROUND((COUNT(*) FILTER (WHERE is_recommended = true) * 100.0) / COUNT(*)), 0)
            FROM reviews
            WHERE specialist_id = NEW.specialist_id
        ),
        updated_at = NOW()
    WHERE id = NEW.specialist_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_specialist_rating_trigger
AFTER INSERT OR UPDATE ON reviews
FOR EACH ROW
EXECUTE FUNCTION update_specialist_rating();

CREATE TRIGGER update_specialist_rating_delete_trigger
AFTER DELETE ON reviews
FOR EACH ROW
EXECUTE FUNCTION update_specialist_rating(); 