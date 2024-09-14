CREATE TABLE IF NOT EXISTS user_events (
    user_id BIGINT REFERENCES users(id),
    event_id BIGINT REFERENCES events(id),
    PRIMARY KEY (user_id, event_id)
);
