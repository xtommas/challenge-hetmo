CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    long_description TEXT,
    short_description TEXT,
    date_and_time TIMESTAMP NOT NULL,
    organizer VARCHAR(255),
    location VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'draft'
);